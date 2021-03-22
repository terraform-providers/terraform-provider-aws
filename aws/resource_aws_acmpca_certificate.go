package aws

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/acmpca/waiter"
)

func resourceAwsAcmpcaCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAcmpcaCertificateCreate,
		Read:   resourceAwsAcmpcaCertificateRead,
		Delete: resourceAwsAcmpcaCertificateRevoke,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				re := regexp.MustCompile(`arn:.+:certificate-authority/[^/]+`)
				authorityArn := re.FindString(d.Id())
				if authorityArn == "" {
					return nil, fmt.Errorf("Unexpected format for ID (%q), expected ACM PCA Certificate ARN", d.Id())
				}

				d.Set("certificate_authority_arn", authorityArn)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_chain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_authority_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"certificate_signing_request": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"signing_algorithm": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(acmpca.SigningAlgorithm_Values(), false),
			},
			"validity": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(acmpca.ValidityPeriodType_Values(), false),
						},
						"value": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateTypeStringIsDateOrPositiveInt,
						},
					},
				},
			},
			"template_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateAcmPcaTemplateArn,
			},
		},
	}
}

func resourceAwsAcmpcaCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn

	certificateAuthorityArn := d.Get("certificate_authority_arn").(string)
	input := &acmpca.IssueCertificateInput{
		CertificateAuthorityArn: aws.String(certificateAuthorityArn),
		Csr:                     []byte(d.Get("certificate_signing_request").(string)),
		IdempotencyToken:        aws.String(resource.UniqueId()),
		SigningAlgorithm:        aws.String(d.Get("signing_algorithm").(string)),
	}
	validity, err := expandAcmpcaValidity(d.Get("validity").([]interface{}))
	if err != nil {
		return fmt.Errorf("error issuing ACM PCA Certificate with Certificate Authority (%s): %w", certificateAuthorityArn, err)
	}
	input.Validity = validity

	if v, ok := d.Get("template_arn").(string); ok && v != "" {
		input.TemplateArn = aws.String(v)
	}

	var output *acmpca.IssueCertificateOutput
	err = resource.Retry(waiter.CertificateAuthorityActiveTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.IssueCertificate(input)
		if tfawserr.ErrMessageContains(err, acmpca.ErrCodeInvalidStateException, "The certificate authority is not in a valid state for issuing certificates") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		output, err = conn.IssueCertificate(input)
	}

	if err != nil {
		return fmt.Errorf("error issuing ACM PCA Certificate with Certificate Authority (%s): %w", certificateAuthorityArn, err)
	}

	d.SetId(aws.StringValue(output.CertificateArn))

	getCertificateInput := &acmpca.GetCertificateInput{
		CertificateArn:          output.CertificateArn,
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
	}

	err = conn.WaitUntilCertificateIssued(getCertificateInput)
	if err != nil {
		return fmt.Errorf("error waiting for ACM PCA Certificate Authority (%s) to issue Certificate (%s), error: %w", certificateAuthorityArn, d.Id(), err)
	}

	return resourceAwsAcmpcaCertificateRead(d, meta)
}

func resourceAwsAcmpcaCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn

	getCertificateInput := &acmpca.GetCertificateInput{
		CertificateArn:          aws.String(d.Id()),
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
	}

	log.Printf("[DEBUG] Reading ACM PCA Certificate: %s", getCertificateInput)

	certificateOutput, err := conn.GetCertificate(getCertificateInput)
	if err != nil {
		if isAWSErr(err, acmpca.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] ACM PCA Certificate (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading ACM PCA Certificate: %s", err)
	}

	d.Set("arn", d.Id())
	d.Set("certificate", aws.StringValue(certificateOutput.Certificate))
	d.Set("certificate_chain", aws.StringValue(certificateOutput.CertificateChain))

	return nil
}

func resourceAwsAcmpcaCertificateRevoke(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmpcaconn

	block, _ := pem.Decode([]byte(d.Get("certificate").(string)))
	if block == nil {
		log.Printf("[WARN] Failed to parse ACM PCA Certificate (%s)", d.Id())
		return nil
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("Failed to parse ACM PCA Certificate (%s): %w", d.Id(), err)
	}

	input := &acmpca.RevokeCertificateInput{
		CertificateAuthorityArn: aws.String(d.Get("certificate_authority_arn").(string)),
		CertificateSerial:       aws.String(fmt.Sprintf("%x", cert.SerialNumber)),
		RevocationReason:        aws.String(acmpca.RevocationReasonUnspecified),
	}
	_, err = conn.RevokeCertificate(input)

	if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrCodeEquals(err, acmpca.ErrCodeRequestAlreadyProcessedException) ||
		tfawserr.ErrCodeEquals(err, acmpca.ErrCodeRequestInProgressException) ||
		tfawserr.ErrMessageContains(err, acmpca.ErrCodeInvalidRequestException, "Self-signed certificate can not be revoked") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error revoking ACM PCA Certificate (%s): %w", d.Id(), err)
	}

	return nil
}

func validateAcmPcaTemplateArn(v interface{}, k string) (ws []string, errors []error) {
	wsARN, errorsARN := validateArn(v, k)
	ws = append(ws, wsARN...)
	errors = append(errors, errorsARN...)

	if len(errors) == 0 {
		value := v.(string)
		parsedARN, _ := arn.Parse(value)

		if parsedARN.Service != acmpca.ServiceName {
			errors = append(errors, fmt.Errorf("%q (%s) is not a valid ACM PCA template ARN: service must be \""+acmpca.ServiceName+"\", was %q)", k, value, parsedARN.Service))
		}

		if parsedARN.Region != "" {
			errors = append(errors, fmt.Errorf("%q (%s) is not a valid ACM PCA template ARN: region must be empty, was %q)", k, value, parsedARN.Region))
		}

		if parsedARN.AccountID != "" {
			errors = append(errors, fmt.Errorf("%q (%s) is not a valid ACM PCA template ARN: account ID must be empty, was %q)", k, value, parsedARN.AccountID))
		}

		if !strings.HasPrefix(parsedARN.Resource, "template/") {
			errors = append(errors, fmt.Errorf("%q (%s) is not a valid ACM PCA template ARN: expected resource to start with \"template/\", was %q)", k, value, parsedARN.Resource))
		}
	}

	return ws, errors
}

func expandAcmpcaValidity(l []interface{}) (*acmpca.Validity, error) {
	if len(l) == 0 {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	valueType := m["type"].(string)
	result := &acmpca.Validity{
		Type: aws.String(valueType),
	}

	i, err := expandAcmpcaValidityValue(valueType, m["value"].(string))
	if err != nil {
		return nil, fmt.Errorf("error parsing value %q: %w", m["value"].(string), err)
	}
	result.Value = aws.Int64(i)

	return result, nil
}

func expandAcmpcaValidityValue(valueType, v string) (int64, error) {
	if valueType == acmpca.ValidityPeriodTypeEndDate {
		date, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return 0, err
		}
		v = date.UTC().Format("20060102150405") // YYYYMMDDHHMMSS
	}

	return strconv.ParseInt(v, 10, 64)
}
