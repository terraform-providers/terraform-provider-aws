package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
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

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service": {
				Type:     schema.TypeString,
				Required: true,
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
				Computed: true,
			},

			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"network_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			// If you are using the CodeDeploy or an external deployment controller,
			// multiple target groups are not supported.
			// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/register-multiple-targetgroups.html
			"load_balancers": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"elb_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"target_group_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateArn,
						},
						"container_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"container_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsPortNumber,
						},
					},
				},
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
							ValidateFunc: validation.IsPortNumber,
						},
						"port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"registry_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},

			"launch_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					ecs.LaunchTypeEc2,
					ecs.LaunchTypeFargate,
				}, false),
			},

			"capacity_provider_strategy": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 100000),
							ForceNew:     true,
						},

						"capacity_provider": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 1000),
							ForceNew:     true,
						},
					},
				},
			},

			"platform_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"scale": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unit": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  ecs.ScaleUnitPercent,
							ValidateFunc: validation.StringInSlice([]string{
								ecs.ScaleUnitPercent,
							}, false),
						},
						"value": {
							Type:         schema.TypeFloat,
							Optional:     true,
							ValidateFunc: validation.FloatBetween(0.0, 100.0),
						},
					},
				},
			},

			"force_delete": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"wait_until_stable": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"wait_until_stable_timeout": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					duration, err := time.ParseDuration(value)
					if err != nil {
						errors = append(errors, fmt.Errorf(
							"%q cannot be parsed as a duration: %s", k, err))
					}
					if duration < 0 {
						errors = append(errors, fmt.Errorf(
							"%q must be greater than zero", k))
					}
					return
				},
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsEcsTaskSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	cluster := d.Get("cluster").(string)
	service := d.Get("service").(string)
	input := ecs.CreateTaskSetInput{
		ClientToken:    aws.String(resource.UniqueId()),
		Cluster:        aws.String(cluster),
		Service:        aws.String(service),
		TaskDefinition: aws.String(d.Get("task_definition").(string)),
		Tags:           keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().EcsTags(),
	}

	if v, ok := d.GetOk("external_id"); ok {
		input.ExternalId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("launch_type"); ok {
		input.LaunchType = aws.String(v.(string))
	}

	input.CapacityProviderStrategy = expandEcsCapacityProviderStrategy(d.Get("capacity_provider_strategy").(*schema.Set))

	loadBalancers := expandEcsLoadBalancers(d.Get("load_balancers").([]interface{}))
	if len(loadBalancers) > 0 {
		log.Printf("[DEBUG] Adding ECS load balancers: %s", loadBalancers)
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

	log.Printf("[DEBUG] Creating ECS Task set: %s", input)

	// Retry due to AWS IAM & ECS eventual consistency
	var out *ecs.CreateTaskSetOutput
	var err error
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		out, err = conn.CreateTaskSet(&input)

		if err != nil {
			if isAWSErr(err, ecs.ErrCodeClusterNotFoundException, "") ||
				isAWSErr(err, ecs.ErrCodeServiceNotFoundException, "") ||
				isAWSErr(err, ecs.ErrCodeTaskSetNotFoundException, "") ||
				isAWSErr(err, ecs.ErrCodeInvalidParameterException, "does not have an associated load balancer") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		out, err = conn.CreateTaskSet(&input)
	}

	if err != nil {
		return fmt.Errorf("Error creating ECS TaskSet: %s", err)
	}

	taskSet := *out.TaskSet

	log.Printf("[DEBUG] ECS Task set created: %s", aws.StringValue(taskSet.Id))
	d.SetId(aws.StringValue(taskSet.Id))

	if d.Get("wait_until_stable").(bool) {
		waitUntilStableTimeOut := d.Timeout(schema.TimeoutCreate)
		if v, ok := d.GetOk("wait_until_stable_timeout"); ok && v.(string) != "" {
			timeout, err := time.ParseDuration(v.(string))
			if err != nil {
				return err
			}
			waitUntilStableTimeOut = timeout
		}

		// Wait until it's stable
		wait := resource.StateChangeConf{
			Pending: []string{ecs.StabilityStatusStabilizing},
			Target:  []string{ecs.StabilityStatusSteadyState},
			Timeout: waitUntilStableTimeOut,
			Delay:   10 * time.Second,
			Refresh: func() (interface{}, string, error) {
				log.Printf("[DEBUG] Checking if ECS task set %s is set to %s", d.Id(), ecs.StabilityStatusSteadyState)
				resp, err := conn.DescribeTaskSets(&ecs.DescribeTaskSetsInput{
					TaskSets: []*string{aws.String(d.Id())},
					Cluster:  aws.String(d.Get("cluster").(string)),
					Service:  aws.String(d.Get("service").(string)),
				})
				if err != nil {
					return resp, "FAILED", err
				}

				log.Printf("[DEBUG] ECS task set (%s) is currently %s", d.Id(), aws.StringValue(resp.TaskSets[0].StabilityStatus))
				return resp, aws.StringValue(resp.TaskSets[0].StabilityStatus), nil
			},
		}

		_, err = wait.WaitForState()
		if err != nil {
			return err
		}
	}

	return resourceAwsEcsTaskSetRead(d, meta)
}

func resourceAwsEcsTaskSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	log.Printf("[DEBUG] Reading ECS task set %s", d.Id())

	cluster := d.Get("cluster").(string)
	service := d.Get("service").(string)
	input := ecs.DescribeTaskSetsInput{
		Cluster:  aws.String(cluster),
		Service:  aws.String(service),
		TaskSets: []*string{aws.String(d.Id())},
	}

	var out *ecs.DescribeTaskSetsOutput
	err := resource.Retry(d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		var err error
		out, err = conn.DescribeTaskSets(&input)
		if err != nil {
			if d.IsNewResource() &&
				isAWSErr(err, ecs.ErrCodeServiceNotFoundException, "") ||
				isAWSErr(err, ecs.ErrCodeClusterNotFoundException, "") ||
				isAWSErr(err, ecs.ErrCodeTaskSetNotFoundException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		if len(out.TaskSets) < 1 {
			if d.IsNewResource() {
				return resource.RetryableError(fmt.Errorf("ECS task set not created yet: %q", d.Id()))
			}
			log.Printf("[WARN] ECS Task Set %s not found, removing from state.", d.Id())
			d.SetId("")
			return nil
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		out, err = conn.DescribeTaskSets(&input)
	}

	// after retrying
	if err != nil {
		if isAWSErr(err, ecs.ErrCodeClusterNotFoundException, "") ||
			isAWSErr(err, ecs.ErrCodeServiceNotFoundException, "") ||
			isAWSErr(err, ecs.ErrCodeTaskSetNotFoundException, "") {
			log.Printf("[WARN] ECS TaskSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if len(out.TaskSets) < 1 {
		if d.IsNewResource() {
			return fmt.Errorf("ECS TaskSet not created: %q", d.Id())
		}
		log.Printf("[WARN] Removing ECS task set %s because it's gone", d.Id())
		d.SetId("")
		return nil
	}

	if len(out.TaskSets) != 1 {
		return fmt.Errorf("Error reading # of ECS TaskSet (%s) expected 1, got %d", d.Id(), len(out.TaskSets))
	}

	taskSet := out.TaskSets[0]

	log.Printf("[DEBUG] Received ECS task set %s", taskSet)

	d.SetId(aws.StringValue(taskSet.Id))
	d.Set("arn", taskSet.TaskSetArn)
	d.Set("launch_type", taskSet.LaunchType)
	d.Set("platform_version", taskSet.PlatformVersion)
	d.Set("external_id", taskSet.ExternalId)

	// Save cluster in the same format
	if strings.HasPrefix(d.Get("cluster").(string), "arn:"+meta.(*AWSClient).partition+":ecs:") {
		d.Set("cluster", taskSet.ClusterArn)
	} else {
		clusterARN := getNameFromARN(*taskSet.ClusterArn)
		d.Set("cluster", clusterARN)
	}

	// Save task definition in the same format
	if strings.HasPrefix(d.Get("task_definition").(string), "arn:"+meta.(*AWSClient).partition+":ecs:") {
		d.Set("task_definition", taskSet.TaskDefinition)
	} else {
		taskDefinition := buildFamilyAndRevisionFromARN(*taskSet.TaskDefinition)
		d.Set("task_definition", taskDefinition)
	}

	if taskSet.LoadBalancers != nil {
		d.Set("load_balancers", flattenEcsLoadBalancers(taskSet.LoadBalancers))
	}

	if err := d.Set("scale", flattenAwsEcsScale(taskSet.Scale)); err != nil {
		return fmt.Errorf("Error setting scale for (%s): %s", d.Id(), err)
	}

	if err := d.Set("capacity_provider_strategy", flattenEcsCapacityProviderStrategy(taskSet.CapacityProviderStrategy)); err != nil {
		return fmt.Errorf("error setting capacity_provider_strategy: %s", err)
	}

	if err := d.Set("network_configuration", flattenEcsNetworkConfiguration(taskSet.NetworkConfiguration)); err != nil {
		return fmt.Errorf("Error setting network_configuration for (%s): %s", d.Id(), err)
	}

	if err := d.Set("service_registries", flattenServiceRegistries(taskSet.ServiceRegistries)); err != nil {
		return fmt.Errorf("Error setting service_registries for (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsEcsTaskSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn
	updateTaskset := false

	input := ecs.UpdateTaskSetInput{
		Cluster: aws.String(d.Get("cluster").(string)),
		Service: aws.String(d.Get("service").(string)),
		TaskSet: aws.String(d.Id()),
	}

	if d.HasChange("scale") {
		scale := d.Get("scale").([]interface{})
		if len(scale) > 0 {
			updateTaskset = true
			input.Scale = expandAwsEcsScale(scale[0].(map[string]interface{}))
		}
	}

	if updateTaskset {
		log.Printf("[DEBUG] Updating ECS Task Set (%s): %s", d.Id(), input)
		// Retry due to IAM eventual consistency
		err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
			_, err := conn.UpdateTaskSet(&input)
			if err != nil {
				if isAWSErr(err, ecs.ErrCodeClusterNotFoundException, "") ||
					isAWSErr(err, ecs.ErrCodeServiceNotFoundException, "") ||
					isAWSErr(err, ecs.ErrCodeTaskSetNotFoundException, "") ||
					isAWSErr(err, ecs.ErrCodeInvalidParameterException, "does not have an associated load balancer") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if isResourceTimeoutError(err) {
			_, err = conn.UpdateTaskSet(&input)
		}
		if err != nil {
			return fmt.Errorf("Error updating ECS Task set (%s): %s", d.Id(), err)
		}

		if d.Get("wait_until_stable").(bool) {
			waitUntilStableTimeOut := d.Timeout(schema.TimeoutUpdate)
			if v, ok := d.GetOk("wait_until_stable_timeout"); ok && v.(string) != "" {
				timeout, err := time.ParseDuration(v.(string))
				if err != nil {
					return err
				}
				waitUntilStableTimeOut = timeout
			}

			// Wait until it's stable
			wait := resource.StateChangeConf{
				Pending: []string{ecs.StabilityStatusStabilizing},
				Target:  []string{ecs.StabilityStatusSteadyState},
				Timeout: waitUntilStableTimeOut,
				Delay:   10 * time.Second,
				Refresh: func() (interface{}, string, error) {
					log.Printf("[DEBUG] Checking if ECS task set %s is set to %s", d.Id(), ecs.StabilityStatusSteadyState)
					resp, err := conn.DescribeTaskSets(&ecs.DescribeTaskSetsInput{
						TaskSets: []*string{aws.String(d.Id())},
						Cluster:  aws.String(d.Get("cluster").(string)),
						Service:  aws.String(d.Get("service").(string)),
					})
					if err != nil {
						return resp, "FAILED", err
					}

					log.Printf("[DEBUG] ECS task set (%s) is currently %q", d.Id(), *resp.TaskSets[0].StabilityStatus)
					return resp, *resp.TaskSets[0].StabilityStatus, nil
				},
			}

			_, err = wait.WaitForState()
			if err != nil {
				return err
			}
		}

	}

	return resourceAwsEcsTaskSetRead(d, meta)
}

func resourceAwsEcsTaskSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	// Check if it's not already gone
	resp, err := conn.DescribeTaskSets(&ecs.DescribeTaskSetsInput{
		TaskSets: []*string{aws.String(d.Id())},
		Service:  aws.String(d.Get("service").(string)),
		Cluster:  aws.String(d.Get("cluster").(string)),
	})

	if err != nil {
		if isAWSErr(err, ecs.ErrCodeTaskSetNotFoundException, "") {
			log.Printf("[DEBUG] Removing ECS Task set from state, %q is already gone", d.Id())
			return nil
		}
		return err
	}

	if len(resp.TaskSets) == 0 {
		log.Printf("[DEBUG] Removing ECS Task set from state, %q is already gone", d.Id())
		return nil
	}

	log.Printf("[DEBUG] ECS TaskSet %s is currently %s", d.Id(), aws.StringValue(resp.TaskSets[0].Status))

	input := ecs.DeleteTaskSetInput{
		Cluster: aws.String(d.Get("cluster").(string)),
		Service: aws.String(d.Get("service").(string)),
		TaskSet: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("force_delete"); ok && v.(bool) {
		input.Force = aws.Bool(v.(bool))
	}

	// Wait until the ECS task set is drained
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		log.Printf("[DEBUG] Trying to delete ECS task set %s", input)
		_, err := conn.DeleteTaskSet(&input)
		if err != nil {
			if isAWSErr(err, ecs.ErrCodeTaskSetNotFoundException, "") {
				return nil
			}
			if isAWSErr(err, ecs.ErrCodeInvalidParameterException, "The service cannot be stopped while deployments are active.") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteTaskSet(&input)
	}

	if err != nil {
		if isAWSErr(err, ecs.ErrCodeTaskSetNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting ECS task set: %s", err)
	}

	// Wait until it's deleted
	wait := resource.StateChangeConf{
		Pending: []string{"ACTIVE", "PRIMARY", "DRAINING"},
		Target:  []string{"INACTIVE"},
		Timeout: d.Timeout(schema.TimeoutDelete),
		Refresh: func() (interface{}, string, error) {
			log.Printf("[DEBUG] Checking if ECS task set %s is INACTIVE", d.Id())
			resp, err := conn.DescribeTaskSets(&ecs.DescribeTaskSetsInput{
				TaskSets: []*string{aws.String(d.Id())},
				Cluster:  aws.String(d.Get("cluster").(string)),
				Service:  aws.String(d.Get("service").(string)),
			})

			if err != nil {
				return resp, "FAILED", err
			}

			// task set is already gone
			if len(resp.TaskSets) == 0 {
				return resp, "INACTIVE", nil
			}

			log.Printf("[DEBUG] ECS task set (%s) is currently %s", d.Id(), aws.StringValue(resp.TaskSets[0].Status))
			return resp, aws.StringValue(resp.TaskSets[0].Status), nil
		},
	}

	_, err = wait.WaitForState()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] ECS TaskSet %s deleted.", d.Id())
	return nil
}

func expandAwsEcsServiceRegistries(d []interface{}) []*ecs.ServiceRegistry {
	if len(d) == 0 {
		return nil
	}

	result := make([]*ecs.ServiceRegistry, 0, len(d))
	for _, v := range d {
		m := v.(map[string]interface{})
		sr := &ecs.ServiceRegistry{
			RegistryArn: aws.String(m["registry_arn"].(string)),
		}
		if raw, ok := m["container_name"].(string); ok && raw != "" {
			sr.ContainerName = aws.String(raw)
		}
		if raw, ok := m["container_port"].(int); ok && raw != 0 {
			sr.ContainerPort = aws.Int64(int64(raw))
		}
		if raw, ok := m["port"].(int); ok && raw != 0 {
			sr.Port = aws.Int64(int64(raw))
		}
		result = append(result, sr)
	}

	return result
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
