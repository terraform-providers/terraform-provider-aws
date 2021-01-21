package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsSecurityHubActionTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecurityHubActionTargetCreate,
		Read:   resourceAwsSecurityHubActionTargetRead,
		Update: resourceAwsSecurityHubActionTargetUpdate,
		Delete: resourceAwsSecurityHubActionTargetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 20),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]+$`), "must contain only alphanumeric characters"),
				),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 20),
				),
			},
		},
	}
}

func resourceAwsSecurityHubActionTargetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	description := d.Get("description").(string)
	name := d.Get("name").(string)
	identifier := d.Get("identifier").(string)

	log.Printf("[DEBUG] Creating Security Hub custom action target %s", identifier)

	resp, err := conn.CreateActionTarget(&securityhub.CreateActionTargetInput{
		Description: aws.String(description),
		Id:          aws.String(identifier),
		Name:        aws.String(name),
	})

	if err != nil {
		return fmt.Errorf("Error creating Security Hub custom action target %s: %s", identifier, err)
	}

	d.SetId(aws.StringValue(resp.ActionTargetArn))

	return resourceAwsSecurityHubActionTargetRead(d, meta)
}

func resourceAwsSecurityHubActionTargetParseIdentifier(identifier string) (string, error) {
	parts := strings.Split(identifier, "/")

	if len(parts) != 3 {
		return "", fmt.Errorf("Expected Security Hub Custom action ARN, received: %s", identifier)
	}

	return parts[2], nil
}

func resourceAwsSecurityHubActionTargetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	log.Printf("[DEBUG] Reading Security Hub custom action targets to find %s", d.Id())

	actionTargetIdentifier, err := resourceAwsSecurityHubActionTargetParseIdentifier(d.Id())

	if err != nil {
		return err
	}

	actionTarget, err := resourceAwsSecurityHubActionTargetCheckExists(conn, d.Id())

	if err != nil {
		return fmt.Errorf("Error reading Security Hub custom action targets to find %s: %s", d.Id(), err)
	}

	if actionTarget == nil {
		log.Printf("[WARN] Security Hub custom action target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("identifier", actionTargetIdentifier)
	d.Set("description", actionTarget.Description)
	d.Set("arn", actionTarget.ActionTargetArn)
	d.Set("name", actionTarget.Name)

	return nil
}

func resourceAwsSecurityHubActionTargetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	input := &securityhub.UpdateActionTargetInput{
		ActionTargetArn: aws.String(d.Id()),
		Description:     aws.String(d.Get("description").(string)),
		Name:            aws.String(d.Get("name").(string)),
	}
	if _, err := conn.UpdateActionTarget(input); err != nil {
		return fmt.Errorf("error updating Security Hub Action Target (%s): %w", d.Id(), err)
	}
	return nil
}

func resourceAwsSecurityHubActionTargetCheckExists(conn *securityhub.SecurityHub, actionTargetArn string) (*securityhub.ActionTarget, error) {
	input := &securityhub.DescribeActionTargetsInput{
		ActionTargetArns: aws.StringSlice([]string{actionTargetArn}),
	}
	var found *securityhub.ActionTarget = nil
	err := conn.DescribeActionTargetsPages(input, func(page *securityhub.DescribeActionTargetsOutput, lastPage bool) bool {
		for _, actionTarget := range page.ActionTargets {
			if aws.StringValue(actionTarget.ActionTargetArn) == actionTargetArn {
				found = actionTarget
				return false
			}
		}
		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return found, nil
}

func resourceAwsSecurityHubActionTargetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Printf("[DEBUG] Deleting Security Hub custom action target %s", d.Id())

	_, err := conn.DeleteActionTarget(&securityhub.DeleteActionTargetInput{
		ActionTargetArn: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("Error deleting Security Hub custom action target %s: %s", d.Id(), err)
	}

	return nil
}
