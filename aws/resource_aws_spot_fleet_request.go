package aws

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsSpotFleetRequest() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAwsSpotFleetRequestCreate,
		Read:   resourceAwsSpotFleetRequestRead,
		Delete: resourceAwsSpotFleetRequestDelete,
		Update: resourceAwsSpotFleetRequestUpdate,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("instance_pools_to_use_count", 1)
				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		SchemaVersion: 1,
		MigrateState:  resourceAwsSpotFleetRequestMigrateState,

		Schema: map[string]*schema.Schema{
			"iam_fleet_role": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"replace_unhealthy_instances": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"wait_for_fulfillment": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  false,
			},
			// http://docs.aws.amazon.com/sdk-for-go/api/service/ec2.html#type-SpotFleetLaunchSpecification
			// http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_SpotFleetLaunchSpecification.html
			"launch_specification": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vpc_security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"associate_public_ip_address": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"ebs_block_device": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delete_on_termination": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},
									"device_name": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"encrypted": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"iops": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"kms_key_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"snapshot_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"volume_size": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"volume_type": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
										ValidateFunc: validation.StringInSlice([]string{
											ec2.VolumeTypeStandard,
											ec2.VolumeTypeIo1,
											ec2.VolumeTypeGp2,
											ec2.VolumeTypeSc1,
											ec2.VolumeTypeSt1,
										}, false),
									},
								},
							},
							Set: hashEbsBlockDevice,
						},
						"ephemeral_block_device": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"device_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"virtual_name": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							Set: hashEphemeralBlockDevice,
						},
						"root_block_device": {
							// TODO: This is a set because we don't support singleton
							//       sub-resources today. We'll enforce that the set only ever has
							//       length zero or one below. When TF gains support for
							//       sub-resources this can be converted.
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								// "You can only modify the volume size, volume type, and Delete on
								// Termination flag on the block device mapping entry for the root
								// device volume." - bit.ly/ec2bdmap
								Schema: map[string]*schema.Schema{
									"delete_on_termination": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},
									"encrypted": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"iops": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"kms_key_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"volume_size": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"volume_type": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
										ValidateFunc: validation.StringInSlice([]string{
											ec2.VolumeTypeStandard,
											ec2.VolumeTypeIo1,
											ec2.VolumeTypeGp2,
											ec2.VolumeTypeSc1,
											ec2.VolumeTypeSt1,
										}, false),
									},
								},
							},
							Set: hashRootBlockDevice,
						},
						"ebs_optimized": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"iam_instance_profile": {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"iam_instance_profile_arn": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"ami": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"key_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Computed:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"monitoring": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"placement_group": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"placement_tenancy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								ec2.TenancyDefault,
								ec2.TenancyDedicated,
								ec2.TenancyHost,
							}, false),
						},
						"spot_price": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"user_data": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							StateFunc: func(v interface{}) string {
								switch v := v.(type) {
								case string:
									return userDataHashSum(v)
								default:
									return ""
								}
							},
						},
						"weighted_capacity": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"availability_zone": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"tags": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
				Set:          hashLaunchSpecification,
				ExactlyOneOf: []string{"launch_specification", "launch_template_config"},
			},
			"launch_template_config": {
				Type:         schema.TypeSet,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"launch_specification", "launch_template_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"launch_template_specification": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validateLaunchTemplateId,
									},
									"name": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validateLaunchTemplateName,
									},
									"version": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
								},
							},
						},
						"overrides": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"availability_zone": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"instance_type": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"spot_price": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"weighted_capacity": {
										Type:     schema.TypeFloat,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"priority": {
										Type:     schema.TypeFloat,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
								},
							},
							Set: hashLaunchTemplateOverrides,
						},
					},
				},
			},
			// Everything on a spot fleet is ForceNew except target_capacity and excess_capacity_termination_policy,
			// see https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#ModifySpotFleetRequestInput
			"target_capacity": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: false,
			},
			"allocation_strategy": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ec2.AllocationStrategyLowestPrice,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.AllocationStrategyLowestPrice,
					ec2.AllocationStrategyDiversified,
					ec2.AllocationStrategyCapacityOptimized,
				}, false),
			},
			"instance_pools_to_use_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
				ForceNew: true,
			},
			// Provided constants do not have the correct casing so going with hard-coded values.
			"excess_capacity_termination_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Default",
				ForceNew: false,
				ValidateFunc: validation.StringInSlice([]string{
					"Default",
					"NoTermination",
				}, false),
			},
			"instance_interruption_behaviour": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ec2.InstanceInterruptionBehaviorTerminate,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.InstanceInterruptionBehaviorTerminate,
					ec2.InstanceInterruptionBehaviorStop,
					ec2.InstanceInterruptionBehaviorHibernate,
				}, false),
			},
			"spot_price": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"terminate_instances_with_expiration": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"valid_from": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"valid_until": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"fleet_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ec2.FleetTypeMaintain,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.FleetTypeMaintain,
					ec2.FleetTypeRequest,
					ec2.FleetTypeInstant,
				}, false),
			},
			"spot_request_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"load_balancers": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"target_group_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
				Set: schema.HashString,
			},
			"tags": tagsSchema(),
		},
	}
}

