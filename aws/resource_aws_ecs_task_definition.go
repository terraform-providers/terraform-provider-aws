package aws

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEcsTaskDefinition() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAwsEcsTaskDefinitionCreate,
		Read:   resourceAwsEcsTaskDefinitionRead,
		Update: resourceAwsEcsTaskDefinitionUpdate,
		Delete: resourceAwsEcsTaskDefinitionDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

				var familyRevision string

				resARN, err := arn.Parse(d.Id())
				if err == nil {
					log.Printf("[DEBUG] import aws_ecs_task_definition resource by arn: %s", d.Id())
					d.Set("arn", d.Id())
					familyRevision = strings.TrimPrefix(resARN.Resource, "task-definition/")
				} else {
					log.Printf("[DEBUG] import aws_ecs_task_definition resource by FAMILY[:REVISION]: %s", d.Id())
					familyRevision = d.Id()
				}
				idErr := fmt.Errorf("Expected ID in format of either arn:PARTITION:ecs:REGION:ACCOUNTID:task-definition/FAMILY:REVISION, FAMILY:REVISION, or FAMILY, and provided: %s", d.Id())
				familyRevisionParts := strings.Split(familyRevision, ":")
				if len(familyRevisionParts) > 2 {
					return nil, idErr
				} else if len(familyRevisionParts) == 2 {
					// import resource by either ARN or FAMILY:REVISION
					d.Set("revision", familyRevisionParts[1])
				}
				d.SetId(familyRevisionParts[0])
				d.Set("family", familyRevisionParts[0])

				return []*schema.ResourceData{d}, nil
			},
		},

		CustomizeDiff: SetTagsDiff,

		SchemaVersion: 1,
		MigrateState:  resourceAwsEcsTaskDefinitionMigrateState,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cpu": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"container_definitions": {
				Type:     schema.TypeString,
				Required: true,
				StateFunc: func(v interface{}) string {
					// Sort the lists of environment variables as they are serialized to state, so we won't get
					// spurious reorderings in plans (diff is suppressed if the environment variables haven't changed,
					// but they still show in the plan if some other property changes).
					orderedCDs, _ := expandEcsContainerDefinitions(v.(string))
					containerDefinitions(orderedCDs).OrderEnvironmentVariables()
					unnormalizedJson, _ := flattenEcsContainerDefinitions(orderedCDs)
					json, _ := structure.NormalizeJsonString(unnormalizedJson)
					return json
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					networkMode, ok := d.GetOk("network_mode")
					isAWSVPC := ok && networkMode.(string) == ecs.NetworkModeAwsvpc
					equal, _ := EcsContainerDefinitionsAreEquivalent(old, new, isAWSVPC)
					return equal
				},
				ValidateFunc: validateAwsEcsTaskDefinitionContainerDefinitions,
			},

			"task_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},

			"execution_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},

			"memory": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"network_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					ecs.NetworkModeBridge,
					ecs.NetworkModeHost,
					ecs.NetworkModeAwsvpc,
					ecs.NetworkModeNone,
				}, false),
			},

			"volume": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"host_path": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"docker_volume_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"scope": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.StringInSlice([]string{
											ecs.ScopeShared,
											ecs.ScopeTask,
										}, false),
										Default: ecs.ScopeTask,
									},
									"autoprovision": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"driver": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "local",
									},
									"driver_opts": {
										Type:     schema.TypeMap,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Optional: true,
									},
									"labels": {
										Type:     schema.TypeMap,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Optional: true,
									},
								},
							},
						},
						"efs_volume_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"file_system_id": {
										Type:     schema.TypeString,
										ForceNew: true,
										Required: true,
									},
									"root_directory": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
										Default:  "/",
									},
									"transit_encryption": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
										ValidateFunc: validation.StringInSlice([]string{
											ecs.EFSTransitEncryptionEnabled,
											ecs.EFSTransitEncryptionDisabled,
										}, false),
									},
									"transit_encryption_port": {
										Type:         schema.TypeInt,
										ForceNew:     true,
										Optional:     true,
										ValidateFunc: validation.IsPortNumber,
									},
									"authorization_config": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_point_id": {
													Type:     schema.TypeString,
													ForceNew: true,
													Optional: true,
												},
												"iam": {
													Type:     schema.TypeString,
													ForceNew: true,
													Optional: true,
													ValidateFunc: validation.StringInSlice([]string{
														ecs.EFSAuthorizationConfigIAMEnabled,
														ecs.EFSAuthorizationConfigIAMDisabled,
													}, false),
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Set: resourceAwsEcsTaskDefinitionVolumeHash,
			},

			"placement_constraints": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								ecs.TaskDefinitionPlacementConstraintTypeMemberOf,
							}, false),
						},
						"expression": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"requires_compatibilities": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"ipc_mode": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					ecs.IpcModeHost,
					ecs.IpcModeNone,
					ecs.IpcModeTask,
				}, false),
			},

			"pid_mode": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					ecs.PidModeHost,
					ecs.PidModeTask,
				}, false),
			},

			"proxy_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"properties": {
							Type:     schema.TypeMap,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Default:  ecs.ProxyConfigurationTypeAppmesh,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								ecs.ProxyConfigurationTypeAppmesh,
							}, false),
						},
					},
				},
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"inference_accelerator": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"device_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
		},
		CustomizeDiff: resourceAwsEcsTaskDefinitionCustomDiff,
	}
}

