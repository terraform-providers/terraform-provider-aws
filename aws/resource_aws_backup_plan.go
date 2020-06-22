package aws

import (
	"bytes"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsBackupPlan() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBackupPlanCreate,
		Read:   resourceAwsBackupPlanRead,
		Update: resourceAwsBackupPlanUpdate,
		Delete: resourceAwsBackupPlanDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rule": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target_vault_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"schedule": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"start_window": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  60,
						},
						"completion_window": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  180,
						},
						"lifecycle": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cold_storage_after": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"delete_after": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"copy_action": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"lifecycle": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cold_storage_after": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"delete_after": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
									"destination_vault_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},
						"recovery_point_tags": tagsSchema(),
					},
				},
				Set: backupBackupPlanHash,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsBackupPlanCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	input := &backup.CreateBackupPlanInput{
		BackupPlan: &backup.PlanInput{
			BackupPlanName: aws.String(d.Get("name").(string)),
			Rules:          expandBackupPlanRules(d.Get("rule").(*schema.Set)),
		},
		BackupPlanTags: keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().BackupTags(),
	}

	log.Printf("[DEBUG] Creating Backup Plan: %#v", input)
	resp, err := conn.CreateBackupPlan(input)
	if err != nil {
		return fmt.Errorf("error creating Backup Plan: %s", err)
	}

	d.SetId(aws.StringValue(resp.BackupPlanId))

	return resourceAwsBackupPlanRead(d, meta)
}

func resourceAwsBackupPlanRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.GetBackupPlan(&backup.GetBackupPlanInput{
		BackupPlanId: aws.String(d.Id()),
	})
	if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Backup Plan (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading Backup Plan (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.BackupPlanArn)
	d.Set("name", resp.BackupPlan.BackupPlanName)
	d.Set("version", resp.VersionId)

	if err := d.Set("rule", flattenBackupPlanRules(resp.BackupPlan.Rules)); err != nil {
		return fmt.Errorf("error setting rule: %s", err)
	}

	tags, err := keyvaluetags.BackupListTags(conn, d.Get("arn").(string))
	if err != nil {
		return fmt.Errorf("error listing tags for Backup Plan (%s): %s", d.Id(), err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsBackupPlanUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	if d.HasChange("rule") {
		input := &backup.UpdateBackupPlanInput{
			BackupPlanId: aws.String(d.Id()),
			BackupPlan: &backup.PlanInput{
				BackupPlanName: aws.String(d.Get("name").(string)),
				Rules:          expandBackupPlanRules(d.Get("rule").(*schema.Set)),
			},
		}

		log.Printf("[DEBUG] Updating Backup Plan: %#v", input)
		_, err := conn.UpdateBackupPlan(input)
		if err != nil {
			return fmt.Errorf("error updating Backup Plan (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.BackupUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags for Backup Plan (%s): %s", d.Id(), err)
		}
	}

	return resourceAwsBackupPlanRead(d, meta)
}

func resourceAwsBackupPlanDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).backupconn

	log.Printf("[DEBUG] Deleting Backup Plan: %s", d.Id())
	_, err := conn.DeleteBackupPlan(&backup.DeleteBackupPlanInput{
		BackupPlanId: aws.String(d.Id()),
	})
	if isAWSErr(err, backup.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Backup Plan (%s): %s", d.Id(), err)
	}

	return nil
}

func expandBackupPlanRules(vRules *schema.Set) []*backup.RuleInput {
	rules := []*backup.RuleInput{}

	for _, vRule := range vRules.List() {
		rule := &backup.RuleInput{}

		mRule := vRule.(map[string]interface{})

		if vRuleName, ok := mRule["rule_name"].(string); ok && vRuleName != "" {
			rule.RuleName = aws.String(vRuleName)
		} else {
			continue
		}
		if vTargetVaultName, ok := mRule["target_vault_name"].(string); ok && vTargetVaultName != "" {
			rule.TargetBackupVaultName = aws.String(vTargetVaultName)
		}
		if vSchedule, ok := mRule["schedule"].(string); ok && vSchedule != "" {
			rule.ScheduleExpression = aws.String(vSchedule)
		}
		if vStartWindow, ok := mRule["start_window"].(int); ok {
			rule.StartWindowMinutes = aws.Int64(int64(vStartWindow))
		}
		if vCompletionWindow, ok := mRule["completion_window"].(int); ok {
			rule.CompletionWindowMinutes = aws.Int64(int64(vCompletionWindow))
		}

		if vRecoveryPointTags, ok := mRule["recovery_point_tags"].(map[string]interface{}); ok && len(vRecoveryPointTags) > 0 {
			rule.RecoveryPointTags = keyvaluetags.New(vRecoveryPointTags).IgnoreAws().BackupTags()
		}

		if vLifecycle, ok := mRule["lifecycle"].([]interface{}); ok && len(vLifecycle) > 0 && vLifecycle[0] != nil {
			lifecycle := &backup.Lifecycle{}

			mLifecycle := vLifecycle[0].(map[string]interface{})

			if vDeleteAfter, ok := mLifecycle["delete_after"].(int); ok && vDeleteAfter > 0 {
				lifecycle.DeleteAfterDays = aws.Int64(int64(vDeleteAfter))
			}
			if vColdStorageAfter, ok := mLifecycle["cold_storage_after"].(int); ok && vColdStorageAfter > 0 {
				lifecycle.MoveToColdStorageAfterDays = aws.Int64(int64(vColdStorageAfter))
			}

			rule.Lifecycle = lifecycle
		}

		if vCopyActions := expandBackupPlanCopyActions(mRule["copy_action"].(*schema.Set).List()); len(vCopyActions) > 0 {
			rule.CopyActions = vCopyActions
		}

		rules = append(rules, rule)
	}

	return rules
}

