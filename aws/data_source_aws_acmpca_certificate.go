package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsAcmpcaCertificate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAcmpcaCertificateRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"certificate_authority_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_chain": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsAcmpcaCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn
	certificateArn := d.Get("arn").(string)

	getCertificateInput := &acmpca.GetCertificateInput{
		CertificateArn:          aws.String(certificateArn),
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
	}

	log.Printf("[DEBUG] Reading ACM PCA Certificate: %s", getCertificateInput)

	certificateOutput, err := conn.GetCertificate(getCertificateInput)
	if err != nil {
		return fmt.Errorf("error reading ACM PCA Certificate (%s): %w", certificateArn, err)
	}

	d.SetId(certificateArn)
	d.Set("certificate", aws.StringValue(certificateOutput.Certificate))
	d.Set("certificate_chain", aws.StringValue(certificateOutput.CertificateChain))

	return nil
}