func buildSpotFleetLaunchSpecification(d map[string]interface{}, meta interface{}) (*ec2.SpotFleetLaunchSpecification, error) {
	conn := meta.(*AWSClient).ec2conn

	opts := &ec2.SpotFleetLaunchSpecification{
		ImageId:      aws.String(d["ami"].(string)),
		InstanceType: aws.String(d["instance_type"].(string)),
		SpotPrice:    aws.String(d["spot_price"].(string)),
	}

	placement := new(ec2.SpotPlacement)
	if v, ok := d["availability_zone"]; ok {
		placement.AvailabilityZone = aws.String(v.(string))
		opts.Placement = placement
	}

	if v, ok := d["placement_tenancy"]; ok {
		placement.Tenancy = aws.String(v.(string))
		opts.Placement = placement
	}

	if v, ok := d["placement_group"]; ok {
		if v.(string) != "" {
			// If instanceInterruptionBehavior is set to STOP, this can't be set at all, even to an empty string, so check for "" to avoid those errors
			placement.GroupName = aws.String(v.(string))
			opts.Placement = placement
		}
	}

	if v, ok := d["ebs_optimized"]; ok {
		opts.EbsOptimized = aws.Bool(v.(bool))
	}

	if v, ok := d["monitoring"]; ok {
		opts.Monitoring = &ec2.SpotFleetMonitoring{
			Enabled: aws.Bool(v.(bool)),
		}
	}

	if v, ok := d["iam_instance_profile"]; ok {
		opts.IamInstanceProfile = &ec2.IamInstanceProfileSpecification{
			Name: aws.String(v.(string)),
		}
	}

	if v, ok := d["iam_instance_profile_arn"]; ok && v.(string) != "" {
		opts.IamInstanceProfile = &ec2.IamInstanceProfileSpecification{
			Arn: aws.String(v.(string)),
		}
	}

	if v, ok := d["user_data"]; ok {
		opts.UserData = aws.String(base64Encode([]byte(v.(string))))
	}

	if v, ok := d["key_name"]; ok && v != "" {
		opts.KeyName = aws.String(v.(string))
	}

	if v, ok := d["weighted_capacity"]; ok && v != "" {
		wc, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return nil, err
		}
		opts.WeightedCapacity = aws.Float64(wc)
	}

	var securityGroupIds []*string
	if v, ok := d["vpc_security_group_ids"]; ok {
		if s := v.(*schema.Set); s.Len() > 0 {
			for _, v := range s.List() {
				securityGroupIds = append(securityGroupIds, aws.String(v.(string)))
			}
		}
	}

	if m, ok := d["tags"].(map[string]interface{}); ok && len(m) > 0 {
		tagsSpec := make([]*ec2.SpotFleetTagSpecification, 0)

		tags := keyvaluetags.New(m).IgnoreAws().Ec2Tags()

		spec := &ec2.SpotFleetTagSpecification{
			ResourceType: aws.String(ec2.ResourceTypeInstance),
			Tags:         tags,
		}

		tagsSpec = append(tagsSpec, spec)

		opts.TagSpecifications = tagsSpec
	}

	subnetId, hasSubnetId := d["subnet_id"]
	if hasSubnetId {
		opts.SubnetId = aws.String(subnetId.(string))
	}

	associatePublicIpAddress, hasPublicIpAddress := d["associate_public_ip_address"]
	if hasPublicIpAddress && associatePublicIpAddress.(bool) && hasSubnetId {

		// If we have a non-default VPC / Subnet specified, we can flag
		// AssociatePublicIpAddress to get a Public IP assigned. By default these are not provided.
		// You cannot specify both SubnetId and the NetworkInterface.0.* parameters though, otherwise
		// you get: Network interfaces and an instance-level subnet ID may not be specified on the same request
		// You also need to attach Security Groups to the NetworkInterface instead of the instance,
		// to avoid: Network interfaces and an instance-level security groups may not be specified on
		// the same request
		ni := &ec2.InstanceNetworkInterfaceSpecification{
			AssociatePublicIpAddress: aws.Bool(true),
			DeleteOnTermination:      aws.Bool(true),
			DeviceIndex:              aws.Int64(int64(0)),
			SubnetId:                 aws.String(subnetId.(string)),
			Groups:                   securityGroupIds,
		}

		opts.NetworkInterfaces = []*ec2.InstanceNetworkInterfaceSpecification{ni}
		opts.SubnetId = aws.String("")
	} else {
		for _, id := range securityGroupIds {
			opts.SecurityGroups = append(opts.SecurityGroups, &ec2.GroupIdentifier{GroupId: id})
		}
	}

	blockDevices, err := readSpotFleetBlockDeviceMappingsFromConfig(d, conn)
	if err != nil {
		return nil, err
	}
	if len(blockDevices) > 0 {
		opts.BlockDeviceMappings = blockDevices
	}

	return opts, nil
}

