package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsS3BucketOwnershipControls() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3BucketOwnershipControlsCreate,
		Read:   resourceAwsS3BucketOwnershipControlsRead,
		Update: resourceAwsS3BucketOwnershipControlsUpdate,
		Delete: resourceAwsS3BucketOwnershipControlsDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"object_ownership": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(s3.ObjectOwnership_Values(), false),
						},
					},
				},
			},
		},
	}
}

func resourceAwsS3BucketOwnershipControlsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)

	input := &s3.PutBucketOwnershipControlsInput{
		Bucket: aws.String(bucket),
		OwnershipControls: &s3.OwnershipControls{
			Rules: expandS3OwnershipControlsRules(d.Get("rule").([]interface{})),
		},
	}

	_, err := conn.PutBucketOwnershipControls(input)

	if err != nil {
		return fmt.Errorf("error creating S3 Bucket (%s) Ownership Controls: %w", bucket, err)
	}

	d.SetId(bucket)

	return resourceAwsS3BucketOwnershipControlsRead(d, meta)
}

func resourceAwsS3BucketOwnershipControlsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	input := &s3.GetBucketOwnershipControlsInput{
		Bucket: aws.String(d.Id()),
	}

	output, err := conn.GetBucketOwnershipControls(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket Ownership Controls (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "OwnershipControlsNotFoundError") {
		log.Printf("[WARN] S3 Bucket Ownership Controls (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Bucket (%s) Ownership Controls: %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error reading S3 Bucket (%s) Ownership Controls: empty response", d.Id())
	}

	d.Set("bucket", d.Id())

	if output.OwnershipControls == nil {
		d.Set("rule", nil)
	} else {
		if err := d.Set("rule", flattenS3OwnershipControlsRules(output.OwnershipControls.Rules)); err != nil {
			return fmt.Errorf("error setting rule: %w", err)
		}
	}

	return nil
}

func resourceAwsS3BucketOwnershipControlsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	input := &s3.PutBucketOwnershipControlsInput{
		Bucket: aws.String(d.Id()),
		OwnershipControls: &s3.OwnershipControls{
			Rules: expandS3OwnershipControlsRules(d.Get("rule").([]interface{})),
		},
	}

	_, err := conn.PutBucketOwnershipControls(input)

	if err != nil {
		return fmt.Errorf("error updating S3 Bucket (%s) Ownership Controls: %w", d.Id(), err)
	}

	return resourceAwsS3BucketOwnershipControlsRead(d, meta)
}

func resourceAwsS3BucketOwnershipControlsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	input := &s3.DeleteBucketOwnershipControlsInput{
		Bucket: aws.String(d.Id()),
	}

	_, err := conn.DeleteBucketOwnershipControls(input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, "OwnershipControlsNotFoundError") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Bucket (%s) Ownership Controls: %w", d.Id(), err)
	}

	return nil
}

func expandS3OwnershipControlsRules(tfList []interface{}) []*s3.OwnershipControlsRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []*s3.OwnershipControlsRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandS3OwnershipControlsRule(tfMap))
	}

	return apiObjects
}

func expandS3OwnershipControlsRule(tfMap map[string]interface{}) *s3.OwnershipControlsRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3.OwnershipControlsRule{}

	if v, ok := tfMap["object_ownership"].(string); ok && v != "" {
		apiObject.ObjectOwnership = aws.String(v)
	}

	return apiObject
}

func flattenS3OwnershipControlsRules(apiObjects []*s3.OwnershipControlsRule) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenS3OwnershipControlsRule(apiObject))
	}

	return tfList
}

func flattenS3OwnershipControlsRule(apiObject *s3.OwnershipControlsRule) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ObjectOwnership; v != nil {
		tfMap["object_ownership"] = aws.StringValue(v)
	}

	return tfMap
}
