package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ram/waiter"
)

func resourceAwsRamResourceShare() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRamResourceShareCreate,
		Read:   resourceAwsRamResourceShareRead,
		Update: resourceAwsRamResourceShareUpdate,
		Delete: resourceAwsRamResourceShareDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"allow_external_principals": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsRamResourceShareCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	request := &ram.CreateResourceShareInput{
		Name:                    aws.String(d.Get("name").(string)),
		AllowExternalPrincipals: aws.Bool(d.Get("allow_external_principals").(bool)),
	}

	if len(tags) > 0 {
		request.Tags = tags.IgnoreAws().RamTags()
	}

	log.Println("[DEBUG] Create RAM resource share request:", request)
	createResp, err := conn.CreateResourceShare(request)
	if err != nil {
		return fmt.Errorf("Error creating RAM resource share: %s", err)
	}

	d.SetId(aws.StringValue(createResp.ResourceShare.ResourceShareArn))

	_, err = waiter.ResourceShareOwnedBySelfActive(conn, d.Id(), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmt.Errorf("Error waiting for RAM resource share (%s) to become ready: %s", d.Id(), err)
	}

	return resourceAwsRamResourceShareRead(d, meta)
}

func resourceAwsRamResourceShareRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	request := &ram.GetResourceSharesInput{
		ResourceShareArns: []*string{aws.String(d.Id())},
		ResourceOwner:     aws.String(ram.ResourceOwnerSelf),
	}

	output, err := conn.GetResourceShares(request)
	if err != nil {
		if isAWSErr(err, ram.ErrCodeUnknownResourceException, "") {
			log.Printf("[WARN] No RAM resource share by ARN (%s) found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading RAM resource share %s: %s", d.Id(), err)
	}

	if len(output.ResourceShares) == 0 {
		log.Printf("[WARN] No RAM resource share by ARN (%s) found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	resourceShare := output.ResourceShares[0]

	if aws.StringValue(resourceShare.Status) != ram.ResourceShareStatusActive {
		log.Printf("[WARN] RAM resource share (%s) delet(ing|ed), removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", resourceShare.ResourceShareArn)
	d.Set("name", resourceShare.Name)
	d.Set("allow_external_principals", resourceShare.AllowExternalPrincipals)

	tags := keyvaluetags.RamKeyValueTags(resourceShare.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsRamResourceShareUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	if d.HasChanges("name", "allow_external_principals") {
		request := &ram.UpdateResourceShareInput{
			ResourceShareArn:        aws.String(d.Id()),
			Name:                    aws.String(d.Get("name").(string)),
			AllowExternalPrincipals: aws.Bool(d.Get("allow_external_principals").(bool)),
		}

		log.Println("[DEBUG] Update RAM resource share request:", request)
		_, err := conn.UpdateResourceShare(request)
		if err != nil {
			if isAWSErr(err, ram.ErrCodeUnknownResourceException, "") {
				log.Printf("[WARN] No RAM resource share by ARN (%s) found", d.Id())
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error updating RAM resource share %s: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.RamUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating RAM resource share (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsRamResourceShareRead(d, meta)
}

func resourceAwsRamResourceShareDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	deleteResourceShareInput := &ram.DeleteResourceShareInput{
		ResourceShareArn: aws.String(d.Id()),
	}

	log.Println("[DEBUG] Delete RAM resource share request:", deleteResourceShareInput)
	_, err := conn.DeleteResourceShare(deleteResourceShareInput)
	if err != nil {
		if isAWSErr(err, ram.ErrCodeUnknownResourceException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting RAM resource share %s: %s", d.Id(), err)
	}

	_, err = waiter.ResourceShareOwnedBySelfDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("Error waiting for RAM resource share (%s) to become ready: %s", d.Id(), err)
	}

	return nil
}
