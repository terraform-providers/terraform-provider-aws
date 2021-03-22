package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudformation/waiter"
)

func resourceAwsCloudFormationStack() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudFormationStackCreate,
		Read:   resourceAwsCloudFormationStackRead,
		Update: resourceAwsCloudFormationStackUpdate,
		Delete: resourceAwsCloudFormationStackDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(waiter.StackCreatedDefaultTimeout),
			Update: schema.DefaultTimeout(waiter.StackUpdatedDefaultTimeout),
			Delete: schema.DefaultTimeout(waiter.StackDeletedDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"template_body": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateStringIsJsonOrYaml,
				StateFunc: func(v interface{}) string {
					template, _ := normalizeJsonOrYamlString(v)
					return template
				},
			},
			"template_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"capabilities": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(cloudformation.Capability_Values(), false),
				},
				Set: schema.HashString,
			},
			"disable_rollback": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"notification_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"on_failure": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(cloudformation.OnFailure_Values(), false),
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"outputs": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"policy_body": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"policy_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"timeout_in_minutes": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"tags": tagsSchema(),
			"iam_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsCloudFormationStackCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn

	requestToken := resource.UniqueId()
	input := cloudformation.CreateStackInput{
		StackName:          aws.String(d.Get("name").(string)),
		ClientRequestToken: aws.String(requestToken),
	}
	if v, ok := d.GetOk("template_body"); ok {
		template, err := normalizeJsonOrYamlString(v)
		if err != nil {
			return fmt.Errorf("template body contains an invalid JSON or YAML: %s", err)
		}
		input.TemplateBody = aws.String(template)
	}
	if v, ok := d.GetOk("template_url"); ok {
		input.TemplateURL = aws.String(v.(string))
	}
	if v, ok := d.GetOk("capabilities"); ok {
		input.Capabilities = expandStringSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("disable_rollback"); ok {
		input.DisableRollback = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("notification_arns"); ok {
		input.NotificationARNs = expandStringSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("on_failure"); ok {
		input.OnFailure = aws.String(v.(string))
	}
	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = expandCloudFormationParameters(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("policy_body"); ok {
		policy, err := structure.NormalizeJsonString(v)
		if err != nil {
			return fmt.Errorf("policy body contains an invalid JSON: %s", err)
		}
		input.StackPolicyBody = aws.String(policy)
	}
	if v, ok := d.GetOk("policy_url"); ok {
		input.StackPolicyURL = aws.String(v.(string))
	}
	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().CloudformationTags()
	}
	if v, ok := d.GetOk("timeout_in_minutes"); ok {
		m := int64(v.(int))
		input.TimeoutInMinutes = aws.Int64(m)
	}
	if v, ok := d.GetOk("iam_role_arn"); ok {
		input.RoleARN = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating CloudFormation Stack: %s", input)
	resp, err := conn.CreateStack(&input)
	if err != nil {
		return fmt.Errorf("creating CloudFormation stack failed: %w", err)
	}

	d.SetId(aws.StringValue(resp.StackId))

	stack, err := waiter.StackCreated(conn, d.Id(), requestToken, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		if stack != nil {
			status := aws.StringValue(stack.StackStatus)
			if status == cloudformation.StackStatusDeleteComplete || status == cloudformation.StackStatusDeleteFailed {
				// Need to validate if this is actually necessary
				d.SetId("")
			}
		}
		return fmt.Errorf("error waiting for CloudFormation Stack creation: %w", err)
	}

	log.Printf("[INFO] CloudFormation Stack %q created", d.Id())

	return resourceAwsCloudFormationStackRead(d, meta)
}

func resourceAwsCloudFormationStackRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(d.Id()),
	}
	resp, err := conn.DescribeStacks(input)
	if tfawserr.ErrCodeEquals(err, "ValidationError") {
		log.Printf("[WARN] CloudFormation stack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return err
	}

	stacks := resp.Stacks
	if len(stacks) < 1 {
		log.Printf("[WARN] CloudFormation stack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	stack := stacks[0]
	if aws.StringValue(stack.StackStatus) == cloudformation.StackStatusDeleteComplete {
		log.Printf("[WARN] CloudFormation stack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	tInput := cloudformation.GetTemplateInput{
		StackName:     aws.String(d.Id()),
		TemplateStage: aws.String("Original"),
	}
	out, err := conn.GetTemplate(&tInput)
	if err != nil {
		return err
	}

	template, err := normalizeJsonOrYamlString(*out.TemplateBody)
	if err != nil {
		return fmt.Errorf("template body contains an invalid JSON or YAML: %s", err)
	}
	d.Set("template_body", template)

	log.Printf("[DEBUG] Received CloudFormation stack: %s", stack)

	d.Set("name", stack.StackName)
	d.Set("iam_role_arn", stack.RoleARN)

	if stack.TimeoutInMinutes != nil {
		d.Set("timeout_in_minutes", int(*stack.TimeoutInMinutes))
	}
	if stack.Description != nil {
		d.Set("description", stack.Description)
	}
	if stack.DisableRollback != nil {
		d.Set("disable_rollback", stack.DisableRollback)
	}
	if len(stack.NotificationARNs) > 0 {
		err = d.Set("notification_arns", flattenStringSet(stack.NotificationARNs))
		if err != nil {
			return err
		}
	}

	originalParams := d.Get("parameters").(map[string]interface{})
	err = d.Set("parameters", flattenCloudFormationParameters(stack.Parameters, originalParams))
	if err != nil {
		return err
	}

	if err := d.Set("tags", keyvaluetags.CloudformationKeyValueTags(stack.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	err = d.Set("outputs", flattenCloudFormationOutputs(stack.Outputs))
	if err != nil {
		return err
	}

	if len(stack.Capabilities) > 0 {
		err = d.Set("capabilities", flattenStringSet(stack.Capabilities))
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceAwsCloudFormationStackUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn

	requestToken := resource.UniqueId()
	input := &cloudformation.UpdateStackInput{
		StackName:          aws.String(d.Id()),
		ClientRequestToken: aws.String(requestToken),
	}

	// Either TemplateBody, TemplateURL or UsePreviousTemplate are required
	if v, ok := d.GetOk("template_url"); ok {
		input.TemplateURL = aws.String(v.(string))
	}
	if v, ok := d.GetOk("template_body"); ok && input.TemplateURL == nil {
		template, err := normalizeJsonOrYamlString(v)
		if err != nil {
			return fmt.Errorf("template body contains an invalid JSON or YAML: %s", err)
		}
		input.TemplateBody = aws.String(template)
	}

	// Capabilities must be present whether they are changed or not
	if v, ok := d.GetOk("capabilities"); ok {
		input.Capabilities = expandStringSet(v.(*schema.Set))
	}

	if d.HasChange("notification_arns") {
		input.NotificationARNs = expandStringSet(d.Get("notification_arns").(*schema.Set))
	}

	// Parameters must be present whether they are changed or not
	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = expandCloudFormationParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().CloudformationTags()
	}

	if d.HasChange("policy_body") {
		policy, err := structure.NormalizeJsonString(d.Get("policy_body"))
		if err != nil {
			return fmt.Errorf("policy body contains an invalid JSON: %s", err)
		}
		input.StackPolicyBody = aws.String(policy)
	}
	if d.HasChange("policy_url") {
		input.StackPolicyURL = aws.String(d.Get("policy_url").(string))
	}

	if d.HasChange("iam_role_arn") {
		input.RoleARN = aws.String(d.Get("iam_role_arn").(string))
	}

	log.Printf("[DEBUG] Updating CloudFormation stack: %s", input)
	_, err := conn.UpdateStack(input)
	if tfawserr.ErrMessageContains(err, "ValidationError", "No updates are to be performed.") {
		log.Printf("[DEBUG] Current CloudFormation stack has no updates")
	} else if err != nil {
		return fmt.Errorf("error updating CloudFormation stack (%s): %w", d.Id(), err)
	}

	_, err = waiter.StackUpdated(conn, d.Id(), requestToken, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return fmt.Errorf("error waiting for CloudFormation Stack update: %w", err)
	}

	log.Printf("[INFO] CloudFormation stack (%s) updated", d.Id())

	return resourceAwsCloudFormationStackRead(d, meta)
}

func resourceAwsCloudFormationStackDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn

	requestToken := resource.UniqueId()
	input := &cloudformation.DeleteStackInput{
		StackName:          aws.String(d.Id()),
		ClientRequestToken: aws.String(requestToken),
	}
	log.Printf("[DEBUG] Deleting CloudFormation stack %s", input)
	_, err := conn.DeleteStack(input)
	if tfawserr.ErrCodeEquals(err, "ValidationError") {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = waiter.StackDeleted(conn, d.Id(), requestToken, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return fmt.Errorf("error waiting for CloudFormation Stack deletion: %w", err)
	}

	log.Printf("[INFO] CloudFormation stack (%s) deleted", d.Id())

	return nil
}