func resourceAwsEcsTaskDefinitionCustomDiff(d *schema.ResourceDiff, meta interface{}) error {
	for _, key := range [...]string{
		"cpu",
		"task_role_arn",
		"execution_role_arn",
		"memory",
		"network_mode",
		"volume",
		"placement_constraints",
		"requires_compatibilities",
		"ipc_mode",
		"pid_mode",
		"proxy_configuration",
	} {
		if d.HasChange(key) {
			log.Printf("[DEBUG] change to %s will trigger new revision/arn for resource %s", key, d.Id())
			err := d.SetNewComputed("arn")
			if err != nil {
				return err
			}
			return d.SetNewComputed("revision")
		}
	}

	// check for ECS container definitions changes
	networkMode, ok := d.GetOk("network_mode")
	isAWSVPC := ok && networkMode.(string) == ecs.NetworkModeAwsvpc
	old, new := d.GetChange("container_definitions")
	equal, _ := EcsContainerDefinitionsAreEquivalent(old.(string), new.(string), isAWSVPC)
	if !equal {
		log.Printf("[DEBUG] change to container_definitions will trigger new revision/arn for resource %s", d.Id())
		err := d.SetNewComputed("arn")
		if err != nil {
			return err
		}
		return d.SetNewComputed("revision")
	}
	return nil
}

func validateAwsEcsTaskDefinitionContainerDefinitions(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := expandEcsContainerDefinitions(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("ECS Task Definition container_definitions is invalid: %s", err))
	}
	return
}

func resourceAwsEcsTaskDefinitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	rawDefinitions := d.Get("container_definitions").(string)
	definitions, err := expandEcsContainerDefinitions(rawDefinitions)
	if err != nil {
		return err
	}

	input := ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: definitions,
		Family:               aws.String(d.Get("family").(string)),
	}

	// ClientException: Tags can not be empty.
	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().EcsTags()
	}

	if v, ok := d.GetOk("task_role_arn"); ok {
		input.TaskRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("execution_role_arn"); ok {
		input.ExecutionRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cpu"); ok {
		input.Cpu = aws.String(v.(string))
	}

	if v, ok := d.GetOk("memory"); ok {
		input.Memory = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_mode"); ok {
		input.NetworkMode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ipc_mode"); ok {
		input.IpcMode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pid_mode"); ok {
		input.PidMode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("volume"); ok {
		volumes := expandEcsVolumes(v.(*schema.Set).List())
		input.Volumes = volumes
	}

	if v, ok := d.GetOk("inference_accelerator"); ok {
		input.InferenceAccelerators = expandEcsInferenceAccelerators(v.(*schema.Set).List())
	}

	constraints := d.Get("placement_constraints").(*schema.Set).List()
	if len(constraints) > 0 {
		var pc []*ecs.TaskDefinitionPlacementConstraint
		for _, raw := range constraints {
			p := raw.(map[string]interface{})
			t := p["type"].(string)
			e := p["expression"].(string)
			if err := validateAwsEcsPlacementConstraint(t, e); err != nil {
				return err
			}
			pc = append(pc, &ecs.TaskDefinitionPlacementConstraint{
				Type:       aws.String(t),
				Expression: aws.String(e),
			})
		}
		input.PlacementConstraints = pc
	}

	if v, ok := d.GetOk("requires_compatibilities"); ok && v.(*schema.Set).Len() > 0 {
		input.RequiresCompatibilities = expandStringSet(v.(*schema.Set))
	}

	proxyConfigs := d.Get("proxy_configuration").([]interface{})
	if len(proxyConfigs) > 0 {
		proxyConfig := proxyConfigs[0]
		configMap := proxyConfig.(map[string]interface{})

		containerName := configMap["container_name"].(string)
		proxyType := configMap["type"].(string)

		rawProperties := configMap["properties"].(map[string]interface{})

		properties := make([]*ecs.KeyValuePair, len(rawProperties))
		i := 0
		for name, value := range rawProperties {
			properties[i] = &ecs.KeyValuePair{
				Name:  aws.String(name),
				Value: aws.String(value.(string)),
			}
			i++
		}

		var ecsProxyConfig ecs.ProxyConfiguration
		ecsProxyConfig.ContainerName = aws.String(containerName)
		ecsProxyConfig.Type = aws.String(proxyType)
		ecsProxyConfig.Properties = properties

		input.ProxyConfiguration = &ecsProxyConfig
	}

	log.Printf("[DEBUG] Registering ECS task definition: %s", input)
	out, err := conn.RegisterTaskDefinition(&input)
	if err != nil {
		return err
	}

	taskDefinition := *out.TaskDefinition

	log.Printf("[DEBUG] ECS task definition registered: %q (rev. %d)",
		aws.StringValue(taskDefinition.TaskDefinitionArn), aws.Int64Value(taskDefinition.Revision))

	d.SetId(aws.StringValue(taskDefinition.Family))
	d.Set("arn", taskDefinition.TaskDefinitionArn)

	return resourceAwsEcsTaskDefinitionRead(d, meta)
}

func resourceAwsEcsTaskDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading task definition %s", d.Id())
	out, err := conn.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(d.Get("family").(string)),
		Include:        []*string{aws.String(ecs.TaskDefinitionFieldTags)},
	})
	if err != nil {
		// If the task definition only has INACTIVE revisions, we will get a ClientException with a message "Unable to describe task definition."
		// This is the same response we would get if there was no task definition at all.
		if !isAWSErr(err, ecs.ErrCodeClientException, "Unable to describe task definition.") {
			return err
		}
		log.Printf("[DEBUG] Removing ECS task definition %s because it has no ACTIVE revisions", d.Get("family"))
		d.SetId("")
		return nil
	}
	log.Printf("[DEBUG] Received task definition %s, status:%s\n %s", aws.StringValue(out.TaskDefinition.Family),
		aws.StringValue(out.TaskDefinition.Status), out)

	taskDefinition := out.TaskDefinition

	if aws.StringValue(taskDefinition.Status) == ecs.TaskDefinitionStatusInactive {
		log.Printf("[DEBUG] Removing ECS task definition %s because it's INACTIVE", aws.StringValue(out.TaskDefinition.Family))
		d.SetId("")
		return nil
	}

	d.SetId(aws.StringValue(taskDefinition.Family))
	d.Set("arn", taskDefinition.TaskDefinitionArn)
	d.Set("family", taskDefinition.Family)
	d.Set("revision", taskDefinition.Revision)

	// Sort the lists of environment variables as they come in, so we won't get spurious reorderings in plans
	// (diff is suppressed if the environment variables haven't changed, but they still show in the plan if
	// some other property changes).
	containerDefinitions(taskDefinition.ContainerDefinitions).OrderEnvironmentVariables()

	defs, err := flattenEcsContainerDefinitions(taskDefinition.ContainerDefinitions)
	if err != nil {
		return err
	}
	err = d.Set("container_definitions", defs)
	if err != nil {
		return err
	}

	d.Set("task_role_arn", taskDefinition.TaskRoleArn)
	d.Set("execution_role_arn", taskDefinition.ExecutionRoleArn)
	d.Set("cpu", taskDefinition.Cpu)
	d.Set("memory", taskDefinition.Memory)
	d.Set("network_mode", taskDefinition.NetworkMode)
	d.Set("ipc_mode", taskDefinition.IpcMode)
	d.Set("pid_mode", taskDefinition.PidMode)

	tags := keyvaluetags.EcsKeyValueTags(out.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	if err := d.Set("volume", flattenEcsVolumes(taskDefinition.Volumes)); err != nil {
		return fmt.Errorf("error setting volume: %s", err)
	}

	if err := d.Set("inference_accelerator", flattenEcsInferenceAccelerators(taskDefinition.InferenceAccelerators)); err != nil {
		return fmt.Errorf("error setting inference accelerators: %s", err)
	}

	if err := d.Set("placement_constraints", flattenPlacementConstraints(taskDefinition.PlacementConstraints)); err != nil {
		log.Printf("[ERR] Error setting placement_constraints for (%s): %s", d.Id(), err)
	}

	if err := d.Set("requires_compatibilities", flattenStringList(taskDefinition.RequiresCompatibilities)); err != nil {
		return fmt.Errorf("error setting requires_compatibilities: %s", err)
	}

	if err := d.Set("proxy_configuration", flattenProxyConfiguration(taskDefinition.ProxyConfiguration)); err != nil {
		return fmt.Errorf("error setting proxy_configuration: %s", err)
	}

	return nil
}

