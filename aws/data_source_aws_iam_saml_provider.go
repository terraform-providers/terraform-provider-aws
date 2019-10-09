package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsIAMSamlProvider() *schema.Resource {
	return &schema.Resouce{
		Read: dataSourceAwsIAMSamlProviderRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			}
			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"saml_metadata_document": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"valid_until": {
				Type:	  schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsIAMSamlProviderRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	arn = d.Get("arn").(string)

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

	d.SetId(*req.SAMLProviderArn)
	d.Set("create_date", resp.CreateDate)
	d.Set("saml_metadata_document", resp.SAMLMetadataDocument)
	d.Set("valid_until", resp.ValidUntil)

	return nil
}
