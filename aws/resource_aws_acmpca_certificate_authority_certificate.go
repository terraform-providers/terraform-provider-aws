package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/acmpca/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsAcmpcaCertificateAuthorityCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAcmpcaCertificateAuthorityCertificateCreate,
		Read:   resourceAwsAcmpcaCertificateAuthorityCertificateRead,
		Delete: schema.Noop,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"certificate": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32768),
			},
			"certificate_authority_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"certificate_chain": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 2097152),
			},
		},
	}
}

func resourceAwsAcmpcaCertificateAuthorityCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn

	certificateAuthorityArn := d.Get("certificate_authority_arn").(string)

	input := &acmpca.ImportCertificateAuthorityCertificateInput{
		Certificate:             []byte(d.Get("certificate").(string)),
		CertificateAuthorityArn: aws.String(certificateAuthorityArn),
	}
	if v, ok := d.Get("certificate_chain").(string); ok && v != "" {
		input.CertificateChain = []byte(v)
	}

	_, err := conn.ImportCertificateAuthorityCertificate(input)
	if err != nil {
		return fmt.Errorf("error associating ACM PCA Certificate with Certificate Authority (%s): %w", certificateAuthorityArn, err)
	}

	d.SetId(certificateAuthorityArn)

	return resourceAwsAcmpcaCertificateAuthorityCertificateRead(d, meta)
}

func resourceAwsAcmpcaCertificateAuthorityCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn

	output, err := finder.CertificateAuthorityCertificateByARN(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ACM PCA Certificate Authority Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading ACM PCA Certificate Authority Certificate (%s): %w", d.Id(), err)
	}

	d.Set("certificate_authority_arn", d.Id())
	d.Set("certificate", aws.StringValue(output.Certificate))
	d.Set("certificate_chain", aws.StringValue(output.CertificateChain))

	return nil
}
