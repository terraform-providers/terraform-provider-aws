package aws

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	// Maximum amount of time for ACM Certificate cross-service reference propagation.
	// Removal of ACM Certificates from API Gateway Custom Domains can take >15 minutes.
	AcmCertificateCrossServicePropagationTimeout = 20 * time.Minute

	// Maximum amount of time for ACM Certificate asynchronous DNS validation record assignment.
	// This timeout is unrelated to any creation or validation of those assigned DNS records.
	AcmCertificateDnsValidationAssignmentTimeout = 5 * time.Minute
)

func resourceAwsAcmCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAcmCertificateCreate,
		Read:   resourceAwsAcmCertificateRead,
		Update: resourceAwsAcmCertificateUpdate,
		Delete: resourceAwsAcmCertificateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"certificate_body": {
				Type:      schema.TypeString,
				Optional:  true,
				StateFunc: normalizeCert,
			},

			"certificate_chain": {
				Type:      schema.TypeString,
				Optional:  true,
				StateFunc: normalizeCert,
			},
			"private_key": {
				Type:      schema.TypeString,
				Optional:  true,
				StateFunc: normalizeCert,
				Sensitive: true,
			},
			"certificate_authority_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"domain_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"private_key", "certificate_body", "certificate_chain"},
				StateFunc: func(v interface{}) string {
					// AWS Provider 1.42.0+ aws_route53_zone references may contain a
					// trailing period, which generates an ACM API error
					return strings.TrimSuffix(v.(string), ".")
				},
			},
			"subject_alternative_names": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"private_key", "certificate_body", "certificate_chain"},
				Elem: &schema.Schema{
					Type: schema.TypeString,
					StateFunc: func(v interface{}) string {
						// AWS Provider 1.42.0+ aws_route53_zone references may contain a
						// trailing period, which generates an ACM API error
						return strings.TrimSuffix(v.(string), ".")
					},
				},
			},
			"validation_method": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"private_key", "certificate_body", "certificate_chain", "certificate_authority_arn"},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_validation_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_record_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_record_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_record_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"validation_emails": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if _, ok := d.GetOk("private_key"); ok {
						// ignore diffs for imported certs; they have a different logging preference
						// default to requested certs which can't be changed by the ImportCertificate API
						return true
					}
					// behave just like suppressMissingOptionalConfigurationBlock() for requested certs
					return old == "1" && new == "0"
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_transparency_logging_preference": {
							Type:          schema.TypeString,
							Optional:      true,
							Default:       acm.CertificateTransparencyLoggingPreferenceEnabled,
							ForceNew:      true,
							ConflictsWith: []string{"private_key", "certificate_body", "certificate_chain"},
							ValidateFunc: validation.StringInSlice([]string{
								acm.CertificateTransparencyLoggingPreferenceEnabled,
								acm.CertificateTransparencyLoggingPreferenceDisabled,
							}, false),
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsAcmCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	if _, ok := d.GetOk("domain_name"); ok {
		if _, ok := d.GetOk("certificate_authority_arn"); ok {
			return resourceAwsAcmCertificateCreateRequested(d, meta)
		}

		if _, ok := d.GetOk("validation_method"); !ok {
			return errors.New("validation_method must be set when creating a certificate")
		}
		return resourceAwsAcmCertificateCreateRequested(d, meta)
	} else if _, ok := d.GetOk("private_key"); ok {
		if _, ok := d.GetOk("certificate_body"); !ok {
			return errors.New("certificate_body must be set when importing a certificate with private_key")
		}
		return resourceAwsAcmCertificateCreateImported(d, meta)
	}
	return errors.New("certificate must be imported (private_key) or created (domain_name)")
}

func resourceAwsAcmCertificateCreateImported(d *schema.ResourceData, meta interface{}) error {
	acmconn := meta.(*AWSClient).acmconn
	resp, err := resourceAwsAcmCertificateImport(acmconn, d, false)
	if err != nil {
		return fmt.Errorf("Error importing certificate: %s", err)
	}

	d.SetId(*resp.CertificateArn)

	return resourceAwsAcmCertificateRead(d, meta)
}

