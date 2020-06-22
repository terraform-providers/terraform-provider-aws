package aws

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsInstance() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsInstanceRead,

		Schema: map[string]*schema.Schema{
			"filter":        dataSourceFiltersSchema(),
			"tags":          tagsSchemaComputed(),
			"instance_tags": tagsSchemaComputed(),
			"instance_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ami": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"placement_group": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tenancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"get_password_data": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"password_data": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_dns": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iam_instance_profile": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"associate_public_ip_address": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ebs_optimized": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"source_dest_check": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"monitoring": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"get_user_data": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"user_data": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_data_base64": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vpc_security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ephemeral_block_device": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"virtual_name": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"no_device": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"ebs_block_device": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_on_termination": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"encrypted": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"iops": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"kms_key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"snapshot_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"volume_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"volume_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"volume_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				// This should not be necessary, but currently is (see #7198)
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["device_name"].(string)))
					buf.WriteString(fmt.Sprintf("%s-", m["snapshot_id"].(string)))
					return hashcode.String(buf.String())
				},
			},
			"root_block_device": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_on_termination": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"encrypted": {
							Type:     schema.TypeBool,
							Computed: true,
						},

						"iops": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"kms_key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"volume_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"volume_type": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"volume_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"credit_specification": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu_credits": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"metadata_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_endpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"http_tokens": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"http_put_response_hop_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"disable_api_termination": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

// dataSourceAwsInstanceRead performs the instanceID lookup
func dataSourceAwsInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	filters, filtersOk := d.GetOk("filter")
	instanceID, instanceIDOk := d.GetOk("instance_id")
	tags, tagsOk := d.GetOk("instance_tags")

	if !filtersOk && !instanceIDOk && !tagsOk {
		return fmt.Errorf("One of filters, instance_tags, or instance_id must be assigned")
	}

	// Build up search parameters
	params := &ec2.DescribeInstancesInput{}
	if filtersOk {
		params.Filters = buildAwsDataSourceFilters(filters.(*schema.Set))
	}
	if instanceIDOk {
		params.InstanceIds = []*string{aws.String(instanceID.(string))}
	}
	if tagsOk {
		params.Filters = append(params.Filters, ec2TagFiltersFromMap(tags.(map[string]interface{}))...)
	}

	log.Printf("[DEBUG] Reading IAM Instance: %s", params)
	resp, err := conn.DescribeInstances(params)
	if err != nil {
		return err
	}

	// If no instances were returned, return
	if len(resp.Reservations) == 0 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	var filteredInstances []*ec2.Instance

	// loop through reservations, and remove terminated instances, populate instance slice
	for _, res := range resp.Reservations {
		for _, instance := range res.Instances {
			if instance.State != nil && *instance.State.Name != "terminated" {
				filteredInstances = append(filteredInstances, instance)
			}
		}
	}

	var instance *ec2.Instance
	if len(filteredInstances) < 1 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	// (TODO: Support a list of instances to be returned)
	// Possibly with a different data source that returns a list of individual instance data sources
	if len(filteredInstances) > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more " +
			"specific search criteria.")
	} else {
		instance = filteredInstances[0]
	}

	log.Printf("[DEBUG] aws_instance - Single Instance ID found: %s", *instance.InstanceId)
	if err := instanceDescriptionAttributes(d, instance, conn, ignoreTagsConfig); err != nil {
		return err
	}

	if d.Get("get_password_data").(bool) {
		passwordData, err := getAwsEc2InstancePasswordData(*instance.InstanceId, conn)
		if err != nil {
			return err
		}
		d.Set("password_data", passwordData)
	}

	// ARN
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "ec2",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("instance/%s", d.Id()),
	}
	d.Set("arn", arn.String())

	return nil
}

