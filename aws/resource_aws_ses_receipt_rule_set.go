package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsSesReceiptRuleSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesReceiptRuleSetCreate,
		Read:   resourceAwsSesReceiptRuleSetRead,
		Delete: resourceAwsSesReceiptRuleSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"rule_set_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsSesReceiptRuleSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	ruleSetName := d.Get("rule_set_name").(string)

	createOpts := &ses.CreateReceiptRuleSetInput{
		RuleSetName: aws.String(ruleSetName),
	}

	_, err := conn.CreateReceiptRuleSet(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating SES rule set: %s", err)
	}

	d.SetId(ruleSetName)

	return resourceAwsSesReceiptRuleSetRead(d, meta)
}

func resourceAwsSesReceiptRuleSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	input := &ses.DescribeReceiptRuleSetInput{
		RuleSetName: aws.String(d.Id()),
	}

	_, err := conn.DescribeReceiptRuleSet(input)

	if isAWSErr(err, ses.ErrCodeRuleSetDoesNotExistException, "") {
		log.Printf("[WARN] SES Receipt Rule Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing SES Receipt Rule Set (%s): %s", d.Id(), err)
	}

	d.Set("rule_set_name", d.Id())

	return nil
}

func resourceAwsSesReceiptRuleSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	log.Printf("[DEBUG] SES Delete Receipt Rule Set: %s", d.Id())
	_, err := conn.DeleteReceiptRuleSet(&ses.DeleteReceiptRuleSetInput{
		RuleSetName: aws.String(d.Id()),
	})

	return err
}
