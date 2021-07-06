package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
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
				d.Set("arn", d.Id())

				idErr := fmt.Errorf("Expected ID in format of arn:PARTITION:ecs:REGION:ACCOUNTID:task-definition/FAMILY:REVISION and provided: %s", d.Id())
				resARN, err := arn.Parse(d.Id())
				if err != nil {
					return nil, idErr
				}
				familyRevision := strings.TrimPrefix(resARN.Resource, "task-definition/")
				familyRevisionParts := strings.Split(familyRevision, ":")
				if len(familyRevisionParts) != 2 {
					return nil, idErr
				}
				d.SetId(familyRevisionParts[0])

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
				ForceNew: true,
			},

			"family": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"container_definitions": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			"ephemeral_storage": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size_in_gib": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(21, 200),
						},
					},
				},
			},
			"task_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"execution_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"memory": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"network_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ecs.NetworkMode_Values(), false),
			},

			"volume": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"host_path": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},

						"docker_volume_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"scope": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(ecs.Scope_Values(), false),
									},
									"autoprovision": {
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
										Default:  false,
									},
									"driver": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
									"driver_opts": {
										Type:     schema.TypeMap,
										Elem:     &schema.Schema{Type: schema.TypeString},
										ForceNew: true,
										Optional: true,
									},
									"labels": {
										Type:     schema.TypeMap,
										Elem:     &schema.Schema{Type: schema.TypeString},
										ForceNew: true,
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
										Type:         schema.TypeString,
										ForceNew:     true,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(ecs.EFSTransitEncryption_Values(), false),
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
													Type:         schema.TypeString,
													ForceNew:     true,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(ecs.EFSAuthorizationConfigIAM_Values(), false),
												},
											},
										},
									},
								},
							},
						},
						"fsx_windows_file_server_volume_configuration": {
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
										Required: true,
									},
									"authorization_config": {
										Type:     schema.TypeList,
										Required: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"credentials_parameter": {
													Type:         schema.TypeString,
													ForceNew:     true,
													Required:     true,
													ValidateFunc: validateArn,
												},
												"domain": {
													Type:     schema.TypeString,
													ForceNew: true,
													Required: true,
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
				ForceNew: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ecs.TaskDefinitionPlacementConstraintType_Values(), false),
						},
						"expression": {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
						},
					},
				},
			},

			"requires_compatibilities": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"EC2",
						"FARGATE",
						"EXTERNAL",
					}, false),
				},
			},

			"ipc_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ecs.IpcMode_Values(), false),
			},

			"pid_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ecs.PidMode_Values(), false),
			},

			"proxy_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"properties": {
							Type:     schema.TypeMap,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
							ForceNew: true,
						},
						"type": {
							Type:         schema.TypeString,
							Default:      ecs.ProxyConfigurationTypeAppmesh,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(ecs.ProxyConfigurationType_Values(), false),
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
	}
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
		cons, err := expandEcsTaskDefinitionPlacementConstraints(constraints)
		if err != nil {
			return err
		}
		input.PlacementConstraints = cons
	}

	if v, ok := d.GetOk("requires_compatibilities"); ok && v.(*schema.Set).Len() > 0 {
		input.RequiresCompatibilities = expandStringSet(v.(*schema.Set))
	}

	proxyConfigs := d.Get("proxy_configuration").([]interface{})
	if len(proxyConfigs) > 0 {
		input.ProxyConfiguration = expandEcsTaskDefinitionProxyConfiguration(proxyConfigs)
	}

	if v, ok := d.GetOk("ephemeral_storage"); ok && len(v.([]interface{})) > 0 {
		input.EphemeralStorage = expandEcsTaskDefinitionEphemeralStorage(v.([]interface{}))
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
		TaskDefinition: aws.String(d.Get("arn").(string)),
		Include:        []*string{aws.String(ecs.TaskDefinitionFieldTags)},
	})
	if err != nil {
		return err
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
		return fmt.Errorf("error setting volume: %w", err)
	}

	if err := d.Set("inference_accelerator", flattenEcsInferenceAccelerators(taskDefinition.InferenceAccelerators)); err != nil {
		return fmt.Errorf("error setting inference accelerators: %w", err)
	}

	if err := d.Set("placement_constraints", flattenPlacementConstraints(taskDefinition.PlacementConstraints)); err != nil {
		return fmt.Errorf("error setting placement_constraints: %w", err)
	}

	if err := d.Set("requires_compatibilities", flattenStringList(taskDefinition.RequiresCompatibilities)); err != nil {
		return fmt.Errorf("error setting requires_compatibilities: %w", err)
	}

	if err := d.Set("proxy_configuration", flattenProxyConfiguration(taskDefinition.ProxyConfiguration)); err != nil {
		return fmt.Errorf("error setting proxy_configuration: %w", err)
	}

	if err := d.Set("ephemeral_storage", flattenEcsTaskDefinitionEphemeralStorage(taskDefinition.EphemeralStorage)); err != nil {
		return fmt.Errorf("error setting ephemeral_storage: %w", err)
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
		c["type"] = aws.StringValue(pc.Type)
		c["expression"] = aws.StringValue(pc.Expression)
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

	if v, ok := m["fsx_windows_file_server_volume_configuration"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		m := v.([]interface{})[0].(map[string]interface{})

		if v, ok := m["file_system_id"]; ok && v.(string) != "" {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}

		if v, ok := m["root_directory"]; ok && v.(string) != "" {
			buf.WriteString(fmt.Sprintf("%s-", v.(string)))
		}

		if v, ok := m["authorization_config"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			m := v.([]interface{})[0].(map[string]interface{})
			if v, ok := m["credentials_parameter"]; ok && v.(string) != "" {
				buf.WriteString(fmt.Sprintf("%s-", v.(string)))
			}
			if v, ok := m["domain"]; ok && v.(string) != "" {
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
			"device_name": aws.StringValue(iAcc.DeviceName),
			"device_type": aws.StringValue(iAcc.DeviceType),
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

func expandEcsTaskDefinitionPlacementConstraints(constraints []interface{}) ([]*ecs.TaskDefinitionPlacementConstraint, error) {
	var pc []*ecs.TaskDefinitionPlacementConstraint
	for _, raw := range constraints {
		p := raw.(map[string]interface{})
		t := p["type"].(string)
		e := p["expression"].(string)
		if err := validateAwsEcsPlacementConstraint(t, e); err != nil {
			return nil, err
		}
		pc = append(pc, &ecs.TaskDefinitionPlacementConstraint{
			Type:       aws.String(t),
			Expression: aws.String(e),
		})
	}

	return pc, nil
}

func expandEcsTaskDefinitionProxyConfiguration(proxyConfigs []interface{}) *ecs.ProxyConfiguration {
	proxyConfig := proxyConfigs[0]
	configMap := proxyConfig.(map[string]interface{})

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

	ecsProxyConfig := &ecs.ProxyConfiguration{
		ContainerName: aws.String(configMap["container_name"].(string)),
		Type:          aws.String(configMap["type"].(string)),
		Properties:    properties,
	}

	return ecsProxyConfig
}

func expandEcsVolumes(configured []interface{}) []*ecs.Volume {
	volumes := make([]*ecs.Volume, 0, len(configured))

	// Loop over our configured volumes and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := &ecs.Volume{
			Name: aws.String(data["name"].(string)),
		}

		hostPath := data["host_path"].(string)
		if hostPath != "" {
			l.Host = &ecs.HostVolumeProperties{
				SourcePath: aws.String(hostPath),
			}
		}

		if v, ok := data["docker_volume_configuration"].([]interface{}); ok && len(v) > 0 {
			l.DockerVolumeConfiguration = expandEcsVolumesDockerVolume(v)
		}

		if v, ok := data["efs_volume_configuration"].([]interface{}); ok && len(v) > 0 {
			l.EfsVolumeConfiguration = expandEcsVolumesEFSVolume(v)
		}

		if v, ok := data["fsx_windows_file_server_volume_configuration"].([]interface{}); ok && len(v) > 0 {
			l.FsxWindowsFileServerVolumeConfiguration = expandEcsVolumesFsxWinVolume(v)
		}

		volumes = append(volumes, l)
	}

	return volumes
}

func expandEcsVolumesDockerVolume(configList []interface{}) *ecs.DockerVolumeConfiguration {
	config := configList[0].(map[string]interface{})
	dockerVol := &ecs.DockerVolumeConfiguration{}

	if v, ok := config["scope"].(string); ok && v != "" {
		dockerVol.Scope = aws.String(v)
	}

	if v, ok := config["autoprovision"]; ok && v != "" {
		scope := dockerVol.Scope
		if scope == nil || *scope != ecs.ScopeTask || v.(bool) {
			dockerVol.Autoprovision = aws.Bool(v.(bool))
		}
	}

	if v, ok := config["driver"].(string); ok && v != "" {
		dockerVol.Driver = aws.String(v)
	}

	if v, ok := config["driver_opts"].(map[string]interface{}); ok && len(v) > 0 {
		dockerVol.DriverOpts = expandStringMap(v)
	}

	if v, ok := config["labels"].(map[string]interface{}); ok && len(v) > 0 {
		dockerVol.Labels = expandStringMap(v)
	}

	return dockerVol
}

func expandEcsVolumesEFSVolume(efsConfig []interface{}) *ecs.EFSVolumeConfiguration {
	config := efsConfig[0].(map[string]interface{})
	efsVol := &ecs.EFSVolumeConfiguration{}

	if v, ok := config["file_system_id"].(string); ok && v != "" {
		efsVol.FileSystemId = aws.String(v)
	}

	if v, ok := config["root_directory"].(string); ok && v != "" {
		efsVol.RootDirectory = aws.String(v)
	}
	if v, ok := config["transit_encryption"].(string); ok && v != "" {
		efsVol.TransitEncryption = aws.String(v)
	}

	if v, ok := config["transit_encryption_port"].(int); ok && v > 0 {
		efsVol.TransitEncryptionPort = aws.Int64(int64(v))
	}
	if v, ok := config["authorization_config"].([]interface{}); ok && len(v) > 0 {
		efsVol.RootDirectory = nil
		efsVol.AuthorizationConfig = expandEcsVolumesEFSVolumeAuthorizationConfig(v)
	}

	return efsVol
}

func expandEcsVolumesEFSVolumeAuthorizationConfig(efsConfig []interface{}) *ecs.EFSAuthorizationConfig {
	authconfig := efsConfig[0].(map[string]interface{})
	auth := &ecs.EFSAuthorizationConfig{}

	if v, ok := authconfig["access_point_id"].(string); ok && v != "" {
		auth.AccessPointId = aws.String(v)
	}

	if v, ok := authconfig["iam"].(string); ok && v != "" {
		auth.Iam = aws.String(v)
	}

	return auth
}

func expandEcsVolumesFsxWinVolume(fsxWinConfig []interface{}) *ecs.FSxWindowsFileServerVolumeConfiguration {
	config := fsxWinConfig[0].(map[string]interface{})
	fsxVol := &ecs.FSxWindowsFileServerVolumeConfiguration{}

	if v, ok := config["file_system_id"].(string); ok && v != "" {
		fsxVol.FileSystemId = aws.String(v)
	}

	if v, ok := config["root_directory"].(string); ok && v != "" {
		fsxVol.RootDirectory = aws.String(v)
	}

	if v, ok := config["authorization_config"].([]interface{}); ok && len(v) > 0 {
		fsxVol.AuthorizationConfig = expandEcsVolumesFsxWinVolumeAuthorizationConfig(v)
	}

	return fsxVol
}

func expandEcsVolumesFsxWinVolumeAuthorizationConfig(config []interface{}) *ecs.FSxWindowsFileServerAuthorizationConfig {
	authconfig := config[0].(map[string]interface{})
	auth := &ecs.FSxWindowsFileServerAuthorizationConfig{}

	if v, ok := authconfig["credentials_parameter"].(string); ok && v != "" {
		auth.CredentialsParameter = aws.String(v)
	}

	if v, ok := authconfig["domain"].(string); ok && v != "" {
		auth.Domain = aws.String(v)
	}

	return auth
}

func flattenEcsVolumes(list []*ecs.Volume) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, volume := range list {
		l := map[string]interface{}{
			"name": aws.StringValue(volume.Name),
		}

		if volume.Host != nil && volume.Host.SourcePath != nil {
			l["host_path"] = aws.StringValue(volume.Host.SourcePath)
		}

		if volume.DockerVolumeConfiguration != nil {
			l["docker_volume_configuration"] = flattenDockerVolumeConfiguration(volume.DockerVolumeConfiguration)
		}

		if volume.EfsVolumeConfiguration != nil {
			l["efs_volume_configuration"] = flattenEFSVolumeConfiguration(volume.EfsVolumeConfiguration)
		}

		if volume.FsxWindowsFileServerVolumeConfiguration != nil {
			l["fsx_windows_file_server_volume_configuration"] = flattenFsxWinVolumeConfiguration(volume.FsxWindowsFileServerVolumeConfiguration)
		}

		result = append(result, l)
	}
	return result
}

func flattenDockerVolumeConfiguration(config *ecs.DockerVolumeConfiguration) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})

	if v := config.Scope; v != nil {
		m["scope"] = aws.StringValue(v)
	}

	if v := config.Autoprovision; v != nil {
		m["autoprovision"] = aws.BoolValue(v)
	}

	if v := config.Driver; v != nil {
		m["driver"] = aws.StringValue(v)
	}

	if config.DriverOpts != nil {
		m["driver_opts"] = pointersMapToStringList(config.DriverOpts)
	}

	if v := config.Labels; v != nil {
		m["labels"] = pointersMapToStringList(v)
	}

	items = append(items, m)
	return items
}

