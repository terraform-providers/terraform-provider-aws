package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsXrayGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsXrayGroupCreate,
		Read:   resourceAwsXrayGroupRead,
		Update: resourceAwsXrayGroupUpdate,
		Delete: resourceAwsXrayGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"filter_expression": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsXrayGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).xrayconn
	input := &xray.CreateGroupInput{
		GroupName:        aws.String(d.Get("group_name").(string)),
		FilterExpression: aws.String(d.Get("filter_expression").(string)),
		Tags:             keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().XrayTags(),
	}

	out, err := conn.CreateGroup(input)
	if err != nil {
		return fmt.Errorf("error creating XRay Group: %w", err)
	}

	d.SetId(aws.StringValue(out.Group.GroupARN))

	return resourceAwsXrayGroupRead(d, meta)
}

func resourceAwsXrayGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).xrayconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &xray.GetGroupInput{
		GroupARN: aws.String(d.Id()),
	}

	group, err := conn.GetGroup(input)

	if err != nil {
		if isAWSErr(err, xray.ErrCodeInvalidRequestException, "Group not found") {
			log.Printf("[WARN] XRay Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading XRay Group (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(group.Group.GroupARN)
	d.Set("arn", arn)
	d.Set("group_name", group.Group.GroupName)
	d.Set("filter_expression", group.Group.FilterExpression)

	tags, err := keyvaluetags.XrayListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for Xray Group (%q): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsXrayGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).xrayconn

	if d.HasChange("filter_expression") {
		input := &xray.UpdateGroupInput{
			GroupARN:         aws.String(d.Id()),
			FilterExpression: aws.String(d.Get("filter_expression").(string)),
		}

		_, err := conn.UpdateGroup(input)
		if err != nil {
			return fmt.Errorf("error updating XRay Group (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.XrayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsXrayGroupRead(d, meta)
}

func resourceAwsXrayGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).xrayconn

	log.Printf("[INFO] Deleting XRay Group: %s", d.Id())

	params := &xray.DeleteGroupInput{
		GroupARN: aws.String(d.Id()),
	}
	_, err := conn.DeleteGroup(params)
	if err != nil {
		return fmt.Errorf("error deleting XRay Group (%s): %w", d.Id(), err)
	}

	return nil
}
