package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsGlueResourcePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueResourcePolicyPut(glue.ExistConditionNotExist),
		Read:   resourceAwsGlueResourcePolicyRead,
		Update: resourceAwsGlueResourcePolicyPut(glue.ExistConditionMustExist),
		Delete: resourceAwsGlueResourcePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
		},
	}
}

func resourceAwsGlueResourcePolicyPut(condition string) func(d *schema.ResourceData, meta interface{}) error {
	return func(d *schema.ResourceData, meta interface{}) error {
		conn := meta.(*AWSClient).glueconn

		_, err := conn.PutResourcePolicy(&glue.PutResourcePolicyInput{
			PolicyInJson:          aws.String(d.Get("policy").(string)),
			PolicyExistsCondition: aws.String(condition),
		})
		if err != nil {
			return fmt.Errorf("error putting policy request: %s", err)
		}
		d.SetId(meta.(*AWSClient).region)
		return resourceAwsGlueResourcePolicyRead(d, meta)
	}
}

func resourceAwsGlueResourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	resourcePolicy, err := conn.GetResourcePolicy(&glue.GetResourcePolicyInput{})
	if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
		log.Printf("[WARN] Glue Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading policy request: %s", err)
	}

	if *resourcePolicy.PolicyInJson == "" {
		//Since the glue resource policy is global we expect it to be deleted when the policy is empty
		d.SetId("")
	} else {
		d.Set("policy", resourcePolicy.PolicyInJson)
	}
	return nil
}

func resourceAwsGlueResourcePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	_, err := conn.DeleteResourcePolicy(&glue.DeleteResourcePolicyInput{})
	if err != nil {
		return fmt.Errorf("error deleting policy request: %s", err)
	}

	return nil
}