func readSpotFleetBlockDeviceMappingsFromConfig(
	d map[string]interface{}, conn *ec2.EC2) ([]*ec2.BlockDeviceMapping, error) {
	blockDevices := make([]*ec2.BlockDeviceMapping, 0)

	if v, ok := d["ebs_block_device"]; ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			bd := v.(map[string]interface{})
			ebs := &ec2.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(bd["delete_on_termination"].(bool)),
			}

			if v, ok := bd["snapshot_id"].(string); ok && v != "" {
				ebs.SnapshotId = aws.String(v)
			}

			if v, ok := bd["encrypted"].(bool); ok && v {
				ebs.Encrypted = aws.Bool(v)
			}

			if v, ok := bd["kms_key_id"].(string); ok && v != "" {
				ebs.KmsKeyId = aws.String(v)
			}

			if v, ok := bd["volume_size"].(int); ok && v != 0 {
				ebs.VolumeSize = aws.Int64(int64(v))
			}

			if v, ok := bd["volume_type"].(string); ok && v != "" {
				ebs.VolumeType = aws.String(v)
			}

			if v, ok := bd["iops"].(int); ok && v > 0 {
				ebs.Iops = aws.Int64(int64(v))
			}

			blockDevices = append(blockDevices, &ec2.BlockDeviceMapping{
				DeviceName: aws.String(bd["device_name"].(string)),
				Ebs:        ebs,
			})
		}
	}

	if v, ok := d["ephemeral_block_device"]; ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			bd := v.(map[string]interface{})
			blockDevices = append(blockDevices, &ec2.BlockDeviceMapping{
				DeviceName:  aws.String(bd["device_name"].(string)),
				VirtualName: aws.String(bd["virtual_name"].(string)),
			})
		}
	}

	if v, ok := d["root_block_device"]; ok {
		vL := v.(*schema.Set).List()
		if len(vL) > 1 {
			return nil, fmt.Errorf("Cannot specify more than one root_block_device.")
		}
		for _, v := range vL {
			bd := v.(map[string]interface{})
			ebs := &ec2.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(bd["delete_on_termination"].(bool)),
			}

			if v, ok := bd["encrypted"].(bool); ok && v {
				ebs.Encrypted = aws.Bool(v)
			}

			if v, ok := bd["kms_key_id"].(string); ok && v != "" {
				ebs.KmsKeyId = aws.String(v)
			}

			if v, ok := bd["volume_size"].(int); ok && v != 0 {
				ebs.VolumeSize = aws.Int64(int64(v))
			}

			if v, ok := bd["volume_type"].(string); ok && v != "" {
				ebs.VolumeType = aws.String(v)
			}

			if v, ok := bd["iops"].(int); ok && v > 0 {
				ebs.Iops = aws.Int64(int64(v))
			}

			if dn, err := fetchRootDeviceName(d["ami"].(string), conn); err == nil {
				if dn == nil {
					return nil, fmt.Errorf(
						"Expected 1 AMI for ID: %s, got none",
						d["ami"].(string))
				}

				blockDevices = append(blockDevices, &ec2.BlockDeviceMapping{
					DeviceName: dn,
					Ebs:        ebs,
				})
			} else {
				return nil, err
			}
		}
	}

	return blockDevices, nil
}

func buildAwsSpotFleetLaunchSpecifications(
	d *schema.ResourceData, meta interface{}) ([]*ec2.SpotFleetLaunchSpecification, error) {

	userSpecs := d.Get("launch_specification").(*schema.Set).List()
	specs := make([]*ec2.SpotFleetLaunchSpecification, len(userSpecs))
	for i, userSpec := range userSpecs {
		userSpecMap := userSpec.(map[string]interface{})
		// panic: interface conversion: interface {} is map[string]interface {}, not *schema.ResourceData
		opts, err := buildSpotFleetLaunchSpecification(userSpecMap, meta)
		if err != nil {
			return nil, err
		}
		specs[i] = opts
	}

	return specs, nil
}

func buildLaunchTemplateConfigs(d *schema.ResourceData) []*ec2.LaunchTemplateConfig {
	launchTemplateConfigs := d.Get("launch_template_config").(*schema.Set)
	configs := make([]*ec2.LaunchTemplateConfig, 0)

	for _, launchTemplateConfig := range launchTemplateConfigs.List() {

		ltc := &ec2.LaunchTemplateConfig{}

		ltcMap := launchTemplateConfig.(map[string]interface{})

		//launch template spec
		if v, ok := ltcMap["launch_template_specification"]; ok {
			vL := v.([]interface{})
			lts := vL[0].(map[string]interface{})

			flts := &ec2.FleetLaunchTemplateSpecification{}

			if v, ok := lts["id"].(string); ok && v != "" {
				flts.LaunchTemplateId = aws.String(v)
			}

			if v, ok := lts["name"].(string); ok && v != "" {
				flts.LaunchTemplateName = aws.String(v)
			}

			if v, ok := lts["version"].(string); ok && v != "" {
				flts.Version = aws.String(v)
			}

			ltc.LaunchTemplateSpecification = flts

		}

		if v, ok := ltcMap["overrides"]; ok && v.(*schema.Set).Len() > 0 {
			vL := v.(*schema.Set).List()
			overrides := make([]*ec2.LaunchTemplateOverrides, 0)

			for _, v := range vL {
				ors := v.(map[string]interface{})
				lto := &ec2.LaunchTemplateOverrides{}

				if v, ok := ors["availability_zone"].(string); ok && v != "" {
					lto.AvailabilityZone = aws.String(v)
				}

				if v, ok := ors["instance_type"].(string); ok && v != "" {
					lto.InstanceType = aws.String(v)
				}

				if v, ok := ors["spot_price"].(string); ok && v != "" {
					lto.SpotPrice = aws.String(v)
				}

				if v, ok := ors["subnet_id"].(string); ok && v != "" {
					lto.SubnetId = aws.String(v)
				}

				if v, ok := ors["weighted_capacity"].(float64); ok && v > 0 {
					lto.WeightedCapacity = aws.Float64(v)
				}

				if v, ok := ors["priority"].(float64); ok {
					lto.Priority = aws.Float64(v)
				}

				overrides = append(overrides, lto)
			}

			ltc.Overrides = overrides
		}

		configs = append(configs, ltc)
	}

	return configs
}

