package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsIotPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotPolicyCreate,
		Read:   resourceAwsIotPolicyRead,
		Update: resourceAwsIotPolicyUpdate,
		Delete: resourceAwsIotPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsIotPolicyCreate(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*AWSClient).iotconn

	out, err := conn.CreatePolicy(&iot.CreatePolicyInput{
		PolicyName:     aws.String(d.Get("name").(string)),
		PolicyDocument: aws.String(d.Get("policy").(string)),
	})

	if err != nil {
		return fmt.Errorf("error creating IoT Policy: %s", err)
	}

	d.SetId(aws.StringValue(out.PolicyName))

	return resourceAwsIotPolicyRead(d, meta)
}

func resourceAwsIotPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	out, err := conn.GetPolicy(&iot.GetPolicyInput{
		PolicyName: aws.String(d.Id()),
	})

	if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] IoT Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IoT Policy (%s): %s", d.Id(), err)
	}

	d.Set("arn", out.PolicyArn)
	d.Set("default_version_id", out.DefaultVersionId)
	d.Set("name", out.PolicyName)
	d.Set("policy", out.PolicyDocument)

	return nil
}

func resourceAwsIotPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	if d.HasChange("policy") {
		_, err := conn.CreatePolicyVersion(&iot.CreatePolicyVersionInput{
			PolicyName:     aws.String(d.Id()),
			PolicyDocument: aws.String(d.Get("policy").(string)),
			SetAsDefault:   aws.Bool(true),
		})

		if err != nil {
			return fmt.Errorf("error updating IoT Policy (%s): %s", d.Id(), err)
		}
	}

	return resourceAwsIotPolicyRead(d, meta)
}

func resourceAwsIotPolicyDelete(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*AWSClient).iotconn

	out, err := conn.ListPolicyVersions(&iot.ListPolicyVersionsInput{
		PolicyName: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error listing IoT Policy (%s) versions: %s", d.Id(), err)
	}

	// Delete all non-default versions of the policy
	for _, ver := range out.PolicyVersions {
		if !aws.BoolValue(ver.IsDefaultVersion) {
			_, err = conn.DeletePolicyVersion(&iot.DeletePolicyVersionInput{
				PolicyName:      aws.String(d.Id()),
				PolicyVersionId: ver.VersionId,
			})

			if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
				continue
			}

			if err != nil {
				return fmt.Errorf("error deleting IoT Policy (%s) version (%s): %s", d.Id(), aws.StringValue(ver.VersionId), err)
			}
		}
	}

	//Delete default policy version
	_, err = conn.DeletePolicy(&iot.DeletePolicyInput{
		PolicyName: aws.String(d.Id()),
	})

	if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting IoT Policy (%s): %s", d.Id(), err)
	}

	return nil
}
