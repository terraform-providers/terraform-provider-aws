package aws

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsS3DownloadBucketObject() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsS3DownloadBucketObjectRead,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"filename": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAwsS3DownloadBucketObjectRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	uniqueID := bucket + "/" + key

	input := s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if v, ok := d.GetOk("version_id"); ok {
		input.VersionId = aws.String(v.(string))
	}
	out, err := conn.GetObject(&input)
	if err != nil {
		return fmt.Errorf("Failed getting S3 object: %s", err)
	}

	s3ObjectBytes, err := ioutil.ReadAll(out.Body)
	if err != nil {
		return fmt.Errorf("Failed reading content of S3 object (%s): %s", uniqueID, err)
	}
	log.Printf("[INFO] Saving %d bytes from S3 object %s", len(s3ObjectBytes), uniqueID)

	fileName := d.Get("filename").(string)
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("Failed to create a file at %s: %s", fileName, err)
	}
	defer file.Close()

	if _, err = file.Write(s3ObjectBytes); err != nil {
		return fmt.Errorf("Error writing content from s3://%s to %s: %s", uniqueID, fileName, err)
	}

	d.SetId(fileName)
	d.Set("version_id", out.VersionId)

	return nil
}
