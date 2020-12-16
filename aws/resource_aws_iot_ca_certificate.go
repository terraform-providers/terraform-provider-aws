package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsIotCACertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotCACertificateCreate,
		Read:   resourceAwsIotCACertificateRead,
		Update: resourceAwsIotCACertificateUpdate,
		Delete: resourceAwsIotCACertificateDelete,
		Schema: map[string]*schema.Schema{
			"active": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_registration": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"ca_certificate_pem": {
				Type:     schema.TypeString,
				Required: true,
			},
			"verification_certificate_pem": {
				Type:     schema.TypeString,
				Required: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsIotCACertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	out, err := conn.RegisterCACertificate(&iot.RegisterCACertificateInput{
		AllowAutoRegistration:   aws.Bool(d.Get("auto_registration").(bool)),
		CaCertificate:           aws.String(d.Get("ca_certificate_pem").(string)),
		SetAsActive:             aws.Bool(d.Get("active").(bool)),
		VerificationCertificate: aws.String(d.Get("verification_certificate_pem").(string)),
	})
	if err != nil {
		return fmt.Errorf("error registering ca certificate: %v", err)
	}

	d.SetId(aws.StringValue(out.CertificateId))

	return resourceAwsIotCACertificateRead(d, meta)
}

func resourceAwsIotCACertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	out, err := conn.DescribeCACertificate(&iot.DescribeCACertificateInput{
		CertificateId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error reading ca certificate details: %v", err)
	}

	d.Set("active", aws.Bool(*out.CertificateDescription.Status == iot.CACertificateStatusActive))
	d.Set("auto_registration", aws.Bool(*out.CertificateDescription.AutoRegistrationStatus == iot.AutoRegistrationStatusEnable))
	d.Set("arn", out.CertificateDescription.CertificateArn)

	return nil
}

func resourceAwsIotCACertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	if d.HasChange("active") {
		status := iot.CACertificateStatusInactive
		if d.Get("active").(bool) {
			status = iot.CACertificateStatusActive
		}
		_, err := conn.UpdateCACertificate(&iot.UpdateCACertificateInput{
			CertificateId: aws.String(d.Id()),
			NewStatus:     aws.String(status),
		})
		if err != nil {
			return fmt.Errorf("error updating certificate: %v", err)
		}
	}

	if d.HasChange("auto_registration") {
		status := iot.AutoRegistrationStatusDisable
		if d.Get("active").(bool) {
			status = iot.AutoRegistrationStatusEnable
		}

		_, err := conn.UpdateCACertificate(&iot.UpdateCACertificateInput{
			CertificateId:             aws.String(d.Id()),
			NewAutoRegistrationStatus: aws.String(status),
		})
		if err != nil {
			return fmt.Errorf("error updating certificate: %v", err)
		}
	}

	return resourceAwsIotCACertificateRead(d, meta)
}

func resourceAwsIotCACertificateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	_, err := conn.UpdateCACertificate(&iot.UpdateCACertificateInput{
		CertificateId: aws.String(d.Id()),
		NewStatus:     aws.String(iot.CACertificateStatusInactive),
	})
	if err != nil {
		return fmt.Errorf("error inactivating ca certificate: %v", err)
	}

	_, err = conn.DeleteCACertificate(&iot.DeleteCACertificateInput{
		CertificateId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error deleting ca certificate: %v", err)
	}

	return nil
}
