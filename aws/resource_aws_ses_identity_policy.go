package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsSesIdentityPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesIdentityPolicyCreate,
		Read:   resourceAwsSesIdentityPolicyRead,
		Update: resourceAwsSesIdentityPolicyUpdate,
		Delete: resourceAwsSesIdentityPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"identity": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\-\_]+$`), "must contain only alphanumeric characters, dashes, and underscores"),
				),
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
		},
	}
}

func resourceAwsSesIdentityPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	identity := d.Get("identity").(string)
	policyName := d.Get("name").(string)

	input := &ses.PutIdentityPolicyInput{
		Identity:   aws.String(identity),
		PolicyName: aws.String(policyName),
		Policy:     aws.String(d.Get("policy").(string)),
	}

	_, err := conn.PutIdentityPolicy(input)
	if err != nil {
		return fmt.Errorf("error creating SES Identity (%s) Policy: %s", identity, err)
	}

	d.SetId(fmt.Sprintf("%s|%s", identity, policyName))

	return resourceAwsSesIdentityPolicyRead(d, meta)
}

func resourceAwsSesIdentityPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	identity, policyName, err := resourceAwsSesIdentityPolicyParseID(d.Id())
	if err != nil {
		return err
	}

	req := ses.PutIdentityPolicyInput{
		Identity:   aws.String(identity),
		PolicyName: aws.String(policyName),
		Policy:     aws.String(d.Get("policy").(string)),
	}

	_, err = conn.PutIdentityPolicy(&req)
	if err != nil {
		return fmt.Errorf("error updating SES Identity (%s) Policy (%s): %s", identity, policyName, err)
	}

	return resourceAwsSesIdentityPolicyRead(d, meta)
}

func resourceAwsSesIdentityPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	identity, policyName, err := resourceAwsSesIdentityPolicyParseID(d.Id())
	if err != nil {
		return err
	}

	input := &ses.GetIdentityPoliciesInput{
		Identity:    aws.String(identity),
		PolicyNames: aws.StringSlice([]string{policyName}),
	}

	output, err := conn.GetIdentityPolicies(input)

	if err != nil {
		return fmt.Errorf("error getting SES Identity (%s) Policy (%s): %s", identity, policyName, err)
	}

	if output == nil {
		return fmt.Errorf("error getting SES Identity (%s) Policy (%s): empty result", identity, policyName)
	}

	if len(output.Policies) == 0 {
		log.Printf("[WARN] SES Identity (%s) Policy (%s) not found, removing from state", identity, policyName)
		d.SetId("")
		return nil
	}

	policy, ok := output.Policies[policyName]
	if !ok {
		log.Printf("[WARN] SES Identity (%s) Policy (%s) not found, removing from state", identity, policyName)
		d.SetId("")
		return nil
	}

	d.Set("identity", identity)
	d.Set("name", policyName)
	d.Set("policy", policy)

	return nil
}

func resourceAwsSesIdentityPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	identity, policyName, err := resourceAwsSesIdentityPolicyParseID(d.Id())
	if err != nil {
		return err
	}

	input := &ses.DeleteIdentityPolicyInput{
		Identity:   aws.String(identity),
		PolicyName: aws.String(policyName),
	}

	log.Printf("[DEBUG] Deleting SES Identity Policy: %s", input)
	_, err = conn.DeleteIdentityPolicy(input)

	if err != nil {
		return fmt.Errorf("error deleting SES Identity (%s) Policy (%s): %s", identity, policyName, err)
	}

	return nil
}

func resourceAwsSesIdentityPolicyParseID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "|", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected IDENTITY|NAME", id)
	}
	return idParts[0], idParts[1], nil
}
