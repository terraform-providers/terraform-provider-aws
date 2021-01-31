package aws

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/emrcontainers/waiter"
)

func resourceAwsEMRContainersVirtualCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEMRContainersVirtualClusterCreate,
		Read:   resourceAwsEMRContainersVirtualClusterRead,
		Delete: resourceAwsEMRContainersVirtualClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"container_provider": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"info": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"eks_info": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"namespace": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
												},
											},
										},
									},
								},
							},
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice(emrcontainers.ContainerProviderType_Values(), false),
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEMRContainersVirtualClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrcontainersconn

	input := emrcontainers.CreateVirtualClusterInput{
		ContainerProvider: expandEMRContainersContainerProvider(d.Get("container_provider").([]interface{})),
		Name: aws.String(d.Get("name").(string)),
	}

	log.Printf("[INFO] Creating EMR containers virtual cluster: %s", input)
	out, err := conn.CreateVirtualCluster(&input)
	if err != nil {
		return fmt.Errorf("error creating EMR containers virtual cluster: %w", err)
	}

	if _, err := waiter.VirtualClusterCreated(conn, aws.StringValue(out.Id)); err != nil {
		return fmt.Errorf("error waiting for EMR containers virtual cluster (%s) creation: %w", d.Id(), err)
	}

	return resourceAwsEMRContainersVirtualClusterRead(d, meta)
}

