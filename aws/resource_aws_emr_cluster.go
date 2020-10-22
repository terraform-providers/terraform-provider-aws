package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEMRCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEMRClusterCreate,
		Read:   resourceAwsEMRClusterRead,
		Update: resourceAwsEMRClusterUpdate,
		Delete: resourceAwsEMRClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"release_label": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"additional_info": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"cluster_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"log_uri": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// EMR uses a proprietary filesystem called EMRFS
					// and both s3n & s3 protocols are mapped to that FS
					// so they're equvivalent in this context (confirmed by AWS support)
					old = strings.Replace(old, "s3n://", "s3://", -1)
					return old == new
				},
			},
			"master_public_dns": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"applications": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"termination_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"keep_job_flow_alive_when_no_steps": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"ec2_attributes": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"additional_master_security_groups": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"additional_slave_security_groups": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"emr_managed_master_security_group": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"emr_managed_slave_security_group": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
						"instance_profile": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"service_access_security_group": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
					},
				},
			},
			"kerberos_attributes": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ad_domain_join_password": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
							ForceNew:  true,
						},
						"ad_domain_join_user": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"cross_realm_trust_principal_password": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
							ForceNew:  true,
						},
						"kdc_admin_password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
							ForceNew:  true,
						},
						"realm": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"core_instance_group": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
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
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntAtLeast(1),
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
					},
				},
			},
			"master_instance_group": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bid_price": {
							Type:     schema.TypeString,
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
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							Default:      1,
							ValidateFunc: validation.IntInSlice([]int{1, 3}),
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
					},
				},
			},
			"master_instance_fleet": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				MaxItems:      1,
				ConflictsWith: []string{"core_instance_group", "master_instance_group"},
				Elem:          InstanceFleetConfigSchema(),
			},
			"core_instance_fleet": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				Computed:      true,
				MaxItems:      1,
				ConflictsWith: []string{"core_instance_group", "master_instance_group"},
				Elem:          InstanceFleetConfigSchema(),
			},

			"bootstrap_action": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"args": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"step": {
				Type:       schema.TypeList,
				Optional:   true,
				Computed:   true,
				ForceNew:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action_on_failure": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								emr.ActionOnFailureCancelAndWait,
								emr.ActionOnFailureContinue,
								emr.ActionOnFailureTerminateCluster,
								emr.ActionOnFailureTerminateJobFlow,
							}, false),
						},
						"hadoop_jar_step": {
							Type:       schema.TypeList,
							MaxItems:   1,
							Required:   true,
							ForceNew:   true,
							ConfigMode: schema.SchemaConfigModeAttr,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"args": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"jar": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"main_class": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"properties": {
										Type:     schema.TypeMap,
										Optional: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
			"configurations": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"configurations_json"},
			},
			"configurations_json": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ConflictsWith:    []string{"configurations"},
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"service_role": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"scale_down_behavior": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					emr.ScaleDownBehaviorTerminateAtInstanceHour,
					emr.ScaleDownBehaviorTerminateAtTaskCompletion,
				}, false),
			},
			"security_configuration": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"autoscaling_role": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"visible_to_all_users": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"ebs_root_volume_size": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
			},
			"custom_ami_id": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ValidateFunc: validateAwsEmrCustomAmiId,
			},
			"step_concurrency_level": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 256),
			},
		},
	}
}

func InstanceFleetConfigSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_type_configs": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bid_price": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"bid_price_as_percentage_of_on_demand_price": {
							Type:     schema.TypeFloat,
							Optional: true,
							ForceNew: true,
							Default:  100,
						},
						"configurations": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"classification": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"properties": {
										Type:     schema.TypeMap,
										Optional: true,
										ForceNew: true,
										Elem:     schema.TypeString,
									},
								},
							},
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
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"weighted_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
							Default:  1,
						},
					},
				},
				Set: resourceAwsEMRInstanceTypeConfigHash,
			},
			"launch_specifications": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"on_demand_specification": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"allocation_strategy": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(emr.OnDemandProvisioningAllocationStrategy_Values(), false),
									},
								},
							},
						},
						"spot_specification": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"allocation_strategy": {
										Type:         schema.TypeString,
										ForceNew:     true,
										Required:     true,
										ValidateFunc: validation.StringInSlice(emr.SpotProvisioningAllocationStrategy_Values(), false),
									},
									"block_duration_minutes": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
										Default:  0,
									},
									"timeout_action": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(emr.SpotProvisioningTimeoutAction_Values(), false),
									},
									"timeout_duration_minutes": {
										Type:     schema.TypeInt,
										ForceNew: true,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"target_on_demand_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  0,
			},
			"target_spot_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  0,
			},
			"provisioned_on_demand_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"provisioned_spot_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAwsEMRClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	log.Printf("[DEBUG] Creating EMR cluster")
	applications := d.Get("applications").(*schema.Set).List()

	keepJobFlowAliveWhenNoSteps := true
	if v, ok := d.GetOkExists("keep_job_flow_alive_when_no_steps"); ok {
		keepJobFlowAliveWhenNoSteps = v.(bool)
	}

	// For multiple master nodes, EMR automatically enables
	// termination protection and ignores this configuration at launch.
	// There is additional handling after the job flow is running
	// to potentially disable termination protection to match the
	// desired Terraform configuration.
	terminationProtection := false
	if v, ok := d.GetOk("termination_protection"); ok {
		terminationProtection = v.(bool)
	}
	instanceConfig := &emr.JobFlowInstancesConfig{
		KeepJobFlowAliveWhenNoSteps: aws.Bool(keepJobFlowAliveWhenNoSteps),
		TerminationProtected:        aws.Bool(terminationProtection),
	}

	if l := d.Get("master_instance_group").([]interface{}); len(l) > 0 && l[0] != nil {
		m := l[0].(map[string]interface{})

		instanceGroup := &emr.InstanceGroupConfig{
			InstanceCount: aws.Int64(int64(m["instance_count"].(int))),
			InstanceRole:  aws.String(emr.InstanceRoleTypeMaster),
			InstanceType:  aws.String(m["instance_type"].(string)),
			Market:        aws.String(emr.MarketTypeOnDemand),
			Name:          aws.String(m["name"].(string)),
		}

		if v, ok := m["bid_price"]; ok && v.(string) != "" {
			instanceGroup.BidPrice = aws.String(v.(string))
			instanceGroup.Market = aws.String(emr.MarketTypeSpot)
		}

		expandEbsConfig(m, instanceGroup)

		instanceConfig.InstanceGroups = append(instanceConfig.InstanceGroups, instanceGroup)
	}

	if l := d.Get("core_instance_group").([]interface{}); len(l) > 0 && l[0] != nil {
		m := l[0].(map[string]interface{})

		instanceGroup := &emr.InstanceGroupConfig{
			InstanceCount: aws.Int64(int64(m["instance_count"].(int))),
			InstanceRole:  aws.String(emr.InstanceRoleTypeCore),
			InstanceType:  aws.String(m["instance_type"].(string)),
			Market:        aws.String(emr.MarketTypeOnDemand),
			Name:          aws.String(m["name"].(string)),
		}

		if v, ok := m["autoscaling_policy"]; ok && v.(string) != "" {
			var autoScalingPolicy *emr.AutoScalingPolicy

			if err := json.Unmarshal([]byte(v.(string)), &autoScalingPolicy); err != nil {
				return fmt.Errorf("error parsing core_instance_group Auto Scaling Policy JSON: %s", err)
			}

			instanceGroup.AutoScalingPolicy = autoScalingPolicy
		}

		if v, ok := m["bid_price"]; ok && v.(string) != "" {
			instanceGroup.BidPrice = aws.String(v.(string))
			instanceGroup.Market = aws.String(emr.MarketTypeSpot)
		}

		expandEbsConfig(m, instanceGroup)

		instanceConfig.InstanceGroups = append(instanceConfig.InstanceGroups, instanceGroup)
	}

	if l := d.Get("master_instance_fleet").([]interface{}); len(l) > 0 && l[0] != nil {
		instanceFleetConfig := readInstanceFleetConfig(l[0].(map[string]interface{}), emr.InstanceFleetTypeMaster)
		instanceConfig.InstanceFleets = append(instanceConfig.InstanceFleets, instanceFleetConfig)
	}

	if l := d.Get("core_instance_fleet").([]interface{}); len(l) > 0 && l[0] != nil {
		instanceFleetConfig := readInstanceFleetConfig(l[0].(map[string]interface{}), emr.InstanceFleetTypeCore)
		instanceConfig.InstanceFleets = append(instanceConfig.InstanceFleets, instanceFleetConfig)
	}

	var instanceProfile string
	if a, ok := d.GetOk("ec2_attributes"); ok {
		ec2Attributes := a.([]interface{})
		attributes := ec2Attributes[0].(map[string]interface{})

		if v, ok := attributes["key_name"]; ok {
			instanceConfig.Ec2KeyName = aws.String(v.(string))
		}
		if v, ok := attributes["subnet_id"]; ok {
			instanceConfig.Ec2SubnetId = aws.String(v.(string))
		}

		if v, ok := attributes["additional_master_security_groups"]; ok {
			strSlice := strings.Split(v.(string), ",")
			for i, s := range strSlice {
				strSlice[i] = strings.TrimSpace(s)
			}
			instanceConfig.AdditionalMasterSecurityGroups = aws.StringSlice(strSlice)
		}

		if v, ok := attributes["additional_slave_security_groups"]; ok {
			strSlice := strings.Split(v.(string), ",")
			for i, s := range strSlice {
				strSlice[i] = strings.TrimSpace(s)
			}
			instanceConfig.AdditionalSlaveSecurityGroups = aws.StringSlice(strSlice)
		}

		if v, ok := attributes["emr_managed_master_security_group"]; ok {
			instanceConfig.EmrManagedMasterSecurityGroup = aws.String(v.(string))
		}
		if v, ok := attributes["emr_managed_slave_security_group"]; ok {
			instanceConfig.EmrManagedSlaveSecurityGroup = aws.String(v.(string))
		}

		if len(strings.TrimSpace(attributes["instance_profile"].(string))) != 0 {
			instanceProfile = strings.TrimSpace(attributes["instance_profile"].(string))
		}

		if v, ok := attributes["service_access_security_group"]; ok {
			instanceConfig.ServiceAccessSecurityGroup = aws.String(v.(string))
		}
	}

	emrApps := expandApplications(applications)

	params := &emr.RunJobFlowInput{
		Instances:    instanceConfig,
		Name:         aws.String(d.Get("name").(string)),
		Applications: emrApps,

		ReleaseLabel:      aws.String(d.Get("release_label").(string)),
		ServiceRole:       aws.String(d.Get("service_role").(string)),
		VisibleToAllUsers: aws.Bool(d.Get("visible_to_all_users").(bool)),
		Tags:              keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().EmrTags(),
	}

	if v, ok := d.GetOk("additional_info"); ok {
		info, err := structure.NormalizeJsonString(v)
		if err != nil {
			return fmt.Errorf("Additional Info contains an invalid JSON: %v", err)
		}
		params.AdditionalInfo = aws.String(info)
	}

	if v, ok := d.GetOk("log_uri"); ok {
		params.LogUri = aws.String(v.(string))
	}

	if v, ok := d.GetOk("autoscaling_role"); ok {
		params.AutoScalingRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("scale_down_behavior"); ok {
		params.ScaleDownBehavior = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_configuration"); ok {
		params.SecurityConfiguration = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ebs_root_volume_size"); ok {
		params.EbsRootVolumeSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("custom_ami_id"); ok {
		params.CustomAmiId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("step_concurrency_level"); ok {
		params.StepConcurrencyLevel = aws.Int64(int64(v.(int)))
	}

	if instanceProfile != "" {
		params.JobFlowRole = aws.String(instanceProfile)
	}

	if v, ok := d.GetOk("bootstrap_action"); ok {
		bootstrapActions := v.([]interface{})
		params.BootstrapActions = expandBootstrapActions(bootstrapActions)
	}
	if v, ok := d.GetOk("step"); ok {
		steps := v.([]interface{})
		params.Steps = expandEmrStepConfigs(steps)
	}
	if v, ok := d.GetOk("configurations"); ok {
		confUrl := v.(string)
		params.Configurations = expandConfigures(confUrl)
	}

	if v, ok := d.GetOk("configurations_json"); ok {
		info, err := structure.NormalizeJsonString(v)
		if err != nil {
			return fmt.Errorf("configurations_json contains an invalid JSON: %v", err)
		}
		params.Configurations, err = expandConfigurationJson(info)
		if err != nil {
			return fmt.Errorf("Error reading EMR configurations_json: %s", err)
		}
	}

	if v, ok := d.GetOk("kerberos_attributes"); ok {
		kerberosAttributesList := v.([]interface{})
		kerberosAttributesMap := kerberosAttributesList[0].(map[string]interface{})
		params.KerberosAttributes = expandEmrKerberosAttributes(kerberosAttributesMap)
	}

	log.Printf("[DEBUG] EMR Cluster create options: %s", params)

	var resp *emr.RunJobFlowOutput
	err := resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error
		resp, err = conn.RunJobFlow(params)
		if err != nil {
			if isAWSErr(err, "ValidationException", "Invalid InstanceProfile:") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, "AccessDeniedException", "Failed to authorize instance profile") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		resp, err = conn.RunJobFlow(params)
	}
	if err != nil {
		return fmt.Errorf("error running EMR Job Flow: %s", err)
	}

	d.SetId(*resp.JobFlowId)
	// This value can only be obtained through a deprecated function
	d.Set("keep_job_flow_alive_when_no_steps", params.Instances.KeepJobFlowAliveWhenNoSteps)

	log.Println("[INFO] Waiting for EMR Cluster to be available")

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			emr.ClusterStateBootstrapping,
			emr.ClusterStateStarting,
		},
		Target: []string{
			emr.ClusterStateRunning,
			emr.ClusterStateWaiting,
		},
		Refresh:    resourceAwsEMRClusterStateRefreshFunc(d, meta),
		Timeout:    75 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	clusterRaw, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for EMR Cluster state to be \"WAITING\" or \"RUNNING\": %s", err)
	}

	// For multiple master nodes, EMR automatically enables
	// termination protection and ignores the configuration at launch.
	// This additional handling is to potentially disable termination
	// protection to match the desired Terraform configuration.
	cluster := clusterRaw.(*emr.Cluster)

	if aws.BoolValue(cluster.TerminationProtected) != terminationProtection {
		input := &emr.SetTerminationProtectionInput{
			JobFlowIds:           []*string{aws.String(d.Id())},
			TerminationProtected: aws.Bool(terminationProtection),
		}

		if _, err := conn.SetTerminationProtection(input); err != nil {
			return fmt.Errorf("error setting EMR Cluster (%s) termination protection to match configuration: %s", d.Id(), err)
		}
	}

	return resourceAwsEMRClusterRead(d, meta)
}

