package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsRoute53TrafficPolicyInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53TrafficPolicyInstanceCreate,
		Read:   resourceAwsRoute53TrafficPolicyInstanceRead,
		Update: resourceAwsRoute53TrafficPolicyInstanceUpdate,
		Delete: resourceAwsRoute53TrafficPolicyInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"hosted_zone_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
				StateFunc: func(v interface{}) string {
					value := strings.TrimSuffix(v.(string), ".")
					return strings.ToLower(value)
				},
			},
			"traffic_policy_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 36),
			},
			"traffic_policy_version": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtMost(1000),
			},
			"ttl": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtMost(2147483647),
			},
		},
	}
}

func resourceAwsRoute53TrafficPolicyInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	request := &route53.CreateTrafficPolicyInstanceInput{
		HostedZoneId:         aws.String(d.Get("hosted_zone_id").(string)),
		Name:                 aws.String(d.Get("name").(string)),
		TrafficPolicyId:      aws.String(d.Get("traffic_policy_id").(string)),
		TrafficPolicyVersion: aws.Int64(int64(d.Get("traffic_policy_version").(int))),
		TTL:                  aws.Int64(int64(d.Get("ttl").(int))),
	}

	response, err := conn.CreateTrafficPolicyInstance(request)
	if err != nil {
		return fmt.Errorf("Error creating Route53 Traffic Policy Instance %s: %s", d.Get("name").(string), err)
	}

	d.SetId(*response.TrafficPolicyInstance.Id)

	return resourceAwsRoute53TrafficPolicyInstanceRead(d, meta)
}

func resourceAwsRoute53TrafficPolicyInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	request := &route53.GetTrafficPolicyInstanceInput{
		Id: aws.String(d.Id()),
	}

	response, err := conn.GetTrafficPolicyInstance(request)
	if err != nil {
		return fmt.Errorf("Error reading Route53 Traffic Policy Instance %s: %s", d.Get("name").(string), err)
	}

	err = d.Set("name", strings.TrimSuffix(*response.TrafficPolicyInstance.Name, "."))
	if err != nil {
		return fmt.Errorf("Error setting name for: %s, error: %#v", d.Id(), err)
	}

	err = d.Set("hosted_zone_id", response.TrafficPolicyInstance.HostedZoneId)
	if err != nil {
		return fmt.Errorf("Error setting hosted_zone_id for: %s, error: %#v", d.Id(), err)
	}

	err = d.Set("traffic_policy_id", response.TrafficPolicyInstance.TrafficPolicyId)
	if err != nil {
		return fmt.Errorf("Error setting traffic_policy_id for: %s, error: %#v", d.Id(), err)
	}

	err = d.Set("traffic_policy_version", response.TrafficPolicyInstance.TrafficPolicyVersion)
	if err != nil {
		return fmt.Errorf("Error setting traffic_policy_version for: %s, error: %#v", d.Id(), err)
	}

	err = d.Set("ttl", response.TrafficPolicyInstance.TTL)
	if err != nil {
		return fmt.Errorf("Error setting ttl for: %s, error: %#v", d.Id(), err)
	}

	return nil
}

func resourceAwsRoute53TrafficPolicyInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	request := &route53.UpdateTrafficPolicyInstanceInput{
		Id:                   aws.String(d.Id()),
		TrafficPolicyId:      aws.String(d.Get("traffic_policy_id").(string)),
		TrafficPolicyVersion: aws.Int64(int64(d.Get("traffic_policy_version").(int))),
		TTL:                  aws.Int64(int64(d.Get("ttl").(int))),
	}

	_, err := conn.UpdateTrafficPolicyInstance(request)
	if err != nil {
		return fmt.Errorf("Error updating Route53 Traffic Policy Instance %s: %s", d.Get("name").(string), err)
	}

	return resourceAwsRoute53TrafficPolicyInstanceRead(d, meta)
}

func resourceAwsRoute53TrafficPolicyInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	request := &route53.DeleteTrafficPolicyInstanceInput{
		Id: aws.String(d.Id()),
	}

	_, err := conn.DeleteTrafficPolicyInstance(request)
	if err != nil {
		return fmt.Errorf("Error deleting Route53 Traffic Policy Instance %s: %s", d.Get("name").(string), err)
	}

	return nil
}
