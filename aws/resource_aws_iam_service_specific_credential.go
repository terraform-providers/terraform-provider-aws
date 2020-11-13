package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsIamServiceSpecificCredential() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamServiceSpecificCredentialCreate,
		Read:   resourceAwsIamServiceSpecificCredentialRead,
		Delete: resourceAwsIamServiceSpecificCredentialDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"service_password": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
			"service_user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_specific_credential_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsIamServiceSpecificCredentialCreate(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	input := &iam.CreateServiceSpecificCredentialInput{
		ServiceName: aws.String(d.Get("service_name").(string)),
		UserName:    aws.String(d.Get("user_name").(string)),
	}

	out, err := iamconn.CreateServiceSpecificCredential(input)
	if err != nil {
		return fmt.Errorf("error creating IAM Service Specific Credential: %w", err)
	}

	cred := out.ServiceSpecificCredential

	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(cred.ServiceName), aws.StringValue(cred.UserName)))
	d.Set("service_password", cred.ServicePassword)

	return resourceAwsIamServiceSpecificCredentialRead(d, meta)
}

func resourceAwsIamServiceSpecificCredentialRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	serviceName, userName, err := decodeAwsIamServiceSpecificCredential(d.Id())
	if err != nil {
		return err
	}

	input := &iam.ListServiceSpecificCredentialsInput{
		ServiceName: aws.String(serviceName),
		UserName:    aws.String(userName),
	}

	out, err := conn.ListServiceSpecificCredentials(input)
	if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
		log.Printf("[WARN] IAM Service Specific Credential (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading IAM Service Specific Credential (%s): %w", d.Id(), err)
	}

	if out == nil || len(out.ServiceSpecificCredentials) == 0 {
		log.Printf("[WARN] IAM Service Specific Credential (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if len(out.ServiceSpecificCredentials) > 1 {
		return fmt.Errorf("error reading IAM Service Specific Credential: multiple results found, try adjusting search criteria")
	}

	cred := out.ServiceSpecificCredentials[0]

	d.Set("service_specific_credential_id", cred.ServiceSpecificCredentialId)
	d.Set("service_user_name", cred.ServiceUserName)
	d.Set("service_name", cred.ServiceName)
	d.Set("user_name", cred.UserName)
	d.Set("status", cred.Status)

	return nil
}

func resourceAwsIamServiceSpecificCredentialDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	_, userName, err := decodeAwsIamServiceSpecificCredential(d.Id())
	if err != nil {
		return err
	}

	input := &iam.DeleteServiceSpecificCredentialInput{
		ServiceSpecificCredentialId: aws.String(d.Get("service_specific_credential_id").(string)),
		UserName:                    aws.String(userName),
	}

	_, err = conn.DeleteServiceSpecificCredential(input)
	if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting IAM Service Specific Credential (%s): %w", d.Id(), err)
	}

	return nil
}

func decodeAwsIamServiceSpecificCredential(id string) (string, string, error) {
	creds := strings.Split(string(id), ":")
	if len(creds) != 2 {
		return "", "", fmt.Errorf("unknown IAM Service Specific Credential ID format")
	}
	serviceName := creds[0]
	userName := creds[1]

	return serviceName, userName, nil
}