func resourceAwsEMRClusterRead(d *schema.ResourceData, meta interface{}) error {
	emrconn := meta.(*AWSClient).emrconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &emr.DescribeClusterInput{
		ClusterId: aws.String(d.Id()),
	}

	resp, err := emrconn.DescribeCluster(req)
	if err != nil {
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

func resourceAwsEMRClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	if d.HasChange("visible_to_all_users") {
		_, errModify := conn.SetVisibleToAllUsers(&emr.SetVisibleToAllUsersInput{
			JobFlowIds:        []*string{aws.String(d.Id())},
			VisibleToAllUsers: aws.Bool(d.Get("visible_to_all_users").(bool)),
		})
		if errModify != nil {
			log.Printf("[ERROR] %s", errModify)
			return errModify
		}
	}

	if d.HasChange("termination_protection") {
		_, errModify := conn.SetTerminationProtection(&emr.SetTerminationProtectionInput{
			JobFlowIds:           []*string{aws.String(d.Id())},
			TerminationProtected: aws.Bool(d.Get("termination_protection").(bool)),
		})
		if errModify != nil {
			log.Printf("[ERROR] %s", errModify)
			return errModify
		}
	}

	if d.HasChange("core_instance_group.0.autoscaling_policy") {
		autoscalingPolicyStr := d.Get("core_instance_group.0.autoscaling_policy").(string)
		instanceGroupID := d.Get("core_instance_group.0.id").(string)

		if autoscalingPolicyStr != "" {
			var autoScalingPolicy *emr.AutoScalingPolicy

			if err := json.Unmarshal([]byte(autoscalingPolicyStr), &autoScalingPolicy); err != nil {
				return fmt.Errorf("error parsing core_instance_group Auto Scaling Policy JSON: %s", err)
			}

			input := &emr.PutAutoScalingPolicyInput{
				ClusterId:         aws.String(d.Id()),
				AutoScalingPolicy: autoScalingPolicy,
				InstanceGroupId:   aws.String(instanceGroupID),
			}

			if _, err := conn.PutAutoScalingPolicy(input); err != nil {
				return fmt.Errorf("error updating EMR Cluster (%s) Instance Group (%s) Auto Scaling Policy: %s", d.Id(), instanceGroupID, err)
			}
		} else {
			input := &emr.RemoveAutoScalingPolicyInput{
				ClusterId:       aws.String(d.Id()),
				InstanceGroupId: aws.String(instanceGroupID),
			}

			if _, err := conn.RemoveAutoScalingPolicy(input); err != nil {
				return fmt.Errorf("error removing EMR Cluster (%s) Instance Group (%s) Auto Scaling Policy: %s", d.Id(), instanceGroupID, err)
			}

			// RemoveAutoScalingPolicy seems to have eventual consistency.
			// Retry reading Instance Group configuration until the policy is removed.
			err := resource.Retry(1*time.Minute, func() *resource.RetryError {
				autoscalingPolicy, err := getEmrCoreInstanceGroupAutoscalingPolicy(conn, d.Id())

				if err != nil {
					return resource.NonRetryableError(err)
				}

				if autoscalingPolicy != nil {
					return resource.RetryableError(fmt.Errorf("EMR Cluster (%s) Instance Group (%s) Auto Scaling Policy still exists", d.Id(), instanceGroupID))
				}

				return nil
			})

			if isResourceTimeoutError(err) {
				var autoscalingPolicy *emr.AutoScalingPolicyDescription

				autoscalingPolicy, err = getEmrCoreInstanceGroupAutoscalingPolicy(conn, d.Id())

				if autoscalingPolicy != nil {
					err = fmt.Errorf("EMR Cluster (%s) Instance Group (%s) Auto Scaling Policy still exists", d.Id(), instanceGroupID)
				}
			}

			if err != nil {
				return fmt.Errorf("error waiting for EMR Cluster (%s) Instance Group (%s) Auto Scaling Policy removal: %s", d.Id(), instanceGroupID, err)
			}
		}
	}

	if d.HasChange("core_instance_group.0.instance_count") {
		instanceGroupID := d.Get("core_instance_group.0.id").(string)

		input := &emr.ModifyInstanceGroupsInput{
			InstanceGroups: []*emr.InstanceGroupModifyConfig{
				{
					InstanceGroupId: aws.String(instanceGroupID),
					InstanceCount:   aws.Int64(int64(d.Get("core_instance_group.0.instance_count").(int))),
				},
			},
		}

		if _, err := conn.ModifyInstanceGroups(input); err != nil {
			return fmt.Errorf("error modifying EMR Cluster (%s) Instance Group (%s): %s", d.Id(), instanceGroupID, err)
		}

		stateConf := &resource.StateChangeConf{
			Pending: []string{
				emr.InstanceGroupStateBootstrapping,
				emr.InstanceGroupStateProvisioning,
				emr.InstanceGroupStateReconfiguring,
				emr.InstanceGroupStateResizing,
			},
			Target:  []string{emr.InstanceGroupStateRunning},
			Refresh: instanceGroupStateRefresh(conn, d.Id(), instanceGroupID),
			Timeout: 20 * time.Minute,
			Delay:   10 * time.Second,
		}

		if _, err := stateConf.WaitForState(); err != nil {
			return fmt.Errorf("error waiting for EMR Cluster (%s) Instance Group (%s) modification: %s", d.Id(), instanceGroupID, err)
		}
	}

	if d.HasChange("instance_group") {
		o, n := d.GetChange("instance_group")
		oSet := o.(*schema.Set).List()
		nSet := n.(*schema.Set).List()
		for _, currInstanceGroup := range oSet {
			for _, nextInstanceGroup := range nSet {
				oInstanceGroup := currInstanceGroup.(map[string]interface{})
				nInstanceGroup := nextInstanceGroup.(map[string]interface{})

				if oInstanceGroup["instance_role"].(string) != nInstanceGroup["instance_role"].(string) || oInstanceGroup["name"].(string) != nInstanceGroup["name"].(string) {
					continue
				}

				// Prevent duplicate PutAutoScalingPolicy from earlier update logic
				if nInstanceGroup["id"] == d.Get("core_instance_group.0.id").(string) && d.HasChange("core_instance_group.0.autoscaling_policy") {
					continue
				}

				if v, ok := nInstanceGroup["autoscaling_policy"]; ok && v.(string) != "" {
					var autoScalingPolicy *emr.AutoScalingPolicy

					err := json.Unmarshal([]byte(v.(string)), &autoScalingPolicy)
					if err != nil {
						return fmt.Errorf("error parsing EMR Auto Scaling Policy JSON for update: \n\n%s\n\n%s", v.(string), err)
					}

					putAutoScalingPolicy := &emr.PutAutoScalingPolicyInput{
						ClusterId:         aws.String(d.Id()),
						AutoScalingPolicy: autoScalingPolicy,
						InstanceGroupId:   aws.String(oInstanceGroup["id"].(string)),
					}

					_, errModify := conn.PutAutoScalingPolicy(putAutoScalingPolicy)
					if errModify != nil {
						return fmt.Errorf("error updating autoscaling policy for instance group %q: %s", oInstanceGroup["id"].(string), errModify)
					}

					break
				}
			}

		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.EmrUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EMR Cluster (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChange("step_concurrency_level") {
		_, errModify := conn.ModifyCluster(&emr.ModifyClusterInput{
			ClusterId:            aws.String(d.Id()),
			StepConcurrencyLevel: aws.Int64(int64(d.Get("step_concurrency_level").(int))),
		})
		if errModify != nil {
			log.Printf("[ERROR] %s", errModify)
			return errModify
		}
	}

	return resourceAwsEMRClusterRead(d, meta)
}

func resourceAwsEMRClusterDelete(d *schema.ResourceData, meta interface{}) error {
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

func countEMRRemainingInstances(resp *emr.ListInstancesOutput, emrClusterId string) int {
	if resp == nil {
		log.Printf("[ERROR] response is nil")
		return 0
	}

	instanceCount := len(resp.Instances)

	if instanceCount == 0 {
		log.Printf("[DEBUG] No instances found for EMR Cluster (%s)", emrClusterId)
		return 0
	}

	// Collect instance status states, wait for all instances to be terminated
	// before moving on
	var terminated []string
	for j, i := range resp.Instances {
		if i.Status != nil {
			if aws.StringValue(i.Status.State) == emr.InstanceStateTerminated {
				terminated = append(terminated, *i.Ec2InstanceId)
			}
		} else {
			log.Printf("[DEBUG] Cluster instance (%d : %s) has no status", j, *i.Ec2InstanceId)
		}
	}
	if len(terminated) == instanceCount {
		log.Printf("[DEBUG] All (%d) EMR Cluster (%s) Instances terminated", instanceCount, emrClusterId)
		return 0
	}
	return len(resp.Instances)
}

func expandApplications(apps []interface{}) []*emr.Application {
	appOut := make([]*emr.Application, 0, len(apps))

	for _, appName := range expandStringList(apps) {
		app := &emr.Application{
			Name: appName,
		}
		appOut = append(appOut, app)
	}
	return appOut
}

func flattenApplications(apps []*emr.Application) []interface{} {
	appOut := make([]interface{}, 0, len(apps))

	for _, app := range apps {
		appOut = append(appOut, *app.Name)
	}
	return appOut
}

func flattenEc2Attributes(ia *emr.Ec2InstanceAttributes) []map[string]interface{} {
	attrs := map[string]interface{}{}
	result := make([]map[string]interface{}, 0)

	if ia.Ec2KeyName != nil {
		attrs["key_name"] = *ia.Ec2KeyName
	}
	if ia.Ec2SubnetId != nil {
		attrs["subnet_id"] = *ia.Ec2SubnetId
	}
	if ia.IamInstanceProfile != nil {
		attrs["instance_profile"] = *ia.IamInstanceProfile
	}
	if ia.EmrManagedMasterSecurityGroup != nil {
		attrs["emr_managed_master_security_group"] = *ia.EmrManagedMasterSecurityGroup
	}
	if ia.EmrManagedSlaveSecurityGroup != nil {
		attrs["emr_managed_slave_security_group"] = *ia.EmrManagedSlaveSecurityGroup
	}

	if len(ia.AdditionalMasterSecurityGroups) > 0 {
		strs := aws.StringValueSlice(ia.AdditionalMasterSecurityGroups)
		attrs["additional_master_security_groups"] = strings.Join(strs, ",")
	}
	if len(ia.AdditionalSlaveSecurityGroups) > 0 {
		strs := aws.StringValueSlice(ia.AdditionalSlaveSecurityGroups)
		attrs["additional_slave_security_groups"] = strings.Join(strs, ",")
	}

	if ia.ServiceAccessSecurityGroup != nil {
		attrs["service_access_security_group"] = *ia.ServiceAccessSecurityGroup
	}

	result = append(result, attrs)

	return result
}

func flattenEmrAutoScalingPolicyDescription(policy *emr.AutoScalingPolicyDescription) (string, error) {
	if policy == nil {
		return "", nil
	}

	// AutoScalingPolicy has an additional Status field and null values that are causing a new hashcode to be generated
	// for `instance_group`.
	// We are purposefully omitting that field and the null values here when we flatten the autoscaling policy string
	// for the statefile.
	for i, rule := range policy.Rules {
		for j, dimension := range rule.Trigger.CloudWatchAlarmDefinition.Dimensions {
			if *dimension.Key == "JobFlowId" {
				tmpDimensions := append(policy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions[:j], policy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions[j+1:]...)
				policy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions = tmpDimensions
			}
		}
		if len(policy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions) == 0 {
			policy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions = nil
		}
	}

	tmpAutoScalingPolicy := emr.AutoScalingPolicy{
		Constraints: policy.Constraints,
		Rules:       policy.Rules,
	}
	autoscalingPolicyConstraintsBytes, err := json.Marshal(tmpAutoScalingPolicy.Constraints)
	if err != nil {
		return "", fmt.Errorf("error parsing EMR Cluster Instance Groups AutoScalingPolicy Constraints: %s", err)
	}
	autoscalingPolicyConstraintsString := string(autoscalingPolicyConstraintsBytes)

	autoscalingPolicyRulesBytes, err := json.Marshal(tmpAutoScalingPolicy.Rules)
	if err != nil {
		return "", fmt.Errorf("error parsing EMR Cluster Instance Groups AutoScalingPolicy Rules: %s", err)
	}

	var rules []map[string]interface{}
	if err := json.Unmarshal(autoscalingPolicyRulesBytes, &rules); err != nil {
		return "", err
	}

	var cleanRules []map[string]interface{}
	for _, rule := range rules {
		cleanRules = append(cleanRules, removeNil(rule))
	}

	withoutNulls, err := json.Marshal(cleanRules)
	if err != nil {
		return "", err
	}
	autoscalingPolicyRulesString := string(withoutNulls)

	autoscalingPolicyString := fmt.Sprintf("{\"Constraints\":%s,\"Rules\":%s}", autoscalingPolicyConstraintsString, autoscalingPolicyRulesString)

	return autoscalingPolicyString, nil
}

func flattenEmrCoreInstanceGroup(instanceGroup *emr.InstanceGroup) ([]interface{}, error) {
	if instanceGroup == nil {
		return []interface{}{}, nil
	}

	autoscalingPolicy, err := flattenEmrAutoScalingPolicyDescription(instanceGroup.AutoScalingPolicy)

	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{
		"autoscaling_policy": autoscalingPolicy,
		"bid_price":          aws.StringValue(instanceGroup.BidPrice),
		"ebs_config":         flattenEBSConfig(instanceGroup.EbsBlockDevices),
		"id":                 aws.StringValue(instanceGroup.Id),
		"instance_count":     aws.Int64Value(instanceGroup.RequestedInstanceCount),
		"instance_type":      aws.StringValue(instanceGroup.InstanceType),
		"name":               aws.StringValue(instanceGroup.Name),
	}

	return []interface{}{m}, nil
}

func flattenEmrMasterInstanceGroup(instanceGroup *emr.InstanceGroup) []interface{} {
	if instanceGroup == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"bid_price":      aws.StringValue(instanceGroup.BidPrice),
		"ebs_config":     flattenEBSConfig(instanceGroup.EbsBlockDevices),
		"id":             aws.StringValue(instanceGroup.Id),
		"instance_count": aws.Int64Value(instanceGroup.RequestedInstanceCount),
		"instance_type":  aws.StringValue(instanceGroup.InstanceType),
		"name":           aws.StringValue(instanceGroup.Name),
	}

	return []interface{}{m}
}

func flattenEmrKerberosAttributes(d *schema.ResourceData, kerberosAttributes *emr.KerberosAttributes) []map[string]interface{} {
	l := make([]map[string]interface{}, 0)

	if kerberosAttributes == nil || kerberosAttributes.Realm == nil {
		return l
	}

	// Do not set from API:
	// * ad_domain_join_password
	// * ad_domain_join_user
	// * cross_realm_trust_principal_password
	// * kdc_admin_password

	m := map[string]interface{}{
		"kdc_admin_password": d.Get("kerberos_attributes.0.kdc_admin_password").(string),
		"realm":              *kerberosAttributes.Realm,
	}

	if v, ok := d.GetOk("kerberos_attributes.0.ad_domain_join_password"); ok {
		m["ad_domain_join_password"] = v.(string)
	}

	if v, ok := d.GetOk("kerberos_attributes.0.ad_domain_join_user"); ok {
		m["ad_domain_join_user"] = v.(string)
	}

	if v, ok := d.GetOk("kerberos_attributes.0.cross_realm_trust_principal_password"); ok {
		m["cross_realm_trust_principal_password"] = v.(string)
	}

	l = append(l, m)

	return l
}

func flattenEmrHadoopStepConfig(config *emr.HadoopStepConfig) map[string]interface{} {
	if config == nil {
		return nil
	}

	m := map[string]interface{}{
		"args":       aws.StringValueSlice(config.Args),
		"jar":        aws.StringValue(config.Jar),
		"main_class": aws.StringValue(config.MainClass),
		"properties": aws.StringValueMap(config.Properties),
	}

	return m
}

func flattenEmrStepSummaries(stepSummaries []*emr.StepSummary) []map[string]interface{} {
	l := make([]map[string]interface{}, 0)

	if len(stepSummaries) == 0 {
		return l
	}

	for _, stepSummary := range stepSummaries {
		l = append(l, flattenEmrStepSummary(stepSummary))
	}

	return l
}

func flattenEmrStepSummary(stepSummary *emr.StepSummary) map[string]interface{} {
	if stepSummary == nil {
		return nil
	}

	m := map[string]interface{}{
		"action_on_failure": aws.StringValue(stepSummary.ActionOnFailure),
		"hadoop_jar_step":   []map[string]interface{}{flattenEmrHadoopStepConfig(stepSummary.Config)},
		"name":              aws.StringValue(stepSummary.Name),
	}

	return m
}

func flattenEBSConfig(ebsBlockDevices []*emr.EbsBlockDevice) *schema.Set {
	uniqueEBS := make(map[int]int)
	ebsConfig := make([]interface{}, 0)
	for _, ebs := range ebsBlockDevices {
		ebsAttrs := make(map[string]interface{})
		if ebs.VolumeSpecification.Iops != nil {
			ebsAttrs["iops"] = int(*ebs.VolumeSpecification.Iops)
		}
		if ebs.VolumeSpecification.SizeInGB != nil {
			ebsAttrs["size"] = int(*ebs.VolumeSpecification.SizeInGB)
		}
		if ebs.VolumeSpecification.VolumeType != nil {
			ebsAttrs["type"] = *ebs.VolumeSpecification.VolumeType
		}
		ebsAttrs["volumes_per_instance"] = 1
		uniqueEBS[resourceAwsEMRClusterEBSConfigHash(ebsAttrs)] += 1
		ebsConfig = append(ebsConfig, ebsAttrs)
	}

	for _, ebs := range ebsConfig {
		ebs.(map[string]interface{})["volumes_per_instance"] = uniqueEBS[resourceAwsEMRClusterEBSConfigHash(ebs)]
	}
	return schema.NewSet(resourceAwsEMRClusterEBSConfigHash, ebsConfig)
}

func flattenBootstrapArguments(actions []*emr.Command) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, b := range actions {
		attrs := make(map[string]interface{})
		attrs["name"] = *b.Name
		attrs["path"] = *b.ScriptPath
		attrs["args"] = flattenStringList(b.Args)
		result = append(result, attrs)
	}

	return result
}

func emrCoreInstanceGroup(grps []*emr.InstanceGroup) *emr.InstanceGroup {
	for _, grp := range grps {
		if aws.StringValue(grp.InstanceGroupType) == emr.InstanceGroupTypeCore {
			return grp
		}
	}
	return nil
}

func expandBootstrapActions(bootstrapActions []interface{}) []*emr.BootstrapActionConfig {
	actionsOut := []*emr.BootstrapActionConfig{}

	for _, raw := range bootstrapActions {
		actionAttributes := raw.(map[string]interface{})
		actionName := actionAttributes["name"].(string)
		actionPath := actionAttributes["path"].(string)
		actionArgs := actionAttributes["args"].([]interface{})

		action := &emr.BootstrapActionConfig{
			Name: aws.String(actionName),
			ScriptBootstrapAction: &emr.ScriptBootstrapActionConfig{
				Path: aws.String(actionPath),
				Args: expandStringList(actionArgs),
			},
		}
		actionsOut = append(actionsOut, action)
	}

	return actionsOut
}

func expandEmrHadoopJarStepConfig(m map[string]interface{}) *emr.HadoopJarStepConfig {
	hadoopJarStepConfig := &emr.HadoopJarStepConfig{
		Jar: aws.String(m["jar"].(string)),
	}

	if v, ok := m["args"]; ok {
		hadoopJarStepConfig.Args = expandStringList(v.([]interface{}))
	}

	if v, ok := m["main_class"]; ok {
		hadoopJarStepConfig.MainClass = aws.String(v.(string))
	}

	if v, ok := m["properties"]; ok {
		hadoopJarStepConfig.Properties = expandEmrKeyValues(v.(map[string]interface{}))
	}

	return hadoopJarStepConfig
}

func expandEmrKeyValues(m map[string]interface{}) []*emr.KeyValue {
	keyValues := make([]*emr.KeyValue, 0)

	for k, v := range m {
		keyValue := &emr.KeyValue{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		keyValues = append(keyValues, keyValue)
	}

	return keyValues
}

func expandEmrKerberosAttributes(m map[string]interface{}) *emr.KerberosAttributes {
	kerberosAttributes := &emr.KerberosAttributes{
		KdcAdminPassword: aws.String(m["kdc_admin_password"].(string)),
		Realm:            aws.String(m["realm"].(string)),
	}
	if v, ok := m["ad_domain_join_password"]; ok && v.(string) != "" {
		kerberosAttributes.ADDomainJoinPassword = aws.String(v.(string))
	}
	if v, ok := m["ad_domain_join_user"]; ok && v.(string) != "" {
		kerberosAttributes.ADDomainJoinUser = aws.String(v.(string))
	}
	if v, ok := m["cross_realm_trust_principal_password"]; ok && v.(string) != "" {
		kerberosAttributes.CrossRealmTrustPrincipalPassword = aws.String(v.(string))
	}
	return kerberosAttributes
}

func expandEmrStepConfig(m map[string]interface{}) *emr.StepConfig {
	hadoopJarStepList := m["hadoop_jar_step"].([]interface{})
	hadoopJarStepMap := hadoopJarStepList[0].(map[string]interface{})

	stepConfig := &emr.StepConfig{
		ActionOnFailure: aws.String(m["action_on_failure"].(string)),
		HadoopJarStep:   expandEmrHadoopJarStepConfig(hadoopJarStepMap),
		Name:            aws.String(m["name"].(string)),
	}

	return stepConfig
}

func expandEmrStepConfigs(l []interface{}) []*emr.StepConfig {
	stepConfigs := []*emr.StepConfig{}

	for _, raw := range l {
		m := raw.(map[string]interface{})
		stepConfigs = append(stepConfigs, expandEmrStepConfig(m))
	}

	return stepConfigs
}

func expandEbsConfig(configAttributes map[string]interface{}, config *emr.InstanceGroupConfig) {
	if rawEbsConfigs, ok := configAttributes["ebs_config"]; ok {
		ebsConfig := &emr.EbsConfiguration{}

		ebsBlockDeviceConfigs := make([]*emr.EbsBlockDeviceConfig, 0)
		for _, rawEbsConfig := range rawEbsConfigs.(*schema.Set).List() {
			rawEbsConfig := rawEbsConfig.(map[string]interface{})
			ebsBlockDeviceConfig := &emr.EbsBlockDeviceConfig{
				VolumesPerInstance: aws.Int64(int64(rawEbsConfig["volumes_per_instance"].(int))),
				VolumeSpecification: &emr.VolumeSpecification{
					SizeInGB:   aws.Int64(int64(rawEbsConfig["size"].(int))),
					VolumeType: aws.String(rawEbsConfig["type"].(string)),
				},
			}
			if v, ok := rawEbsConfig["iops"].(int); ok && v != 0 {
				ebsBlockDeviceConfig.VolumeSpecification.Iops = aws.Int64(int64(v))
			}
			ebsBlockDeviceConfigs = append(ebsBlockDeviceConfigs, ebsBlockDeviceConfig)
		}
		ebsConfig.EbsBlockDeviceConfigs = ebsBlockDeviceConfigs

		config.EbsConfiguration = ebsConfig
	}
}

func expandConfigurationJson(input string) ([]*emr.Configuration, error) {
	configsOut := []*emr.Configuration{}
	err := json.Unmarshal([]byte(input), &configsOut)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Expanded EMR Configurations %s", configsOut)

	return configsOut, nil
}

func flattenConfigurationJson(config []*emr.Configuration) (string, error) {
	out, err := jsonutil.BuildJSON(config)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func expandConfigures(input string) []*emr.Configuration {
	configsOut := []*emr.Configuration{}
	if strings.HasPrefix(input, "http") {
		if err := readHttpJson(input, &configsOut); err != nil {
			log.Printf("[ERR] Error reading HTTP JSON: %s", err)
		}
	} else if strings.HasSuffix(input, ".json") {
		if err := readLocalJson(input, &configsOut); err != nil {
			log.Printf("[ERR] Error reading local JSON: %s", err)
		}
	} else {
		if err := readBodyJson(input, &configsOut); err != nil {
			log.Printf("[ERR] Error reading body JSON: %s", err)
		}
	}
	log.Printf("[DEBUG] Expanded EMR Configurations %s", configsOut)

	return configsOut
}

func readHttpJson(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func readLocalJson(localFile string, target interface{}) error {
	file, e := ioutil.ReadFile(localFile)
	if e != nil {
		log.Printf("[ERROR] %s", e)
		return e
	}

	return json.Unmarshal(file, target)
}

func readBodyJson(body string, target interface{}) error {
	log.Printf("[DEBUG] Raw Body %s\n", body)
	err := json.Unmarshal([]byte(body), target)
	if err != nil {
		log.Printf("[ERROR] parsing JSON %s", err)
		return err
	}
	return nil
}

func resourceAwsEMRClusterStateRefreshFunc(d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		conn := meta.(*AWSClient).emrconn

		log.Printf("[INFO] Reading EMR Cluster Information: %s", d.Id())
		params := &emr.DescribeClusterInput{
			ClusterId: aws.String(d.Id()),
		}

		resp, err := conn.DescribeCluster(params)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "ClusterNotFound" {
					return 42, "destroyed", nil
				}
			}
			log.Printf("[WARN] Error on retrieving EMR Cluster (%s) when waiting: %s", d.Id(), err)
			return nil, "", err
		}

		if resp.Cluster == nil {
			return 42, "destroyed", nil
		}

		if resp.Cluster.Status == nil {
			return resp.Cluster, "", fmt.Errorf("cluster status not provided")
		}

		state := aws.StringValue(resp.Cluster.Status.State)
		log.Printf("[DEBUG] EMR Cluster status (%s): %s", d.Id(), state)

		if state == emr.ClusterStateTerminating || state == emr.ClusterStateTerminatedWithErrors {
			reason := resp.Cluster.Status.StateChangeReason
			if reason == nil {
				return resp.Cluster, state, fmt.Errorf("%s: reason code and message not provided", state)
			}
			return resp.Cluster, state, fmt.Errorf("%s: %s: %s", state, aws.StringValue(reason.Code), aws.StringValue(reason.Message))
		}

		return resp.Cluster, state, nil
	}
}

func findMasterGroup(instanceGroups []*emr.InstanceGroup) *emr.InstanceGroup {
	for _, group := range instanceGroups {
		if *group.InstanceGroupType == emr.InstanceRoleTypeMaster {
			return group
		}
	}
	return nil
}

func resourceAwsEMRClusterEBSConfigHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m["size"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["type"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["volumes_per_instance"].(int)))
	if v, ok := m["iops"]; ok && v.(int) != 0 {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	return hashcode.String(buf.String())
}

func getEmrCoreInstanceGroupAutoscalingPolicy(conn *emr.EMR, clusterID string) (*emr.AutoScalingPolicyDescription, error) {
	instanceGroups, err := fetchAllEMRInstanceGroups(conn, clusterID)

	if err != nil {
		return nil, err
	}

	coreGroup := emrCoreInstanceGroup(instanceGroups)

	if coreGroup == nil {
		return nil, fmt.Errorf("EMR Cluster (%s) Core Instance Group not found", clusterID)
	}

	return coreGroup.AutoScalingPolicy, nil
}

func fetchAllEMRInstanceGroups(conn *emr.EMR, clusterID string) ([]*emr.InstanceGroup, error) {
	input := &emr.ListInstanceGroupsInput{
		ClusterId: aws.String(clusterID),
	}
	var groups []*emr.InstanceGroup

	err := conn.ListInstanceGroupsPages(input, func(page *emr.ListInstanceGroupsOutput, lastPage bool) bool {
		groups = append(groups, page.InstanceGroups...)

		return !lastPage
	})

	return groups, err
}

func readInstanceFleetConfig(data map[string]interface{}, InstanceFleetType string) *emr.InstanceFleetConfig {

	config := &emr.InstanceFleetConfig{
		InstanceFleetType:      &InstanceFleetType,
		Name:                   aws.String(data["name"].(string)),
		TargetOnDemandCapacity: aws.Int64(int64(data["target_on_demand_capacity"].(int))),
		TargetSpotCapacity:     aws.Int64(int64(data["target_spot_capacity"].(int))),
	}

	if v, ok := data["instance_type_configs"].(*schema.Set); ok && v.Len() > 0 {
		config.InstanceTypeConfigs = expandInstanceTypeConfigs(v.List())
	}

	if v, ok := data["launch_specifications"].([]interface{}); ok && len(v) == 1 {
		config.LaunchSpecifications = expandLaunchSpecification(v[0].(map[string]interface{}))
	}

	return config
}

func fetchAllEMRInstanceFleets(conn *emr.EMR, clusterID string) ([]*emr.InstanceFleet, error) {
	input := &emr.ListInstanceFleetsInput{
		ClusterId: aws.String(clusterID),
	}
	var fleets []*emr.InstanceFleet

	err := conn.ListInstanceFleetsPages(input, func(page *emr.ListInstanceFleetsOutput, lastPage bool) bool {
		fleets = append(fleets, page.InstanceFleets...)

		return !lastPage
	})

	return fleets, err
}

func findInstanceFleet(instanceFleets []*emr.InstanceFleet, instanceRoleType string) *emr.InstanceFleet {
	for _, instanceFleet := range instanceFleets {
		if instanceFleet.InstanceFleetType != nil {
			if aws.StringValue(instanceFleet.InstanceFleetType) == instanceRoleType {
				return instanceFleet
			}
		}
	}
	return nil
}

func flattenInstanceFleet(instanceFleet *emr.InstanceFleet) []interface{} {
	if instanceFleet == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"id":                             aws.StringValue(instanceFleet.Id),
		"name":                           aws.StringValue(instanceFleet.Name),
		"target_on_demand_capacity":      aws.Int64Value(instanceFleet.TargetOnDemandCapacity),
		"target_spot_capacity":           aws.Int64Value(instanceFleet.TargetSpotCapacity),
		"provisioned_on_demand_capacity": aws.Int64Value(instanceFleet.ProvisionedOnDemandCapacity),
		"provisioned_spot_capacity":      aws.Int64Value(instanceFleet.ProvisionedSpotCapacity),
		"instance_type_configs":          flatteninstanceTypeConfigs(instanceFleet.InstanceTypeSpecifications),
		"launch_specifications":          flattenLaunchSpecifications(instanceFleet.LaunchSpecifications),
	}

	return []interface{}{m}

}

func flatteninstanceTypeConfigs(instanceTypeSpecifications []*emr.InstanceTypeSpecification) *schema.Set {
	instanceTypeConfigs := make([]interface{}, 0)

	for _, itc := range instanceTypeSpecifications {
		flattenTypeConfig := make(map[string]interface{})

		if itc.BidPrice != nil {
			flattenTypeConfig["bid_price"] = aws.StringValue(itc.BidPrice)
		}

		if itc.BidPriceAsPercentageOfOnDemandPrice != nil {
			flattenTypeConfig["bid_price_as_percentage_of_on_demand_price"] = aws.Float64Value(itc.BidPriceAsPercentageOfOnDemandPrice)
		}

		flattenTypeConfig["instance_type"] = aws.StringValue(itc.InstanceType)
		flattenTypeConfig["weighted_capacity"] = int(aws.Int64Value(itc.WeightedCapacity))

		flattenTypeConfig["ebs_config"] = flattenEBSConfig(itc.EbsBlockDevices)

		instanceTypeConfigs = append(instanceTypeConfigs, flattenTypeConfig)
	}

	return schema.NewSet(resourceAwsEMRInstanceTypeConfigHash, instanceTypeConfigs)
}

func flattenLaunchSpecifications(launchSpecifications *emr.InstanceFleetProvisioningSpecifications) []interface{} {
	if launchSpecifications == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"on_demand_specification": flattenOnDemandSpecification(launchSpecifications.OnDemandSpecification),
		"spot_specification":      flattenSpotSpecification(launchSpecifications.SpotSpecification),
	}
	return []interface{}{m}
}

func flattenOnDemandSpecification(onDemandSpecification *emr.OnDemandProvisioningSpecification) []interface{} {
	if onDemandSpecification == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		// The return value from api is wrong. it return "LOWEST_PRICE" instead of "lowest-price"
		// "allocation_strategy": aws.StringValue(onDemandSpecification.AllocationStrategy),
		"allocation_strategy": "lowest-price",
	}
	return []interface{}{m}
}

