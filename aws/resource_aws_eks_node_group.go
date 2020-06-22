package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEksNodeGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEksNodeGroupCreate,
		Read:   resourceAwsEksNodeGroupRead,
		Update: resourceAwsEksNodeGroupUpdate,
		Delete: resourceAwsEksNodeGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"ami_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					eks.AMITypesAl2X8664,
					eks.AMITypesAl2X8664Gpu,
				}, false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"disk_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"force_update_version": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"instance_types": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				// Multiple instance types returns an API error currently:
				// InvalidParameterException: Instance type list not valid, only one instance type is supported!
				MaxItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"node_group_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"node_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"release_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"remote_access": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ec2_ssh_key": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"source_security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"autoscaling_groups": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"remote_access_security_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"scaling_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"max_size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"min_size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchema(),
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceAwsEksNodeGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn
	clusterName := d.Get("cluster_name").(string)
	nodeGroupName := d.Get("node_group_name").(string)

	input := &eks.CreateNodegroupInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		ClusterName:        aws.String(clusterName),
		NodegroupName:      aws.String(nodeGroupName),
		NodeRole:           aws.String(d.Get("node_role_arn").(string)),
		Subnets:            expandStringSet(d.Get("subnet_ids").(*schema.Set)),
	}

	if v, ok := d.GetOk("ami_type"); ok {
		input.AmiType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disk_size"); ok {
		input.DiskSize = aws.Int64(int64(v.(int)))
	}

	if v := d.Get("instance_types").([]interface{}); len(v) > 0 {
		input.InstanceTypes = expandStringList(v)
	}

	if v := d.Get("labels").(map[string]interface{}); len(v) > 0 {
		input.Labels = stringMapToPointers(v)
	}

	if v, ok := d.GetOk("release_version"); ok {
		input.ReleaseVersion = aws.String(v.(string))
	}

	if v := d.Get("remote_access").([]interface{}); len(v) > 0 {
		input.RemoteAccess = expandEksRemoteAccessConfig(v)
	}

	if v := d.Get("scaling_config").([]interface{}); len(v) > 0 {
		input.ScalingConfig = expandEksNodegroupScalingConfig(v)
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().EksTags()
	}

	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}

	_, err := conn.CreateNodegroup(input)

	id := fmt.Sprintf("%s:%s", clusterName, nodeGroupName)

	if err != nil {
		return fmt.Errorf("error creating EKS Node Group (%s): %s", id, err)
	}

	d.SetId(id)

	stateConf := resource.StateChangeConf{
		Pending: []string{eks.NodegroupStatusCreating},
		Target:  []string{eks.NodegroupStatusActive},
		Timeout: d.Timeout(schema.TimeoutCreate),
		Refresh: refreshEksNodeGroupStatus(conn, clusterName, nodeGroupName),
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for EKS Node Group (%s) creation: %s", d.Id(), err)
	}

	return resourceAwsEksNodeGroupRead(d, meta)
}

func resourceAwsEksNodeGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	clusterName, nodeGroupName, err := resourceAwsEksNodeGroupParseId(d.Id())
	if err != nil {
		return err
	}

	input := &eks.DescribeNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodeGroupName),
	}

	output, err := conn.DescribeNodegroup(input)

	if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] EKS Node Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EKS Node Group (%s): %s", d.Id(), err)
	}

	nodeGroup := output.Nodegroup
	if nodeGroup == nil {
		log.Printf("[WARN] EKS Node Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("ami_type", nodeGroup.AmiType)
	d.Set("arn", nodeGroup.NodegroupArn)
	d.Set("cluster_name", nodeGroup.ClusterName)
	d.Set("disk_size", nodeGroup.DiskSize)

	if err := d.Set("instance_types", aws.StringValueSlice(nodeGroup.InstanceTypes)); err != nil {
		return fmt.Errorf("error setting instance_types: %s", err)
	}

	if err := d.Set("labels", aws.StringValueMap(nodeGroup.Labels)); err != nil {
		return fmt.Errorf("error setting labels: %s", err)
	}

	d.Set("node_group_name", nodeGroup.NodegroupName)
	d.Set("node_role_arn", nodeGroup.NodeRole)
	d.Set("release_version", nodeGroup.ReleaseVersion)

	if err := d.Set("remote_access", flattenEksRemoteAccessConfig(nodeGroup.RemoteAccess)); err != nil {
		return fmt.Errorf("error setting remote_access: %s", err)
	}

	if err := d.Set("resources", flattenEksNodeGroupResources(nodeGroup.Resources)); err != nil {
		return fmt.Errorf("error setting resources: %s", err)
	}

	if err := d.Set("scaling_config", flattenEksNodeGroupScalingConfig(nodeGroup.ScalingConfig)); err != nil {
		return fmt.Errorf("error setting scaling_config: %s", err)
	}

	d.Set("status", nodeGroup.Status)

	if err := d.Set("subnet_ids", aws.StringValueSlice(nodeGroup.Subnets)); err != nil {
		return fmt.Errorf("error setting subnets: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.EksKeyValueTags(nodeGroup.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("version", nodeGroup.Version)

	return nil
}

func resourceAwsEksNodeGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn

	clusterName, nodeGroupName, err := resourceAwsEksNodeGroupParseId(d.Id())
	if err != nil {
		return err
	}

	if d.HasChanges("labels", "scaling_config") {
		oldLabelsRaw, newLabelsRaw := d.GetChange("labels")

		input := &eks.UpdateNodegroupConfigInput{
			ClientRequestToken: aws.String(resource.UniqueId()),
			ClusterName:        aws.String(clusterName),
			Labels:             expandEksUpdateLabelsPayload(oldLabelsRaw, newLabelsRaw),
			NodegroupName:      aws.String(nodeGroupName),
		}

		if v := d.Get("scaling_config").([]interface{}); len(v) > 0 {
			input.ScalingConfig = expandEksNodegroupScalingConfig(v)
		}

		output, err := conn.UpdateNodegroupConfig(input)

		if err != nil {
			return fmt.Errorf("error updating EKS Node Group (%s) config: %s", d.Id(), err)
		}

		if output == nil || output.Update == nil || output.Update.Id == nil {
			return fmt.Errorf("error determining EKS Node Group (%s) config update ID: empty response", d.Id())
		}

		updateID := aws.StringValue(output.Update.Id)

		err = waitForEksNodeGroupUpdate(conn, clusterName, nodeGroupName, updateID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for EKS Node Group (%s) config update (%s): %s", d.Id(), updateID, err)
		}
	}

	if d.HasChanges("release_version", "version") {
		input := &eks.UpdateNodegroupVersionInput{
			ClientRequestToken: aws.String(resource.UniqueId()),
			ClusterName:        aws.String(clusterName),
			Force:              aws.Bool(d.Get("force_update_version").(bool)),
			NodegroupName:      aws.String(nodeGroupName),
		}

		if v, ok := d.GetOk("release_version"); ok && d.HasChange("release_version") {
			input.ReleaseVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("version"); ok {
			input.Version = aws.String(v.(string))
		}

		output, err := conn.UpdateNodegroupVersion(input)

		if err != nil {
			return fmt.Errorf("error updating EKS Node Group (%s) version: %s", d.Id(), err)
		}

		if output == nil || output.Update == nil || output.Update.Id == nil {
			return fmt.Errorf("error determining EKS Node Group (%s) version update ID: empty response", d.Id())
		}

		updateID := aws.StringValue(output.Update.Id)

		err = waitForEksNodeGroupUpdate(conn, clusterName, nodeGroupName, updateID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return fmt.Errorf("error waiting for EKS Node Group (%s) version update (%s): %s", d.Id(), updateID, err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.EksUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsEksNodeGroupRead(d, meta)
}

func resourceAwsEksNodeGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn

	clusterName, nodeGroupName, err := resourceAwsEksNodeGroupParseId(d.Id())
	if err != nil {
		return err
	}

	input := &eks.DeleteNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodeGroupName),
	}

	_, err = conn.DeleteNodegroup(input)

	if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EKS Node Group (%s): %s", d.Id(), err)
	}

	if err := waitForEksNodeGroupDeletion(conn, clusterName, nodeGroupName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for EKS Node Group (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func expandEksNodegroupScalingConfig(l []interface{}) *eks.NodegroupScalingConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &eks.NodegroupScalingConfig{}

	if v, ok := m["desired_size"].(int); ok && v != 0 {
		config.DesiredSize = aws.Int64(int64(v))
	}

	if v, ok := m["max_size"].(int); ok && v != 0 {
		config.MaxSize = aws.Int64(int64(v))
	}

	if v, ok := m["min_size"].(int); ok && v != 0 {
		config.MinSize = aws.Int64(int64(v))
	}

	return config
}

func expandEksRemoteAccessConfig(l []interface{}) *eks.RemoteAccessConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &eks.RemoteAccessConfig{}

	if v, ok := m["ec2_ssh_key"].(string); ok && v != "" {
		config.Ec2SshKey = aws.String(v)
	}

	if v, ok := m["source_security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		config.SourceSecurityGroups = expandStringSet(v)
	}

	return config
}

func expandEksUpdateLabelsPayload(oldLabelsMap, newLabelsMap interface{}) *eks.UpdateLabelsPayload {
	// EKS Labels operate similarly to keyvaluetags
	oldLabels := keyvaluetags.New(oldLabelsMap)
	newLabels := keyvaluetags.New(newLabelsMap)

	removedLabels := oldLabels.Removed(newLabels)
	updatedLabels := oldLabels.Updated(newLabels)

	if len(removedLabels) == 0 && len(updatedLabels) == 0 {
		return nil
	}

	updateLabelsPayload := &eks.UpdateLabelsPayload{}

	if len(removedLabels) > 0 {
		updateLabelsPayload.RemoveLabels = aws.StringSlice(removedLabels.Keys())
	}

	if len(updatedLabels) > 0 {
		updateLabelsPayload.AddOrUpdateLabels = aws.StringMap(updatedLabels.Map())
	}

	return updateLabelsPayload
}

func flattenEksAutoScalingGroups(autoScalingGroups []*eks.AutoScalingGroup) []map[string]interface{} {
	if len(autoScalingGroups) == 0 {
		return []map[string]interface{}{}
	}

	l := make([]map[string]interface{}, 0, len(autoScalingGroups))

	for _, autoScalingGroup := range autoScalingGroups {
		m := map[string]interface{}{
			"name": aws.StringValue(autoScalingGroup.Name),
		}

		l = append(l, m)
	}

	return l
}

func flattenEksNodeGroupResources(resources *eks.NodegroupResources) []map[string]interface{} {
	if resources == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"autoscaling_groups":              flattenEksAutoScalingGroups(resources.AutoScalingGroups),
		"remote_access_security_group_id": aws.StringValue(resources.RemoteAccessSecurityGroup),
	}

	return []map[string]interface{}{m}
}

func flattenEksNodeGroupScalingConfig(config *eks.NodegroupScalingConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"desired_size": aws.Int64Value(config.DesiredSize),
		"max_size":     aws.Int64Value(config.MaxSize),
		"min_size":     aws.Int64Value(config.MinSize),
	}

	return []map[string]interface{}{m}
}

func flattenEksRemoteAccessConfig(config *eks.RemoteAccessConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"ec2_ssh_key":               aws.StringValue(config.Ec2SshKey),
		"source_security_group_ids": aws.StringValueSlice(config.SourceSecurityGroups),
	}

	return []map[string]interface{}{m}
}

