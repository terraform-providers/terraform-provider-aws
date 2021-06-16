package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfbudgets "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/budgets"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/budgets/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/budgets/waiter"
)

func resourceAwsBudgetsBudgetAction() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBudgetsBudgetActionCreate,
		Read:   resourceAwsBudgetsBudgetActionRead,
		Update: resourceAwsBudgetsBudgetActionUpdate,
		Delete: resourceAwsBudgetsBudgetActionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"account_id": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"action_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"action_threshold": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action_threshold_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(budgets.ThresholdType_Values(), false),
						},
						"action_threshold_value": {
							Type:         schema.TypeFloat,
							Required:     true,
							ValidateFunc: validation.FloatBetween(0, 40000000000),
						},
					},
				},
			},
			"action_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(budgets.ActionType_Values(), false),
			},
			"approval_model": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(budgets.ApprovalModel_Values(), false),
			},
			"budget_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexp.MustCompile(`[^:\\]+`), "The ':' and '\\' characters aren't allowed."),
				),
			},
			"definition": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iam_action_definition": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"policy_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
									"groups": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 100,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"roles": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 100,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"users": {
										Type:     schema.TypeSet,
										Optional: true,
										MaxItems: 100,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"ssm_action_definition": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action_sub_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(budgets.ActionSubType_Values(), false),
									},
									"instance_ids": {
										Type:     schema.TypeSet,
										Required: true,
										MaxItems: 100,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"region": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"scp_action_definition": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"policy_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"target_ids": {
										Type:     schema.TypeSet,
										Required: true,
										MaxItems: 100,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"execution_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"notification_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(budgets.NotificationType_Values(), false),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subscriber": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 11,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 2147483647),
								validation.StringMatch(regexp.MustCompile(`(.*[\n\r\t\f\ ]?)*`), "Can't contain line breaks."),
							)},
						"subscription_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(budgets.SubscriptionType_Values(), false),
						},
					},
				},
			},
		},
	}
}

func resourceAwsBudgetsBudgetActionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).budgetconn

	var accountID string
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	} else {
		accountID = meta.(*AWSClient).accountid
	}

	input := &budgets.CreateBudgetActionInput{
		AccountId:        aws.String(accountID),
		BudgetName:       aws.String(d.Get("budget_name").(string)),
		ActionType:       aws.String(d.Get("action_type").(string)),
		ApprovalModel:    aws.String(d.Get("approval_model").(string)),
		ExecutionRoleArn: aws.String(d.Get("execution_role_arn").(string)),
		NotificationType: aws.String(d.Get("notification_type").(string)),
		ActionThreshold:  expandAwsBudgetsBudgetActionActionThreshold(d.Get("action_threshold").([]interface{})),
		Subscribers:      expandAwsBudgetsBudgetActionSubscriber(d.Get("subscriber").(*schema.Set)),
		Definition:       expandAwsBudgetsBudgetActionActionDefinition(d.Get("definition").([]interface{})),
	}

	var output *budgets.CreateBudgetActionOutput
	_, err := retryOnAwsCode(budgets.ErrCodeAccessDeniedException, func() (interface{}, error) {
		var err error
		output, err = conn.CreateBudgetAction(input)
		return output, err
	})
	if err != nil {
		return fmt.Errorf("create Budget Action failed: %v", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", aws.StringValue(output.AccountId), aws.StringValue(output.ActionId), aws.StringValue(output.BudgetName)))

	if _, err := waiter.ActionAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Budget Action (%s) creation: %w", d.Id(), err)
	}

	return resourceAwsBudgetsBudgetActionRead(d, meta)
}

func resourceAwsBudgetsBudgetActionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).budgetconn
	out, err := finder.ActionById(conn, d.Id())
	if isAWSErr(err, budgets.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Budget Action %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("describe Budget Action failed: %w", err)
	}

	action := out.Action
	if action == nil {
		log.Printf("[WARN] Budget Action %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	budgetName := aws.StringValue(out.BudgetName)
	actId := aws.StringValue(action.ActionId)
	d.Set("account_id", out.AccountId)
	d.Set("budget_name", budgetName)
	d.Set("action_id", actId)
	d.Set("action_type", action.ActionType)
	d.Set("approval_model", action.ApprovalModel)
	d.Set("execution_role_arn", action.ExecutionRoleArn)
	d.Set("notification_type", action.NotificationType)
	d.Set("status", action.Status)

	if err := d.Set("subscriber", flattenAwsBudgetsBudgetActionSubscriber(action.Subscribers)); err != nil {
		return fmt.Errorf("error setting subscriber: %w", err)
	}

	if err := d.Set("definition", flattenAwsBudgetsBudgetActionDefinition(action.Definition)); err != nil {
		return fmt.Errorf("error setting definition: %w", err)
	}

	if err := d.Set("action_threshold", flattenAwsBudgetsBudgetActionActionThreshold(action.ActionThreshold)); err != nil {
		return fmt.Errorf("error setting action_threshold: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "budgetservice",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("budget/%s/action/%s", budgetName, actId),
	}
	d.Set("arn", arn.String())

	return nil
}

func resourceAwsBudgetsBudgetActionUpdate(d *schema.ResourceData, meta interface{}) error {
	accountID, actionID, budgetName, err := tfbudgets.DecodeBudgetsBudgetActionID(d.Id())
	if err != nil {
		return err
	}

	conn := meta.(*AWSClient).budgetconn
	input := &budgets.UpdateBudgetActionInput{
		BudgetName: aws.String(budgetName),
		AccountId:  aws.String(accountID),
		ActionId:   aws.String(actionID),
	}

	if d.HasChange("approval_model") {
		input.ApprovalModel = aws.String(d.Get("approval_model").(string))
	}

	if d.HasChange("execution_role_arn") {
		input.ExecutionRoleArn = aws.String(d.Get("execution_role_arn").(string))
	}

	if d.HasChange("notification_type") {
		input.NotificationType = aws.String(d.Get("notification_type").(string))
	}

	if d.HasChange("action_threshold") {
		input.ActionThreshold = expandAwsBudgetsBudgetActionActionThreshold(d.Get("action_threshold").([]interface{}))
	}

	if d.HasChange("subscriber") {
		input.Subscribers = expandAwsBudgetsBudgetActionSubscriber(d.Get("subscriber").(*schema.Set))
	}

	if d.HasChange("definition") {
		input.Definition = expandAwsBudgetsBudgetActionActionDefinition(d.Get("definition").([]interface{}))
	}

	_, err = conn.UpdateBudgetAction(input)
	if err != nil {
		return fmt.Errorf("Updating Budget Action failed: %w", err)
	}

	if _, err := waiter.ActionAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Budget Action (%s) update: %w", d.Id(), err)
	}

	return resourceAwsBudgetsBudgetActionRead(d, meta)
}

func resourceAwsBudgetsBudgetActionDelete(d *schema.ResourceData, meta interface{}) error {
	accountID, actionID, budgetName, err := tfbudgets.DecodeBudgetsBudgetActionID(d.Id())
	if err != nil {
		return err
	}

	conn := meta.(*AWSClient).budgetconn
	_, err = conn.DeleteBudgetAction(&budgets.DeleteBudgetActionInput{
		BudgetName: aws.String(budgetName),
		AccountId:  aws.String(accountID),
		ActionId:   aws.String(actionID),
	})
	if err != nil {
		if isAWSErr(err, budgets.ErrCodeNotFoundException, "") {
			log.Printf("[INFO] Budget Action %s could not be found. skipping delete.", d.Id())
			return nil
		}

		return fmt.Errorf("Deleting Budget Action failed: %w", err)
	}

	return nil
}

func expandAwsBudgetsBudgetActionActionThreshold(l []interface{}) *budgets.ActionThreshold {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &budgets.ActionThreshold{}

	if v, ok := m["action_threshold_type"].(string); ok && v != "" {
		config.ActionThresholdType = aws.String(v)
	}

	if v, ok := m["action_threshold_value"].(float64); ok {
		config.ActionThresholdValue = aws.Float64(v)
	}

	return config
}

func expandAwsBudgetsBudgetActionSubscriber(l *schema.Set) []*budgets.Subscriber {
	if l.Len() == 0 {
		return []*budgets.Subscriber{}
	}

	items := []*budgets.Subscriber{}

	for _, m := range l.List() {
		config := &budgets.Subscriber{}
		raw := m.(map[string]interface{})

		if v, ok := raw["address"].(string); ok && v != "" {
			config.Address = aws.String(v)
		}

		if v, ok := raw["subscription_type"].(string); ok {
			config.SubscriptionType = aws.String(v)
		}

		items = append(items, config)
	}

	return items
}

func expandAwsBudgetsBudgetActionActionDefinition(l []interface{}) *budgets.Definition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &budgets.Definition{}

	if v, ok := m["ssm_action_definition"].([]interface{}); ok && len(v) > 0 {
		config.SsmActionDefinition = expandAwsBudgetsBudgetActionActionSsmActionDefinition(v)
	}

	if v, ok := m["scp_action_definition"].([]interface{}); ok && len(v) > 0 {
		config.ScpActionDefinition = expandAwsBudgetsBudgetActionActionScpActionDefinition(v)
	}

	if v, ok := m["iam_action_definition"].([]interface{}); ok && len(v) > 0 {
		config.IamActionDefinition = expandAwsBudgetsBudgetActionActionIamActionDefinition(v)
	}

	return config
}

