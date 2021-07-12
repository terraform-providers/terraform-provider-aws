package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
)

func resourceAwsInternetGatewayAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsInternetGatewayAttachmentCreate,
		Read:   resourceAwsInternetGatewayAttachmentRead,
		Delete: resourceAwsInternetGatewayAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"internet_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsInternetGatewayAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	igwID := d.Get("internet_gateway_id").(string)
	vpcID := d.Get("vpc_id").(string)

	input := &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(igwID),
		VpcId:             aws.String(vpcID),
	}

	log.Printf("[DEBUG] Creating internet gateway attachment")
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := conn.AttachInternetGateway(input)
		if err == nil {
			return nil
		}
		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidInternetGatewayIDNotFound) {
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err)
	})
	if isResourceTimeoutError(err) {
		_, err = conn.AttachInternetGateway(input)
	}
	if err != nil {
		return fmt.Errorf("Error creating Internet Gateway attachment: %w", err)
	}

	d.SetId(tfec2.InternetGatewayAttachmentCreateID(vpcID, igwID))

	_, err = waiter.InternetGatewayAttachmentCreated(conn, igwID, vpcID)
	if err != nil {
		return fmt.Errorf("error waiting for Internet Gateway attachment %q to be created: %w", d.Id(), err)
	}

	return resourceAwsInternetGatewayAttachmentRead(d, meta)
}

func resourceAwsInternetGatewayAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcID, igwID, err := tfec2.InternetGatewayAttachmentParseID(d.Id())
	if err != nil {
		return err
	}

	resp, err := finder.InternetGatewayAttachmentByID(conn, igwID, vpcID)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidInternetGatewayIDNotFound) {
			log.Printf("[WARN] Internet Gateway Attachment (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
	}

	if resp == nil {
		log.Printf("[WARN] Internet Gateway Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(resp.VpcId) != vpcID {
		log.Printf("[WARN] Internet Gateway Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("vpc_id", vpcID)
	d.Set("internet_gateway_id", igwID)

	return nil
}

func resourceAwsInternetGatewayAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcID, igwID, err := tfec2.InternetGatewayAttachmentParseID(d.Id())
	if err != nil {
		return err
	}

	input := &ec2.DetachInternetGatewayInput{
		InternetGatewayId: aws.String(igwID),
		VpcId:             aws.String(vpcID),
	}

	log.Printf("[DEBUG] deleting internet gateway attachment")
	err = resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, err := conn.DetachInternetGateway(input)
		if err == nil {
			return nil
		}

		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidInternetGatewayIDNotFound) {
			return nil
		}

		if tfawserr.ErrCodeEquals(err, "DependencyViolation") {
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err)
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DetachInternetGateway(input)
	}
	if err != nil {
		return fmt.Errorf("Error deleting internet gateway attchment: %w", err)
	}

	_, err = waiter.InternetGatewayAttachmentDeleted(conn, igwID, vpcID)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidInternetGatewayIDNotFound) {
			return nil
		}
		return fmt.Errorf("error waiting for Internet Gateway attachment %q to be deleted: %w", d.Id(), err)
	}

	return nil
}