func flattenSpotSpecification(spotSpecification *emr.SpotProvisioningSpecification) []interface{} {
	if spotSpecification == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"timeout_action":           aws.StringValue(spotSpecification.TimeoutAction),
		"timeout_duration_minutes": aws.Int64Value(spotSpecification.TimeoutDurationMinutes),
	}
	if spotSpecification.BlockDurationMinutes != nil {
		m["block_duration_minutes"] = aws.Int64Value(spotSpecification.BlockDurationMinutes)
	}
	if spotSpecification.AllocationStrategy != nil {
		// The return value from api is wrong. It return "CAPACITY_OPTIMIZED" instead of "capacity-optimized"
		// m["allocation_strategy"] = aws.StringValue(spotSpecification.AllocationStrategy)
		m["allocation_strategy"] = "capacity-optimized"
	}

	return []interface{}{m}
}

func expandEbsConfiguration(ebsConfigurations []interface{}) *emr.EbsConfiguration {
	ebsConfig := &emr.EbsConfiguration{}
	ebsConfigs := make([]*emr.EbsBlockDeviceConfig, 0)
	for _, ebsConfiguration := range ebsConfigurations {
		cfg := ebsConfiguration.(map[string]interface{})
		ebsBlockDeviceConfig := &emr.EbsBlockDeviceConfig{
			VolumesPerInstance: aws.Int64(int64(cfg["volumes_per_instance"].(int))),
			VolumeSpecification: &emr.VolumeSpecification{
				SizeInGB:   aws.Int64(int64(cfg["size"].(int))),
				VolumeType: aws.String(cfg["type"].(string)),
			},
		}
		if v, ok := cfg["iops"].(int); ok && v != 0 {
			ebsBlockDeviceConfig.VolumeSpecification.Iops = aws.Int64(int64(v))
		}
		ebsConfigs = append(ebsConfigs, ebsBlockDeviceConfig)
	}
	ebsConfig.EbsBlockDeviceConfigs = ebsConfigs
	return ebsConfig
}

