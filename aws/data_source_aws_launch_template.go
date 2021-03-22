package aws

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsLaunchTemplate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLaunchTemplateRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"default_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"latest_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"block_device_mappings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"no_device": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"virtual_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ebs": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delete_on_termination": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"encrypted": {
										Type:     schema.TypeString,
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
									"throughput": {
										Type:     schema.TypeInt,
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
								},
							},
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
			"disable_api_termination": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ebs_optimized": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"elastic_gpu_specifications": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"iam_instance_profile": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"image_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"instance_initiated_shutdown_behavior": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_market_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"market_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"spot_options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"block_duration_minutes": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"instance_interruption_behavior": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"max_price": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"spot_instance_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"valid_until": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kernel_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_name": {
				Type:     schema.TypeString,
				Computed: true,
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
			"enclave_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"monitoring": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"network_interfaces": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"associate_carrier_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"associate_public_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"delete_on_termination": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_index": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"security_groups": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ipv6_address_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ipv6_addresses": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"network_interface_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ipv4_addresses": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ipv4_address_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"placement": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"affinity": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"availability_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"host_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"spread_domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tenancy": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"partition_number": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"ram_disk_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_security_group_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tag_specifications": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tags": tagsSchemaComputed(),
					},
				},
			},
			"user_data": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hibernation_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"configured": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"tags":   tagsSchemaComputed(),
			"filter": dataSourceFiltersSchema(),
		},
	}
}

func dataSourceAwsLaunchTemplateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	filters, filtersOk := d.GetOk("filter")
	id, idOk := d.GetOk("id")
	name, nameOk := d.GetOk("name")
	tags, tagsOk := d.GetOk("tags")

	params := &ec2.DescribeLaunchTemplatesInput{}
	if filtersOk {
		params.Filters = buildAwsDataSourceFilters(filters.(*schema.Set))
	}
	if idOk {
		params.LaunchTemplateIds = []*string{aws.String(id.(string))}
	}
	if nameOk {
		params.LaunchTemplateNames = []*string{aws.String(name.(string))}
	}
	if tagsOk {
		params.Filters = append(params.Filters, ec2TagFiltersFromMap(tags.(map[string]interface{}))...)
	}

	dlt, err := conn.DescribeLaunchTemplates(params)

	if err != nil {
		if isAWSErr(err, "InvalidLaunchTemplateId.NotFound", "") ||
			isAWSErr(err, "InvalidLaunchTemplateName.NotFoundException", "") {
			return fmt.Errorf("Launch Template not found")
		}
		return fmt.Errorf("Error getting launch template: %w", err)
	}

	if dlt == nil || len(dlt.LaunchTemplates) == 0 {
		return fmt.Errorf("error reading launch template: empty output")
	}

	if len(dlt.LaunchTemplates) > 1 {
		return fmt.Errorf("Search returned %d result(s), please revise so only one is returned", len(dlt.LaunchTemplates))
	}

	lt := dlt.LaunchTemplates[0]
	log.Printf("[DEBUG] Found launch template %s", aws.StringValue(lt.LaunchTemplateId))

	d.SetId(aws.StringValue(lt.LaunchTemplateId))
	d.Set("name", lt.LaunchTemplateName)
	d.Set("latest_version", lt.LatestVersionNumber)
	d.Set("default_version", lt.DefaultVersionNumber)
	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(lt.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("launch-template/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	version := strconv.Itoa(int(*lt.LatestVersionNumber))
	dltv, err := conn.DescribeLaunchTemplateVersions(&ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateId: aws.String(d.Id()),
		Versions:         []*string{aws.String(version)},
	})
	if err != nil {
		return fmt.Errorf("error reading launch template version (%s) for launch template (%s): %w", version, d.Id(), err)
	}

	if dltv == nil || len(dltv.LaunchTemplateVersions) == 0 {
		return fmt.Errorf("error reading launch template version (%s) for launch template (%s): empty output", version, d.Id())
	}

	log.Printf("[DEBUG] Received launch template version %q (version %d)", d.Id(), aws.Int64Value(lt.LatestVersionNumber))

	ltData := dltv.LaunchTemplateVersions[0].LaunchTemplateData

	d.Set("disable_api_termination", ltData.DisableApiTermination)
	d.Set("image_id", ltData.ImageId)
	d.Set("instance_initiated_shutdown_behavior", ltData.InstanceInitiatedShutdownBehavior)
	d.Set("instance_type", ltData.InstanceType)
	d.Set("kernel_id", ltData.KernelId)
	d.Set("key_name", ltData.KeyName)
	d.Set("ram_disk_id", ltData.RamDiskId)
	if err := d.Set("security_group_names", aws.StringValueSlice(ltData.SecurityGroups)); err != nil {
		return fmt.Errorf("error setting security_group_names: %w", err)
	}
	d.Set("user_data", ltData.UserData)
	if err := d.Set("vpc_security_group_ids", aws.StringValueSlice(ltData.SecurityGroupIds)); err != nil {
		return fmt.Errorf("error setting vpc_security_group_ids: %w", err)
	}
	d.Set("ebs_optimized", "")

	if ltData.EbsOptimized != nil {
		d.Set("ebs_optimized", strconv.FormatBool(aws.BoolValue(ltData.EbsOptimized)))
	}

	if err := d.Set("block_device_mappings", getBlockDeviceMappings(ltData.BlockDeviceMappings)); err != nil {
		return fmt.Errorf("error setting block_device_mappings: %w", err)
	}

	if strings.HasPrefix(aws.StringValue(ltData.InstanceType), "t2") || strings.HasPrefix(aws.StringValue(ltData.InstanceType), "t3") {
		if err := d.Set("credit_specification", getCreditSpecification(ltData.CreditSpecification)); err != nil {
			return fmt.Errorf("error setting credit_specification: %w", err)
		}
	}

	if err := d.Set("elastic_gpu_specifications", getElasticGpuSpecifications(ltData.ElasticGpuSpecifications)); err != nil {
		return fmt.Errorf("error setting elastic_gpu_specifications: %w", err)
	}

	if err := d.Set("iam_instance_profile", getIamInstanceProfile(ltData.IamInstanceProfile)); err != nil {
		return fmt.Errorf("error setting iam_instance_profile: %w", err)
	}

	if err := d.Set("instance_market_options", getInstanceMarketOptions(ltData.InstanceMarketOptions)); err != nil {
		return fmt.Errorf("error setting instance_market_options: %w", err)
	}

	if err := d.Set("metadata_options", flattenLaunchTemplateInstanceMetadataOptions(ltData.MetadataOptions)); err != nil {
		return fmt.Errorf("error setting metadata_options: %w", err)
	}

	if err := d.Set("enclave_options", getEnclaveOptions(ltData.EnclaveOptions)); err != nil {
		return fmt.Errorf("error setting enclave_options: %w", err)
	}

	if err := d.Set("monitoring", getMonitoring(ltData.Monitoring)); err != nil {
		return fmt.Errorf("error setting monitoring: %w", err)
	}

	if err := d.Set("network_interfaces", getNetworkInterfaces(ltData.NetworkInterfaces)); err != nil {
		return fmt.Errorf("error setting network_interfaces: %w", err)
	}

	if err := d.Set("placement", getPlacement(ltData.Placement)); err != nil {
		return fmt.Errorf("error setting placement: %w", err)
	}

	if err := d.Set("hibernation_options", flattenLaunchTemplateHibernationOptions(ltData.HibernationOptions)); err != nil {
		return fmt.Errorf("error setting hibernation_options: %w", err)
	}

	if err := d.Set("tag_specifications", getTagSpecifications(ltData.TagSpecifications)); err != nil {
		return fmt.Errorf("error setting tag_specifications: %w", err)
	}

	return nil
}
