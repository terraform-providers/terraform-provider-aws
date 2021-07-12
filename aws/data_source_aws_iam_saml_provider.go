package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsIAMSamlProvider() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIAMSamlProviderRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"saml_metadata_document": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"valid_until": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsIAMSamlProviderRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	arn := d.Get("arn").(string)

	req := &iam.GetSAMLProviderInput{
		SAMLProviderArn: aws.String(arn),
	}

	log.Printf("[DEBUG] Reading IAM SAML Provider: %s", req)
	resp, err := iamconn.GetSAMLProvider(req)
	if err != nil {
		return fmt.Errorf("Error getting SAML provider: %s", err)
	}
	if resp == nil {
		return fmt.Errorf("no SAML provider found")
	}

	d.SetId(aws.StringValue(req.SAMLProviderArn))

	validUntil := resp.ValidUntil.Format(time.RFC1123)
	dateCreated := resp.CreateDate.Format(time.RFC1123)

	d.Set("create_date", dateCreated)
	d.Set("saml_metadata_document", resp.SAMLMetadataDocument)
	d.Set("valid_until", validUntil)

	return nil
}