func refreshEksNodeGroupStatus(conn *eks.EKS, clusterName string, nodeGroupName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &eks.DescribeNodegroupInput{
			ClusterName:   aws.String(clusterName),
			NodegroupName: aws.String(nodeGroupName),
		}

		output, err := conn.DescribeNodegroup(input)

		if err != nil {
			return "", "", err
		}

		nodeGroup := output.Nodegroup

		if nodeGroup == nil {
			return nodeGroup, "", fmt.Errorf("EKS Node Group (%s:%s) missing", clusterName, nodeGroupName)
		}

		status := aws.StringValue(nodeGroup.Status)

		// Return enhanced error messaging if available, instead of:
		// unexpected state 'CREATE_FAILED', wanted target 'ACTIVE'. last error: %!s(<nil>)
		if status == eks.NodegroupStatusCreateFailed || status == eks.NodegroupStatusDeleteFailed {
			if nodeGroup.Health == nil || len(nodeGroup.Health.Issues) == 0 || nodeGroup.Health.Issues[0] == nil {
				return nodeGroup, status, fmt.Errorf("unable to find additional information about %s status, check EKS Node Group (%s:%s) health", status, clusterName, nodeGroupName)
			}

			issue := nodeGroup.Health.Issues[0]

			return nodeGroup, status, fmt.Errorf("%s: %s. Resource IDs: %v", aws.StringValue(issue.Code), aws.StringValue(issue.Message), aws.StringValueSlice(issue.ResourceIds))
		}

		return nodeGroup, status, nil
	}
}

