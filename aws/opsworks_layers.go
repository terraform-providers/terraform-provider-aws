package aws

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

// OpsWorks has a single concept of "layer" which represents several different
// layer types. The differences between these are in some extra properties that
// get packed into an "Attributes" map, but in the OpsWorks UI these are presented
// as first-class options, and so Terraform prefers to expose them this way and
// hide the implementation detail that they are all packed into a single type
// in the underlying API.
//
// This file contains utilities that are shared between all of the concrete
// layer resource types, which have names matching aws_opsworks_*_layer .

type opsworksLayerTypeAttribute struct {
	AttrName  string
	Type      schema.ValueType
	Default   interface{}
	Required  bool
	WriteOnly bool
}

type opsworksLayerType struct {
	TypeName         string
	DefaultLayerName string
	Attributes       map[string]*opsworksLayerTypeAttribute
	CustomShortName  bool
}

var (
	opsworksTrueString  = "true"
	opsworksFalseString = "false"
)

func (lt *opsworksLayerType) SchemaResource() *schema.Resource {
	resourceSchema := map[string]*schema.Schema{
		"auto_assign_elastic_ips": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},

		"auto_assign_public_ips": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},

		"custom_instance_profile_arn": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validateArn,
		},

		"elastic_load_balancer": {
			Type:     schema.TypeString,
			Optional: true,
		},

		"custom_setup_recipes": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},

		"custom_configure_recipes": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},

		"custom_deploy_recipes": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},

		"custom_undeploy_recipes": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},

		"custom_shutdown_recipes": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},

		"custom_security_group_ids": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Set:      schema.HashString,
		},

		"custom_json": {
			Type: schema.TypeString,
			StateFunc: func(v interface{}) string {
				json, _ := structure.NormalizeJsonString(v)
				return json
			},
			Optional: true,
		},

		"auto_healing": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},

		"install_updates_on_boot": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},

		"instance_shutdown_timeout": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  120,
		},

		"drain_elb_on_shutdown": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},

		"system_packages": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Set:      schema.HashString,
		},

		"stack_id": {
			Type:     schema.TypeString,
			ForceNew: true,
			Required: true,
		},

		"use_ebs_optimized_instances": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},

		"ebs_volume": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{

					"iops": {
						Type:     schema.TypeInt,
						Optional: true,
						Default:  0,
					},

					"mount_point": {
						Type:     schema.TypeString,
						Required: true,
					},

					"number_of_disks": {
						Type:     schema.TypeInt,
						Required: true,
					},

					"raid_level": {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "",
					},

					"size": {
						Type:     schema.TypeInt,
						Required: true,
					},

					"type": {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "standard",
						ValidateFunc: validation.StringInSlice([]string{
							"standard",
							"io1",
							"gp2",
							"st1",
							"sc1",
						}, false),
					},

					"encrypted": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
				},
			},
			Set: func(v interface{}) int {
				m := v.(map[string]interface{})
				return hashcode.String(m["mount_point"].(string))
			},
		},
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"tags":     tagsSchema(),
		"tags_all": tagsSchemaComputed(),
		"cloudwatch_configuration": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"enabled": {
						Type:     schema.TypeBool,
						Optional: true,
					},
					"log_streams": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"batch_count": {
									Type:         schema.TypeInt,
									Default:      1000,
									Optional:     true,
									ValidateFunc: validation.IntAtMost(10000),
								},
								"batch_size": {
									Type:         schema.TypeInt,
									Default:      32768,
									Optional:     true,
									ValidateFunc: validation.IntAtMost(1048576),
								},
								"buffer_duration": {
									Type:         schema.TypeInt,
									Default:      5000,
									Optional:     true,
									ValidateFunc: validation.IntAtLeast(5000),
								},
								"datetime_format": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"encoding": {
									Type:     schema.TypeString,
									Optional: true,
									Default:  opsworks.CloudWatchLogsEncodingUtf8,
								},
								"file": {
									Type:     schema.TypeString,
									Required: true,
								},
								"file_fingerprint_lines": {
									Type:     schema.TypeString,
									Optional: true,
									Default:  "1",
								},
								"initial_position": {
									Type:     schema.TypeString,
									Optional: true,
									Default:  opsworks.CloudWatchLogsInitialPositionStartOfFile,
									ValidateFunc: validation.StringInSlice([]string{
										opsworks.CloudWatchLogsInitialPositionEndOfFile,
										opsworks.CloudWatchLogsInitialPositionStartOfFile,
									}, false),
								},
								"log_group_name": {
									Type:     schema.TypeString,
									Required: true,
								},
								"multiline_start_pattern": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"timezone": {
									Type:     schema.TypeString,
									Optional: true,
									ValidateFunc: validation.StringInSlice([]string{
										opsworks.CloudWatchLogsTimeZoneLocal,
										opsworks.CloudWatchLogsTimeZoneUtc,
									}, false),
								},
							},
						},
					},
				},
			},
		},
	}

	if lt.CustomShortName {
		resourceSchema["short_name"] = &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		}
	}

	if lt.DefaultLayerName != "" {
		resourceSchema["name"] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Default:  lt.DefaultLayerName,
		}
	} else {
		resourceSchema["name"] = &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		}
	}

	for key, def := range lt.Attributes {
		resourceSchema[key] = &schema.Schema{
			Type:     def.Type,
			Default:  def.Default,
			Required: def.Required,
			Optional: !def.Required,
		}
	}

	return &schema.Resource{
		Read: func(d *schema.ResourceData, meta interface{}) error {
			return lt.Read(d, meta)
		},
		Create: func(d *schema.ResourceData, meta interface{}) error {
			return lt.Create(d, meta)
		},
		Update: func(d *schema.ResourceData, meta interface{}) error {
			return lt.Update(d, meta)
		},
		Delete: func(d *schema.ResourceData, meta interface{}) error {
			return lt.Delete(d, meta)
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: resourceSchema,

		CustomizeDiff: SetTagsDiff,
	}
}

