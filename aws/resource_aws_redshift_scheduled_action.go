package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"time"
)

func resourceAwsRedshiftScheduledAction() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRedshiftScheduledActionCreate,
		Read:   resourceAwsRedshiftScheduledActionRead,
		Update: resourceAwsRedshiftScheduledActionUpdate,
		Delete: resourceAwsRedshiftScheduledActionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"start_time": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"end_time": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"schedule": {
				Type:     schema.TypeString,
				Required: true,
			},
			"iam_role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"target_action": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								redshift.ScheduledActionTypeValuesResumeCluster,
								redshift.ScheduledActionTypeValuesPauseCluster,
								redshift.ScheduledActionTypeValuesResizeCluster,
							}, false),
						},
						"cluster_identifier": {
							Type:     schema.TypeString,
							Required: true,
						},
						"classic": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"cluster_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"node_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"number_of_nodes": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsRedshiftScheduledActionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn
	name := d.Get("name").(string)
	createOpts := &redshift.CreateScheduledActionInput{
		ScheduledActionName: aws.String(name),
		Schedule:            aws.String(d.Get("schedule").(string)),
		IamRole:             aws.String(d.Get("iam_role").(string)),
		TargetAction:        expandRedshiftScheduledActionTargetAction(d.Get("target_action")),
	}
	if attr, ok := d.GetOk("description"); ok {
		createOpts.ScheduledActionDescription = aws.String(attr.(string))
	}
	if attr, ok := d.GetOk("active"); ok {
		createOpts.Enable = aws.Bool(attr.(bool))
	}
	if attr, ok := d.GetOk("start_time"); ok {
		t, _ := time.Parse(time.RFC3339, attr.(string))
		createOpts.StartTime = aws.Time(t)
	}
	if attr, ok := d.GetOk("end_time"); ok {
		t, _ := time.Parse(time.RFC3339, attr.(string))
		createOpts.EndTime = aws.Time(t)
	}

	log.Printf("[DEBUG] Creating Redshift Scheduled Action: %s", createOpts)

	// Retry for IAM eventual consistency
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateScheduledAction(createOpts)

		// InvalidParameterValue: If you create iam role same time, you must wait the role will be valid
		if isAWSErr(err, "InvalidParameterValue", "The IAM role must delegate access to Amazon Redshift scheduler") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error creating Redshift Scheduled Action (%s): %s", name, err)
	}

	d.SetId(name)

	return resourceAwsRedshiftScheduledActionRead(d, meta)
}

func resourceAwsRedshiftScheduledActionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn
	name := d.Id()

	descOpts := &redshift.DescribeScheduledActionsInput{
		ScheduledActionName: aws.String(name),
	}

	resp, err := conn.DescribeScheduledActions(descOpts)
	if err != nil {
		return fmt.Errorf("Error describing Redshift Scheduled Action %s: %s", d.Id(), err)
	}

	if resp.ScheduledActions == nil || len(resp.ScheduledActions) != 1 {
		log.Printf("[WARN] Unable to find Redshift Scheduled Action (%s)", d.Id())
		d.SetId("")
		return nil
	}

	scheduledAction := resp.ScheduledActions[0]

	d.Set("name", scheduledAction.ScheduledActionName)
	d.Set("description", scheduledAction.ScheduledActionDescription)
	d.Set("schedule", scheduledAction.Schedule)
	d.Set("iam_role", scheduledAction.IamRole)

	if aws.StringValue(scheduledAction.State) == redshift.ScheduledActionStateActive {
		d.Set("active", true)
	} else {
		d.Set("active", false)
	}

	if scheduledAction.StartTime != nil {
		d.Set("start_time", aws.TimeValue(scheduledAction.StartTime).Format(time.RFC3339))
	}

	if scheduledAction.EndTime != nil {
		d.Set("end_time", aws.TimeValue(scheduledAction.EndTime).Format(time.RFC3339))
	}

	if err := d.Set("target_action", flattenRedshiftScheduledActionType(scheduledAction.TargetAction)); err != nil {
		return fmt.Errorf("Error setting definitions: %s", err)
	}

	return nil
}

func resourceAwsRedshiftScheduledActionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn

	modifyOpts := &redshift.ModifyScheduledActionInput{
		ScheduledActionName:        aws.String(d.Get("name").(string)),
		Schedule:                   aws.String(d.Get("schedule").(string)),
		IamRole:                    aws.String(d.Get("iam_role").(string)),
		TargetAction:               expandRedshiftScheduledActionTargetAction(d.Get("target_action")),
		Enable:                     aws.Bool(d.Get("active").(bool)),
		ScheduledActionDescription: aws.String(d.Get("description").(string)),
	}

	if attr, ok := d.GetOk("start_time"); ok {
		t, _ := time.Parse(time.RFC3339, attr.(string))
		modifyOpts.StartTime = aws.Time(t)
	}
	if attr, ok := d.GetOk("end_time"); ok {
		t, _ := time.Parse(time.RFC3339, attr.(string))
		modifyOpts.EndTime = aws.Time(t)
	}

	log.Printf("[DEBUG] Updating Redshift Scheduled Action: %s", modifyOpts)

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.ModifyScheduledAction(modifyOpts)

		// InvalidParameterValue: If you create iam role same time, you must wait the role will be valid
		if isAWSErr(err, "InvalidParameterValue", "The IAM role must delegate access to Amazon Redshift scheduler") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error updating Redshift Scheduled Action (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsRedshiftScheduledActionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn

	deleteOpts := &redshift.DeleteScheduledActionInput{
		ScheduledActionName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Redshift Scheduled Action: %s", deleteOpts)
	_, err := conn.DeleteScheduledAction(deleteOpts)
	if err != nil {
		return fmt.Errorf("error deleting Redshift Scheduled Action (%s): %s", d.Id(), err)
	}

	return nil
}

func expandRedshiftScheduledActionTargetAction(configured interface{}) *redshift.ScheduledActionType {
	if configured == nil || len(configured.([]interface{})) == 0 {
		return nil
	}

	p := configured.([]interface{})[0].(map[string]interface{})

	switch p["action"].(string) {
	case redshift.ScheduledActionTypeValuesPauseCluster:
		pauseCluster := redshift.PauseClusterMessage{ClusterIdentifier: aws.String(p["cluster_identifier"].(string))}
		return &redshift.ScheduledActionType{
			PauseCluster: &pauseCluster,
		}
	case redshift.ScheduledActionTypeValuesResumeCluster:
		resumeCluster := redshift.ResumeClusterMessage{ClusterIdentifier: aws.String(p["cluster_identifier"].(string))}
		return &redshift.ScheduledActionType{
			ResumeCluster: &resumeCluster,
		}
	case redshift.ScheduledActionTypeValuesResizeCluster:
		resizeCluster := redshift.ResizeClusterMessage{
			ClusterIdentifier: aws.String(p["cluster_identifier"].(string)),
			Classic:           aws.Bool(p["classic"].(bool)),
			ClusterType:       aws.String(p["cluster_type"].(string)),
			NodeType:          aws.String(p["node_type"].(string)),
			NumberOfNodes:     aws.Int64(int64(p["number_of_nodes"].(int))),
		}
		return &redshift.ScheduledActionType{
			ResizeCluster: &resizeCluster,
		}
	}
	return nil
}

func flattenRedshiftScheduledActionType(scheduledActionType *redshift.ScheduledActionType) []map[string]interface{} {
	if scheduledActionType == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if scheduledActionType.ResumeCluster != nil {
		m = map[string]interface{}{
			"action":             redshift.ScheduledActionTypeValuesResumeCluster,
			"cluster_identifier": aws.StringValue(scheduledActionType.ResumeCluster.ClusterIdentifier),
		}
	}
	if scheduledActionType.PauseCluster != nil {
		m = map[string]interface{}{
			"action":             redshift.ScheduledActionTypeValuesPauseCluster,
			"cluster_identifier": aws.StringValue(scheduledActionType.PauseCluster.ClusterIdentifier),
		}
	}
	if scheduledActionType.ResizeCluster != nil {
		m = map[string]interface{}{
			"action":             redshift.ScheduledActionTypeValuesResizeCluster,
			"cluster_identifier": aws.StringValue(scheduledActionType.ResizeCluster.ClusterIdentifier),
			"classic":            aws.BoolValue(scheduledActionType.ResizeCluster.Classic),
			"cluster_type":       aws.StringValue(scheduledActionType.ResizeCluster.ClusterType),
			"node_type":          aws.StringValue(scheduledActionType.ResizeCluster.NodeType),
			"number_of_nodes":    aws.Int64Value(scheduledActionType.ResizeCluster.NumberOfNodes),
		}
	}
	return []map[string]interface{}{m}
}
