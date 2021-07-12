package aws

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsIotProvisioningTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotProvisioningTemplateCreate,
		Read:   resourceAwsIotProvisioningTemplateRead,
		Update: resourceAwsIotProvisioningTemplateUpdate,
		Delete: resourceAwsIotProvisioningTemplateDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"default_version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"provisioning_role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"template_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"template_body": {
				Type:     schema.TypeString,
				Required: true,
			},
			"template_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsIotProvisioningTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	var out *iot.CreateProvisioningTemplateOutput
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		out, err = conn.CreateProvisioningTemplate(&iot.CreateProvisioningTemplateInput{
			Description:         aws.String(d.Get("description").(string)),
			Enabled:             aws.Bool(d.Get("enabled").(bool)),
			ProvisioningRoleArn: aws.String(d.Get("provisioning_role_arn").(string)),
			TemplateBody:        aws.String(d.Get("template_body").(string)),
			TemplateName:        aws.String(d.Get("template_name").(string)),
		})

		if err != nil {
			// Handle IoT not detecting the provisioning role's assume role policy immediately.
			if isAWSErr(err, iot.ErrCodeInvalidRequestException, "The provisioning role cannot be assumed by AWS IoT") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		log.Printf("[ERROR] %s", err)
		return err
	}

	d.SetId(*out.TemplateName)

	return resourceAwsIotProvisioningTemplateRead(d, meta)
}

func resourceAwsIotProvisioningTemplateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	out, err := conn.DescribeProvisioningTemplate(&iot.DescribeProvisioningTemplateInput{
		TemplateName: aws.String(d.Id()),
	})

	if err != nil {
		log.Printf("[ERROR] %s", err)
		return err
	}

	d.Set("default_version_id", out.DefaultVersionId)
	d.Set("description", out.Description)
	d.Set("enabled", out.Enabled)
	d.Set("provisioning_role_arn", out.ProvisioningRoleArn)
	d.Set("template_arn", out.TemplateArn)
	d.Set("template_body", out.TemplateBody)
	d.Set("template_name", out.TemplateName)

	return nil
}

func resourceAwsIotProvisioningTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	if d.HasChange("template_body") {
		_, err := conn.CreateProvisioningTemplateVersion(&iot.CreateProvisioningTemplateVersionInput{
			TemplateName: aws.String(d.Id()),
			TemplateBody: aws.String(d.Get("template_body").(string)),
			SetAsDefault: aws.Bool(true),
		})

		if err != nil {
			log.Printf("[ERROR] %s", err)
			return err
		}
	}

	if d.HasChanges("description", "enabled", "provisioning_role_arn") {
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.UpdateProvisioningTemplate(&iot.UpdateProvisioningTemplateInput{
				Description:         aws.String(d.Get("description").(string)),
				Enabled:             aws.Bool(d.Get("enabled").(bool)),
				ProvisioningRoleArn: aws.String(d.Get("provisioning_role_arn").(string)),
				TemplateName:        aws.String(d.Id()),
			})

			if err != nil {
				// Handle IoT not detecting the provisioning role's assume role policy immediately.
				if isAWSErr(err, iot.ErrCodeInvalidRequestException, "The provisioning role cannot be assumed by AWS IoT") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if err != nil {
			log.Printf("[ERROR] %s", err)
			return err
		}
	}

	return resourceAwsIotProvisioningTemplateRead(d, meta)
}

func resourceAwsIotProvisioningTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	_, err := conn.DeleteProvisioningTemplate(&iot.DeleteProvisioningTemplateInput{
		TemplateName: aws.String(d.Id()),
	})

	if err != nil {
		log.Printf("[ERROR] %s", err)
		return err
	}

	return nil
}
