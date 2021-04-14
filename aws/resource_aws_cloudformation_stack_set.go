package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudformation/waiter"
)

func resourceAwsCloudFormationStackSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudFormationStackSetCreate,
		Read:   resourceAwsCloudFormationStackSetRead,
		Update: resourceAwsCloudFormationStackSetUpdate,
		Delete: resourceAwsCloudFormationStackSetDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(waiter.StackSetUpdatedDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"administration_role_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"auto_deployment"},
				ValidateFunc:  validateArn,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_deployment": {
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				ConflictsWith: []string{
					"administration_role_arn",
					"execution_role_name",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"retain_stacks_on_account_removal": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"capabilities": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(cloudformation.Capability_Values(), false),
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"execution_role_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"auto_deployment"},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]+$`), "must contain only alphanumeric and hyphen characters"),
				),
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"permission_model": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(cloudformation.PermissionModels_Values(), false),
				Default:      cloudformation.PermissionModelsSelfManaged,
			},
			"stack_set_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
			"template_body": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ConflictsWith:    []string{"template_url"},
				DiffSuppressFunc: suppressEquivalentJsonOrYamlDiffs,
				ValidateFunc:     validateStringIsJsonOrYaml,
			},
			"template_url": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"template_body"},
			},
		},
	}
}

func resourceAwsCloudFormationStackSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn
	name := d.Get("name").(string)

	input := &cloudformation.CreateStackSetInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		StackSetName:       aws.String(name),
	}

	if v, ok := d.GetOk("administration_role_arn"); ok {
		input.AdministrationRoleARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("auto_deployment"); ok {
		input.AutoDeployment = expandAutoDeployment(v.([]interface{}))
	}

	if v, ok := d.GetOk("capabilities"); ok {
		input.Capabilities = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("execution_role_name"); ok {
		input.ExecutionRoleName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = expandCloudFormationParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("permission_model"); ok {
		input.PermissionModel = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().CloudformationTags()
	}

	if v, ok := d.GetOk("template_body"); ok {
		input.TemplateBody = aws.String(v.(string))
	}

	if v, ok := d.GetOk("template_url"); ok {
		input.TemplateURL = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating CloudFormation StackSet: %s", input)
	_, err := conn.CreateStackSet(input)

	if err != nil {
		return fmt.Errorf("error creating CloudFormation StackSet: %s", err)
	}

	d.SetId(name)

	return resourceAwsCloudFormationStackSetRead(d, meta)
}

func resourceAwsCloudFormationStackSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &cloudformation.DescribeStackSetInput{
		StackSetName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading CloudFormation StackSet: %s", d.Id())
	output, err := conn.DescribeStackSet(input)

	if isAWSErr(err, cloudformation.ErrCodeStackSetNotFoundException, "") {
		log.Printf("[WARN] CloudFormation StackSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudFormation StackSet (%s): %s", d.Id(), err)
	}

	if output == nil || output.StackSet == nil {
		return fmt.Errorf("error reading CloudFormation StackSet (%s): empty response", d.Id())
	}

	stackSet := output.StackSet

	d.Set("administration_role_arn", stackSet.AdministrationRoleARN)
	d.Set("arn", stackSet.StackSetARN)

	if err := d.Set("auto_deployment", flattenStackSetAutoDeploymentResponse(stackSet.AutoDeployment)); err != nil {
		return fmt.Errorf("error setting auto_deployment: %s", err)
	}

	if err := d.Set("capabilities", aws.StringValueSlice(stackSet.Capabilities)); err != nil {
		return fmt.Errorf("error setting capabilities: %s", err)
	}

	d.Set("description", stackSet.Description)
	d.Set("execution_role_name", stackSet.ExecutionRoleName)
	d.Set("name", stackSet.StackSetName)
	d.Set("permission_model", stackSet.PermissionModel)

	if err := d.Set("parameters", flattenAllCloudFormationParameters(stackSet.Parameters)); err != nil {
		return fmt.Errorf("error setting parameters: %s", err)
	}

	d.Set("stack_set_id", stackSet.StackSetId)

	if err := d.Set("tags", keyvaluetags.CloudformationKeyValueTags(stackSet.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("template_body", stackSet.TemplateBody)

	return nil
}

func resourceAwsCloudFormationStackSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn

	input := &cloudformation.UpdateStackSetInput{
		OperationId:  aws.String(resource.UniqueId()),
		StackSetName: aws.String(d.Id()),
		Tags:         []*cloudformation.Tag{},
		TemplateBody: aws.String(d.Get("template_body").(string)),
	}

	if v, ok := d.GetOk("administration_role_arn"); ok {
		input.AdministrationRoleARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("capabilities"); ok {
		input.Capabilities = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("execution_role_name"); ok {
		input.ExecutionRoleName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = expandCloudFormationParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("permission_model"); ok {
		input.PermissionModel = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().CloudformationTags()
	}

	if v, ok := d.GetOk("template_url"); ok {
		// ValidationError: Exactly one of TemplateBody or TemplateUrl must be specified
		// TemplateBody is always present when TemplateUrl is used so remove TemplateBody if TemplateUrl is set
		input.TemplateBody = nil
		input.TemplateURL = aws.String(v.(string))
	}

	// When `auto_deployment` is set, ignore `administration_role_arn` and
	// `execution_role_name` fields since it's using the SERVICE_MANAGED
	// permission model
	if v, ok := d.GetOk("auto_deployment"); ok {
		input.AdministrationRoleARN = nil
		input.ExecutionRoleName = nil
		input.AutoDeployment = expandAutoDeployment(v.([]interface{}))
	}

	log.Printf("[DEBUG] Updating CloudFormation StackSet: %s", input)
	output, err := conn.UpdateStackSet(input)

	if err != nil {
		return fmt.Errorf("error updating CloudFormation StackSet (%s): %s", d.Id(), err)
	}

	if err := waiter.StackSetOperationSucceeded(conn, d.Id(), aws.StringValue(output.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return fmt.Errorf("error waiting for CloudFormation StackSet (%s) update: %s", d.Id(), err)
	}

	return resourceAwsCloudFormationStackSetRead(d, meta)
}

func resourceAwsCloudFormationStackSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cfconn

	input := &cloudformation.DeleteStackSetInput{
		StackSetName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting CloudFormation StackSet: %s", d.Id())
	_, err := conn.DeleteStackSet(input)

	if isAWSErr(err, cloudformation.ErrCodeStackSetNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudFormation StackSet (%s): %s", d.Id(), err)
	}

	return nil
}

func listCloudFormationStackSets(conn *cloudformation.CloudFormation) ([]*cloudformation.StackSetSummary, error) {
	input := &cloudformation.ListStackSetsInput{
		Status: aws.String(cloudformation.StackSetStatusActive),
	}
	result := make([]*cloudformation.StackSetSummary, 0)

	for {
		output, err := conn.ListStackSets(input)

		if err != nil {
			return result, err
		}

		result = append(result, output.Summaries...)

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return result, nil
}

func expandAutoDeployment(l []interface{}) *cloudformation.AutoDeployment {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	autoDeployment := &cloudformation.AutoDeployment{
		Enabled:                      aws.Bool(m["enabled"].(bool)),
		RetainStacksOnAccountRemoval: aws.Bool(m["retain_stacks_on_account_removal"].(bool)),
	}

	return autoDeployment
}

func flattenStackSetAutoDeploymentResponse(autoDeployment *cloudformation.AutoDeployment) []map[string]interface{} {
	if autoDeployment == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enabled":                          aws.BoolValue(autoDeployment.Enabled),
		"retain_stacks_on_account_removal": aws.BoolValue(autoDeployment.RetainStacksOnAccountRemoval),
	}

	return []map[string]interface{}{m}
}
