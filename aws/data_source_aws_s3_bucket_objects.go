package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const keyRequestPageSize = 1000

func dataSourceAwsS3BucketObjects() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsS3BucketObjectsRead,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"delimiter": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"encoding_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"max_keys": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1000,
			},
			"start_after": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"fetch_owner": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"keys": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"common_prefixes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"owners": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsS3BucketObjectsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	prefix := d.Get("prefix").(string)

	listInput := s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}

	if prefix != "" {
		listInput.Prefix = aws.String(prefix)
	}

	if s, ok := d.GetOk("delimiter"); ok {
		listInput.Delimiter = aws.String(s.(string))
	}

	if s, ok := d.GetOk("encoding_type"); ok {
		listInput.EncodingType = aws.String(s.(string))
	}

	// "listInput.MaxKeys" refers to max keys returned in a single request
	// (i.e., page size), not the total number of keys returned if you page
	// through the results. "maxKeys" does refer to total keys returned.
	maxKeys := int64(d.Get("max_keys").(int))
	if maxKeys <= keyRequestPageSize {
		listInput.MaxKeys = aws.Int64(maxKeys)
	}

	if s, ok := d.GetOk("start_after"); ok {
		listInput.StartAfter = aws.String(s.(string))
	}

	if b, ok := d.GetOk("fetch_owner"); ok {
		listInput.FetchOwner = aws.Bool(b.(bool))
	}

	var commonPrefixes []string
	var keys []string
	var owners []string

	err := conn.ListObjectsV2Pages(&listInput, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, commonPrefix := range page.CommonPrefixes {
			commonPrefixes = append(commonPrefixes, aws.StringValue(commonPrefix.Prefix))
		}

		for _, object := range page.Contents {
			keys = append(keys, aws.StringValue(object.Key))

			if object.Owner != nil {
				owners = append(owners, aws.StringValue(object.Owner.ID))
			}
		}

		maxKeys = maxKeys - aws.Int64Value(page.KeyCount)

		if maxKeys <= keyRequestPageSize {
			listInput.MaxKeys = aws.Int64(maxKeys)
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing S3 Bucket (%s) Objects: %w", bucket, err)
	}

	d.SetId(bucket)

	if err := d.Set("common_prefixes", commonPrefixes); err != nil {
		return fmt.Errorf("error setting common_prefixes: %w", err)
	}

	if err := d.Set("keys", keys); err != nil {
		return fmt.Errorf("error setting keys: %w", err)
	}

	if err := d.Set("owners", owners); err != nil {
		return fmt.Errorf("error setting owners: %w", err)
	}

	return nil
}
