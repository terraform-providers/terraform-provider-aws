package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsCloudFormationStack() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudFormationStackRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"template_body": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capabilities": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disable_rollback": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"notification_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"outputs": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"timeout_in_minutes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"iam_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsCloudFormationStackRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	}

	log.Printf("[DEBUG] Reading CloudFormation Stack: %s", input)
	out, err := conn.DescribeStacks(input)
	if err != nil {
		return fmt.Errorf("Failed describing CloudFormation stack (%s): %s", name, err)
	}
	if l := len(out.Stacks); l != 1 {
		return fmt.Errorf("Expected 1 CloudFormation stack (%s), found %d", name, l)
	}
	stack := out.Stacks[0]
	d.SetId(*stack.StackId)

	d.Set("description", stack.Description)
	d.Set("disable_rollback", stack.DisableRollback)
	d.Set("timeout_in_minutes", stack.TimeoutInMinutes)
	d.Set("iam_role_arn", stack.RoleARN)

	if len(stack.NotificationARNs) > 0 {
		d.Set("notification_arns", schema.NewSet(schema.HashString, flattenStringList(stack.NotificationARNs)))
	}

	d.Set("parameters", flattenAllCloudFormationParameters(stack.Parameters))
	if err := d.Set("tags", keyvaluetags.CloudformationKeyValueTags(stack.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}
	d.Set("outputs", flattenCloudFormationOutputs(stack.Outputs))

	if len(stack.Capabilities) > 0 {
		d.Set("capabilities", schema.NewSet(schema.HashString, flattenStringList(stack.Capabilities)))
	}

	tInput := cloudformation.GetTemplateInput{
		StackName: aws.String(name),
	}
	tOut, err := conn.GetTemplate(&tInput)
	if err != nil {
		return err
	}

	template, err := normalizeCloudFormationTemplate(*tOut.TemplateBody)
	if err != nil {
		return fmt.Errorf("template body contains an invalid JSON or YAML: %s", err)
	}
	d.Set("template_body", template)

	return nil
}