func resourceAwsEMRContainersVirtualClusterRead(d *schema.ResourceData, meta interface{}) error {
	emrconn := meta.(*AWSClient).emrconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &emr.DescribeClusterInput{
		ClusterId: aws.String(d.Id()),
	}

	resp, err := emrconn.DescribeCluster(req)
	if err != nil {
		// After a Cluster has been terminated for an indeterminate period of time,
		// the EMR API will return this type of error:
		//   InvalidRequestException: Cluster id 'j-XXX' is not valid.
		// If this causes issues with masking other legitimate request errors, the
		// handling should be updated for deeper inspection of the special error type
		// which includes an accurate error code:
		//   ErrorCode: "NoSuchCluster",
		if isAWSErr(err, emr.ErrCodeInvalidRequestException, "is not valid") {
			log.Printf("[DEBUG] EMR Cluster (%s) not found", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading EMR cluster: %s", err)
	}

	if resp.Cluster == nil {
		log.Printf("[DEBUG] EMR Cluster (%s) not found", d.Id())
		d.SetId("")
		return nil
	}

	cluster := resp.Cluster

	if cluster.Status != nil {
		state := aws.StringValue(cluster.Status.State)

		if state == emr.ClusterStateTerminated || state == emr.ClusterStateTerminatedWithErrors {
			log.Printf("[WARN] EMR Cluster (%s) was %s already, removing from state", d.Id(), state)
			d.SetId("")
			return nil
		}

		d.Set("cluster_state", state)

		d.Set("arn", aws.StringValue(cluster.ClusterArn))
	}

	instanceGroups, err := fetchAllEMRInstanceGroups(emrconn, d.Id())

	if err == nil { // find instance group

		coreGroup := emrCoreInstanceGroup(instanceGroups)
		masterGroup := findMasterGroup(instanceGroups)

		flattenedCoreInstanceGroup, err := flattenEmrCoreInstanceGroup(coreGroup)

		if err != nil {
			return fmt.Errorf("error flattening core_instance_group: %s", err)
		}

		if err := d.Set("core_instance_group", flattenedCoreInstanceGroup); err != nil {
			return fmt.Errorf("error setting core_instance_group: %s", err)
		}

		if err := d.Set("master_instance_group", flattenEmrMasterInstanceGroup(masterGroup)); err != nil {
			return fmt.Errorf("error setting master_instance_group: %s", err)
		}
	}

	instanceFleets, err := fetchAllEMRInstanceFleets(emrconn, d.Id())

	if err == nil { // find instance fleets

		coreFleet := findInstanceFleet(instanceFleets, emr.InstanceFleetTypeCore)
		masterFleet := findInstanceFleet(instanceFleets, emr.InstanceFleetTypeMaster)

		flattenedCoreInstanceFleet := flattenInstanceFleet(coreFleet)
		if err := d.Set("core_instance_fleet", flattenedCoreInstanceFleet); err != nil {
			return fmt.Errorf("error setting core_instance_fleet: %s", err)
		}

		flattenedMasterInstanceFleet := flattenInstanceFleet(masterFleet)
		if err := d.Set("master_instance_fleet", flattenedMasterInstanceFleet); err != nil {
			return fmt.Errorf("error setting master_instance_fleet: %s", err)
		}
	}

	if err := d.Set("tags", keyvaluetags.EmrKeyValueTags(cluster.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error settings tags: %s", err)
	}

	d.Set("name", cluster.Name)

	d.Set("service_role", cluster.ServiceRole)
	d.Set("security_configuration", cluster.SecurityConfiguration)
	d.Set("autoscaling_role", cluster.AutoScalingRole)
	d.Set("release_label", cluster.ReleaseLabel)
	d.Set("log_uri", cluster.LogUri)
	d.Set("master_public_dns", cluster.MasterPublicDnsName)
	d.Set("visible_to_all_users", cluster.VisibleToAllUsers)
	d.Set("ebs_root_volume_size", cluster.EbsRootVolumeSize)
	d.Set("scale_down_behavior", cluster.ScaleDownBehavior)
	d.Set("termination_protection", cluster.TerminationProtected)
	d.Set("step_concurrency_level", cluster.StepConcurrencyLevel)

	if cluster.CustomAmiId != nil {
		d.Set("custom_ami_id", cluster.CustomAmiId)
	}

	if err := d.Set("applications", flattenApplications(cluster.Applications)); err != nil {
		return fmt.Errorf("error setting EMR Applications for cluster (%s): %s", d.Id(), err)
	}

	if _, ok := d.GetOk("configurations_json"); ok {
		configOut, err := flattenConfigurationJson(cluster.Configurations)
		if err != nil {
			return fmt.Errorf("Error reading EMR cluster configurations: %s", err)
		}
		if err := d.Set("configurations_json", configOut); err != nil {
			return fmt.Errorf("Error setting EMR configurations_json for cluster (%s): %s", d.Id(), err)
		}
	}

	if err := d.Set("ec2_attributes", flattenEc2Attributes(cluster.Ec2InstanceAttributes)); err != nil {
		return fmt.Errorf("error setting EMR Ec2 Attributes: %s", err)
	}

	if err := d.Set("kerberos_attributes", flattenEmrKerberosAttributes(d, cluster.KerberosAttributes)); err != nil {
		return fmt.Errorf("error setting kerberos_attributes: %s", err)
	}

	respBootstraps, err := emrconn.ListBootstrapActions(&emr.ListBootstrapActionsInput{
		ClusterId: cluster.Id,
	})
	if err != nil {
		return fmt.Errorf("error listing bootstrap actions: %s", err)
	}

	if err := d.Set("bootstrap_action", flattenBootstrapArguments(respBootstraps.BootstrapActions)); err != nil {
		return fmt.Errorf("error setting Bootstrap Actions: %s", err)
	}

	var stepSummaries []*emr.StepSummary
	listStepsInput := &emr.ListStepsInput{
		ClusterId: aws.String(d.Id()),
	}
	err = emrconn.ListStepsPages(listStepsInput, func(page *emr.ListStepsOutput, lastPage bool) bool {
		// ListSteps returns steps in reverse order (newest first)
		for _, step := range page.Steps {
			stepSummaries = append([]*emr.StepSummary{step}, stepSummaries...)
		}
		return !lastPage
	})
	if err != nil {
		return fmt.Errorf("error listing steps: %s", err)
	}
	if err := d.Set("step", flattenEmrStepSummaries(stepSummaries)); err != nil {
		return fmt.Errorf("error setting step: %s", err)
	}

	// AWS provides no other way to read back the additional_info
	if v, ok := d.GetOk("additional_info"); ok {
		info, err := structure.NormalizeJsonString(v)
		if err != nil {
			return fmt.Errorf("Additional Info contains an invalid JSON: %v", err)
		}
		d.Set("additional_info", info)
	}

	return nil
}

func resourceAwsEMRContainersVirtualClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	req := &emr.TerminateJobFlowsInput{
		JobFlowIds: []*string{
			aws.String(d.Id()),
		},
	}

	_, err := conn.TerminateJobFlows(req)
	if err != nil {
		log.Printf("[ERROR], %s", err)
		return err
	}

	input := &emr.ListInstancesInput{
		ClusterId: aws.String(d.Id()),
	}
	var resp *emr.ListInstancesOutput
	var count int
	err = resource.Retry(20*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.ListInstances(input)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		count = countEMRRemainingInstances(resp, d.Id())
		if count != 0 {
			return resource.RetryableError(fmt.Errorf("EMR Cluster (%s) has (%d) Instances remaining", d.Id(), count))
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		resp, err = conn.ListInstances(input)

		if err == nil {
			count = countEMRRemainingInstances(resp, d.Id())
		}
	}

	if count != 0 {
		return fmt.Errorf("EMR Cluster (%s) has (%d) Instances remaining", d.Id(), count)
	}

	if err != nil {
		return fmt.Errorf("error waiting for EMR Cluster (%s) Instances to drain: %s", d.Id(), err)
	}

	return nil
}

func expandEMRContainersContainerProvider(l []interface{}) *emrcontainers.ContainerProvider {
	m := l[0].(map[string]interface{})

	input := emrcontainers.ContainerProvider{
		Id: aws.String(m["id"].(string)),
		Type: aws.String(m["type"].(string)),
	}

	if v, ok := m["info"]; ok {
		input.Info = expandEMRContainersContainerInfo(v.([]interface{}))
	}

	return &input
}

func expandEMRContainersContainerInfo(l []interface{}) *emrcontainers.ContainerInfo {
	m := l[0].(map[string]interface{})

	input := emrcontainers.ContainerInfo{}

	if v, ok := m["eks_info"]; ok {
		input.EksInfo = expandEMRContainersEksInfo(v.([]interface{}))
	}

	return &input
}

func expandEMRContainersEksInfo(l []interface{}) *emrcontainers.EksInfo {
	m := l[0].(map[string]interface{})

	input := emrcontainers.EksInfo{}

	if v, ok := m["namespace"]; ok {
		input.Namespace = aws.String(v.(string))
	}

	return &input
}