func (lt *opsworksLayerType) Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).opsworksconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &opsworks.DescribeLayersInput{
		LayerIds: []*string{
			aws.String(d.Id()),
		},
	}

	log.Printf("[DEBUG] Reading OpsWorks layer: %s", d.Id())

	resp, err := conn.DescribeLayers(req)
	if err != nil {
		if isAWSErr(err, opsworks.ErrCodeResourceNotFoundException, "") {
			d.SetId("")
			return nil
		}
		return err
	}

	layer := resp.Layers[0]
	d.SetId(aws.StringValue(layer.LayerId))
	d.Set("auto_assign_elastic_ips", layer.AutoAssignElasticIps)
	d.Set("auto_assign_public_ips", layer.AutoAssignPublicIps)
	d.Set("custom_instance_profile_arn", layer.CustomInstanceProfileArn)
	d.Set("custom_security_group_ids", flattenStringList(layer.CustomSecurityGroupIds))
	d.Set("auto_healing", layer.EnableAutoHealing)
	d.Set("install_updates_on_boot", layer.InstallUpdatesOnBoot)
	d.Set("name", layer.Name)
	d.Set("system_packages", flattenStringList(layer.Packages))
	d.Set("stack_id", layer.StackId)
	d.Set("use_ebs_optimized_instances", layer.UseEbsOptimizedInstances)

	if err := d.Set("cloudwatch_configuration", flattenOpsworksCloudWatchConfig(layer.CloudWatchLogsConfiguration)); err != nil {
		return fmt.Errorf("error setting cloudwatch_configuration: %s", err)
	}

	if lt.CustomShortName {
		d.Set("short_name", layer.Shortname)
	}

	if layer.CustomJson == nil {
		d.Set("custom_json", "")
	} else {
		policy, err := structure.NormalizeJsonString(*layer.CustomJson)
		if err != nil {
			return fmt.Errorf("policy contains an invalid JSON: %s", err)
		}
		d.Set("custom_json", policy)
	}

	err = lt.SetAttributeMap(d, layer.Attributes)
	if err != nil {
		return err
	}
	lt.SetLifecycleEventConfiguration(d, layer.LifecycleEventConfiguration)
	lt.SetCustomRecipes(d, layer.CustomRecipes)
	lt.SetVolumeConfigurations(d, layer.VolumeConfigurations)

	/* get ELB */
	ebsRequest := &opsworks.DescribeElasticLoadBalancersInput{
		LayerIds: []*string{
			aws.String(d.Id()),
		},
	}
	loadBalancers, err := conn.DescribeElasticLoadBalancers(ebsRequest)
	if err != nil {
		return err
	}

	if loadBalancers.ElasticLoadBalancers == nil || len(loadBalancers.ElasticLoadBalancers) == 0 {
		d.Set("elastic_load_balancer", "")
	} else {
		loadBalancer := loadBalancers.ElasticLoadBalancers[0]
		if loadBalancer != nil {
			d.Set("elastic_load_balancer", loadBalancer.ElasticLoadBalancerName)
		}
	}

	arn := aws.StringValue(layer.Arn)
	d.Set("arn", arn)
	tags, err := keyvaluetags.OpsworksListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Opsworks Layer (%s): %s", arn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func (lt *opsworksLayerType) Create(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).opsworksconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	attributes, err := lt.AttributeMap(d)
	if err != nil {
		return err
	}
	req := &opsworks.CreateLayerInput{
		AutoAssignElasticIps:        aws.Bool(d.Get("auto_assign_elastic_ips").(bool)),
		AutoAssignPublicIps:         aws.Bool(d.Get("auto_assign_public_ips").(bool)),
		CustomInstanceProfileArn:    aws.String(d.Get("custom_instance_profile_arn").(string)),
		CustomRecipes:               lt.CustomRecipes(d),
		CustomSecurityGroupIds:      expandStringSet(d.Get("custom_security_group_ids").(*schema.Set)),
		EnableAutoHealing:           aws.Bool(d.Get("auto_healing").(bool)),
		InstallUpdatesOnBoot:        aws.Bool(d.Get("install_updates_on_boot").(bool)),
		LifecycleEventConfiguration: lt.LifecycleEventConfiguration(d),
		Name:                        aws.String(d.Get("name").(string)),
		Packages:                    expandStringSet(d.Get("system_packages").(*schema.Set)),
		Type:                        aws.String(lt.TypeName),
		StackId:                     aws.String(d.Get("stack_id").(string)),
		UseEbsOptimizedInstances:    aws.Bool(d.Get("use_ebs_optimized_instances").(bool)),
		Attributes:                  attributes,
		VolumeConfigurations:        lt.VolumeConfigurations(d),
	}

	if v, ok := d.GetOk("cloudwatch_configuration"); ok {
		req.CloudWatchLogsConfiguration = expandOpsworksCloudWatchConfig(v.([]interface{}))
	}

	if lt.CustomShortName {
		req.Shortname = aws.String(d.Get("short_name").(string))
	} else {
		req.Shortname = aws.String(lt.TypeName)
	}

	req.CustomJson = aws.String(d.Get("custom_json").(string))

	log.Printf("[DEBUG] Creating OpsWorks layer: %s", d.Id())

	resp, err := conn.CreateLayer(req)
	if err != nil {
		return err
	}

	layerId := *resp.LayerId
	d.SetId(layerId)

	loadBalancer := aws.String(d.Get("elastic_load_balancer").(string))
	if loadBalancer != nil && *loadBalancer != "" {
		log.Printf("[DEBUG] Attaching load balancer: %s", *loadBalancer)
		_, err := conn.AttachElasticLoadBalancer(&opsworks.AttachElasticLoadBalancerInput{
			ElasticLoadBalancerName: loadBalancer,
			LayerId:                 &layerId,
		})
		if err != nil {
			return err
		}
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "opsworks",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("layer/%s", d.Id()),
	}.String()

	if len(tags) > 0 {
		if err := keyvaluetags.OpsworksUpdateTags(conn, arn, nil, tags); err != nil {
			return fmt.Errorf("error updating Opsworks stack (%s) tags: %s", arn, err)
		}
	}

	return lt.Read(d, meta)
}