// Populate instance attribute fields with the returned instance
func instanceDescriptionAttributes(d *schema.ResourceData, instance *ec2.Instance, conn *ec2.EC2, ignoreTagsConfig *keyvaluetags.IgnoreConfig) error {
	d.SetId(*instance.InstanceId)
	// Set the easy attributes
	d.Set("instance_state", instance.State.Name)
	if instance.Placement != nil {
		d.Set("availability_zone", instance.Placement.AvailabilityZone)
	}
	if instance.Placement.GroupName != nil {
		d.Set("placement_group", instance.Placement.GroupName)
	}
	if instance.Placement.Tenancy != nil {
		d.Set("tenancy", instance.Placement.Tenancy)
	}
	if instance.Placement.HostId != nil {
		d.Set("host_id", instance.Placement.HostId)
	}
	d.Set("ami", instance.ImageId)
	d.Set("instance_type", instance.InstanceType)
	d.Set("key_name", instance.KeyName)
	d.Set("public_dns", instance.PublicDnsName)
	d.Set("public_ip", instance.PublicIpAddress)
	d.Set("private_dns", instance.PrivateDnsName)
	d.Set("private_ip", instance.PrivateIpAddress)
	d.Set("outpost_arn", instance.OutpostArn)
	d.Set("iam_instance_profile", iamInstanceProfileArnToName(instance.IamInstanceProfile))

	// iterate through network interfaces, and set subnet, network_interface, public_addr
	if len(instance.NetworkInterfaces) > 0 {
		for _, ni := range instance.NetworkInterfaces {
			if *ni.Attachment.DeviceIndex == 0 {
				d.Set("subnet_id", ni.SubnetId)
				d.Set("network_interface_id", ni.NetworkInterfaceId)
				d.Set("associate_public_ip_address", ni.Association != nil)
			}
		}
	} else {
		d.Set("subnet_id", instance.SubnetId)
		d.Set("network_interface_id", "")
	}

	d.Set("ebs_optimized", instance.EbsOptimized)
	if instance.SubnetId != nil && *instance.SubnetId != "" {
		d.Set("source_dest_check", instance.SourceDestCheck)
	}

	if instance.Monitoring != nil && instance.Monitoring.State != nil {
		monitoringState := *instance.Monitoring.State
		d.Set("monitoring", monitoringState == "enabled" || monitoringState == "pending")
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(instance.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	// Security Groups
	if err := readSecurityGroups(d, instance, conn); err != nil {
		return err
	}

	// Block devices
	if err := readBlockDevices(d, instance, conn); err != nil {
		return err
	}
	if _, ok := d.GetOk("ephemeral_block_device"); !ok {
		d.Set("ephemeral_block_device", []interface{}{})
	}

	// Lookup and Set Instance Attributes
	{
		attr, err := conn.DescribeInstanceAttribute(&ec2.DescribeInstanceAttributeInput{
			Attribute:  aws.String("disableApiTermination"),
			InstanceId: aws.String(d.Id()),
		})
		if err != nil {
			return err
		}
		d.Set("disable_api_termination", attr.DisableApiTermination.Value)
	}
	{
		attr, err := conn.DescribeInstanceAttribute(&ec2.DescribeInstanceAttributeInput{
			Attribute:  aws.String(ec2.InstanceAttributeNameUserData),
			InstanceId: aws.String(d.Id()),
		})
		if err != nil {
			return err
		}
		if attr != nil && attr.UserData != nil && attr.UserData.Value != nil {
			d.Set("user_data", userDataHashSum(aws.StringValue(attr.UserData.Value)))
			if d.Get("get_user_data").(bool) {
				d.Set("user_data_base64", aws.StringValue(attr.UserData.Value))
			}
		}
	}

	var creditSpecifications []map[string]interface{}

	// AWS Standard will return InstanceCreditSpecification.NotSupported errors for EC2 Instance IDs outside T2 and T3 instance types
	// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/8055
	if strings.HasPrefix(aws.StringValue(instance.InstanceType), "t2") || strings.HasPrefix(aws.StringValue(instance.InstanceType), "t3") {
		var err error
		creditSpecifications, err = getCreditSpecifications(conn, d.Id())

		// Ignore UnsupportedOperation errors for AWS China and GovCloud (US)
		// Reference: https://github.com/terraform-providers/terraform-provider-aws/pull/4362
		if err != nil && !isAWSErr(err, "UnsupportedOperation", "") {
			return fmt.Errorf("error getting EC2 Instance (%s) Credit Specifications: %s", d.Id(), err)
		}
	}

	if err := d.Set("credit_specification", creditSpecifications); err != nil {
		return fmt.Errorf("error setting credit_specification: %s", err)
	}

	if err := d.Set("metadata_options", flattenEc2InstanceMetadataOptions(instance.MetadataOptions)); err != nil {
		return fmt.Errorf("error setting metadata_options: %s", err)
	}

	return nil
}