func resourceAwsSpotFleetRequestCreate(d *schema.ResourceData, meta interface{}) error {
	// http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_RequestSpotFleet.html
	conn := meta.(*AWSClient).ec2conn

	_, launchSpecificationOk := d.GetOk("launch_specification")
	_, launchTemplateConfigsOk := d.GetOk("launch_template_config")

	// http://docs.aws.amazon.com/sdk-for-go/api/service/ec2.html#type-SpotFleetRequestConfigData
	spotFleetConfig := &ec2.SpotFleetRequestConfigData{
		IamFleetRole:                     aws.String(d.Get("iam_fleet_role").(string)),
		TargetCapacity:                   aws.Int64(int64(d.Get("target_capacity").(int))),
		ClientToken:                      aws.String(resource.UniqueId()),
		TerminateInstancesWithExpiration: aws.Bool(d.Get("terminate_instances_with_expiration").(bool)),
		ReplaceUnhealthyInstances:        aws.Bool(d.Get("replace_unhealthy_instances").(bool)),
		InstanceInterruptionBehavior:     aws.String(d.Get("instance_interruption_behaviour").(string)),
		Type:                             aws.String(d.Get("fleet_type").(string)),
		TagSpecifications:                ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), ec2.ResourceTypeSpotFleetRequest),
	}

	if launchSpecificationOk {
		launchSpecs, err := buildAwsSpotFleetLaunchSpecifications(d, meta)
		if err != nil {
			return err
		}
		spotFleetConfig.LaunchSpecifications = launchSpecs
	}

	if launchTemplateConfigsOk {
		launchTemplates := buildLaunchTemplateConfigs(d)
		spotFleetConfig.LaunchTemplateConfigs = launchTemplates
	}

	if v, ok := d.GetOk("excess_capacity_termination_policy"); ok {
		spotFleetConfig.ExcessCapacityTerminationPolicy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("allocation_strategy"); ok {
		spotFleetConfig.AllocationStrategy = aws.String(v.(string))
	} else {
		spotFleetConfig.AllocationStrategy = aws.String("lowestPrice")
	}

	if v, ok := d.GetOk("instance_pools_to_use_count"); ok && v.(int) != 1 {
		spotFleetConfig.InstancePoolsToUseCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("spot_price"); ok && v.(string) != "" {
		spotFleetConfig.SpotPrice = aws.String(v.(string))
	}

	if v, ok := d.GetOk("valid_from"); ok {
		validFrom, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return err
		}
		spotFleetConfig.ValidFrom = aws.Time(validFrom)
	}

	if v, ok := d.GetOk("valid_until"); ok {
		validUntil, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return err
		}
		spotFleetConfig.ValidUntil = aws.Time(validUntil)
	} else {
		validUntil := time.Now().Add(24 * time.Hour)
		spotFleetConfig.ValidUntil = aws.Time(validUntil)
	}

	if v, ok := d.GetOk("load_balancers"); ok && v.(*schema.Set).Len() > 0 {
		var elbNames []*ec2.ClassicLoadBalancer
		for _, v := range v.(*schema.Set).List() {
			elbNames = append(elbNames, &ec2.ClassicLoadBalancer{
				Name: aws.String(v.(string)),
			})
		}
		if spotFleetConfig.LoadBalancersConfig == nil {
			spotFleetConfig.LoadBalancersConfig = &ec2.LoadBalancersConfig{}
		}
		spotFleetConfig.LoadBalancersConfig.ClassicLoadBalancersConfig = &ec2.ClassicLoadBalancersConfig{
			ClassicLoadBalancers: elbNames,
		}
	}

	if v, ok := d.GetOk("target_group_arns"); ok && v.(*schema.Set).Len() > 0 {
		var targetGroups []*ec2.TargetGroup
		for _, v := range v.(*schema.Set).List() {
			targetGroups = append(targetGroups, &ec2.TargetGroup{
				Arn: aws.String(v.(string)),
			})
		}
		if spotFleetConfig.LoadBalancersConfig == nil {
			spotFleetConfig.LoadBalancersConfig = &ec2.LoadBalancersConfig{}
		}
		spotFleetConfig.LoadBalancersConfig.TargetGroupsConfig = &ec2.TargetGroupsConfig{
			TargetGroups: targetGroups,
		}
	}

	// http://docs.aws.amazon.com/sdk-for-go/api/service/ec2.html#type-RequestSpotFleetInput
	spotFleetOpts := &ec2.RequestSpotFleetInput{
		SpotFleetRequestConfig: spotFleetConfig,
		DryRun:                 aws.Bool(false),
	}

	log.Printf("[DEBUG] Requesting spot fleet with these opts: %+v", spotFleetOpts)

	// Since IAM is eventually consistent, we retry creation as a newly created role may not
	// take effect immediately, resulting in an InvalidSpotFleetRequestConfig error
	var resp *ec2.RequestSpotFleetOutput
	err := resource.Retry(10*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.RequestSpotFleet(spotFleetOpts)

		if isAWSErr(err, "InvalidSpotFleetRequestConfig", "Duplicate: Parameter combination") {
			return resource.NonRetryableError(fmt.Errorf("Error creating Spot fleet request: %s", err))
		}
		if isAWSErr(err, "InvalidSpotFleetRequestConfig", "") {
			return resource.RetryableError(fmt.Errorf("Error creating Spot fleet request, retrying: %s", err))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		resp, err = conn.RequestSpotFleet(spotFleetOpts)
	}

	if err != nil {
		return fmt.Errorf("Error requesting spot fleet: %s", err)
	}

	d.SetId(*resp.SpotFleetRequestId)

	log.Printf("[INFO] Spot Fleet Request ID: %s", d.Id())
	log.Println("[INFO] Waiting for Spot Fleet Request to be active")
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.BatchStateSubmitted},
		Target:     []string{ec2.BatchStateActive},
		Refresh:    resourceAwsSpotFleetRequestStateRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutCreate), //10 * time.Minute,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	if d.Get("wait_for_fulfillment").(bool) {
		log.Println("[INFO] Waiting for Spot Fleet Request to be fulfilled")
		spotStateConf := &resource.StateChangeConf{
			Pending:    []string{ec2.ActivityStatusPendingFulfillment},
			Target:     []string{ec2.ActivityStatusFulfilled},
			Refresh:    resourceAwsSpotFleetRequestFulfillmentRefreshFunc(d.Id(), meta.(*AWSClient).ec2conn),
			Timeout:    d.Timeout(schema.TimeoutCreate),
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = spotStateConf.WaitForState()

		if err != nil {
			return err
		}
	}

	return resourceAwsSpotFleetRequestRead(d, meta)
}