func (lt *opsworksLayerType) Update(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).opsworksconn

	attributes, err := lt.AttributeMap(d)
	if err != nil {
		return err
	}
	req := &opsworks.UpdateLayerInput{
		LayerId:                     aws.String(d.Id()),
		AutoAssignElasticIps:        aws.Bool(d.Get("auto_assign_elastic_ips").(bool)),
		AutoAssignPublicIps:         aws.Bool(d.Get("auto_assign_public_ips").(bool)),
		CustomInstanceProfileArn:    aws.String(d.Get("custom_instance_profile_arn").(string)),
		CustomRecipes:               lt.CustomRecipes(d),
		CustomSecurityGroupIds:      expandStringSet(d.Get("custom_security_group_ids").(*schema.Set)),
		EnableAutoHealing:           aws.Bool(d.Get("auto_healing").(bool)),
		InstallUpdatesOnBoot:        aws.Bool(d.Get("install_updates_on_boot").(bool)),
		LifecycleEventConfiguration: lt.LifecycleEventConfiguration(d),
		Name:                        aws.String(d.Get("name").(string)),
		Packages:                    expandStringSet(d.Get("system_packages").(*schema.Set)),
		UseEbsOptimizedInstances:    aws.Bool(d.Get("use_ebs_optimized_instances").(bool)),
		Attributes:                  attributes,
		VolumeConfigurations:        lt.VolumeConfigurations(d),
	}

	if v, ok := d.GetOk("cloudwatch_configuration"); ok {
		req.CloudWatchLogsConfiguration = expandOpsworksCloudWatchConfig(v.([]interface{}))
	}

	if lt.CustomShortName {
		req.Shortname = aws.String(d.Get("short_name").(string))
	} else {
		req.Shortname = aws.String(lt.TypeName)
	}

	req.CustomJson = aws.String(d.Get("custom_json").(string))

	log.Printf("[DEBUG] Updating OpsWorks layer: %s", d.Id())

	if d.HasChange("elastic_load_balancer") {
		lbo, lbn := d.GetChange("elastic_load_balancer")

		loadBalancerOld := aws.String(lbo.(string))
		loadBalancerNew := aws.String(lbn.(string))

		if loadBalancerOld != nil && *loadBalancerOld != "" {
			log.Printf("[DEBUG] Dettaching load balancer: %s", *loadBalancerOld)
			_, err := conn.DetachElasticLoadBalancer(&opsworks.DetachElasticLoadBalancerInput{
				ElasticLoadBalancerName: loadBalancerOld,
				LayerId:                 aws.String(d.Id()),
			})
			if err != nil {
				return err
			}
		}

		if loadBalancerNew != nil && *loadBalancerNew != "" {
			log.Printf("[DEBUG] Attaching load balancer: %s", *loadBalancerNew)
			_, err := conn.AttachElasticLoadBalancer(&opsworks.AttachElasticLoadBalancerInput{
				ElasticLoadBalancerName: loadBalancerNew,
				LayerId:                 aws.String(d.Id()),
			})
			if err != nil {
				return err
			}
		}
	}

	_, err = conn.UpdateLayer(req)
	if err != nil {
		return err
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		arn := d.Get("arn").(string)
		if err := keyvaluetags.OpsworksUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Opsworks Layer (%s) tags: %s", arn, err)
		}
	}

	return lt.Read(d, meta)
}

