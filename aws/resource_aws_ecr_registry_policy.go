package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsEcrRegistryPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEcrRegistryPolicyPut,
		Read:   resourceAwsEcrRegistryPolicyRead,
		Update: resourceAwsEcrRegistryPolicyPut,
		Delete: resourceAwsEcrRegistryPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
				ValidateFunc:     validation.StringIsJSON,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEcrRegistryPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	input := ecr.PutRegistryPolicyInput{
		PolicyText: aws.String(d.Get("policy").(string)),
	}

	out, err := conn.PutRegistryPolicy(&input)
	if err != nil {
		return fmt.Errorf("Error creating ECR Registry Policy: %w", err)
	}

	regID := aws.StringValue(out.RegistryId)

	d.SetId(regID)

	return resourceAwsEcrRegistryPolicyRead(d, meta)
}

func resourceAwsEcrRegistryPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	log.Printf("[DEBUG] Reading registry policy %s", d.Id())
	out, err := conn.GetRegistryPolicy(&ecr.GetRegistryPolicyInput{})
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ecr.ErrCodeRegistryPolicyNotFoundException) {
			log.Printf("[WARN] ECR Registry (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("registry_id", out.RegistryId)
	d.Set("policy", out.PolicyText)

	return nil
}

func resourceAwsEcrRegistryPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	_, err := conn.DeleteRegistryPolicy(&ecr.DeleteRegistryPolicyInput{})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRegistryPolicyNotFoundException) {
			return nil
		}
		return err
	}

	return nil
}
