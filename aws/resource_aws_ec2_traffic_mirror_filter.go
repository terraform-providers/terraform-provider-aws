package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEc2TrafficMirrorFilter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TrafficMirrorFilterCreate,
		Read:   resourceAwsEc2TrafficMirrorFilterRead,
		Update: resourceAwsEc2TrafficMirrorFilterUpdate,
		Delete: resourceAwsEc2TrafficMirrorFilterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"network_services": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"amazon-dns",
					}, false),
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsEc2TrafficMirrorFilterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.CreateTrafficMirrorFilterInput{}

	if description, ok := d.GetOk("description"); ok {
		input.Description = aws.String(description.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.TagSpecifications = ec2TagSpecificationsFromMap(v.(map[string]interface{}), ec2.ResourceTypeTrafficMirrorFilter)
	}

	out, err := conn.CreateTrafficMirrorFilter(input)
	if err != nil {
		return fmt.Errorf("Error while creating traffic filter %s", err)
	}

	d.SetId(*out.TrafficMirrorFilter.TrafficMirrorFilterId)

	if v, ok := d.GetOk("network_services"); ok {
		input := &ec2.ModifyTrafficMirrorFilterNetworkServicesInput{
			TrafficMirrorFilterId: aws.String(d.Id()),
			AddNetworkServices:    expandStringSet(v.(*schema.Set)),
		}

		_, err := conn.ModifyTrafficMirrorFilterNetworkServices(input)
		if err != nil {
			return fmt.Errorf("error modifying EC2 Traffic Mirror Filter (%s) network services: %w", d.Id(), err)
		}

	}

	return resourceAwsEc2TrafficMirrorFilterRead(d, meta)
}

func resourceAwsEc2TrafficMirrorFilterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("network_services") {
		input := &ec2.ModifyTrafficMirrorFilterNetworkServicesInput{
			TrafficMirrorFilterId: aws.String(d.Id()),
		}

		o, n := d.GetChange("network_services")
		newServices := n.(*schema.Set).Difference(o.(*schema.Set)).List()
		if len(newServices) > 0 {
			input.AddNetworkServices = expandStringList(newServices)
		}

		removeServices := o.(*schema.Set).Difference(n.(*schema.Set)).List()
		if len(removeServices) > 0 {
			input.RemoveNetworkServices = expandStringList(removeServices)
		}

		_, err := conn.ModifyTrafficMirrorFilterNetworkServices(input)
		if err != nil {
			return fmt.Errorf("error modifying EC2 Traffic Mirror Filter (%s) network services: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Traffic Mirror Filter (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsEc2TrafficMirrorFilterRead(d, meta)
}

func resourceAwsEc2TrafficMirrorFilterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTrafficMirrorFiltersInput{
		TrafficMirrorFilterIds: aws.StringSlice([]string{d.Id()}),
	}

	out, err := conn.DescribeTrafficMirrorFilters(input)

	if isAWSErr(err, "InvalidTrafficMirrorFilterId.NotFound", "") {
		log.Printf("[WARN] EC2 Traffic Mirror Filter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error describing traffic mirror filter %v: %v", d.Id(), err)
	}

	if len(out.TrafficMirrorFilters) == 0 {
		log.Printf("[WARN] EC2 Traffic Mirror Filter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	trafficMirrorFilter := out.TrafficMirrorFilters[0]
	d.Set("description", trafficMirrorFilter.Description)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(trafficMirrorFilter.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("network_services", aws.StringValueSlice(trafficMirrorFilter.NetworkServices)); err != nil {
		return fmt.Errorf("error setting network_services for filter %v: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsEc2TrafficMirrorFilterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DeleteTrafficMirrorFilterInput{
		TrafficMirrorFilterId: aws.String(d.Id()),
	}

	_, err := conn.DeleteTrafficMirrorFilter(input)
	if err != nil {
		return fmt.Errorf("Error deleting traffic mirror filter %v: %v", d.Id(), err)
	}

	return nil
}