func (lt *opsworksLayerType) Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).opsworksconn

	req := &opsworks.DeleteLayerInput{
		LayerId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting OpsWorks layer: %s", d.Id())

	_, err := conn.DeleteLayer(req)
	return err
}

func (lt *opsworksLayerType) AttributeMap(d *schema.ResourceData) (map[string]*string, error) {
	attrs := map[string]*string{}

	for key, def := range lt.Attributes {
		value := d.Get(key)
		switch def.Type {
		case schema.TypeString:
			strValue := value.(string)
			attrs[def.AttrName] = &strValue
		case schema.TypeInt:
			intValue := value.(int)
			strValue := strconv.Itoa(intValue)
			attrs[def.AttrName] = &strValue
		case schema.TypeBool:
			boolValue := value.(bool)
			if boolValue {
				attrs[def.AttrName] = &opsworksTrueString
			} else {
				attrs[def.AttrName] = &opsworksFalseString
			}
		default:
			// should never happen
			return nil, fmt.Errorf("Unsupported OpsWorks layer attribute type: %s", def.Type)
		}
	}

	return attrs, nil
}

func (lt *opsworksLayerType) SetAttributeMap(d *schema.ResourceData, attrs map[string]*string) error {
	for key, def := range lt.Attributes {
		// Ignore write-only attributes; we'll just keep what we already have stored.
		// (The AWS API returns garbage placeholder values for these.)
		if def.WriteOnly {
			continue
		}

		if strPtr, ok := attrs[def.AttrName]; ok && strPtr != nil {
			strValue := *strPtr

			switch def.Type {
			case schema.TypeString:
				d.Set(key, strValue)
			case schema.TypeInt:
				intValue, err := strconv.Atoi(strValue)
				if err == nil {
					d.Set(key, intValue)
				} else {
					// Got garbage from the AWS API
					d.Set(key, nil)
				}
			case schema.TypeBool:
				boolValue := true
				if strValue == opsworksFalseString {
					boolValue = false
				}
				d.Set(key, boolValue)
			default:
				// should never happen
				return fmt.Errorf("Unsupported OpsWorks layer attribute type: %s", def.Type)
			}
			return nil

		} else {
			d.Set(key, nil)
		}
	}
	return nil
}