func expandBackupPlanCopyActions(actionList []interface{}) []*backup.CopyAction {
	actions := []*backup.CopyAction{}

	for _, i := range actionList {
		item := i.(map[string]interface{})
		action := &backup.CopyAction{}

		action.DestinationBackupVaultArn = aws.String(item["destination_vault_arn"].(string))
		action.Lifecycle = expandBackupPlanLifecycle(item["lifecycle"].([]interface{}))

		actions = append(actions, action)
	}

	return actions
}

func expandBackupPlanLifecycle(l []interface{}) *backup.Lifecycle {
	lifecycle := new(backup.Lifecycle)

	for _, i := range l {
		lc := i.(map[string]interface{})
		if vDeleteAfter, ok := lc["delete_after"]; ok && vDeleteAfter.(int) > 0 {
			lifecycle.DeleteAfterDays = aws.Int64(int64(vDeleteAfter.(int)))
		}
		if vMoveToColdStorageAfterDays, ok := lc["cold_storage_after"]; ok && vMoveToColdStorageAfterDays.(int) > 0 {
			lifecycle.MoveToColdStorageAfterDays = aws.Int64(int64(vMoveToColdStorageAfterDays.(int)))
		}
	}

	return lifecycle
}

func flattenBackupPlanRules(rules []*backup.Rule) *schema.Set {
	vRules := []interface{}{}

	for _, rule := range rules {
		mRule := map[string]interface{}{
			"rule_name":           aws.StringValue(rule.RuleName),
			"target_vault_name":   aws.StringValue(rule.TargetBackupVaultName),
			"schedule":            aws.StringValue(rule.ScheduleExpression),
			"start_window":        int(aws.Int64Value(rule.StartWindowMinutes)),
			"completion_window":   int(aws.Int64Value(rule.CompletionWindowMinutes)),
			"recovery_point_tags": keyvaluetags.BackupKeyValueTags(rule.RecoveryPointTags).IgnoreAws().Map(),
		}

		if lifecycle := rule.Lifecycle; lifecycle != nil {
			mRule["lifecycle"] = []interface{}{
				map[string]interface{}{
					"delete_after":       int(aws.Int64Value(lifecycle.DeleteAfterDays)),
					"cold_storage_after": int(aws.Int64Value(lifecycle.MoveToColdStorageAfterDays)),
				},
			}
		}

		mRule["copy_action"] = flattenBackupPlanCopyActions(rule.CopyActions)

		vRules = append(vRules, mRule)
	}

	return schema.NewSet(backupBackupPlanHash, vRules)
}

func flattenBackupPlanCopyActions(copyActions []*backup.CopyAction) []interface{} {
	if len(copyActions) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, copyAction := range copyActions {
		if copyAction == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"destination_vault_arn": aws.StringValue(copyAction.DestinationBackupVaultArn),
			"lifecycle":             flattenBackupPlanCopyActionLifecycle(copyAction.Lifecycle),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenBackupPlanCopyActionLifecycle(copyActionLifecycle *backup.Lifecycle) []interface{} {
	if copyActionLifecycle == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"delete_after":       aws.Int64Value(copyActionLifecycle.DeleteAfterDays),
		"cold_storage_after": aws.Int64Value(copyActionLifecycle.MoveToColdStorageAfterDays),
	}

	return []interface{}{m}
}

func backupBackupPlanHash(vRule interface{}) int {
	var buf bytes.Buffer

	mRule := vRule.(map[string]interface{})

	if v, ok := mRule["rule_name"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := mRule["target_vault_name"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := mRule["schedule"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := mRule["start_window"].(int); ok {
		buf.WriteString(fmt.Sprintf("%d-", v))
	}
	if v, ok := mRule["completion_window"].(int); ok {
		buf.WriteString(fmt.Sprintf("%d-", v))
	}

	if vRecoveryPointTags, ok := mRule["recovery_point_tags"].(map[string]interface{}); ok && len(vRecoveryPointTags) > 0 {
		buf.WriteString(fmt.Sprintf("%d-", keyvaluetags.New(vRecoveryPointTags).Hash()))
	}

	if vLifecycle, ok := mRule["lifecycle"].([]interface{}); ok && len(vLifecycle) > 0 && vLifecycle[0] != nil {
		mLifecycle := vLifecycle[0].(map[string]interface{})

		if v, ok := mLifecycle["delete_after"].(int); ok {
			buf.WriteString(fmt.Sprintf("%d-", v))
		}
		if v, ok := mLifecycle["cold_storage_after"].(int); ok {
			buf.WriteString(fmt.Sprintf("%d-", v))
		}
	}

	if vCopyActions, ok := mRule["copy_action"].(*schema.Set); ok && vCopyActions.Len() > 0 {
		for _, a := range vCopyActions.List() {
			action := a.(map[string]interface{})
			if mLifecycle, ok := action["lifecycle"].([]interface{}); ok {
				for _, l := range mLifecycle {
					lifecycle := l.(map[string]interface{})
					if v, ok := lifecycle["delete_after"].(int); ok {
						buf.WriteString(fmt.Sprintf("%d-", v))
					}
					if v, ok := lifecycle["cold_storage_after"].(int); ok {
						buf.WriteString(fmt.Sprintf("%d-", v))
					}
				}
			}

			if v, ok := action["destination_vault_arn"].(string); ok {
				buf.WriteString(fmt.Sprintf("%s-", v))
			}
		}
	}

	return hashcode.String(buf.String())
}
