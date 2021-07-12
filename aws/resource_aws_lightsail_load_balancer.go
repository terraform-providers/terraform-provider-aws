package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/lightsail/waiter"
)

func resourceAwsLightsailLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLightsailLoadBalancerCreate,
		Read:   resourceAwsLightsailLoadBalancerRead,
		Update: resourceAwsLightsailLoadBalancerUpdate,
		Delete: resourceAwsLightsailLoadBalancerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with an alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]+[^._\-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
				),
			},
			"health_check_path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
			},
			"instance_port": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 65535),
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_ports": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func resourceAwsLightsailLoadBalancerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn

	req := lightsail.CreateLoadBalancerInput{
		HealthCheckPath:  aws.String(d.Get("health_check_path").(string)),
		InstancePort:     aws.Int64(int64(d.Get("instance_port").(int))),
		LoadBalancerName: aws.String(d.Get("name").(string)),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		req.Tags = keyvaluetags.New(v).IgnoreAws().LightsailTags()
	}

	resp, err := conn.CreateLoadBalancer(&req)
	if err != nil {
		return err
	}

	if len(resp.Operations) == 0 {
		return fmt.Errorf("No operations found for CreateInstance request")
	}

	op := resp.Operations[0]
	d.SetId(d.Get("name").(string))

	_, err = waiter.OperationCreated(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for load balancer (%s) to become ready: %s", d.Id(), err)
	}

	return resourceAwsLightsailLoadBalancerRead(d, meta)
}

func resourceAwsLightsailLoadBalancerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.GetLoadBalancer(&lightsail.GetLoadBalancerInput{
		LoadBalancerName: aws.String(d.Id()),
	})

	lb := resp.LoadBalancer

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NotFoundException" {
				log.Printf("[WARN] Lightsail load balancer (%s) not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			return err
		}
		return err
	}

	d.Set("arn", lb.Arn)
	d.Set("created_at", lb.CreatedAt.Format(time.RFC3339))
	d.Set("health_check_path", lb.HealthCheckPath)
	d.Set("instance_port", lb.InstancePort)
	d.Set("name", lb.Name)
	d.Set("protocol", lb.Protocol)
	d.Set("public_ports", lb.PublicPorts)
	d.Set("dns_name", lb.DnsName)

	if err := d.Set("tags", keyvaluetags.LightsailKeyValueTags(lb.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsLightsailLoadBalancerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn
	resp, err := conn.DeleteLoadBalancer(&lightsail.DeleteLoadBalancerInput{
		LoadBalancerName: aws.String(d.Id()),
	})

	op := resp.Operations[0]

	if err != nil {
		return err
	}

	_, err = waiter.OperationCreated(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for load balancer (%s) to become destroyed: %s", d.Id(), err)
	}

	return err
}

func resourceAwsLightsailLoadBalancerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn

	if d.HasChange("health_check_path") {
		_, err := conn.UpdateLoadBalancerAttribute(&lightsail.UpdateLoadBalancerAttributeInput{
			AttributeName:    aws.String("HealthCheckPath"),
			AttributeValue:   aws.String(d.Get("health_check_path").(string)),
			LoadBalancerName: aws.String(d.Get("name").(string)),
		})
		d.Set("health_check_path", d.Get("health_check_path").(string))
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.LightsailUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Lightsail Instance (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsLightsailLoadBalancerRead(d, meta)
}
