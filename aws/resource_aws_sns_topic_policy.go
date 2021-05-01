package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsSnsTopicPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSnsTopicPolicyUpsert,
		Read:   resourceAwsSnsTopicPolicyRead,
		Update: resourceAwsSnsTopicPolicyUpsert,
		Delete: resourceAwsSnsTopicPolicyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsSnsTopicPolicyUpsert(d *schema.ResourceData, meta interface{}) error {
	arn := d.Get("arn").(string)
	req := sns.SetTopicAttributesInput{
		TopicArn:       aws.String(arn),
		AttributeName:  aws.String("Policy"),
		AttributeValue: aws.String(d.Get("policy").(string)),
	}

	d.SetId(arn)

	// Retry the update in the event of an eventually consistent style of
	// error, where say an IAM resource is successfully created but not
	// actually available. See https://github.com/hashicorp/terraform/issues/3660
	conn := meta.(*AWSClient).snsconn
	_, err := retryOnAwsCode("InvalidParameter", func() (interface{}, error) {
		return conn.SetTopicAttributes(&req)
	})
	if err != nil {
		return err
	}

	return resourceAwsSnsTopicPolicyRead(d, meta)
}

func resourceAwsSnsTopicPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).snsconn

	attributeOutput, err := conn.GetTopicAttributes(&sns.GetTopicAttributesInput{
		TopicArn: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, sns.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] SNS Topic (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	if attributeOutput.Attributes == nil {
		log.Printf("[WARN] SNS Topic (%q) attributes not found (nil), removing from state", d.Id())
		d.SetId("")
		return nil
	}
	attrmap := attributeOutput.Attributes

	policy, ok := attrmap["Policy"]
	if !ok {
		log.Printf("[WARN] SNS Topic (%q) policy not found in attributes, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("policy", policy)
	d.Set("arn", attrmap["TopicArn"])
	d.Set("owner", attrmap["Owner"])

	return nil
}

func resourceAwsSnsTopicPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	req := sns.SetTopicAttributesInput{
		TopicArn:      aws.String(d.Id()),
		AttributeName: aws.String("Policy"),
		// It is impossible to delete a policy or set to empty
		// (confirmed by AWS Support representative)
		// so we instead set it back to the default one
		AttributeValue: aws.String(buildDefaultSnsTopicPolicy(d.Id(), d.Get("owner").(string))),
	}

	// Retry the update in the event of an eventually consistent style of
	// error, where say an IAM resource is successfully created but not
	// actually available. See https://github.com/hashicorp/terraform/issues/3660
	log.Printf("[DEBUG] Resetting SNS Topic Policy to default: %s", req)
	conn := meta.(*AWSClient).snsconn
	_, err := retryOnAwsCode("InvalidParameter", func() (interface{}, error) {
		return conn.SetTopicAttributes(&req)
	})
	return err
}

func buildDefaultSnsTopicPolicy(topicArn, accountId string) string {
	return fmt.Sprintf(`{
  "Version": "2008-10-17",
  "Id": "__default_policy_ID",
  "Statement": [
    {
      "Sid": "__default_statement_ID",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": [
        "SNS:GetTopicAttributes",
        "SNS:SetTopicAttributes",
        "SNS:AddPermission",
        "SNS:RemovePermission",
        "SNS:DeleteTopic",
        "SNS:Subscribe",
        "SNS:ListSubscriptionsByTopic",
        "SNS:Publish",
        "SNS:Receive"
      ],
      "Resource": "%s",
      "Condition": {
        "StringEquals": {
          "AWS:SourceOwner": "%s"
        }
      }
    }
  ]
}
`, topicArn, accountId)
}