func expandInstanceTypeConfigs(instanceTypeConfigs []interface{}) []*emr.InstanceTypeConfig {
	configsOut := []*emr.InstanceTypeConfig{}

	for _, raw := range instanceTypeConfigs {
		configAttributes := raw.(map[string]interface{})

		config := &emr.InstanceTypeConfig{
			InstanceType: aws.String(configAttributes["instance_type"].(string)),
		}

		if bidPrice, ok := configAttributes["bid_price"]; ok {
			if bidPrice != "" {
				config.BidPrice = aws.String(bidPrice.(string))
			}
		}

		if v, ok := configAttributes["bid_price_as_percentage_of_on_demand_price"].(float64); ok && v != 0 {
			config.BidPriceAsPercentageOfOnDemandPrice = aws.Float64(v)
		}

		if v, ok := configAttributes["weighted_capacity"].(int); ok {
			config.WeightedCapacity = aws.Int64(int64(v))
		}

		if v, ok := configAttributes["configurations"].(*schema.Set); ok && v.Len() > 0 {
			config.Configurations = expandConfigurations(v.List())
		}

		if v, ok := configAttributes["ebs_config"].(*schema.Set); ok && v.Len() == 1 {
			config.EbsConfiguration = expandEbsConfiguration(v.List())
		}

		configsOut = append(configsOut, config)
	}

	return configsOut
}

