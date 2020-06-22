package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsIamSamlProvider() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamSamlProviderCreate,
		Read:   resourceAwsIamSamlProviderRead,
		Update: resourceAwsIamSamlProviderUpdate,
		Delete: resourceAwsIamSamlProviderDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"valid_until": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"saml_metadata_document": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsIamSamlProviderCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	input := &iam.CreateSAMLProviderInput{
		Name:                 aws.String(d.Get("name").(string)),
		SAMLMetadataDocument: aws.String(d.Get("saml_metadata_document").(string)),
	}

	out, err := conn.CreateSAMLProvider(input)
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(out.SAMLProviderArn))

	return resourceAwsIamSamlProviderRead(d, meta)
}

func resourceAwsIamSamlProviderRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	input := &iam.GetSAMLProviderInput{
		SAMLProviderArn: aws.String(d.Id()),
	}
	out, err := conn.GetSAMLProvider(input)
	if err != nil {
		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			log.Printf("[WARN] IAM SAML Provider %q not found.", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("arn", d.Id())
	name, err := extractNameFromIAMSamlProviderArn(d.Id())
	if err != nil {
		return err
	}
	d.Set("name", name)
	d.Set("valid_until", out.ValidUntil.Format(time.RFC1123))
	d.Set("saml_metadata_document", out.SAMLMetadataDocument)

	return nil
}

func resourceAwsIamSamlProviderUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	input := &iam.UpdateSAMLProviderInput{
		SAMLProviderArn:      aws.String(d.Id()),
		SAMLMetadataDocument: aws.String(d.Get("saml_metadata_document").(string)),
	}
	_, err := conn.UpdateSAMLProvider(input)
	if err != nil {
		return err
	}

	return resourceAwsIamSamlProviderRead(d, meta)
}

func resourceAwsIamSamlProviderDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	input := &iam.DeleteSAMLProviderInput{
		SAMLProviderArn: aws.String(d.Id()),
	}
	_, err := conn.DeleteSAMLProvider(input)

	return err
}

func extractNameFromIAMSamlProviderArn(samlArn string) (string, error) {
	parsedArn, err := arn.Parse(samlArn)
	if err != nil {
		return "", fmt.Errorf("Unable to extract name from a given ARN: %q", samlArn)
	}

	name := strings.TrimPrefix(parsedArn.Resource, "saml-provider/")

	return name, nil
}
