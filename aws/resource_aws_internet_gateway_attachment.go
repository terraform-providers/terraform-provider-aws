package aws

import (
	"fmt"
	"log"
	"strings"
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
		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInternetGatewayIDNotFound) {
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

	d.SetId(fmt.Sprintf("%s:%s", vpcID, igwID))

	_, err = waiter.InternetGatewayAttchmentCreated(conn, igwID, vpcID)
	if err != nil {
		return fmt.Errorf("error waiting for Internet Gateway attachmen %q to be created: %w", d.Id(), err)
	}

	return resourceAwsInternetGatewayAttachmentRead(d, meta)
}

func resourceAwsInternetGatewayAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcID, igwID, err := decodeInternetGatewayAttachmentID(d.Id())
	if err != nil {
		return err
	}

	resp, err := finder.InternetGatewayAttachmentByID(conn, igwID, vpcID)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInternetGatewayIDNotFound) {
			log.Printf("[WARN] Internet Gateway Attachment (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
	}

	if resp == nil || len(resp.InternetGateways) == 0 {
		log.Printf("[WARN] Internet Gateway Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	igw := resp.InternetGateways[0]
	if len(igw.Attachments) == 0 {
		log.Printf("[WARN] Internet Gateway Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	attachment := igw.Attachments[0]
	if aws.StringValue(attachment.VpcId) != vpcID {
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

	vpcID, igwID, err := decodeInternetGatewayAttachmentID(d.Id())
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

		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInternetGatewayIDNotFound) {
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

	_, err = waiter.InternetGatewayAttchmentDeleted(conn, igwID, vpcID)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInternetGatewayIDNotFound) {
			return nil
		}
		return fmt.Errorf("error waiting for Internet Gateway attachment %q to be deleted: %w", d.Id(), err)
	}

	return nil
}

func decodeInternetGatewayAttachmentID(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected VPC-ID:IGW-ID", id)
	}

	return parts[0], parts[1], nil
}