func expandLaunchSpecification(launchSpecification map[string]interface{}) *emr.InstanceFleetProvisioningSpecifications {
	onDemandSpecification := launchSpecification["on_demand_specification"].([]interface{})
	spotSpecification := launchSpecification["spot_specification"].([]interface{})

	fleetSpecification := &emr.InstanceFleetProvisioningSpecifications{}

	if len(onDemandSpecification) > 0 {
		fleetSpecification.OnDemandSpecification = &emr.OnDemandProvisioningSpecification{
			AllocationStrategy: aws.String(onDemandSpecification[0].(map[string]interface{})["allocation_strategy"].(string)),
		}
	}

	if len(spotSpecification) > 0 {
		configAttributes := spotSpecification[0].(map[string]interface{})
		spotProvisioning := &emr.SpotProvisioningSpecification{
			TimeoutAction:          aws.String(configAttributes["timeout_action"].(string)),
			TimeoutDurationMinutes: aws.Int64(int64(configAttributes["timeout_duration_minutes"].(int))),
		}
		if v, ok := configAttributes["block_duration_minutes"]; ok && v != 0 {
			spotProvisioning.BlockDurationMinutes = aws.Int64(int64(v.(int)))
		}
		if v, ok := configAttributes["allocation_strategy"]; ok {

			spotProvisioning.AllocationStrategy = aws.String(v.(string))
		}

		fleetSpecification.SpotSpecification = spotProvisioning
	}

	return fleetSpecification
}

func expandConfigurations(configurations []interface{}) []*emr.Configuration {
	configsOut := []*emr.Configuration{}

	for _, raw := range configurations {
		configAttributes := raw.(map[string]interface{})

		config := &emr.Configuration{}

		if v, ok := configAttributes["classification"].(string); ok {
			config.Classification = aws.String(v)
		}

		if v, ok := configAttributes["configurations"].([]interface{}); ok {
			config.Configurations = expandConfigurations(v)
		}

		if v, ok := configAttributes["properties"].(map[string]interface{}); ok {
			properties := make(map[string]string)
			for k, pv := range v {
				properties[k] = pv.(string)
			}
			config.Properties = aws.StringMap(properties)
		}

		configsOut = append(configsOut, config)
	}

	return configsOut
}

func resourceAwsEMRInstanceTypeConfigHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	if v, ok := m["bid_price"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["weighted_capacity"]; ok && v.(int) > 0 {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["bid_price_as_percentage_of_on_demand_price"]; ok && v.(float64) != 0 {
		buf.WriteString(fmt.Sprintf("%f-", v.(float64)))
	}
	return hashcode.String(buf.String())
}