func (lt *opsworksLayerType) LifecycleEventConfiguration(d *schema.ResourceData) *opsworks.LifecycleEventConfiguration {
	return &opsworks.LifecycleEventConfiguration{
		Shutdown: &opsworks.ShutdownEventConfiguration{
			DelayUntilElbConnectionsDrained: aws.Bool(d.Get("drain_elb_on_shutdown").(bool)),
			ExecutionTimeout:                aws.Int64(int64(d.Get("instance_shutdown_timeout").(int))),
		},
	}
}

func (lt *opsworksLayerType) SetLifecycleEventConfiguration(d *schema.ResourceData, v *opsworks.LifecycleEventConfiguration) {
	if v == nil || v.Shutdown == nil {
		d.Set("drain_elb_on_shutdown", nil)
		d.Set("instance_shutdown_timeout", nil)
	} else {
		d.Set("drain_elb_on_shutdown", v.Shutdown.DelayUntilElbConnectionsDrained)
		d.Set("instance_shutdown_timeout", v.Shutdown.ExecutionTimeout)
	}
}

func (lt *opsworksLayerType) CustomRecipes(d *schema.ResourceData) *opsworks.Recipes {
	return &opsworks.Recipes{
		Configure: expandStringList(d.Get("custom_configure_recipes").([]interface{})),
		Deploy:    expandStringList(d.Get("custom_deploy_recipes").([]interface{})),
		Setup:     expandStringList(d.Get("custom_setup_recipes").([]interface{})),
		Shutdown:  expandStringList(d.Get("custom_shutdown_recipes").([]interface{})),
		Undeploy:  expandStringList(d.Get("custom_undeploy_recipes").([]interface{})),
	}
}

