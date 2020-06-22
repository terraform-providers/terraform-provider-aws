package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ecs/waiter"
)

func resourceAwsEcsCapacityProvider() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEcsCapacityProviderCreate,
		Read:   resourceAwsEcsCapacityProviderRead,
		Update: resourceAwsEcsCapacityProviderUpdate,
		Delete: resourceAwsEcsCapacityProviderDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsEcsCapacityProviderImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_scaling_group_provider": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_scaling_group_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
							ForceNew:     true,
						},
						"managed_termination_protection": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								ecs.ManagedTerminationProtectionEnabled,
								ecs.ManagedTerminationProtectionDisabled,
							}, false),
						},
						"managed_scaling": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"maximum_scaling_step_size": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 10000),
									},
									"minimum_scaling_step_size": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 10000),
									},
									"status": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
										ValidateFunc: validation.StringInSlice([]string{
											ecs.ManagedScalingStatusEnabled,
											ecs.ManagedScalingStatusDisabled,
										}, false)},
									"target_capacity": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ForceNew:     true,
										ValidateFunc: validation.IntBetween(1, 100),
									},
								},
							},
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsEcsCapacityProviderCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	input := ecs.CreateCapacityProviderInput{
		Name:                     aws.String(d.Get("name").(string)),
		AutoScalingGroupProvider: expandAutoScalingGroupProvider(d.Get("auto_scaling_group_provider")),
	}

	// `CreateCapacityProviderInput` does not accept an empty array of tags
	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().EcsTags()
	}

	out, err := conn.CreateCapacityProvider(&input)

	if err != nil {
		return fmt.Errorf("error creating capacity provider: %s", err)
	}

	provider := *out.CapacityProvider

	log.Printf("[DEBUG] ECS Capacity Provider created: %s", aws.StringValue(provider.CapacityProviderArn))
	d.SetId(aws.StringValue(provider.CapacityProviderArn))

	return resourceAwsEcsCapacityProviderRead(d, meta)
}

func resourceAwsEcsCapacityProviderRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &ecs.DescribeCapacityProvidersInput{
		CapacityProviders: []*string{aws.String(d.Id())},
		Include:           []*string{aws.String(ecs.CapacityProviderFieldTags)},
	}

	output, err := conn.DescribeCapacityProviders(input)

	if err != nil {
		return fmt.Errorf("error reading ECS Capacity Provider (%s): %s", d.Id(), err)
	}

	var provider *ecs.CapacityProvider
	for _, cp := range output.CapacityProviders {
		if aws.StringValue(cp.CapacityProviderArn) == d.Id() {
			provider = cp
			break
		}
	}

	if provider == nil {
		log.Printf("[WARN] ECS Capacity Provider (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(provider.Status) == ecs.CapacityProviderStatusInactive {
		log.Printf("[WARN] ECS Capacity Provider (%s) is INACTIVE, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", provider.CapacityProviderArn)
	d.Set("name", provider.Name)

	if err := d.Set("tags", keyvaluetags.EcsKeyValueTags(provider.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("auto_scaling_group_provider", flattenAutoScalingGroupProvider(provider.AutoScalingGroupProvider)); err != nil {
		return fmt.Errorf("error setting autoscaling group provider: %s", err)
	}

	return nil
}

func resourceAwsEcsCapacityProviderUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.EcsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating ECS Cluster (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsEcsCapacityProviderDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	input := &ecs.DeleteCapacityProviderInput{
		CapacityProvider: aws.String(d.Id()),
	}

	_, err := conn.DeleteCapacityProvider(input)

	if err != nil {
		return fmt.Errorf("error deleting ECS Capacity Provider (%s): %w", d.Id(), err)
	}

	if _, err := waiter.CapacityProviderInactive(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for ECS Capacity Provider (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func resourceAwsEcsCapacityProviderImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("name", d.Id())
	d.SetId(arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Service:   "ecs",
		Resource:  fmt.Sprintf("capacity-provider/%s", d.Id()),
	}.String())
	return []*schema.ResourceData{d}, nil
}

func expandAutoScalingGroupProvider(configured interface{}) *ecs.AutoScalingGroupProvider {
	if configured == nil {
		return nil
	}

	if configured.([]interface{}) == nil || len(configured.([]interface{})) == 0 {
		return nil
	}

	prov := ecs.AutoScalingGroupProvider{}
	p := configured.([]interface{})[0].(map[string]interface{})
	arn := p["auto_scaling_group_arn"].(string)
	prov.AutoScalingGroupArn = aws.String(arn)

	if mtp := p["managed_termination_protection"].(string); len(mtp) > 0 {
		prov.ManagedTerminationProtection = aws.String(mtp)
	}

	if v := p["managed_scaling"].([]interface{}); len(v) > 0 && v[0].(map[string]interface{}) != nil {
		ms := v[0].(map[string]interface{})
		managedScaling := ecs.ManagedScaling{}

		if val, ok := ms["maximum_scaling_step_size"].(int); ok && val != 0 {
			managedScaling.MaximumScalingStepSize = aws.Int64(int64(val))
		}
		if val, ok := ms["minimum_scaling_step_size"].(int); ok && val != 0 {
			managedScaling.MinimumScalingStepSize = aws.Int64(int64(val))
		}
		if val, ok := ms["status"].(string); ok && len(val) > 0 {
			managedScaling.Status = aws.String(val)
		}
		if val, ok := ms["target_capacity"].(int); ok && val != 0 {
			managedScaling.TargetCapacity = aws.Int64(int64(val))
		}
		prov.ManagedScaling = &managedScaling
	}

	return &prov
}

func flattenAutoScalingGroupProvider(provider *ecs.AutoScalingGroupProvider) []map[string]interface{} {
	if provider == nil {
		return nil
	}

	p := map[string]interface{}{
		"auto_scaling_group_arn":         aws.StringValue(provider.AutoScalingGroupArn),
		"managed_termination_protection": aws.StringValue(provider.ManagedTerminationProtection),
		"managed_scaling":                []map[string]interface{}{},
	}

	if provider.ManagedScaling != nil {
		m := map[string]interface{}{
			"maximum_scaling_step_size": aws.Int64Value(provider.ManagedScaling.MaximumScalingStepSize),
			"minimum_scaling_step_size": aws.Int64Value(provider.ManagedScaling.MinimumScalingStepSize),
			"status":                    aws.StringValue(provider.ManagedScaling.Status),
			"target_capacity":           aws.Int64Value(provider.ManagedScaling.TargetCapacity),
		}

		p["managed_scaling"] = []map[string]interface{}{m}
	}

	result := []map[string]interface{}{p}
	return result
}