func resourceAwsSpotFleetRequestStateRefreshFunc(d *schema.ResourceData, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		conn := meta.(*AWSClient).ec2conn
		req := &ec2.DescribeSpotFleetRequestsInput{
			SpotFleetRequestIds: []*string{aws.String(d.Id())},
		}
		resp, err := conn.DescribeSpotFleetRequests(req)

		if err != nil {
			log.Printf("Error on retrieving Spot Fleet Request when waiting: %s", err)
			return nil, "", nil
		}

		if resp == nil {
			return nil, "", nil
		}

		if len(resp.SpotFleetRequestConfigs) == 0 {
			return nil, "", nil
		}

		spotFleetRequest := resp.SpotFleetRequestConfigs[0]

		return spotFleetRequest, *spotFleetRequest.SpotFleetRequestState, nil
	}
}

func resourceAwsSpotFleetRequestFulfillmentRefreshFunc(id string, conn *ec2.EC2) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		req := &ec2.DescribeSpotFleetRequestsInput{
			SpotFleetRequestIds: []*string{aws.String(id)},
		}
		resp, err := conn.DescribeSpotFleetRequests(req)

		if err != nil {
			log.Printf("Error on retrieving Spot Fleet Request when waiting: %s", err)
			return nil, "", nil
		}

		if resp == nil {
			return nil, "", nil
		}

		if len(resp.SpotFleetRequestConfigs) == 0 {
			return nil, "", nil
		}

		cfg := resp.SpotFleetRequestConfigs[0]
		status := *cfg.ActivityStatus

		var fleetError error
		if status == ec2.ActivityStatusError {
			var events []*ec2.HistoryRecord

			// Query "information" events (e.g. launchSpecUnusable b/c low bid price)
			out, err := conn.DescribeSpotFleetRequestHistory(&ec2.DescribeSpotFleetRequestHistoryInput{
				EventType:          aws.String(ec2.EventTypeInformation),
				SpotFleetRequestId: aws.String(id),
				StartTime:          cfg.CreateTime,
			})
			if err != nil {
				log.Printf("[ERROR] Failed to get the reason of 'error' state: %s", err)
			}
			if len(out.HistoryRecords) > 0 {
				events = out.HistoryRecords
			}

			out, err = conn.DescribeSpotFleetRequestHistory(&ec2.DescribeSpotFleetRequestHistoryInput{
				EventType:          aws.String(ec2.EventTypeError),
				SpotFleetRequestId: aws.String(id),
				StartTime:          cfg.CreateTime,
			})
			if err != nil {
				log.Printf("[ERROR] Failed to get the reason of 'error' state: %s", err)
			}
			if len(out.HistoryRecords) > 0 {
				events = append(events, out.HistoryRecords...)
			}

			if len(events) > 0 {
				fleetError = fmt.Errorf("Last events: %v", events)
			}
		}

		return cfg, status, fleetError
	}
}