func (lt *opsworksLayerType) SetCustomRecipes(d *schema.ResourceData, v *opsworks.Recipes) {
	// Null out everything first, and then we'll consider what to put back.
	d.Set("custom_configure_recipes", nil)
	d.Set("custom_deploy_recipes", nil)
	d.Set("custom_setup_recipes", nil)
	d.Set("custom_shutdown_recipes", nil)
	d.Set("custom_undeploy_recipes", nil)

	if v == nil {
		return
	}

	d.Set("custom_configure_recipes", flattenStringList(v.Configure))
	d.Set("custom_deploy_recipes", flattenStringList(v.Deploy))
	d.Set("custom_setup_recipes", flattenStringList(v.Setup))
	d.Set("custom_shutdown_recipes", flattenStringList(v.Shutdown))
	d.Set("custom_undeploy_recipes", flattenStringList(v.Undeploy))
}

func (lt *opsworksLayerType) VolumeConfigurations(d *schema.ResourceData) []*opsworks.VolumeConfiguration {
	configuredVolumes := d.Get("ebs_volume").(*schema.Set).List()
	result := make([]*opsworks.VolumeConfiguration, len(configuredVolumes))

	for i := 0; i < len(configuredVolumes); i++ {
		volumeData := configuredVolumes[i].(map[string]interface{})

		result[i] = &opsworks.VolumeConfiguration{
			MountPoint:    aws.String(volumeData["mount_point"].(string)),
			NumberOfDisks: aws.Int64(int64(volumeData["number_of_disks"].(int))),
			Size:          aws.Int64(int64(volumeData["size"].(int))),
			VolumeType:    aws.String(volumeData["type"].(string)),
			Encrypted:     aws.Bool(volumeData["encrypted"].(bool)),
		}

		iops := int64(volumeData["iops"].(int))
		if iops != 0 {
			result[i].Iops = aws.Int64(iops)
		}

		raidLevelStr := volumeData["raid_level"].(string)
		if raidLevelStr != "" {
			raidLevel, err := strconv.Atoi(raidLevelStr)
			if err == nil {
				result[i].RaidLevel = aws.Int64(int64(raidLevel))
			}
		}
	}

	return result
}

func (lt *opsworksLayerType) SetVolumeConfigurations(d *schema.ResourceData, v []*opsworks.VolumeConfiguration) {
	newValue := make([]*map[string]interface{}, len(v))

	for i := 0; i < len(v); i++ {
		config := v[i]
		data := make(map[string]interface{})
		newValue[i] = &data

		if config.Iops != nil {
			data["iops"] = int(*config.Iops)
		} else {
			data["iops"] = 0
		}
		if config.MountPoint != nil {
			data["mount_point"] = *config.MountPoint
		}
		if config.NumberOfDisks != nil {
			data["number_of_disks"] = int(*config.NumberOfDisks)
		}
		if config.RaidLevel != nil {
			data["raid_level"] = strconv.Itoa(int(*config.RaidLevel))
		}
		if config.Size != nil {
			data["size"] = int(*config.Size)
		}
		if config.VolumeType != nil {
			data["type"] = *config.VolumeType
		}
		if config.Encrypted != nil {
			data["encrypted"] = *config.Encrypted
		}
	}

	d.Set("ebs_volume", newValue)
}

func expandOpsworksCloudWatchConfig(l []interface{}) *opsworks.CloudWatchLogsConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &opsworks.CloudWatchLogsConfiguration{
		Enabled:    aws.Bool(m["enabled"].(bool)),
		LogStreams: expandOpsworksCloudWatchConfigLogStream(m["log_streams"].([]interface{})),
	}

	return config
}

