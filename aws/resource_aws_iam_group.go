package aws

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsIamGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamGroupCreate,
		Read:   resourceAwsIamGroupRead,
		Update: resourceAwsIamGroupUpdate,
		Delete: resourceAwsIamGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(`^[0-9A-Za-z=,.@\-_+]+$`),
					fmt.Sprintf("must only contain alphanumeric characters, hyphens, underscores, commas, periods, @ symbols, plus and equals signs"),
				),
			},
			"path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
			},
		},
	}
}

func resourceAwsIamGroupCreate(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn
	name := d.Get("name").(string)
	path := d.Get("path").(string)

	request := &iam.CreateGroupInput{
		Path:      aws.String(path),
		GroupName: aws.String(name),
	}

	createResp, err := iamconn.CreateGroup(request)
	if err != nil {
		return fmt.Errorf("Error creating IAM Group %s: %s", name, err)
	}
	d.SetId(*createResp.Group.GroupName)

	return resourceAwsIamGroupReadResult(d, createResp.Group)
}

func resourceAwsIamGroupRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	request := &iam.GetGroupInput{
		GroupName: aws.String(d.Id()),
	}

	getResp, err := iamconn.GetGroup(request)
	if err != nil {
		if iamerr, ok := err.(awserr.Error); ok && iamerr.Code() == "NoSuchEntity" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading IAM Group %s: %s", d.Id(), err)
	}
	return resourceAwsIamGroupReadResult(d, getResp.Group)
}

func resourceAwsIamGroupReadResult(d *schema.ResourceData, group *iam.Group) error {
	if err := d.Set("name", group.GroupName); err != nil {
		return err
	}
	if err := d.Set("arn", group.Arn); err != nil {
		return err
	}
	if err := d.Set("path", group.Path); err != nil {
		return err
	}
	if err := d.Set("unique_id", group.GroupId); err != nil {
		return err
	}
	return nil
}

func resourceAwsIamGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.HasChanges("name", "path") {
		iamconn := meta.(*AWSClient).iamconn
		on, nn := d.GetChange("name")
		_, np := d.GetChange("path")

		request := &iam.UpdateGroupInput{
			GroupName:    aws.String(on.(string)),
			NewGroupName: aws.String(nn.(string)),
			NewPath:      aws.String(np.(string)),
		}
		_, err := iamconn.UpdateGroup(request)
		if err != nil {
			return fmt.Errorf("Error updating IAM Group %s: %s", d.Id(), err)
		}
		d.SetId(nn.(string))
		return resourceAwsIamGroupRead(d, meta)
	}
	return nil
}

func resourceAwsIamGroupDelete(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	request := &iam.DeleteGroupInput{
		GroupName: aws.String(d.Id()),
	}

	if _, err := iamconn.DeleteGroup(request); err != nil {
		return fmt.Errorf("Error deleting IAM Group %s: %s", d.Id(), err)
	}
	return nil
}
