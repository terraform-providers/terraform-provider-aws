package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsLakeFormationPolicyTag() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLakeFormationPolicyTagCreate,
		Read:   resourceAwsLakeFormationPolicyTagRead,
		Update: resourceAwsLakeFormationPolicyTagUpdate,
		Delete: resourceAwsLakeFormationPolicyTagDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"values": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 15,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 255),
						validation.StringMatch(regexp.MustCompile(`^([\p{L}\p{Z}\p{N}_.:\*\/=+\-@%]*)$`), ""),
					),
				},
				Set: schema.HashString,
			},
			"catalog_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceAwsLakeFormationPolicyTagCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn

	tagKey := d.Get("key").(string)
	tagValues := d.Get("values").(*schema.Set)

	var catalogID string
	if v, ok := d.GetOk("catalog_id"); ok {
		catalogID = v.(string)
	} else {
		catalogID = meta.(*AWSClient).accountid
	}

	input := &lakeformation.CreateLFTagInput{
		CatalogId: aws.String(catalogID),
		TagKey:    aws.String(tagKey),
		TagValues: expandStringSet(tagValues),
	}

	_, err := conn.CreateLFTag(input)
	if err != nil {
		return fmt.Errorf("Error creating Lake Formation Policy Tag: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", catalogID, tagKey))

	return resourceAwsLakeFormationPolicyTagRead(d, meta)
}

func resourceAwsLakeFormationPolicyTagRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn

	catalogID, tagKey, err := readPolicyTagID(d.Id())
	if err != nil {
		return err
	}

	input := &lakeformation.GetLFTagInput{
		CatalogId: aws.String(catalogID),
		TagKey:    aws.String(tagKey),
	}

	output, err := conn.GetLFTag(input)
	if err != nil {
		if isAWSErr(err, lakeformation.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Lake Formation Policy Tag (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error reading Lake Formation Policy Tag: %s", err.Error())
	}

	d.Set("key", output.TagKey)
	d.Set("values", flattenStringList(output.TagValues))
	d.Set("catalog_id", output.CatalogId)

	return nil
}

func resourceAwsLakeFormationPolicyTagUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn

	catalogID, tagKey, err := readPolicyTagID(d.Id())
	if err != nil {
		return err
	}

	o, n := d.GetChange("values")
	os := o.(*schema.Set)
	ns := n.(*schema.Set)
	toAdd := expandStringSet(ns.Difference(os))
	toDelete := expandStringSet(os.Difference(ns))

	input := &lakeformation.UpdateLFTagInput{
		CatalogId: aws.String(catalogID),
		TagKey:    aws.String(tagKey),
	}

	if len(toAdd) > 0 {
		input.TagValuesToAdd = toAdd
	}

	if len(toDelete) > 0 {
		input.TagValuesToDelete = toDelete
	}

	_, err = conn.UpdateLFTag(input)
	if err != nil {
		return fmt.Errorf("Error updating Lake Formation Policy Tag (%s): %w", d.Id(), err)
	}

	return resourceAwsLakeFormationPolicyTagRead(d, meta)
}

func resourceAwsLakeFormationPolicyTagDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn

	catalogID, tagKey, err := readPolicyTagID(d.Id())
	if err != nil {
		return err
	}

	input := &lakeformation.DeleteLFTagInput{
		CatalogId: aws.String(catalogID),
		TagKey:    aws.String(tagKey),
	}

	_, err = conn.DeleteLFTag(input)
	if err != nil {
		return fmt.Errorf("Error deleting Lake Formation Policy Tag (%s): %w", d.Id(), err)
	}

	return nil
}

func readPolicyTagID(id string) (catalogID string, tagKey string, err error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected CATALOG-ID:TAG-KEY", id)
	}
	return idParts[0], idParts[1], nil
}