func resourceAwsAcmCertificateCreateRequested(d *schema.ResourceData, meta interface{}) error {
	acmconn := meta.(*AWSClient).acmconn
	params := &acm.RequestCertificateInput{
		DomainName:       aws.String(strings.TrimSuffix(d.Get("domain_name").(string), ".")),
		IdempotencyToken: aws.String(resource.PrefixedUniqueId("tf")), // 32 character limit
		Options:          expandAcmCertificateOptions(d.Get("options").([]interface{})),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		params.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AcmTags()
	}

	if caARN, ok := d.GetOk("certificate_authority_arn"); ok {
		params.CertificateAuthorityArn = aws.String(caARN.(string))
	}

	if sans, ok := d.GetOk("subject_alternative_names"); ok {
		subjectAlternativeNames := make([]*string, len(sans.([]interface{})))
		for i, sanRaw := range sans.([]interface{}) {
			subjectAlternativeNames[i] = aws.String(strings.TrimSuffix(sanRaw.(string), "."))
		}
		params.SubjectAlternativeNames = subjectAlternativeNames
	}

	if v, ok := d.GetOk("validation_method"); ok {
		params.ValidationMethod = aws.String(v.(string))
	}

	log.Printf("[DEBUG] ACM Certificate Request: %#v", params)
	resp, err := acmconn.RequestCertificate(params)

	if err != nil {
		return fmt.Errorf("Error requesting certificate: %s", err)
	}

	d.SetId(*resp.CertificateArn)

	return resourceAwsAcmCertificateRead(d, meta)
}

func resourceAwsAcmCertificateRead(d *schema.ResourceData, meta interface{}) error {
	acmconn := meta.(*AWSClient).acmconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	params := &acm.DescribeCertificateInput{
		CertificateArn: aws.String(d.Id()),
	}

	return resource.Retry(AcmCertificateDnsValidationAssignmentTimeout, func() *resource.RetryError {
		resp, err := acmconn.DescribeCertificate(params)

		if err != nil {
			if isAWSErr(err, acm.ErrCodeResourceNotFoundException, "") {
				d.SetId("")
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error describing certificate: %s", err))
		}

		d.Set("domain_name", resp.Certificate.DomainName)
		d.Set("arn", resp.Certificate.CertificateArn)
		d.Set("certificate_authority_arn", resp.Certificate.CertificateAuthorityArn)

		if err := d.Set("subject_alternative_names", cleanUpSubjectAlternativeNames(resp.Certificate)); err != nil {
			return resource.NonRetryableError(err)
		}

		domainValidationOptions, emailValidationOptions, err := convertValidationOptions(resp.Certificate)

		if err != nil {
			return resource.RetryableError(err)
		}

		if err := d.Set("domain_validation_options", domainValidationOptions); err != nil {
			return resource.NonRetryableError(err)
		}
		if err := d.Set("validation_emails", emailValidationOptions); err != nil {
			return resource.NonRetryableError(err)
		}

		d.Set("validation_method", resourceAwsAcmCertificateValidationMethod(resp.Certificate))

		if err := d.Set("options", flattenAcmCertificateOptions(resp.Certificate.Options)); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error setting certificate options: %s", err))
		}

		d.Set("status", resp.Certificate.Status)

		tags, err := keyvaluetags.AcmListTags(acmconn, d.Id())

		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error listing tags for ACM Certificate (%s): %s", d.Id(), err))
		}

		if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error setting tags: %s", err))
		}

		return nil
	})
}
func resourceAwsAcmCertificateValidationMethod(certificate *acm.CertificateDetail) string {
	if aws.StringValue(certificate.Type) == acm.CertificateTypeAmazonIssued {
		for _, domainValidation := range certificate.DomainValidationOptions {
			if domainValidation.ValidationMethod != nil {
				return aws.StringValue(domainValidation.ValidationMethod)
			}
		}
	}

	return "NONE"
}

func resourceAwsAcmCertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	acmconn := meta.(*AWSClient).acmconn

	if d.HasChanges("private_key", "certificate_body", "certificate_chain") {
		_, err := resourceAwsAcmCertificateImport(acmconn, d, true)
		if err != nil {
			return fmt.Errorf("Error updating certificate: %s", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.AcmUpdateTags(acmconn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}
	return resourceAwsAcmCertificateRead(d, meta)
}

func cleanUpSubjectAlternativeNames(cert *acm.CertificateDetail) []string {
	sans := cert.SubjectAlternativeNames
	vs := make([]string, 0)
	for _, v := range sans {
		if aws.StringValue(v) != aws.StringValue(cert.DomainName) {
			vs = append(vs, aws.StringValue(v))
		}
	}
	return vs

}

func convertValidationOptions(certificate *acm.CertificateDetail) ([]map[string]interface{}, []string, error) {
	var domainValidationResult []map[string]interface{}
	var emailValidationResult []string

	switch aws.StringValue(certificate.Type) {
	case acm.CertificateTypeAmazonIssued:
		if len(certificate.DomainValidationOptions) == 0 && aws.StringValue(certificate.Status) == acm.DomainStatusPendingValidation {
			log.Printf("[DEBUG] No validation options need to retry.")
			return nil, nil, fmt.Errorf("No validation options need to retry.")
		}
		for _, o := range certificate.DomainValidationOptions {
			if o.ResourceRecord != nil {
				validationOption := map[string]interface{}{
					"domain_name":           *o.DomainName,
					"resource_record_name":  *o.ResourceRecord.Name,
					"resource_record_type":  *o.ResourceRecord.Type,
					"resource_record_value": *o.ResourceRecord.Value,
				}
				domainValidationResult = append(domainValidationResult, validationOption)
			} else if o.ValidationEmails != nil && len(o.ValidationEmails) > 0 {
				for _, validationEmail := range o.ValidationEmails {
					emailValidationResult = append(emailValidationResult, *validationEmail)
				}
			} else if o.ValidationStatus == nil || aws.StringValue(o.ValidationStatus) == acm.DomainStatusPendingValidation {
				log.Printf("[DEBUG] Asynchronous ACM service domain validation assignment not complete, need to retry: %#v", o)
				return nil, nil, fmt.Errorf("asynchronous ACM service domain validation assignment not complete, need to retry: %#v", o)
			}
		}
	case acm.CertificateTypePrivate:
		// While ACM PRIVATE certificates do not need to be validated, there is a slight delay for
		// the API to fill in all certificate details, which is during the PENDING_VALIDATION status.
		if aws.StringValue(certificate.Status) == acm.DomainStatusPendingValidation {
			return nil, nil, fmt.Errorf("certificate still pending issuance")
		}
	}

	return domainValidationResult, emailValidationResult, nil
}

func resourceAwsAcmCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	acmconn := meta.(*AWSClient).acmconn

	log.Printf("[INFO] Deleting ACM Certificate: %s", d.Id())

	params := &acm.DeleteCertificateInput{
		CertificateArn: aws.String(d.Id()),
	}

	err := resource.Retry(AcmCertificateCrossServicePropagationTimeout, func() *resource.RetryError {
		_, err := acmconn.DeleteCertificate(params)
		if err != nil {
			if isAWSErr(err, acm.ErrCodeResourceInUseException, "") {
				log.Printf("[WARN] Conflict deleting certificate in use: %s, retrying", err.Error())
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil && !isAWSErr(err, acm.ErrCodeResourceNotFoundException, "") {
		return fmt.Errorf("Error deleting certificate: %s", err)
	}

	return nil
}

func resourceAwsAcmCertificateImport(conn *acm.ACM, d *schema.ResourceData, update bool) (*acm.ImportCertificateOutput, error) {
	params := &acm.ImportCertificateInput{
		PrivateKey:  []byte(d.Get("private_key").(string)),
		Certificate: []byte(d.Get("certificate_body").(string)),
	}
	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		params.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AcmTags()
	}
	if chain, ok := d.GetOk("certificate_chain"); ok {
		params.CertificateChain = []byte(chain.(string))
	}
	if update {
		params.CertificateArn = aws.String(d.Get("arn").(string))
	}

	log.Printf("[DEBUG] ACM Certificate Import: %#v", params)
	return conn.ImportCertificate(params)
}

func expandAcmCertificateOptions(l []interface{}) *acm.CertificateOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	options := &acm.CertificateOptions{}

	if v, ok := m["certificate_transparency_logging_preference"]; ok {
		options.CertificateTransparencyLoggingPreference = aws.String(v.(string))
	}

	return options
}

func flattenAcmCertificateOptions(co *acm.CertificateOptions) []interface{} {
	m := map[string]interface{}{
		"certificate_transparency_logging_preference": aws.StringValue(co.CertificateTransparencyLoggingPreference),
	}

	return []interface{}{m}
}
