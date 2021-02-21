package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsLbTargetGroupAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLbAttachmentCreate,
		Read:   resourceAwsLbAttachmentRead,
		Delete: resourceAwsLbAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"target_group_arn": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"target_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"port": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
		},
	}
}

func resourceAwsLbAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	target := &elbv2.TargetDescription{
		Id: aws.String(d.Get("target_id").(string)),
	}

	if v, ok := d.GetOk("port"); ok {
		target.Port = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		target.AvailabilityZone = aws.String(v.(string))
	}

	params := &elbv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(d.Get("target_group_arn").(string)),
		Targets:        []*elbv2.TargetDescription{target},
	}

	log.Printf("[INFO] Registering Target %s with Target Group %s", d.Get("target_id").(string),
		d.Get("target_group_arn").(string))

	err := resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, err := elbconn.RegisterTargets(params)

		if isAWSErr(err, "InvalidTarget", "") {
			return resource.RetryableError(fmt.Errorf("Error attaching instance to LB, retrying: %s", err))
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = elbconn.RegisterTargets(params)
	}
	if err != nil {
		return fmt.Errorf("Error registering targets with target group: %s", err)
	}

	//lintignore:R016 // Allow legacy unstable ID usage in managed resource
	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-", d.Get("target_group_arn"))))

	return nil
}

func resourceAwsLbAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	target := &elbv2.TargetDescription{
		Id: aws.String(d.Get("target_id").(string)),
	}

	if v, ok := d.GetOk("port"); ok {
		target.Port = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		target.AvailabilityZone = aws.String(v.(string))
	}

	params := &elbv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(d.Get("target_group_arn").(string)),
		Targets:        []*elbv2.TargetDescription{target},
	}

	_, err := elbconn.DeregisterTargets(params)
	if err != nil && !isAWSErr(err, elbv2.ErrCodeTargetGroupNotFoundException, "") {
		return fmt.Errorf("Error deregistering Targets: %s", err)
	}

	return nil
}

// resourceAwsLbAttachmentRead requires all of the fields in order to describe the correct
// target, so there is no work to do beyond ensuring that the target and group still exist.
func resourceAwsLbAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	target := &elbv2.TargetDescription{
		Id: aws.String(d.Get("target_id").(string)),
	}

	if v, ok := d.GetOk("port"); ok {
		target.Port = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		target.AvailabilityZone = aws.String(v.(string))
	}

	resp, err := elbconn.DescribeTargetHealth(&elbv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(d.Get("target_group_arn").(string)),
		Targets:        []*elbv2.TargetDescription{target},
	})

	if err != nil {
		if isAWSErr(err, elbv2.ErrCodeTargetGroupNotFoundException, "") {
			log.Printf("[WARN] Target group does not exist, removing target attachment %s", d.Id())
			d.SetId("")
			return nil
		}
		if isAWSErr(err, elbv2.ErrCodeInvalidTargetException, "") {
			log.Printf("[WARN] Target does not exist, removing target attachment %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Target Health: %s", err)
	}

	for _, targetDesc := range resp.TargetHealthDescriptions {
		if targetDesc == nil || targetDesc.Target == nil {
			continue
		}

		if aws.StringValue(targetDesc.Target.Id) == d.Get("target_id").(string) {
			// These will catch targets being removed by hand (draining as we plan) or that have been removed for a while
			// without trying to re-create ones that are just not in use. For example, a target can be `unused` if the
			// target group isnt assigned to anything, a scenario where we don't want to continuously recreate the resource.
			if targetDesc.TargetHealth == nil {
				continue
			}

			reason := aws.StringValue(targetDesc.TargetHealth.Reason)

			if reason == elbv2.TargetHealthReasonEnumTargetNotRegistered || reason == elbv2.TargetHealthReasonEnumTargetDeregistrationInProgress {
				log.Printf("[WARN] Target Attachment does not exist, recreating attachment")
				d.SetId("")
				return nil
			}
		}
	}

	if len(resp.TargetHealthDescriptions) != 1 {
		log.Printf("[WARN] Target does not exist, removing target attachment %s", d.Id())
		d.SetId("")
		return nil
	}

	return nil
}