func resourceAwsSpotFleetRequestRead(d *schema.ResourceData, meta interface{}) error {
	// http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSpotFleetRequests.html
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeSpotFleetRequestsInput{
		SpotFleetRequestIds: []*string{aws.String(d.Id())},
	}
	resp, err := conn.DescribeSpotFleetRequests(req)

	if err != nil {
		// If the spot request was not found, return nil so that we can show
		// that it is gone.
		if isAWSErr(err, "InvalidSpotFleetRequestId.NotFound", "") {
			d.SetId("")
			return nil
		}

		// Some other error, report it
		return err
	}

	sfr := resp.SpotFleetRequestConfigs[0]

	// if the request is cancelled, then it is gone
	cancelledStates := map[string]bool{
		ec2.BatchStateCancelled:            true,
		ec2.BatchStateCancelledRunning:     true,
		ec2.BatchStateCancelledTerminating: true,
	}
	if _, ok := cancelledStates[*sfr.SpotFleetRequestState]; ok {
		d.SetId("")
		return nil
	}

	d.SetId(*sfr.SpotFleetRequestId)
	d.Set("spot_request_state", aws.StringValue(sfr.SpotFleetRequestState))

	config := sfr.SpotFleetRequestConfig

	if config.AllocationStrategy != nil {
		d.Set("allocation_strategy", aws.StringValue(config.AllocationStrategy))
	}

	if config.InstancePoolsToUseCount != nil {
		d.Set("instance_pools_to_use_count", aws.Int64Value(config.InstancePoolsToUseCount))
	}

	if config.ClientToken != nil {
		d.Set("client_token", aws.StringValue(config.ClientToken))
	}

	if config.ExcessCapacityTerminationPolicy != nil {
		d.Set("excess_capacity_termination_policy",
			aws.StringValue(config.ExcessCapacityTerminationPolicy))
	}

	if config.IamFleetRole != nil {
		d.Set("iam_fleet_role", aws.StringValue(config.IamFleetRole))
	}

	if config.SpotPrice != nil {
		d.Set("spot_price", aws.StringValue(config.SpotPrice))
	}

	if config.TargetCapacity != nil {
		d.Set("target_capacity", aws.Int64Value(config.TargetCapacity))
	}

	if config.TerminateInstancesWithExpiration != nil {
		d.Set("terminate_instances_with_expiration",
			aws.BoolValue(config.TerminateInstancesWithExpiration))
	}

	if config.ValidFrom != nil {
		d.Set("valid_from",
			aws.TimeValue(config.ValidFrom).Format(time.RFC3339))
	}

	if config.ValidUntil != nil {
		d.Set("valid_until",
			aws.TimeValue(config.ValidUntil).Format(time.RFC3339))
	}

	launchSpec, err := launchSpecsToSet(config.LaunchSpecifications, conn)
	if err != nil {
		return fmt.Errorf("error occurred while reading launch specification: %s", err)
	}

	d.Set("replace_unhealthy_instances", config.ReplaceUnhealthyInstances)
	d.Set("instance_interruption_behaviour", config.InstanceInterruptionBehavior)
	d.Set("fleet_type", config.Type)
	d.Set("launch_specification", launchSpec)
	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(sfr.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if len(config.LaunchTemplateConfigs) > 0 {
		if err := d.Set("launch_template_config", flattenFleetLaunchTemplateConfig(config.LaunchTemplateConfigs)); err != nil {
			return fmt.Errorf("error setting launch_template_config: %s", err)
		}
	}

	if config.LoadBalancersConfig != nil {
		lbConf := config.LoadBalancersConfig

		if lbConf.ClassicLoadBalancersConfig != nil {
			flatLbs := make([]*string, 0)
			for _, lb := range lbConf.ClassicLoadBalancersConfig.ClassicLoadBalancers {
				flatLbs = append(flatLbs, lb.Name)
			}
			if err := d.Set("load_balancers", flattenStringSet(flatLbs)); err != nil {
				return fmt.Errorf("error setting load_balancers: %s", err)
			}
		}

		if lbConf.TargetGroupsConfig != nil {
			flatTgs := make([]*string, 0)
			for _, tg := range lbConf.TargetGroupsConfig.TargetGroups {
				flatTgs = append(flatTgs, tg.Arn)
			}
			if err := d.Set("target_group_arns", flattenStringSet(flatTgs)); err != nil {
				return fmt.Errorf("error setting target_group_arns: %s", err)
			}
		}
	}

	return nil
}

func flattenSpotFleetRequestLaunchTemplateOverrides(override *ec2.LaunchTemplateOverrides) map[string]interface{} {
	m := make(map[string]interface{})

	if override.AvailabilityZone != nil {
		m["availability_zone"] = aws.StringValue(override.AvailabilityZone)
	}
	if override.InstanceType != nil {
		m["instance_type"] = aws.StringValue(override.InstanceType)
	}

	if override.SpotPrice != nil {
		m["spot_price"] = aws.StringValue(override.SpotPrice)
	}

	if override.SubnetId != nil {
		m["subnet_id"] = aws.StringValue(override.SubnetId)
	}

	if override.WeightedCapacity != nil {
		m["weighted_capacity"] = aws.Float64Value(override.WeightedCapacity)
	}

	if override.Priority != nil {
		m["priority"] = aws.Float64Value(override.Priority)
	}

	return m
}

func launchSpecsToSet(launchSpecs []*ec2.SpotFleetLaunchSpecification, conn *ec2.EC2) (*schema.Set, error) {
	specSet := &schema.Set{F: hashLaunchSpecification}
	for _, spec := range launchSpecs {
		rootDeviceName, err := fetchRootDeviceName(aws.StringValue(spec.ImageId), conn)
		if err != nil {
			return nil, err
		}

		specSet.Add(launchSpecToMap(spec, rootDeviceName))
	}
	return specSet, nil
}

func launchSpecToMap(l *ec2.SpotFleetLaunchSpecification, rootDevName *string) map[string]interface{} {
	m := make(map[string]interface{})

	m["root_block_device"] = rootBlockDeviceToSet(l.BlockDeviceMappings, rootDevName)
	m["ebs_block_device"] = ebsBlockDevicesToSet(l.BlockDeviceMappings, rootDevName)
	m["ephemeral_block_device"] = ephemeralBlockDevicesToSet(l.BlockDeviceMappings)

	if l.ImageId != nil {
		m["ami"] = aws.StringValue(l.ImageId)
	}

	if l.InstanceType != nil {
		m["instance_type"] = aws.StringValue(l.InstanceType)
	}

	if l.SpotPrice != nil {
		m["spot_price"] = aws.StringValue(l.SpotPrice)
	}

	if l.EbsOptimized != nil {
		m["ebs_optimized"] = aws.BoolValue(l.EbsOptimized)
	}

	if l.Monitoring != nil && l.Monitoring.Enabled != nil {
		m["monitoring"] = aws.BoolValue(l.Monitoring.Enabled)
	}

	if l.IamInstanceProfile != nil && l.IamInstanceProfile.Name != nil {
		m["iam_instance_profile"] = aws.StringValue(l.IamInstanceProfile.Name)
	}

	if l.IamInstanceProfile != nil && l.IamInstanceProfile.Arn != nil {
		m["iam_instance_profile_arn"] = aws.StringValue(l.IamInstanceProfile.Arn)
	}

	if l.UserData != nil {
		m["user_data"] = userDataHashSum(aws.StringValue(l.UserData))
	}

	if l.KeyName != nil {
		m["key_name"] = aws.StringValue(l.KeyName)
	}

	if l.Placement != nil {
		m["availability_zone"] = aws.StringValue(l.Placement.AvailabilityZone)
	}

	if l.SubnetId != nil {
		m["subnet_id"] = aws.StringValue(l.SubnetId)
	}

	securityGroupIds := &schema.Set{F: schema.HashString}
	if len(l.NetworkInterfaces) > 0 {
		m["associate_public_ip_address"] = aws.BoolValue(l.NetworkInterfaces[0].AssociatePublicIpAddress)
		m["subnet_id"] = aws.StringValue(l.NetworkInterfaces[0].SubnetId)

		for _, group := range l.NetworkInterfaces[0].Groups {
			securityGroupIds.Add(aws.StringValue(group))
		}
	} else {
		for _, group := range l.SecurityGroups {
			securityGroupIds.Add(aws.StringValue(group.GroupId))
		}
	}
	m["vpc_security_group_ids"] = securityGroupIds

	if l.WeightedCapacity != nil {
		m["weighted_capacity"] = strconv.FormatFloat(*l.WeightedCapacity, 'f', 0, 64)
	}

	if l.TagSpecifications != nil {
		for _, tagSpecs := range l.TagSpecifications {
			// only "instance" tags are currently supported: http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_SpotFleetTagSpecification.html
			if aws.StringValue(tagSpecs.ResourceType) == ec2.ResourceTypeInstance {
				m["tags"] = keyvaluetags.Ec2KeyValueTags(tagSpecs.Tags).IgnoreAws().Map()
			}
		}
	}

	return m
}

func ebsBlockDevicesToSet(bdm []*ec2.BlockDeviceMapping, rootDevName *string) *schema.Set {
	set := &schema.Set{F: hashEbsBlockDevice}

	for _, val := range bdm {
		if val.Ebs != nil {
			m := make(map[string]interface{})

			ebs := val.Ebs

			if val.DeviceName != nil {
				if aws.StringValue(rootDevName) == aws.StringValue(val.DeviceName) {
					continue
				}

				m["device_name"] = aws.StringValue(val.DeviceName)
			}

			if ebs.DeleteOnTermination != nil {
				m["delete_on_termination"] = aws.BoolValue(ebs.DeleteOnTermination)
			}

			if ebs.SnapshotId != nil {
				m["snapshot_id"] = aws.StringValue(ebs.SnapshotId)
			}

			if ebs.Encrypted != nil {
				m["encrypted"] = aws.BoolValue(ebs.Encrypted)
			}

			if ebs.KmsKeyId != nil {
				m["kms_key_id"] = aws.StringValue(ebs.KmsKeyId)
			}

			if ebs.VolumeSize != nil {
				m["volume_size"] = aws.Int64Value(ebs.VolumeSize)
			}

			if ebs.VolumeType != nil {
				m["volume_type"] = aws.StringValue(ebs.VolumeType)
			}

			if ebs.Iops != nil {
				m["iops"] = aws.Int64Value(ebs.Iops)
			}

			set.Add(m)
		}
	}

	return set
}

func ephemeralBlockDevicesToSet(bdm []*ec2.BlockDeviceMapping) *schema.Set {
	set := &schema.Set{F: hashEphemeralBlockDevice}

	for _, val := range bdm {
		if val.VirtualName != nil {
			m := make(map[string]interface{})
			m["virtual_name"] = aws.StringValue(val.VirtualName)

			if val.DeviceName != nil {
				m["device_name"] = aws.StringValue(val.DeviceName)
			}

			set.Add(m)
		}
	}

	return set
}

func rootBlockDeviceToSet(
	bdm []*ec2.BlockDeviceMapping,
	rootDevName *string,
) *schema.Set {
	set := &schema.Set{F: hashRootBlockDevice}

	if rootDevName != nil {
		for _, val := range bdm {
			if aws.StringValue(val.DeviceName) == aws.StringValue(rootDevName) {
				m := make(map[string]interface{})
				if val.Ebs.DeleteOnTermination != nil {
					m["delete_on_termination"] = aws.BoolValue(val.Ebs.DeleteOnTermination)
				}

				if val.Ebs.Encrypted != nil {
					m["encrypted"] = aws.BoolValue(val.Ebs.Encrypted)
				}

				if val.Ebs.KmsKeyId != nil {
					m["kms_key_id"] = aws.StringValue(val.Ebs.KmsKeyId)
				}

				if val.Ebs.VolumeSize != nil {
					m["volume_size"] = aws.Int64Value(val.Ebs.VolumeSize)
				}

				if val.Ebs.VolumeType != nil {
					m["volume_type"] = aws.StringValue(val.Ebs.VolumeType)
				}

				if val.Ebs.Iops != nil {
					m["iops"] = aws.Int64Value(val.Ebs.Iops)
				}

				set.Add(m)
			}
		}
	}

	return set
}

func resourceAwsSpotFleetRequestUpdate(d *schema.ResourceData, meta interface{}) error {
	// http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifySpotFleetRequest.html
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.ModifySpotFleetRequestInput{
		SpotFleetRequestId: aws.String(d.Id()),
	}

	updateFlag := false

	if d.HasChange("target_capacity") {
		if val, ok := d.GetOk("target_capacity"); ok {
			req.TargetCapacity = aws.Int64(int64(val.(int)))
		}

		updateFlag = true
	}

	if d.HasChange("excess_capacity_termination_policy") {
		if val, ok := d.GetOk("excess_capacity_termination_policy"); ok {
			req.ExcessCapacityTerminationPolicy = aws.String(val.(string))
		}

		updateFlag = true
	}

	if updateFlag {
		if _, err := conn.ModifySpotFleetRequest(req); err != nil {
			return fmt.Errorf("error updating spot request (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsSpotFleetRequestRead(d, meta)
}

func resourceAwsSpotFleetRequestDelete(d *schema.ResourceData, meta interface{}) error {
	// http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CancelSpotFleetRequests.html
	conn := meta.(*AWSClient).ec2conn
	terminateInstances := d.Get("terminate_instances_with_expiration").(bool)

	log.Printf("[INFO] Cancelling spot fleet request: %s", d.Id())
	err := deleteSpotFleetRequest(d.Id(), terminateInstances, d.Timeout(schema.TimeoutDelete), conn)
	if err != nil {
		return fmt.Errorf("error deleting spot request (%s): %s", d.Id(), err)
	}

	return nil
}

func deleteSpotFleetRequest(spotFleetRequestID string, terminateInstances bool, timeout time.Duration, conn *ec2.EC2) error {
	_, err := conn.CancelSpotFleetRequests(&ec2.CancelSpotFleetRequestsInput{
		SpotFleetRequestIds: []*string{aws.String(spotFleetRequestID)},
		TerminateInstances:  aws.Bool(terminateInstances),
	})
	if err != nil {
		return err
	}

	// Only wait for instance termination if requested
	if !terminateInstances {
		return nil
	}

	activeInstances := func(fleetRequestID string) (int, error) {
		resp, err := conn.DescribeSpotFleetInstances(&ec2.DescribeSpotFleetInstancesInput{
			SpotFleetRequestId: aws.String(fleetRequestID),
		})

		if err != nil || resp == nil {
			return 0, fmt.Errorf("error reading Spot Fleet Instances (%s): %s", spotFleetRequestID, err)
		}

		return len(resp.ActiveInstances), nil
	}

	err = resource.Retry(timeout, func() *resource.RetryError {
		n, err := activeInstances(spotFleetRequestID)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if n > 0 {
			log.Printf("[DEBUG] Active instance count in Spot Fleet Request (%s): %d", spotFleetRequestID, n)
			return resource.RetryableError(fmt.Errorf("fleet still has (%d) running instances", n))
		}

		log.Printf("[DEBUG] Active instance count is 0 for Spot Fleet Request (%s), removing", spotFleetRequestID)
		return nil
	})

	if isResourceTimeoutError(err) {
		n, err := activeInstances(spotFleetRequestID)
		if err != nil {
			return err
		}

		if n > 0 {
			log.Printf("[DEBUG] Active instance count in Spot Fleet Request (%s): %d", spotFleetRequestID, n)
			return fmt.Errorf("fleet still has (%d) running instances", n)
		}
	}

	if err != nil {
		return fmt.Errorf("error reading Spot Fleet Instances (%s): %s", spotFleetRequestID, err)
	}

	return nil
}

func hashEphemeralBlockDevice(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["device_name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["virtual_name"].(string)))
	return hashcode.String(buf.String())
}

func hashRootBlockDevice(v interface{}) int {
	// there can be only one root device; no need to hash anything
	return 0
}

func hashLaunchSpecification(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["ami"].(string)))
	if m["availability_zone"] != "" {
		buf.WriteString(fmt.Sprintf("%s-", m["availability_zone"].(string)))
	}
	if m["subnet_id"] != "" {
		buf.WriteString(fmt.Sprintf("%s-", m["subnet_id"].(string)))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["spot_price"].(string)))
	return hashcode.String(buf.String())
}

func hashLaunchTemplateOverrides(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if m["availability_zone"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["availability_zone"].(string)))
	}
	if m["subnet_id"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["subnet_id"].(string)))
	}
	if m["spot_price"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["spot_price"].(string)))
	}
	if m["instance_type"] != nil {
		buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	}
	if m["weighted_capacity"] != nil {
		buf.WriteString(fmt.Sprintf("%f-", m["weighted_capacity"].(float64)))
	}
	if m["priority"] != nil {
		buf.WriteString(fmt.Sprintf("%f-", m["priority"].(float64)))
	}

	return hashcode.String(buf.String())
}

