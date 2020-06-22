package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsSesDomainMailFrom() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesDomainMailFromSet,
		Read:   resourceAwsSesDomainMailFromRead,
		Update: resourceAwsSesDomainMailFromSet,
		Delete: resourceAwsSesDomainMailFromDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mail_from_domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"behavior_on_mx_failure": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ses.BehaviorOnMXFailureUseDefaultValue,
			},
		},
	}
}

func resourceAwsSesDomainMailFromSet(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	behaviorOnMxFailure := d.Get("behavior_on_mx_failure").(string)
	domainName := d.Get("domain").(string)
	mailFromDomain := d.Get("mail_from_domain").(string)

	input := &ses.SetIdentityMailFromDomainInput{
		BehaviorOnMXFailure: aws.String(behaviorOnMxFailure),
		Identity:            aws.String(domainName),
		MailFromDomain:      aws.String(mailFromDomain),
	}

	_, err := conn.SetIdentityMailFromDomain(input)
	if err != nil {
		return fmt.Errorf("Error setting MAIL FROM domain: %s", err)
	}

	d.SetId(domainName)

	return resourceAwsSesDomainMailFromRead(d, meta)
}

func resourceAwsSesDomainMailFromRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	domainName := d.Id()

	readOpts := &ses.GetIdentityMailFromDomainAttributesInput{
		Identities: []*string{
			aws.String(domainName),
		},
	}

	out, err := conn.GetIdentityMailFromDomainAttributes(readOpts)

	if err != nil {
		return fmt.Errorf("error fetching SES MAIL FROM domain attributes for %s: %s", domainName, err)
	}

	if out == nil {
		return fmt.Errorf("error fetching SES MAIL FROM domain attributes for %s: empty response", domainName)
	}

	attributes, ok := out.MailFromDomainAttributes[domainName]

	if !ok {
		log.Printf("[WARN] SES Domain Identity (%s) not found, removing from state", domainName)
		d.SetId("")
		return nil
	}

	d.Set("behavior_on_mx_failure", attributes.BehaviorOnMXFailure)
	d.Set("domain", domainName)
	d.Set("mail_from_domain", attributes.MailFromDomain)

	return nil
}

func resourceAwsSesDomainMailFromDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	domainName := d.Id()

	deleteOpts := &ses.SetIdentityMailFromDomainInput{
		Identity:       aws.String(domainName),
		MailFromDomain: nil,
	}

	_, err := conn.SetIdentityMailFromDomain(deleteOpts)
	if err != nil {
		return fmt.Errorf("Error deleting SES domain identity: %s", err)
	}

	return nil
}
