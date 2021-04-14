package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
)

func resourceAwsEc2ManagedPrefixList() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2ManagedPrefixListCreate,
		Read:   resourceAwsEc2ManagedPrefixListRead,
		Update: resourceAwsEc2ManagedPrefixListUpdate,
		Delete: resourceAwsEc2ManagedPrefixListDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ComputedIf("version", func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("entry")
			}),
		),

		Schema: map[string]*schema.Schema{
			"address_family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice(
					[]string{"IPv4", "IPv6"},
					false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"entry": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
						},
						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
					},
				},
			},
			"max_entries": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAwsEc2ManagedPrefixListCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.CreateManagedPrefixListInput{}

	if v, ok := d.GetOk("address_family"); ok {
		input.AddressFamily = aws.String(v.(string))
	}

	if v, ok := d.GetOk("entry"); ok && v.(*schema.Set).Len() > 0 {
		input.Entries = expandEc2AddPrefixListEntries(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("max_entries"); ok {
		input.MaxEntries = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("name"); ok {
		input.PrefixListName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok && len(v.(map[string]interface{})) > 0 {
		input.TagSpecifications = ec2TagSpecificationsFromMap(v.(map[string]interface{}), "prefix-list")
	}

	output, err := conn.CreateManagedPrefixList(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Managed Prefix List: %w", err)
	}

	d.SetId(aws.StringValue(output.PrefixList.PrefixListId))

	if _, err := waiter.ManagedPrefixListCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Managed Prefix List (%s) creation: %w", d.Id(), err)
	}

	return resourceAwsEc2ManagedPrefixListRead(d, meta)
}

func resourceAwsEc2ManagedPrefixListRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	pl, err := finder.ManagedPrefixListByID(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidPrefixListIDNotFound) {
		log.Printf("[WARN] EC2 Managed Prefix List %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Managed Prefix List (%s): %w", d.Id(), err)
	}

	if pl == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading EC2 Managed Prefix List (%s): not found", d.Id())
		}

		log.Printf("[WARN] EC2 Managed Prefix List %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	input := &ec2.GetManagedPrefixListEntriesInput{
		PrefixListId: pl.PrefixListId,
	}
	var prefixListEntries []*ec2.PrefixListEntry

	err = conn.GetManagedPrefixListEntriesPages(input, func(page *ec2.GetManagedPrefixListEntriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		prefixListEntries = append(prefixListEntries, page.Entries...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing entries of EC2 Managed Prefix List (%s): %w", d.Id(), err)
	}

	d.Set("address_family", pl.AddressFamily)
	d.Set("arn", pl.PrefixListArn)

	if err := d.Set("entry", flattenEc2PrefixListEntries(prefixListEntries)); err != nil {
		return fmt.Errorf("error setting attribute entry of managed prefix list %s: %w", d.Id(), err)
	}

	d.Set("max_entries", pl.MaxEntries)
	d.Set("name", pl.PrefixListName)
	d.Set("owner_id", pl.OwnerId)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(pl.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error settings attribute tags of managed prefix list %s: %w", d.Id(), err)
	}

	d.Set("version", pl.Version)

	return nil
}

func resourceAwsEc2ManagedPrefixListUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChangeExcept("tags") {
		input := &ec2.ModifyManagedPrefixListInput{
			PrefixListId: aws.String(d.Id()),
		}

		input.PrefixListName = aws.String(d.Get("name").(string))
		currentVersion := int64(d.Get("version").(int))
		wait := false

		oldAttr, newAttr := d.GetChange("entry")
		os := oldAttr.(*schema.Set)
		ns := newAttr.(*schema.Set)

		if addEntries := ns.Difference(os); addEntries.Len() > 0 {
			input.AddEntries = expandEc2AddPrefixListEntries(addEntries.List())
			input.CurrentVersion = aws.Int64(currentVersion)
			wait = true
		}

		if removeEntries := os.Difference(ns); removeEntries.Len() > 0 {
			input.RemoveEntries = expandEc2RemovePrefixListEntries(removeEntries.List())
			input.CurrentVersion = aws.Int64(currentVersion)
			wait = true
		}

		_, err := conn.ModifyManagedPrefixList(input)

		if err != nil {
			return fmt.Errorf("error updating EC2 Managed Prefix List (%s): %w", d.Id(), err)
		}

		if wait {
			if _, err := waiter.ManagedPrefixListModified(conn, d.Id()); err != nil {
				return fmt.Errorf("error waiting for EC2 Managed Prefix List (%s) update: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Managed Prefix List (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsEc2ManagedPrefixListRead(d, meta)
}

func resourceAwsEc2ManagedPrefixListDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DeleteManagedPrefixListInput{
		PrefixListId: aws.String(d.Id()),
	}

	_, err := conn.DeleteManagedPrefixList(input)

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidPrefixListIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Managed Prefix List (%s): %w", d.Id(), err)
	}

	if err := waiter.ManagedPrefixListDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Managed Prefix List (%s) deletion: %w", d.Id(), err)
	}

	return nil
}

func expandEc2AddPrefixListEntry(tfMap map[string]interface{}) *ec2.AddPrefixListEntry {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.AddPrefixListEntry{}

	if v, ok := tfMap["cidr"].(string); ok && v != "" {
		apiObject.Cidr = aws.String(v)
	}

	if v, ok := tfMap["description"].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	return apiObject
}

func expandEc2AddPrefixListEntries(tfList []interface{}) []*ec2.AddPrefixListEntry {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.AddPrefixListEntry

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEc2AddPrefixListEntry(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandEc2RemovePrefixListEntry(tfMap map[string]interface{}) *ec2.RemovePrefixListEntry {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.RemovePrefixListEntry{}

	if v, ok := tfMap["cidr"].(string); ok && v != "" {
		apiObject.Cidr = aws.String(v)
	}

	return apiObject
}

func expandEc2RemovePrefixListEntries(tfList []interface{}) []*ec2.RemovePrefixListEntry {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.RemovePrefixListEntry

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEc2RemovePrefixListEntry(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenEc2PrefixListEntry(apiObject *ec2.PrefixListEntry) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Cidr; v != nil {
		tfMap["cidr"] = aws.StringValue(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap["description"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEc2PrefixListEntries(apiObjects []*ec2.PrefixListEntry) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenEc2PrefixListEntry(apiObject))
	}

	return tfList
}