func expandAwsBudgetsBudgetActionActionScpActionDefinition(l []interface{}) *budgets.ScpActionDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &budgets.ScpActionDefinition{}

	if v, ok := m["policy_id"].(string); ok && v != "" {
		config.PolicyId = aws.String(v)
	}

	if v, ok := m["target_ids"].(*schema.Set); ok && v.Len() > 0 {
		config.TargetIds = expandStringSet(v)
	}

	return config
}

func expandAwsBudgetsBudgetActionActionSsmActionDefinition(l []interface{}) *budgets.SsmActionDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &budgets.SsmActionDefinition{}

	if v, ok := m["action_sub_type"].(string); ok && v != "" {
		config.ActionSubType = aws.String(v)
	}

	if v, ok := m["region"].(string); ok && v != "" {
		config.Region = aws.String(v)
	}

	if v, ok := m["instance_ids"].(*schema.Set); ok && v.Len() > 0 {
		config.InstanceIds = expandStringSet(v)
	}

	return config
}

func expandAwsBudgetsBudgetActionActionIamActionDefinition(l []interface{}) *budgets.IamActionDefinition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &budgets.IamActionDefinition{}

	if v, ok := m["policy_arn"].(string); ok && v != "" {
		config.PolicyArn = aws.String(v)
	}

	if v, ok := m["groups"].(*schema.Set); ok && v.Len() > 0 {
		config.Groups = expandStringSet(v)
	}

	if v, ok := m["roles"].(*schema.Set); ok && v.Len() > 0 {
		config.Roles = expandStringSet(v)
	}

	if v, ok := m["users"].(*schema.Set); ok && v.Len() > 0 {
		config.Users = expandStringSet(v)
	}

	return config
}

func flattenAwsBudgetsBudgetActionSubscriber(configured []*budgets.Subscriber) []map[string]interface{} {
	dataResources := make([]map[string]interface{}, 0, len(configured))

	for _, raw := range configured {
		item := make(map[string]interface{})
		item["address"] = aws.StringValue(raw.Address)
		item["subscription_type"] = aws.StringValue(raw.SubscriptionType)

		dataResources = append(dataResources, item)
	}

	return dataResources
}

func flattenAwsBudgetsBudgetActionActionThreshold(lt *budgets.ActionThreshold) []map[string]interface{} {
	if lt == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"action_threshold_type":  aws.StringValue(lt.ActionThresholdType),
		"action_threshold_value": aws.Float64Value(lt.ActionThresholdValue),
	}

	return []map[string]interface{}{attrs}
}

func flattenAwsBudgetsBudgetActionIamActionDefinition(lt *budgets.IamActionDefinition) []map[string]interface{} {
	if lt == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"policy_arn": aws.StringValue(lt.PolicyArn),
	}

	if lt.Users != nil && len(lt.Users) > 0 {
		attrs["users"] = flattenStringSet(lt.Users)
	}

	if lt.Roles != nil && len(lt.Roles) > 0 {
		attrs["roles"] = flattenStringSet(lt.Roles)
	}

	if lt.Groups != nil && len(lt.Groups) > 0 {
		attrs["groups"] = flattenStringSet(lt.Groups)
	}

	return []map[string]interface{}{attrs}
}

func flattenAwsBudgetsBudgetActionScpActionDefinition(lt *budgets.ScpActionDefinition) []map[string]interface{} {
	if lt == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"policy_id": aws.StringValue(lt.PolicyId),
	}

	if lt.TargetIds != nil && len(lt.TargetIds) > 0 {
		attrs["target_ids"] = flattenStringSet(lt.TargetIds)
	}

	return []map[string]interface{}{attrs}
}

func flattenAwsBudgetsBudgetActionSsmActionDefinition(lt *budgets.SsmActionDefinition) []map[string]interface{} {
	if lt == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{
		"action_sub_type": aws.StringValue(lt.ActionSubType),
		"instance_ids":    flattenStringSet(lt.InstanceIds),
		"region":          aws.StringValue(lt.Region),
	}

	return []map[string]interface{}{attrs}
}

func flattenAwsBudgetsBudgetActionDefinition(lt *budgets.Definition) []map[string]interface{} {
	if lt == nil {
		return []map[string]interface{}{}
	}

	attrs := map[string]interface{}{}

	if lt.SsmActionDefinition != nil {
		attrs["ssm_action_definition"] = flattenAwsBudgetsBudgetActionSsmActionDefinition(lt.SsmActionDefinition)
	}

	if lt.IamActionDefinition != nil {
		attrs["iam_action_definition"] = flattenAwsBudgetsBudgetActionIamActionDefinition(lt.IamActionDefinition)
	}

	if lt.ScpActionDefinition != nil {
		attrs["scp_action_definition"] = flattenAwsBudgetsBudgetActionScpActionDefinition(lt.ScpActionDefinition)
	}

	return []map[string]interface{}{attrs}
}