func flattenPlacementConstraints(pcs []*ecs.TaskDefinitionPlacementConstraint) []map[string]interface{} {
	if len(pcs) == 0 {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, pc := range pcs {
		c := make(map[string]interface{})
		c["type"] = *pc.Type
		c["expression"] = *pc.Expression
		results = append(results, c)
	}
	return results
}

func flattenProxyConfiguration(pc *ecs.ProxyConfiguration) []map[string]interface{} {
	if pc == nil {
		return nil
	}

	meshProperties := make(map[string]string)
	if pc.Properties != nil {
		for _, prop := range pc.Properties {
			meshProperties[aws.StringValue(prop.Name)] = aws.StringValue(prop.Value)
		}
	}

	config := make(map[string]interface{})
	config["container_name"] = aws.StringValue(pc.ContainerName)
	config["type"] = aws.StringValue(pc.Type)
	config["properties"] = meshProperties

	return []map[string]interface{}{
		config,
	}
}

func resourceAwsEcsTaskDefinitionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	if d.HasChanges(
		"cpu",
		"task_role_arn",
		"execution_role_arn",
		"memory",
		"network_mode",
		"volume",
		"placement_constraints",
		"requires_compatibilities",
		"ipc_mode",
		"pid_mode",
		"proxy_configuration",
	) {
		return resourceAwsEcsTaskDefinitionCreate(d, meta)
	}

	// check for ECS container definitions changes
	networkMode, ok := d.GetOk("network_mode")
	isAWSVPC := ok && networkMode.(string) == ecs.NetworkModeAwsvpc
	old, new := d.GetChange("container_definitions")
	equal, _ := EcsContainerDefinitionsAreEquivalent(old.(string), new.(string), isAWSVPC)
	if !equal {
		return resourceAwsEcsTaskDefinitionCreate(d, meta)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.EcsUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating ECS Task Definition (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsEcsTaskDefinitionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	_, err := conn.DeregisterTaskDefinition(&ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: aws.String(d.Get("arn").(string)),
	})
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Task definition %q deregistered.", d.Get("arn").(string))

	return nil
}

func resourceAwsEcsTaskDefinitionVolumeHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["host_path"].(string)))

	if v, ok := m["efs_volume_configuration"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		m := v.([]interface{})[0].(map[string]interface{})

		if v, ok := m["file_system_id"]; ok && v.(string) != "" {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}

		if v, ok := m["root_directory"]; ok && v.(string) != "" {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}

		if v, ok := m["transit_encryption"]; ok && v.(string) != "" {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}
		if v, ok := m["transit_encryption_port"]; ok && v.(int) > 0 {
			buf.WriteString(fmt.Sprintf("%d-", v.(int)))
		}
		if v, ok := m["authorization_config"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			m := v.([]interface{})[0].(map[string]interface{})
			if v, ok := m["access_point_id"]; ok && v.(string) != "" {
				buf.WriteString(fmt.Sprintf("%s-", v.(string)))
			}
			if v, ok := m["iam"]; ok && v.(string) != "" {
				buf.WriteString(fmt.Sprintf("%s-", v.(string)))
			}
		}

	}

	return hashcode.String(buf.String())
}

func flattenEcsInferenceAccelerators(list []*ecs.InferenceAccelerator) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, iAcc := range list {
		l := map[string]interface{}{
			"device_name": *iAcc.DeviceName,
			"device_type": *iAcc.DeviceType,
		}

		result = append(result, l)
	}
	return result
}

func expandEcsInferenceAccelerators(configured []interface{}) []*ecs.InferenceAccelerator {
	iAccs := make([]*ecs.InferenceAccelerator, 0, len(configured))
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})
		l := &ecs.InferenceAccelerator{
			DeviceName: aws.String(data["device_name"].(string)),
			DeviceType: aws.String(data["device_type"].(string)),
		}
		iAccs = append(iAccs, l)
	}

	return iAccs
}