func flattenEFSVolumeConfiguration(config *ecs.EFSVolumeConfiguration) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})
	if config != nil {
		if v := config.FileSystemId; v != nil {
			m["file_system_id"] = aws.StringValue(v)
		}

		if v := config.RootDirectory; v != nil {
			m["root_directory"] = aws.StringValue(v)
		}
		if v := config.TransitEncryption; v != nil {
			m["transit_encryption"] = aws.StringValue(v)
		}

		if v := config.TransitEncryptionPort; v != nil {
			m["transit_encryption_port"] = int(aws.Int64Value(v))
		}

		if v := config.AuthorizationConfig; v != nil {
			m["authorization_config"] = flattenEFSVolumeAuthorizationConfig(v)
		}
	}

	items = append(items, m)
	return items
}

func flattenEFSVolumeAuthorizationConfig(config *ecs.EFSAuthorizationConfig) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})
	if config != nil {
		if v := config.AccessPointId; v != nil {
			m["access_point_id"] = aws.StringValue(v)
		}
		if v := config.Iam; v != nil {
			m["iam"] = aws.StringValue(v)
		}
	}

	items = append(items, m)
	return items
}

func flattenFsxWinVolumeConfiguration(config *ecs.FSxWindowsFileServerVolumeConfiguration) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})
	if config != nil {
		if v := config.FileSystemId; v != nil {
			m["file_system_id"] = aws.StringValue(v)
		}

		if v := config.RootDirectory; v != nil {
			m["root_directory"] = aws.StringValue(v)
		}

		if v := config.AuthorizationConfig; v != nil {
			m["authorization_config"] = flattenFsxWinVolumeAuthorizationConfig(v)
		}
	}

	items = append(items, m)
	return items
}

