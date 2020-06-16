package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAwsAppautoscalingPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAppautoscalingPolicyRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// https://github.com/boto/botocore/blob/9f322b1/botocore/data/autoscaling/2011-01-01/service-2.json#L1862-L1873
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"scalable_dimension": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_namespace": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAwsAppautoscalingPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn
	d.SetId(time.Now().UTC().String())

	policyName := d.Get("name").(string)
	serviceNamespace := d.Get("service_namespace").(string)

	input := &applicationautoscaling.DescribeScalingPoliciesInput{
		ServiceNamespace: aws.String(serviceNamespace),
		PolicyNames: []*string{
			aws.String(policyName),
		},
	}

	log.Printf("[DEBUG] Reading Autoscaling Policy: %s", input)

	result, err := conn.DescribeScalingPolicies(input)

	log.Printf("[DEBUG] Checking for error: %s", err)

	if err != nil {
		return fmt.Errorf("error describing Autoscaling Policy: %s", err)
	}

	log.Printf("[DEBUG] Found Autoscaling Policy: %s", result)

	if len(result.ScalingPolicies) < 1 {
		return fmt.Errorf("Your query did not return any results. Please try a different search criteria.")
	}

	if len(result.ScalingPolicies) > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more " +
			"specific search criteria.")
	}

	// If execution made it to this point, we have exactly one 1 policy returned
	// and this is a safe operation
	policy := result.ScalingPolicies[0]

	log.Printf("[DEBUG] aws_appautoscaling_policy - AutoScaling Policy found: %s", *policy.PolicyName)

	err1 := policyDescriptionAttributes(d, policy)
	return err1
}

// Populate group attribute fields with the returned group
func policyDescriptionAttributes(d *schema.ResourceData, policy *applicationautoscaling.ScalingPolicy) error {
	log.Printf("[DEBUG] Setting attributes: %s", policy)
	d.Set("name", policy.PolicyName)
	d.Set("arn", policy.PolicyARN)
	d.Set("policy_type", policy.PolicyType)
	d.Set("resource_id", policy.ResourceId)
	d.Set("service_namespace", policy.ServiceNamespace)
	d.Set("scalable_dimension", policy.ScalableDimension)
	return nil
}
