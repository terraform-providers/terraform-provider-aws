package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsMediaStoreContainer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMediaStoreContainerCreate,
		Read:   resourceAwsMediaStoreContainerRead,
		Update: resourceAwsMediaStoreContainerUpdate,
		Delete: resourceAwsMediaStoreContainerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^\w+$`), "must contain alphanumeric characters or underscores"),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsMediaStoreContainerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	input := &mediastore.CreateContainerInput{
		ContainerName: aws.String(d.Get("name").(string)),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().MediastoreTags()
	}

	resp, err := conn.CreateContainer(input)
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(resp.Container.Name))

	stateConf := &resource.StateChangeConf{
		Pending:    []string{mediastore.ContainerStatusCreating},
		Target:     []string{mediastore.ContainerStatusActive},
		Refresh:    mediaStoreContainerRefreshStatusFunc(conn, d.Id()),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	return resourceAwsMediaStoreContainerRead(d, meta)
}

func resourceAwsMediaStoreContainerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &mediastore.DescribeContainerInput{
		ContainerName: aws.String(d.Id()),
	}
	resp, err := conn.DescribeContainer(input)
	if isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
		log.Printf("[WARN] No Container found: %s, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error describing media store container %s: %s", d.Id(), err)
	}

	arn := aws.StringValue(resp.Container.ARN)
	d.Set("arn", arn)
	d.Set("name", resp.Container.Name)
	d.Set("endpoint", resp.Container.Endpoint)

	tags, err := keyvaluetags.MediastoreListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for media store container (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsMediaStoreContainerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.MediastoreUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating media store container (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsMediaStoreContainerRead(d, meta)
}

func resourceAwsMediaStoreContainerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	input := &mediastore.DeleteContainerInput{
		ContainerName: aws.String(d.Id()),
	}
	_, err := conn.DeleteContainer(input)
	if err != nil {
		if isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
			return nil
		}
		return err
	}

	dcinput := &mediastore.DescribeContainerInput{
		ContainerName: aws.String(d.Id()),
	}
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DescribeContainer(dcinput)
		if err != nil {
			if isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		return resource.RetryableError(fmt.Errorf("Media Store Container (%s) still exists", d.Id()))
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DescribeContainer(dcinput)
	}
	if err != nil {
		return fmt.Errorf("error waiting for Media Store Container (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func mediaStoreContainerRefreshStatusFunc(conn *mediastore.MediaStore, cn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &mediastore.DescribeContainerInput{
			ContainerName: aws.String(cn),
		}
		resp, err := conn.DescribeContainer(input)
		if err != nil {
			return nil, "failed", err
		}
		return resp, *resp.Container.Status, nil
	}
}
