package aws

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const (
	emrInstanceGroupCreateTimeout = 10 * time.Minute
	emrInstanceGroupUpdateTimeout = 10 * time.Minute
)

func resourceAwsEMRInstanceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEMRInstanceGroupCreate,
		Read:   resourceAwsEMRInstanceGroupRead,
		Update: resourceAwsEMRInstanceGroupUpdate,
		Delete: resourceAwsEMRInstanceGroupDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected cluster-id/ig-id", d.Id())
				}
				clusterID := idParts[0]
				resourceID := idParts[1]
				d.Set("cluster_id", clusterID)
				d.SetId(resourceID)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"autoscaling_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
				ValidateFunc:     validation.StringIsJSON,
			},
			"bid_price": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"configurations_json": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         false,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"ebs_optimized": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"ebs_config": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iops": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateAwsEmrEbsVolumeType(),
						},
						"volumes_per_instance": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
							Default:  1,
						},
					},
				},
				Set: resourceAwsEMRClusterEBSConfigHash,
			},
			"instance_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"running_instance_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEMRInstanceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	instanceRole := emr.InstanceGroupTypeTask
	groupConfig := &emr.InstanceGroupConfig{
		EbsConfiguration: readEmrEBSConfig(d),
		InstanceRole:     aws.String(instanceRole),
		InstanceCount:    aws.Int64(int64(d.Get("instance_count").(int))),
		InstanceType:     aws.String(d.Get("instance_type").(string)),
		Name:             aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("autoscaling_policy"); ok {
		var autoScalingPolicy *emr.AutoScalingPolicy

		if err := json.Unmarshal([]byte(v.(string)), &autoScalingPolicy); err != nil {
			return fmt.Errorf("[DEBUG] error parsing Auto Scaling Policy %s", err)
		}
		groupConfig.AutoScalingPolicy = autoScalingPolicy
	}

	if v, ok := d.GetOk("configurations_json"); ok {
		info, err := structure.NormalizeJsonString(v)
		if err != nil {
			return fmt.Errorf("configurations_json contains an invalid JSON: %s", err)
		}
		groupConfig.Configurations, err = expandConfigurationJson(info)
		if err != nil {
			return fmt.Errorf("Error reading EMR configurations_json: %s", err)
		}
	}

	groupConfig.Market = aws.String(emr.MarketTypeOnDemand)
	if v, ok := d.GetOk("bid_price"); ok {
		groupConfig.BidPrice = aws.String(v.(string))
		groupConfig.Market = aws.String(emr.MarketTypeSpot)
	}

	params := &emr.AddInstanceGroupsInput{
		InstanceGroups: []*emr.InstanceGroupConfig{groupConfig},
		JobFlowId:      aws.String(d.Get("cluster_id").(string)),
	}

	log.Printf("[DEBUG] Creating EMR %s group with the following params: %s", instanceRole, params)
	resp, err := conn.AddInstanceGroups(params)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Created EMR %s group finished: %#v", instanceRole, resp)
	if resp == nil || len(resp.InstanceGroupIds) == 0 {
		return fmt.Errorf("Error creating instance groups: no instance group returned")
	}
	d.SetId(*resp.InstanceGroupIds[0])

	if err := waitForEmrInstanceGroupStateRunning(conn, d.Get("cluster_id").(string), d.Id(), emrInstanceGroupCreateTimeout); err != nil {
		return fmt.Errorf("error waiting for EMR Instance Group (%s) creation: %s", d.Id(), err)
	}

	return resourceAwsEMRInstanceGroupRead(d, meta)
}

func resourceAwsEMRInstanceGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	ig, err := fetchEMRInstanceGroup(conn, d.Get("cluster_id").(string), d.Id())

	if isResourceNotFoundError(err) {
		log.Printf("[DEBUG] EMR Instance Group (%s) not found, removing", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EMR Instance Group (%s): %s", d.Id(), err)
	}

	if ig.Status != nil {
		switch aws.StringValue(ig.Status.State) {
		case emr.InstanceGroupStateTerminating:
			fallthrough
		case emr.InstanceGroupStateTerminated:
			log.Printf("[DEBUG] EMR Instance Group (%s) terminated, removing", d.Id())
			d.SetId("")
			return nil
		}
	}

	switch {
	case len(ig.Configurations) > 0:
		configOut, err := flattenConfigurationJson(ig.Configurations)
		if err != nil {
			return fmt.Errorf("Error reading EMR instance group configurations: %s", err)
		}
		if err := d.Set("configurations_json", configOut); err != nil {
			return fmt.Errorf("Error setting EMR configurations_json for instance group (%s): %s", d.Id(), err)
		}
	default:
		d.Set("configurations_json", "")
	}

	var autoscalingPolicyString string
	if ig.AutoScalingPolicy != nil {
		// AutoScalingPolicy has an additional Status field and null values that are causing a new hashcode to be generated for `instance_group`.
		// We are purposefully omitting that field and the null values here when we flatten the autoscaling policy string for the statefile.
		for i, rule := range ig.AutoScalingPolicy.Rules {
			for j, dimension := range rule.Trigger.CloudWatchAlarmDefinition.Dimensions {
				if aws.StringValue(dimension.Key) == "JobFlowId" {
					tmpDimensions := append(ig.AutoScalingPolicy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions[:j], ig.AutoScalingPolicy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions[j+1:]...)
					ig.AutoScalingPolicy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions = tmpDimensions
				}
			}

			if len(ig.AutoScalingPolicy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions) == 0 {
				ig.AutoScalingPolicy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions = nil
			}
		}

		autoscalingPolicyConstraintsBytes, err := json.Marshal(ig.AutoScalingPolicy.Constraints)
		if err != nil {
			return fmt.Errorf("error parsing EMR Cluster Instance Groups AutoScalingPolicy Constraints: %s", err)
		}

		autoscalingPolicyRulesBytes, err := marshalWithoutNil(ig.AutoScalingPolicy.Rules)
		if err != nil {
			return fmt.Errorf("error parsing EMR Cluster Instance Groups AutoScalingPolicy Rules: %s", err)
		}

		autoscalingPolicyString = fmt.Sprintf("{\"Constraints\":%s,\"Rules\":%s}", string(autoscalingPolicyConstraintsBytes), string(autoscalingPolicyRulesBytes))
	}
	d.Set("autoscaling_policy", autoscalingPolicyString)

	d.Set("bid_price", ig.BidPrice)
	if err := d.Set("ebs_config", flattenEBSConfig(ig.EbsBlockDevices)); err != nil {
		return fmt.Errorf("error setting ebs_config: %s", err)
	}
	d.Set("ebs_optimized", ig.EbsOptimized)
	d.Set("instance_count", ig.RequestedInstanceCount)
	d.Set("instance_type", ig.InstanceType)
	d.Set("name", ig.Name)
	d.Set("running_instance_count", ig.RunningInstanceCount)

	if ig.Status != nil {
		d.Set("status", ig.Status.State)
	}

	return nil
}

func resourceAwsEMRInstanceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	log.Printf("[DEBUG] Modify EMR task group")
	if d.HasChanges("instance_count", "configurations_json") {
		instanceGroupModifyConfig := emr.InstanceGroupModifyConfig{
			InstanceGroupId: aws.String(d.Id()),
		}

		if d.HasChange("instance_count") {
			instanceCount := d.Get("instance_count").(int)
			instanceGroupModifyConfig.InstanceCount = aws.Int64(int64(instanceCount))
		}
		if d.HasChange("configurations_json") {
			if v, ok := d.GetOk("configurations_json"); ok {
				info, err := structure.NormalizeJsonString(v)
				if err != nil {
					return fmt.Errorf("configurations_json contains an invalid JSON: %s", err)
				}
				instanceGroupModifyConfig.Configurations, err = expandConfigurationJson(info)
				if err != nil {
					return fmt.Errorf("Error reading EMR configurations_json: %s", err)
				}
			}
		}
		params := &emr.ModifyInstanceGroupsInput{
			InstanceGroups: []*emr.InstanceGroupModifyConfig{
				&instanceGroupModifyConfig,
			},
		}

		_, err := conn.ModifyInstanceGroups(params)
		if err != nil {
			return fmt.Errorf("error modifying EMR Instance Group (%s): %s", d.Id(), err)
		}

		if err := waitForEmrInstanceGroupStateRunning(conn, d.Get("cluster_id").(string), d.Id(), emrInstanceGroupUpdateTimeout); err != nil {
			return fmt.Errorf("error waiting for EMR Instance Group (%s) modification: %s", d.Id(), err)
		}
	}

	if d.HasChange("autoscaling_policy") {
		var autoScalingPolicy *emr.AutoScalingPolicy

		if err := json.Unmarshal([]byte(d.Get("autoscaling_policy").(string)), &autoScalingPolicy); err != nil {
			return fmt.Errorf("error parsing EMR Auto Scaling Policy JSON for update: %s", err)
		}

		putAutoScalingPolicy := &emr.PutAutoScalingPolicyInput{
			ClusterId:         aws.String(d.Get("cluster_id").(string)),
			AutoScalingPolicy: autoScalingPolicy,
			InstanceGroupId:   aws.String(d.Id()),
		}

		if _, err := conn.PutAutoScalingPolicy(putAutoScalingPolicy); err != nil {
			return fmt.Errorf("error updating autoscaling policy for instance group %q: %s", d.Id(), err)
		}
	}

	return resourceAwsEMRInstanceGroupRead(d, meta)
}

func resourceAwsEMRInstanceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	log.Printf("[WARN] AWS EMR Instance Group does not support DELETE; resizing cluster to zero before removing from state")
	params := &emr.ModifyInstanceGroupsInput{
		InstanceGroups: []*emr.InstanceGroupModifyConfig{
			{
				InstanceGroupId: aws.String(d.Id()),
				InstanceCount:   aws.Int64(0),
			},
		},
	}

	if _, err := conn.ModifyInstanceGroups(params); err != nil {
		return fmt.Errorf("error draining EMR Instance Group (%s): %s", d.Id(), err)
	}
	return nil
}

func instanceGroupStateRefresh(conn *emr.EMR, clusterID, groupID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ig, err := fetchEMRInstanceGroup(conn, clusterID, groupID)
		if err != nil {
			return nil, "Not Found", err
		}

		if ig.Status == nil || ig.Status.State == nil {
			log.Printf("[WARN] ERM Instance Group found, but without state")
			return nil, "Undefined", fmt.Errorf("Undefined EMR Cluster Instance Group state")
		}

		return ig, *ig.Status.State, nil
	}
}

func fetchEMRInstanceGroup(conn *emr.EMR, clusterID, groupID string) (*emr.InstanceGroup, error) {
	input := &emr.ListInstanceGroupsInput{ClusterId: aws.String(clusterID)}

	var groups []*emr.InstanceGroup
	err := conn.ListInstanceGroupsPages(input, func(page *emr.ListInstanceGroupsOutput, lastPage bool) bool {
		groups = append(groups, page.InstanceGroups...)

		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("unable to retrieve EMR Cluster (%q): %s", clusterID, err)
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("no instance groups found for EMR Cluster (%s)", clusterID)
	}

	var ig *emr.InstanceGroup
	for _, group := range groups {
		if groupID == aws.StringValue(group.Id) {
			ig = group
			break
		}
	}

	if ig == nil {
		return nil, &resource.NotFoundError{}
	}

	return ig, nil
}

// readEmrEBSConfig populates an emr.EbsConfiguration struct
func readEmrEBSConfig(d *schema.ResourceData) *emr.EbsConfiguration {
	result := &emr.EbsConfiguration{}
	if v, ok := d.GetOk("ebs_optimized"); ok {
		result.EbsOptimized = aws.Bool(v.(bool))
	}

	ebsConfigs := make([]*emr.EbsBlockDeviceConfig, 0)
	if rawConfig, ok := d.GetOk("ebs_config"); ok {
		configList := rawConfig.(*schema.Set).List()
		for _, config := range configList {
			conf := config.(map[string]interface{})
			ebs := &emr.EbsBlockDeviceConfig{}
			volumeSpec := &emr.VolumeSpecification{
				SizeInGB:   aws.Int64(int64(conf["size"].(int))),
				VolumeType: aws.String(conf["type"].(string)),
			}
			if v, ok := conf["iops"].(int); ok && v != 0 {
				volumeSpec.Iops = aws.Int64(int64(v))
			}
			if v, ok := conf["volumes_per_instance"].(int); ok && v != 0 {
				ebs.VolumesPerInstance = aws.Int64(int64(v))
			}
			ebs.VolumeSpecification = volumeSpec
			ebsConfigs = append(ebsConfigs, ebs)
		}
	}
	result.EbsBlockDeviceConfigs = ebsConfigs
	return result
}

// marshalWithoutNil returns a JSON document of v stripped of any null properties
func marshalWithoutNil(v interface{}) ([]byte, error) {
	//removeNil is a helper for stripping nil values
	removeNil := func(data map[string]interface{}) map[string]interface{} {

		m := make(map[string]interface{})
		for k, v := range data {
			if v == nil {
				continue
			}

			switch v := v.(type) {
			case map[string]interface{}:
				m[k] = removeNil(v)
			default:
				m[k] = v
			}
		}

		return m
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var rules []map[string]interface{}
	if err := json.Unmarshal(b, &rules); err != nil {
		return nil, err
	}

	var cleanRules []map[string]interface{}
	for _, rule := range rules {
		cleanRules = append(cleanRules, removeNil(rule))
	}

	return json.Marshal(cleanRules)
}

func waitForEmrInstanceGroupStateRunning(conn *emr.EMR, clusterID string, instanceGroupID string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			emr.InstanceGroupStateBootstrapping,
			emr.InstanceGroupStateProvisioning,
			emr.InstanceGroupStateReconfiguring,
			emr.InstanceGroupStateResizing,
		},
		Target:     []string{emr.InstanceGroupStateRunning},
		Refresh:    instanceGroupStateRefresh(conn, clusterID, instanceGroupID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}
