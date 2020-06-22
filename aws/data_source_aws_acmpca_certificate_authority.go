package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsAcmpcaCertificateAuthority() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAcmpcaCertificateAuthorityRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_chain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_signing_request": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"not_after": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"not_before": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// https://docs.aws.amazon.com/acm-pca/latest/APIReference/API_RevocationConfiguration.html
			"revocation_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// https://docs.aws.amazon.com/acm-pca/latest/APIReference/API_CrlConfiguration.html
						"crl_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"custom_cname": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"enabled": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"expiration_in_days": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"s3_bucket_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"serial": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsAcmpcaCertificateAuthorityRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	certificateAuthorityArn := d.Get("arn").(string)

	describeCertificateAuthorityInput := &acmpca.DescribeCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(certificateAuthorityArn),
	}

	log.Printf("[DEBUG] Reading ACMPCA Certificate Authority: %s", describeCertificateAuthorityInput)

	describeCertificateAuthorityOutput, err := conn.DescribeCertificateAuthority(describeCertificateAuthorityInput)
	if err != nil {
		return fmt.Errorf("error reading ACMPCA Certificate Authority: %s", err)
	}

	if describeCertificateAuthorityOutput.CertificateAuthority == nil {
		return fmt.Errorf("error reading ACMPCA Certificate Authority: not found")
	}
	certificateAuthority := describeCertificateAuthorityOutput.CertificateAuthority

	d.Set("arn", certificateAuthority.Arn)
	d.Set("not_after", aws.TimeValue(certificateAuthority.NotAfter).Format(time.RFC3339))
	d.Set("not_before", aws.TimeValue(certificateAuthority.NotBefore).Format(time.RFC3339))

	if err := d.Set("revocation_configuration", flattenAcmpcaRevocationConfiguration(certificateAuthority.RevocationConfiguration)); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("serial", certificateAuthority.Serial)
	d.Set("status", certificateAuthority.Status)
	d.Set("type", certificateAuthority.Type)

	getCertificateAuthorityCertificateInput := &acmpca.GetCertificateAuthorityCertificateInput{
		CertificateAuthorityArn: aws.String(certificateAuthorityArn),
	}

	log.Printf("[DEBUG] Reading ACMPCA Certificate Authority Certificate: %s", getCertificateAuthorityCertificateInput)

	getCertificateAuthorityCertificateOutput, err := conn.GetCertificateAuthorityCertificate(getCertificateAuthorityCertificateInput)
	if err != nil {
		// Returned when in PENDING_CERTIFICATE status
		// InvalidStateException: The certificate authority XXXXX is not in the correct state to have a certificate signing request.
		if !isAWSErr(err, acmpca.ErrCodeInvalidStateException, "") {
			return fmt.Errorf("error reading ACMPCA Certificate Authority Certificate: %s", err)
		}
	}

	d.Set("certificate", "")
	d.Set("certificate_chain", "")
	if getCertificateAuthorityCertificateOutput != nil {
		d.Set("certificate", getCertificateAuthorityCertificateOutput.Certificate)
		d.Set("certificate_chain", getCertificateAuthorityCertificateOutput.CertificateChain)
	}

	getCertificateAuthorityCsrInput := &acmpca.GetCertificateAuthorityCsrInput{
		CertificateAuthorityArn: aws.String(certificateAuthorityArn),
	}

	log.Printf("[DEBUG] Reading ACMPCA Certificate Authority Certificate Signing Request: %s", getCertificateAuthorityCsrInput)

	getCertificateAuthorityCsrOutput, err := conn.GetCertificateAuthorityCsr(getCertificateAuthorityCsrInput)
	if err != nil {
		return fmt.Errorf("error reading ACMPCA Certificate Authority Certificate Signing Request: %s", err)
	}

	d.Set("certificate_signing_request", "")
	if getCertificateAuthorityCsrOutput != nil {
		d.Set("certificate_signing_request", getCertificateAuthorityCsrOutput.Csr)
	}

	tags, err := keyvaluetags.AcmpcaListTags(conn, certificateAuthorityArn)

	if err != nil {
		return fmt.Errorf("error listing tags for ACMPCA Certificate Authority (%s): %s", certificateAuthorityArn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.SetId(certificateAuthorityArn)

	return nil
}
