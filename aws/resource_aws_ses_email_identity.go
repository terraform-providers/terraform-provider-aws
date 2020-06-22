package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsSesEmailIdentity() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesEmailIdentityCreate,
		Read:   resourceAwsSesEmailIdentityRead,
		Delete: resourceAwsSesEmailIdentityDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					return strings.TrimSuffix(v.(string), ".")
				},
			},
		},
	}
}

func resourceAwsSesEmailIdentityCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	email := d.Get("email").(string)
	email = strings.TrimSuffix(email, ".")

	createOpts := &ses.VerifyEmailIdentityInput{
		EmailAddress: aws.String(email),
	}

	_, err := conn.VerifyEmailIdentity(createOpts)
	if err != nil {
		return fmt.Errorf("Error requesting SES email identity verification: %s", err)
	}

	d.SetId(email)

	return resourceAwsSesEmailIdentityRead(d, meta)
}

func resourceAwsSesEmailIdentityRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	email := d.Id()
	d.Set("email", email)

	readOpts := &ses.GetIdentityVerificationAttributesInput{
		Identities: []*string{
			aws.String(email),
		},
	}

	response, err := conn.GetIdentityVerificationAttributes(readOpts)
	if err != nil {
		log.Printf("[WARN] Error fetching identity verification attributes for %s: %s", d.Id(), err)
		return err
	}

	_, ok := response.VerificationAttributes[email]
	if !ok {
		log.Printf("[WARN] Email not listed in response when fetching verification attributes for %s", d.Id())
		d.SetId("")
		return nil
	}

	arn := arn.ARN{
		AccountID: meta.(*AWSClient).accountid,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("identity/%s", d.Id()),
		Service:   "ses",
	}.String()
	d.Set("arn", arn)
	return nil
}

func resourceAwsSesEmailIdentityDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	email := d.Get("email").(string)

	deleteOpts := &ses.DeleteIdentityInput{
		Identity: aws.String(email),
	}

	_, err := conn.DeleteIdentity(deleteOpts)
	if err != nil {
		return fmt.Errorf("Error deleting SES email identity: %s", err)
	}

	return nil
}
