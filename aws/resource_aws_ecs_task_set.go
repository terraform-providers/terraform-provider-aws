package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsEcsTaskSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEcsTaskSetCreate,
		Read:   resourceAwsEcsTaskSetRead,
		Update: resourceAwsEcsTaskSetUpdate,
		Delete: resourceAwsEcsTaskSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"client_token": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"cluster": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"external_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"launch_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ecs.LaunchTypeEc2,
					ecs.LaunchTypeFargate,
				}, false),
			},
			"load_balancers": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"container_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 65536),
						},
						"elb_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"target_group_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: resourceAwsEcsLoadBalancerHash,
			},
			"network_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"awsvpc_configuration": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"assign_public_ip": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "DISABLED",
										ValidateFunc: validation.StringInSlice([]string{
											ecs.AssignPublicIpEnabled,
											ecs.AssignPublicIpDisabled,
										}, false),
									},
									"security_groups": {
										Type:     schema.TypeSet,
										MaxItems: 5,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Set:      schema.HashString,
									},
									"subnets": {
										Type:     schema.TypeSet,
										MaxItems: 16,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Set:      schema.HashString,
									},
								},
							},
						},
					},
				},
			},
			"platform_version": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "LATEST",
			},
			"scale": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unit": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								ecs.ScaleUnitPercent,
							}, false),
						},
						"value": {
							Type:     schema.TypeFloat,
							Optional: true,
							ValidateFunc: func(i interface{}, k string) (s []string, es []error) {
								v, ok := i.(float64)
								if !ok {
									es = append(es, fmt.Errorf("expected type of %s to be float64", k))
									return
								}

								if v < 0.0 || v > 100.0 {
									es = append(es, fmt.Errorf("expected %s to be in the range (0.0 - 100.0), got %f", k, v))
									return
								}

								return
							},
						},
					},
				},
			},
			"service": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_registries": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"container_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 65536),
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 65536),
						},
						"registry_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"force_delete": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEcsTaskSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	cluser := d.Get("cluster").(string)
	service := d.Get("service").(string)
	input := &ecs.CreateTaskSetInput{
		Cluster:        aws.String(cluser),
		Service:        aws.String(service),
		TaskDefinition: aws.String(d.Get("task_definition").(string)),
	}

	if v, ok := d.GetOk("client_token"); ok {
		input.ClientToken = aws.String(v.(string))
	}

	if v, ok := d.GetOk("external_id"); ok {
		input.ExternalId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("launch_type"); ok {
		input.LaunchType = aws.String(v.(string))
	}

	loadBalancers := expandEcsLoadBalancers(d.Get("load_balancers").(*schema.Set).List())
	if len(loadBalancers) > 0 {
		input.LoadBalancers = loadBalancers
	}

	input.NetworkConfiguration = expandEcsNetworkConfiguration(d.Get("network_configuration").([]interface{}))

	if v, ok := d.GetOk("platform_version"); ok {
		input.PlatformVersion = aws.String(v.(string))
	}

	scale := d.Get("scale").([]interface{})
	if len(scale) > 0 {
		input.Scale = expandAwsEcsScale(scale[0].(map[string]interface{}))
	}

	serviceRegistries := d.Get("service_registries").([]interface{})
	if len(serviceRegistries) > 0 {
		input.ServiceRegistries = expandAwsEcsServiceRegistries(serviceRegistries)
	}

	resp, err := conn.CreateTaskSet(input)
	if err != nil {
		return fmt.Errorf("error creating ECS TaskSet: %s", err)
	}

	taskSet := resp.TaskSet
	if taskSet == nil {
		return fmt.Errorf("error creating ECS TaskSet: invalid response from AWS")
	}

	d.SetId(fmt.Sprintf("%s|%s|%s", cluser, service, *taskSet.Id))
	return resourceAwsEcsTaskSetRead(d, meta)
}

func resourceAwsEcsTaskSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	cluster, service, taskSetId, err := decodeEcsTaskSetId(d.Id())
	if err != nil {
		return fmt.Errorf("Error reading ECS TaskSet (%s): %s", d.Id(), err)
	}
	input := &ecs.DescribeTaskSetsInput{
		Cluster:  aws.String(cluster),
		Service:  aws.String(service),
		TaskSets: []*string{aws.String(taskSetId)},
	}

	resp, err := conn.DescribeTaskSets(input)
	if err != nil {
		if isAWSErr(err, ecs.ErrCodeClusterNotFoundException, "") {
			log.Printf("[WARN] ECS TaskSet (%s) not found because cluster(%s) isn't found , removing from state", d.Id(), cluster)
			d.SetId("")
			return nil
		}
		if isAWSErr(err, ecs.ErrCodeServiceNotFoundException, "") {
			log.Printf("[WARN] ECS TaskSet (%s) not found because service(%s) isn't found , removing from state", d.Id(), service)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading ECS TaskSet (%s): %s", d.Id(), err)
	}

	if len(resp.TaskSets) != 1 {
		return fmt.Errorf("error reading # of ECS TaskSet (%s) expected 1, got %d", d.Id(), len(resp.TaskSets))
	}

	taskSet := resp.TaskSets[0]
	if taskSet == nil {
		log.Printf("[WARN] ECS TaskSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", taskSet.TaskSetArn)
	d.Set("cluster", cluster)
	d.Set("service", service)
	return nil
}

func resourceAwsEcsTaskSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn
	return resourceAwsEcsTaskSetRead(d, meta)
}

func resourceAwsEcsTaskSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	cluster, service, taskSetId, err := decodeEcsTaskSetId(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting ECS TaskSet (%s): %s", d.Id(), err)
	}
	input := &ecs.DeleteTaskSetInput{
		Cluster: aws.String(cluster),
		Service: aws.String(service),
		TaskSet: aws.String(taskSetId),
	}

	if v, ok := d.GetOk("force_bool"); ok && v.(bool) {
		input.Force = aws.Bool(v.(bool))
	}

	_, err = conn.DeleteTaskSet(input)
	if err != nil {
		if isAWSErr(err, ecs.ErrCodeTaskSetNotFoundException, "") {
			return nil
		}
		return err
	}
	return nil
}

func decodeEcsTaskSetId(id string) (cluster string, service string, taskSetId string, e error) {
	ss := strings.Split(id, "|")
	if len(ss) != 3 {
		e = fmt.Errorf("invalid EcsTaskSet ID: %s", id)
		return
	}
	cluster, service, taskSetId = ss[0], ss[1], ss[2]
	return
}

func expandAwsEcsScale(d map[string]interface{}) *ecs.Scale {
	if len(d) == 0 {
		return nil
	}

	result := &ecs.Scale{}
	if v, ok := d["unit"]; ok && v.(string) != "" {
		result.Unit = aws.String(v.(string))
	}
	if v, ok := d["value"]; ok {
		result.Value = aws.Float64(v.(float64))
	}

	return result
}

func flattenAwsEcsScale(scale *ecs.Scale) []map[string]interface{} {
	if scale == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["unit"] = aws.StringValue(scale.Unit)
	m["value"] = aws.Float64Value(scale.Value)

	return []map[string]interface{}{m}
}

func expandAwsEcsServiceRegistries(d []interface{}) []*ecs.ServiceRegistry {
	if len(d) == 0 {
		return nil
	}

	result := make([]*ecs.ServiceRegistry, 0, len(d))
	for _, v := range d {
		m := v.(map[string]interface{})
		sr := &ecs.ServiceRegistry{}
		if raw, ok := m["container_name"]; ok && raw.(string) != "" {
			sr.ContainerName = aws.String(raw.(string))
		}
		if raw, ok := m["container_port"]; ok {
			sr.ContainerPort = aws.Int64(int64(raw.(int)))
		}
		if raw, ok := m["port"]; ok {
			sr.Port = aws.Int64(int64(raw.(int)))
		}
		if raw, ok := m["registry_arn"]; ok && raw.(string) != "" {
			sr.RegistryArn = aws.String(raw.(string))
		}
		result = append(result, sr)
	}

	return result
}

func flattenAwsEcsServiceRegistries(registories []*ecs.ServiceRegistry) []interface{} {
	if registories == nil || len(registories) == 0 {
		return nil
	}

	result := make([]interface{}, 0, len(registories))

	for _, v := range registories {
		m := make(map[string]interface{})
		m["container_name"] = aws.StringValue(v.ContainerName)
		m["container_port"] = aws.Int64Value(v.ContainerPort)
		m["port"] = aws.Int64Value(v.Port)
		m["registry_arn"] = aws.StringValue(v.RegistryArn)
		result = append(result, m)
	}

	return result
}
