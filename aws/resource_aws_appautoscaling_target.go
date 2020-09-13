package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsAppautoscalingTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppautoscalingTargetPut,
		Read:   resourceAwsAppautoscalingTargetRead,
		Update: resourceAwsAppautoscalingTargetPut,
		Delete: resourceAwsAppautoscalingTargetDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsAppautoscalingTargetImport,
		},

		Schema: map[string]*schema.Schema{
			"max_capacity": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"min_capacity": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"scalable_dimension": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_namespace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsAppautoscalingTargetPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn

	var targetOpts applicationautoscaling.RegisterScalableTargetInput

	targetOpts.MaxCapacity = aws.Int64(int64(d.Get("max_capacity").(int)))
	targetOpts.MinCapacity = aws.Int64(int64(d.Get("min_capacity").(int)))
	targetOpts.ResourceId = aws.String(d.Get("resource_id").(string))
	targetOpts.ScalableDimension = aws.String(d.Get("scalable_dimension").(string))
	targetOpts.ServiceNamespace = aws.String(d.Get("service_namespace").(string))

	if roleArn, exists := d.GetOk("role_arn"); exists {
		targetOpts.RoleARN = aws.String(roleArn.(string))
	}

	log.Printf("[DEBUG] Application autoscaling target create configuration %s", targetOpts)
	var err error
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err = conn.RegisterScalableTarget(&targetOpts)

		if err != nil {
			if isAWSErr(err, applicationautoscaling.ErrCodeValidationException, "Unable to assume IAM role") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, applicationautoscaling.ErrCodeValidationException, "ECS service doesn't exist") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.RegisterScalableTarget(&targetOpts)
	}

	if err != nil {
		return fmt.Errorf("Error creating application autoscaling target: %s", err)
	}

	d.SetId(d.Get("resource_id").(string))
	log.Printf("[INFO] Application AutoScaling Target ID: %s", d.Id())

	return resourceAwsAppautoscalingTargetRead(d, meta)
}

func resourceAwsAppautoscalingTargetRead(d *schema.ResourceData, meta interface{}) error {
	var t *applicationautoscaling.ScalableTarget

	conn := meta.(*AWSClient).appautoscalingconn

	namespace := d.Get("service_namespace").(string)
	dimension := d.Get("scalable_dimension").(string)

	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		t, err = getAwsAppautoscalingTarget(d.Id(), namespace, dimension, conn)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		if d.IsNewResource() && t == nil {
			return resource.RetryableError(&resource.NotFoundError{})
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		t, err = getAwsAppautoscalingTarget(d.Id(), namespace, dimension, conn)
	}

	if err != nil {
		return err
	}
	if t == nil {
		log.Printf("[WARN] Application AutoScaling Target (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("max_capacity", t.MaxCapacity)
	d.Set("min_capacity", t.MinCapacity)
	d.Set("resource_id", t.ResourceId)
	d.Set("role_arn", t.RoleARN)
	d.Set("scalable_dimension", t.ScalableDimension)
	d.Set("service_namespace", t.ServiceNamespace)

	return nil
}

func resourceAwsAppautoscalingTargetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn

	input := &applicationautoscaling.DeregisterScalableTargetInput{
		ResourceId:        aws.String(d.Get("resource_id").(string)),
		ServiceNamespace:  aws.String(d.Get("service_namespace").(string)),
		ScalableDimension: aws.String(d.Get("scalable_dimension").(string)),
	}

	_, err := conn.DeregisterScalableTarget(input)

	if err != nil {
		return fmt.Errorf("error deleting Application AutoScaling Target (%s): %w", d.Id(), err)
	}

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		t, err := getAwsAppautoscalingTarget(d.Get("resource_id").(string), d.Get("service_namespace").(string), d.Get("scalable_dimension").(string), conn)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if t != nil {
			return resource.RetryableError(fmt.Errorf("Application AutoScaling Target (%s) still exists", d.Id()))
		}

		return nil
	})
}

func getAwsAppautoscalingTarget(resourceId, namespace, dimension string,
	conn *applicationautoscaling.ApplicationAutoScaling) (*applicationautoscaling.ScalableTarget, error) {

	describeOpts := applicationautoscaling.DescribeScalableTargetsInput{
		ResourceIds:      []*string{aws.String(resourceId)},
		ServiceNamespace: aws.String(namespace),
	}

	log.Printf("[DEBUG] Application AutoScaling Target describe configuration: %#v", describeOpts)
	describeTargets, err := conn.DescribeScalableTargets(&describeOpts)
	if err != nil {
		// @TODO: We should probably send something else back if we're trying to access an unknown Resource ID
		// targetserr, ok := err.(awserr.Error)
		// if ok && targetserr.Code() == ""
		return nil, fmt.Errorf("Error retrieving Application AutoScaling Target: %s", err)
	}

	for idx, tgt := range describeTargets.ScalableTargets {
		if *tgt.ResourceId == resourceId && *tgt.ScalableDimension == dimension {
			return describeTargets.ScalableTargets[idx], nil
		}
	}

	return nil, nil
}

func resourceAwsAppautoscalingTargetImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")

	if len(idParts) < 3 {
		return nil, fmt.Errorf("unexpected format (%q), expected <service-namespace>/<resource-id>/<scalable-dimension>", d.Id())
	}

	serviceNamespace := idParts[0]
	resourceId := strings.Join(idParts[1:len(idParts)-1], "/")
	scalableDimension := idParts[len(idParts)-1]

	if serviceNamespace == "" || resourceId == "" || scalableDimension == "" {
		return nil, fmt.Errorf("unexpected format (%q), expected <service-namespace>/<resource-id>/<scalable-dimension>", d.Id())
	}

	d.Set("service_namespace", serviceNamespace)
	d.Set("resource_id", resourceId)
	d.Set("scalable_dimension", scalableDimension)
	d.SetId(resourceId)

	return []*schema.ResourceData{d}, nil
}
