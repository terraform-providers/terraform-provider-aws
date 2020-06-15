package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsS3BucketPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsS3BucketPolicyRead,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsS3BucketPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)

	input := &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	}

	log.Printf("[DEBUG] Reading S3 bucket policy: %s", input)
	result, err := conn.GetBucketPolicy(input)

	policy := ""

	if err != nil {
		log.Printf("[DEBUG] Error reading S3 bucket policy: %q", err)

		if reqerr, ok := err.(awserr.RequestFailure); ok {
			log.Printf("[DEBUG] Request failure reading S3 bucket policy: %q", reqerr)

			// ignore error if bucket policy doesn't exist
			if reqerr.StatusCode() != 404 {
				return fmt.Errorf("Failed getting S3 bucket policy: %s Bucket: %q", err, bucket)
			}
		} else {
			return fmt.Errorf("Failed getting S3 bucket policy: %s Bucket: %q", err, bucket)
		}
	} else {
		policy = *result.Policy
	}

	d.SetId(bucket)
	d.Set("policy", policy)

	return nil
}