func flattenFsxWinVolumeAuthorizationConfig(config *ecs.FSxWindowsFileServerAuthorizationConfig) []interface{} {
	var items []interface{}
	m := make(map[string]interface{})
	if config != nil {
		if v := config.CredentialsParameter; v != nil {
			m["credentials_parameter"] = aws.StringValue(v)
		}
		if v := config.Domain; v != nil {
			m["domain"] = aws.StringValue(v)
		}
	}

	items = append(items, m)
	return items
}

func flattenEcsContainerDefinitions(definitions []*ecs.ContainerDefinition) (string, error) {
	b, err := jsonutil.BuildJSON(definitions)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func expandEcsContainerDefinitions(rawDefinitions string) ([]*ecs.ContainerDefinition, error) {
	var definitions []*ecs.ContainerDefinition

	err := json.Unmarshal([]byte(rawDefinitions), &definitions)
	if err != nil {
		return nil, fmt.Errorf("Error decoding JSON: %s", err)
	}

	return definitions, nil
}

func expandEcsTaskDefinitionEphemeralStorage(config []interface{}) *ecs.EphemeralStorage {
	configMap := config[0].(map[string]interface{})

	es := &ecs.EphemeralStorage{
		SizeInGiB: aws.Int64(int64(configMap["size_in_gib"].(int))),
	}

	return es
}

func flattenEcsTaskDefinitionEphemeralStorage(pc *ecs.EphemeralStorage) []map[string]interface{} {
	if pc == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["size_in_gib"] = aws.Int64Value(pc.SizeInGiB)

	return []map[string]interface{}{m}
}