func expandOpsworksCloudWatchConfigLogStream(l []interface{}) []*opsworks.CloudWatchLogsLogStream {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	logStreams := make([]*opsworks.CloudWatchLogsLogStream, 0)

	for _, m := range l {
		item := m.(map[string]interface{})
		logStream := &opsworks.CloudWatchLogsLogStream{}

		if v, ok := item["batch_count"]; ok {
			logStream.BatchCount = aws.Int64(int64(v.(int)))
		}

		if v, ok := item["batch_size"]; ok {
			logStream.BatchSize = aws.Int64(int64(v.(int)))
		}

		if v, ok := item["buffer_duration"]; ok {
			logStream.BufferDuration = aws.Int64(int64(v.(int)))
		}

		if v, ok := item["datetime_format"]; ok {
			logStream.DatetimeFormat = aws.String(v.(string))
		}

		if v, ok := item["encoding"]; ok {
			logStream.Encoding = aws.String(v.(string))
		}

		if v, ok := item["file"]; ok {
			logStream.File = aws.String(v.(string))
		}

		if v, ok := item["file_fingerprint_lines"]; ok {
			logStream.FileFingerprintLines = aws.String(v.(string))
		}

		if v, ok := item["initial_position"]; ok {
			logStream.InitialPosition = aws.String(v.(string))
		}

		if v, ok := item["log_group_name"]; ok {
			logStream.LogGroupName = aws.String(v.(string))
		}

		if v, ok := item["multiline_start_pattern"]; ok {
			logStream.MultiLineStartPattern = aws.String(v.(string))
		}

		if v, ok := item["timezone"]; ok {
			logStream.TimeZone = aws.String(v.(string))
		}

		logStreams = append(logStreams, logStream)
	}

	return logStreams
}

func flattenOpsworksCloudWatchConfig(cloudwatchConfig *opsworks.CloudWatchLogsConfiguration) []map[string]interface{} {
	if cloudwatchConfig == nil {
		return nil
	}

	p := map[string]interface{}{
		"enabled":     aws.BoolValue(cloudwatchConfig.Enabled),
		"log_streams": flattenOpsworksCloudWatchConfigLogStreams(cloudwatchConfig.LogStreams),
	}

	return []map[string]interface{}{p}
}

func flattenOpsworksCloudWatchConfigLogStreams(logStreams []*opsworks.CloudWatchLogsLogStream) []interface{} {
	out := make([]interface{}, len(logStreams))

	for i, logStream := range logStreams {
		m := make(map[string]interface{})

		if logStream.TimeZone != nil {
			m["timezone"] = aws.StringValue(logStream.TimeZone)
		}

		if logStream.MultiLineStartPattern != nil {
			m["multiline_start_pattern"] = aws.StringValue(logStream.MultiLineStartPattern)
		}

		if logStream.Encoding != nil {
			m["encoding"] = aws.StringValue(logStream.Encoding)
		}

		if logStream.LogGroupName != nil {
			m["log_group_name"] = aws.StringValue(logStream.LogGroupName)
		}

		if logStream.File != nil {
			m["file"] = aws.StringValue(logStream.File)
		}

		if logStream.DatetimeFormat != nil {
			m["datetime_format"] = aws.StringValue(logStream.DatetimeFormat)
		}

		if logStream.FileFingerprintLines != nil {
			m["file_fingerprint_lines"] = aws.StringValue(logStream.FileFingerprintLines)
		}

		if logStream.InitialPosition != nil {
			m["initial_position"] = aws.StringValue(logStream.InitialPosition)
		}

		if logStream.BatchSize != nil {
			m["batch_size"] = aws.Int64Value(logStream.BatchSize)
		}

		if logStream.BatchCount != nil {
			m["batch_count"] = aws.Int64Value(logStream.BatchCount)
		}

		if logStream.BufferDuration != nil {
			m["buffer_duration"] = aws.Int64Value(logStream.BufferDuration)
		}

		out[i] = m
	}

	return out
}
