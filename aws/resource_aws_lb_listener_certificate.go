package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfelbv2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elbv2"
)

func resourceAwsLbListenerCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLbListenerCertificateCreate,
		Read:   resourceAwsLbListenerCertificateRead,
		Delete: resourceAwsLbListenerCertificateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"listener_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"certificate_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsLbListenerCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn

	listenerArn := d.Get("listener_arn").(string)
	certificateArn := d.Get("certificate_arn").(string)

	params := &elbv2.AddListenerCertificatesInput{
		ListenerArn: aws.String(listenerArn),
		Certificates: []*elbv2.Certificate{
			{
				CertificateArn: aws.String(certificateArn),
			},
		},
	}

	log.Printf("[DEBUG] Adding certificate: %s of listener: %s", certificateArn, listenerArn)

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.AddListenerCertificates(params)

		// Retry for IAM Server Certificate eventual consistency
		if isAWSErr(err, elbv2.ErrCodeCertificateNotFoundException, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.AddListenerCertificates(params)
	}

	if err != nil {
		return fmt.Errorf("error adding LB Listener Certificate: %w", err)
	}

	d.SetId(tfelbv2.ListenerCertificateCreateID(listenerArn, certificateArn))

	return resourceAwsLbListenerCertificateRead(d, meta)
}

func resourceAwsLbListenerCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn

	listenerArn, certificateArn, err := tfelbv2.ListenerCertificateParseID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing ELBv2 Listener Certificate ID (%s): %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Reading certificate: %s of listener: %s", certificateArn, listenerArn)

	var certificate *elbv2.Certificate
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		certificate, err = findAwsLbListenerCertificate(certificateArn, listenerArn, true, nil, conn)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if certificate == nil {
			err = fmt.Errorf("certificate not found: %s", certificateArn)
			if d.IsNewResource() {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		certificate, err = findAwsLbListenerCertificate(certificateArn, listenerArn, true, nil, conn)
	}
	if err != nil {
		if certificate == nil {
			log.Printf("[WARN] %s - removing from state", err)
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("certificate_arn", certificateArn)
	d.Set("listener_arn", listenerArn)

	return nil
}

func resourceAwsLbListenerCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn

	certificateArn := d.Get("certificate_arn").(string)
	listenerArn := d.Get("listener_arn").(string)

	log.Printf("[DEBUG] Deleting certificate: %s of listener: %s", certificateArn, listenerArn)

	params := &elbv2.RemoveListenerCertificatesInput{
		ListenerArn: aws.String(listenerArn),
		Certificates: []*elbv2.Certificate{
			{
				CertificateArn: aws.String(certificateArn),
			},
		},
	}

	_, err := conn.RemoveListenerCertificates(params)
	if err != nil {
		if isAWSErr(err, elbv2.ErrCodeCertificateNotFoundException, "") {
			return nil
		}
		if isAWSErr(err, elbv2.ErrCodeListenerNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error removing LB Listener Certificate: %w", err)
	}

	return nil
}

func findAwsLbListenerCertificate(certificateArn, listenerArn string, skipDefault bool, nextMarker *string, conn *elbv2.ELBV2) (*elbv2.Certificate, error) {
	params := &elbv2.DescribeListenerCertificatesInput{
		ListenerArn: aws.String(listenerArn),
		PageSize:    aws.Int64(400),
	}
	if nextMarker != nil {
		params.Marker = nextMarker
	}

	resp, err := conn.DescribeListenerCertificates(params)
	if err != nil {
		return nil, err
	}

	for _, cert := range resp.Certificates {
		if skipDefault && aws.BoolValue(cert.IsDefault) {
			continue
		}

		if aws.StringValue(cert.CertificateArn) == certificateArn {
			return cert, nil
		}
	}

	if resp.NextMarker != nil {
		return findAwsLbListenerCertificate(certificateArn, listenerArn, skipDefault, resp.NextMarker, conn)
	}
	return nil, nil
}