func refreshEksNodeGroupUpdateStatus(conn *eks.EKS, clusterName string, nodeGroupName string, updateID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &eks.DescribeUpdateInput{
			Name:          aws.String(clusterName),
			NodegroupName: aws.String(nodeGroupName),
			UpdateId:      aws.String(updateID),
		}

		output, err := conn.DescribeUpdate(input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.Update == nil {
			return nil, "", fmt.Errorf("EKS Node Group (%s:%s) update (%s) missing", clusterName, nodeGroupName, updateID)
		}

		return output.Update, aws.StringValue(output.Update.Status), nil
	}
}

func waitForEksNodeGroupDeletion(conn *eks.EKS, clusterName string, nodeGroupName string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			eks.NodegroupStatusActive,
			eks.NodegroupStatusDeleting,
		},
		Target:  []string{""},
		Timeout: timeout,
		Refresh: refreshEksNodeGroupStatus(conn, clusterName, nodeGroupName),
	}

	_, err := stateConf.WaitForState()

	if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	return err
}

func waitForEksNodeGroupUpdate(conn *eks.EKS, clusterName, nodeGroupName string, updateID string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{eks.UpdateStatusInProgress},
		Target: []string{
			eks.UpdateStatusCancelled,
			eks.UpdateStatusFailed,
			eks.UpdateStatusSuccessful,
		},
		Timeout: timeout,
		Refresh: refreshEksNodeGroupUpdateStatus(conn, clusterName, nodeGroupName, updateID),
	}

	updateRaw, err := stateConf.WaitForState()

	if err != nil {
		return err
	}

	update := updateRaw.(*eks.Update)

	if aws.StringValue(update.Status) == eks.UpdateStatusSuccessful {
		return nil
	}

	var detailedErrors []string
	for i, updateError := range update.Errors {
		detailedErrors = append(detailedErrors, fmt.Sprintf("Error %d: Code: %s / Message: %s", i+1, aws.StringValue(updateError.ErrorCode), aws.StringValue(updateError.ErrorMessage)))
	}

	return fmt.Errorf("EKS Node Group (%s) update (%s) status (%s) not successful: Errors:\n%s", clusterName, updateID, aws.StringValue(update.Status), strings.Join(detailedErrors, "\n"))
}

func resourceAwsEksNodeGroupParseId(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected cluster-name:node-group-name", id)
	}

	return parts[0], parts[1], nil
}
