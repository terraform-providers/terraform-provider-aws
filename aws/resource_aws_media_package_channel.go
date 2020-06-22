package aws

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediapackage"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsMediaPackageChannel() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMediaPackageChannelCreate,
		Read:   resourceAwsMediaPackageChannelRead,
		Update: resourceAwsMediaPackageChannelUpdate,
		Delete: resourceAwsMediaPackageChannelDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"channel_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\w-]+$`), "must only contain alphanumeric characters, dashes or underscores"),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			"hls_ingest": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ingest_endpoints": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"password": {
										Type:      schema.TypeString,
										Computed:  true,
										Sensitive: true,
									},
									"url": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"username": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsMediaPackageChannelCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediapackageconn

	input := &mediapackage.CreateChannelInput{
		Id:          aws.String(d.Get("channel_id").(string)),
		Description: aws.String(d.Get("description").(string)),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().MediapackageTags()
	}

	resp, err := conn.CreateChannel(input)
	if err != nil {
		return fmt.Errorf("error creating MediaPackage Channel: %s", err)
	}

	d.SetId(aws.StringValue(resp.Id))

	return resourceAwsMediaPackageChannelRead(d, meta)
}

func resourceAwsMediaPackageChannelRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediapackageconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &mediapackage.DescribeChannelInput{
		Id: aws.String(d.Id()),
	}
	resp, err := conn.DescribeChannel(input)
	if err != nil {
		return fmt.Errorf("error describing MediaPackage Channel: %s", err)
	}
	d.Set("arn", resp.Arn)
	d.Set("channel_id", resp.Id)
	d.Set("description", resp.Description)

	if err := d.Set("hls_ingest", flattenMediaPackageHLSIngest(resp.HlsIngest)); err != nil {
		return fmt.Errorf("error setting hls_ingest: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.MediapackageKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsMediaPackageChannelUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediapackageconn

	input := &mediapackage.UpdateChannelInput{
		Id:          aws.String(d.Id()),
		Description: aws.String(d.Get("description").(string)),
	}

	_, err := conn.UpdateChannel(input)
	if err != nil {
		return fmt.Errorf("error updating MediaPackage Channel: %s", err)
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.MediapackageUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating MediaPackage Channel (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsMediaPackageChannelRead(d, meta)
}

func resourceAwsMediaPackageChannelDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediapackageconn

	input := &mediapackage.DeleteChannelInput{
		Id: aws.String(d.Id()),
	}
	_, err := conn.DeleteChannel(input)
	if err != nil {
		if isAWSErr(err, mediapackage.ErrCodeNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting MediaPackage Channel: %s", err)
	}

	dcinput := &mediapackage.DescribeChannelInput{
		Id: aws.String(d.Id()),
	}
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DescribeChannel(dcinput)
		if err != nil {
			if isAWSErr(err, mediapackage.ErrCodeNotFoundException, "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		return resource.RetryableError(fmt.Errorf("MediaPackage Channel (%s) still exists", d.Id()))
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DescribeChannel(dcinput)
	}
	if err != nil {
		return fmt.Errorf("error waiting for MediaPackage Channel (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func flattenMediaPackageHLSIngest(h *mediapackage.HlsIngest) []map[string]interface{} {
	if h.IngestEndpoints == nil {
		return []map[string]interface{}{
			{"ingest_endpoints": []map[string]interface{}{}},
		}
	}

	var ingestEndpoints []map[string]interface{}
	for _, e := range h.IngestEndpoints {
		endpoint := map[string]interface{}{
			"password": aws.StringValue(e.Password),
			"url":      aws.StringValue(e.Url),
			"username": aws.StringValue(e.Username),
		}

		ingestEndpoints = append(ingestEndpoints, endpoint)
	}

	return []map[string]interface{}{
		{"ingest_endpoints": ingestEndpoints},
	}
}