func hashEbsBlockDevice(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if name, ok := m["device_name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", name.(string)))
	}
	if id, ok := m["snapshot_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", id.(string)))
	}
	return hashcode.String(buf.String())
}

func flattenFleetLaunchTemplateConfig(ltcs []*ec2.LaunchTemplateConfig) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, ltc := range ltcs {
		ltcRes := map[string]interface{}{}

		if ltc.LaunchTemplateSpecification != nil {
			ltcRes["launch_template_specification"] = flattenFleetLaunchTemplateSpecification(ltc.LaunchTemplateSpecification)
		}

		if ltc.Overrides != nil {
			ltcRes["overrides"] = flattenLaunchTemplateOverrides(ltc.Overrides)
		}

		result = append(result, ltcRes)
	}

	return result
}

func flattenFleetLaunchTemplateSpecification(flt *ec2.FleetLaunchTemplateSpecification) []map[string]interface{} {
	attrs := map[string]interface{}{}
	result := make([]map[string]interface{}, 0)

	// unlike autoscaling.LaunchTemplateConfiguration, FleetLaunchTemplateSpecs only return what was set
	if flt.LaunchTemplateId != nil {
		attrs["id"] = aws.StringValue(flt.LaunchTemplateId)
	}

	if flt.LaunchTemplateName != nil {
		attrs["name"] = aws.StringValue(flt.LaunchTemplateName)
	}

	// version is returned only if it was previously set
	if flt.Version != nil {
		attrs["version"] = aws.StringValue(flt.Version)
	} else {
		attrs["version"] = nil
	}

	result = append(result, attrs)

	return result
}

func flattenLaunchTemplateOverrides(overrides []*ec2.LaunchTemplateOverrides) *schema.Set {
	overrideSet := &schema.Set{F: hashLaunchTemplateOverrides}
	for _, override := range overrides {
		overrideSet.Add(flattenSpotFleetRequestLaunchTemplateOverrides(override))
	}
	return overrideSet
}
