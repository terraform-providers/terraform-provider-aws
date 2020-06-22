package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsAcmCertificate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAcmCertificateRead,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"statuses": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"key_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						acm.KeyAlgorithmEcPrime256v1,
						acm.KeyAlgorithmEcSecp384r1,
						acm.KeyAlgorithmEcSecp521r1,
						acm.KeyAlgorithmRsa1024,
						acm.KeyAlgorithmRsa2048,
						acm.KeyAlgorithmRsa4096,
					}, false),
				},
			},
			"types": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsAcmCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).acmconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	params := &acm.ListCertificatesInput{}

	if v := d.Get("key_types").(*schema.Set); v.Len() > 0 {
		params.Includes = &acm.Filters{
			KeyTypes: expandStringSet(v),
		}
	}

	target := d.Get("domain")
	statuses, ok := d.GetOk("statuses")
	if ok {
		statusStrings := statuses.([]interface{})
		params.CertificateStatuses = expandStringList(statusStrings)
	} else {
		params.CertificateStatuses = []*string{aws.String(acm.CertificateStatusIssued)}
	}

	var arns []*string
	log.Printf("[DEBUG] Reading ACM Certificate: %s", params)
	err := conn.ListCertificatesPages(params, func(page *acm.ListCertificatesOutput, lastPage bool) bool {
		for _, cert := range page.CertificateSummaryList {
			if *cert.DomainName == target {
				arns = append(arns, cert.CertificateArn)
			}
		}

		return true
	})
	if err != nil {
		return fmt.Errorf("Error listing certificates: %q", err)
	}

	if len(arns) == 0 {
		return fmt.Errorf("No certificate for domain %q found in this region", target)
	}

	filterMostRecent := d.Get("most_recent").(bool)
	filterTypes, filterTypesOk := d.GetOk("types")

	var matchedCertificate *acm.CertificateDetail

	if !filterMostRecent && !filterTypesOk && len(arns) > 1 {
		// Multiple certificates have been found and no additional filtering set
		return fmt.Errorf("Multiple certificates for domain %q found in this region", target)
	}

	typesStrings := expandStringList(filterTypes.([]interface{}))

	for _, arn := range arns {
		var err error

		input := &acm.DescribeCertificateInput{
			CertificateArn: aws.String(*arn),
		}
		log.Printf("[DEBUG] Describing ACM Certificate: %s", input)
		output, err := conn.DescribeCertificate(input)
		if err != nil {
			return fmt.Errorf("Error describing ACM certificate: %q", err)
		}
		certificate := output.Certificate

		if filterTypesOk {
			for _, certType := range typesStrings {
				if *certificate.Type == *certType {
					// We do not have a candidate certificate
					if matchedCertificate == nil {
						matchedCertificate = certificate
						break
					}
					// At this point, we already have a candidate certificate
					// Check if we are filtering by most recent and update if necessary
					if filterMostRecent {
						matchedCertificate, err = mostRecentAcmCertificate(certificate, matchedCertificate)
						if err != nil {
							return err
						}
						break
					}
					// Now we have multiple candidate certificates and we only allow one certificate
					return fmt.Errorf("Multiple certificates for domain %q found in this region", target)
				}
			}
			continue
		}
		// We do not have a candidate certificate
		if matchedCertificate == nil {
			matchedCertificate = certificate
			continue
		}
		// At this point, we already have a candidate certificate
		// Check if we are filtering by most recent and update if necessary
		if filterMostRecent {
			matchedCertificate, err = mostRecentAcmCertificate(certificate, matchedCertificate)
			if err != nil {
				return err
			}
			continue
		}
		// Now we have multiple candidate certificates and we only allow one certificate
		return fmt.Errorf("Multiple certificates for domain %q found in this region", target)
	}

	if matchedCertificate == nil {
		return fmt.Errorf("No certificate for domain %q found in this region", target)
	}

	d.SetId(time.Now().UTC().String())
	d.Set("arn", matchedCertificate.CertificateArn)

	tags, err := keyvaluetags.AcmListTags(conn, aws.StringValue(matchedCertificate.CertificateArn))

	if err != nil {
		return fmt.Errorf("error listing tags for ACM Certificate (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func mostRecentAcmCertificate(i, j *acm.CertificateDetail) (*acm.CertificateDetail, error) {
	if *i.Status != *j.Status {
		return nil, fmt.Errorf("most_recent filtering on different ACM certificate statues is not supported")
	}
	// Cover IMPORTED and ISSUED AMAZON_ISSUED certificates
	if *i.Status == acm.CertificateStatusIssued {
		if (*i.NotBefore).After(*j.NotBefore) {
			return i, nil
		}
		return j, nil
	}
	// Cover non-ISSUED AMAZON_ISSUED certificates
	if (*i.CreatedAt).After(*j.CreatedAt) {
		return i, nil
	}
	return j, nil
}
