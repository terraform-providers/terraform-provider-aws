package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsEips() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEipsRead,

		Schema: map[string]*schema.Schema{

			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"public_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsEipsRead(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*AWSClient).ec2conn
	d.SetId(time.Now().UTC().String())

	var rawIds []string
	var rawPublicIps []string

	req := &ec2.DescribeAddressesInput{}

	req.Filters = []*ec2.Filter{}

	req.Filters = append(req.Filters, buildEC2TagFilterList(
		tagsFromMap(d.Get("tags").(map[string]interface{})),
	)...)

	if len(req.Filters) == 0 {
		return fmt.Errorf("No Filters were found")
	}

	log.Printf("[DEBUG] Reading EIP: %s", req)
	resp, err := conn.DescribeAddresses(req)
	if err != nil {
		return fmt.Errorf("error describing EC2 Address: %s", err)
	}

	if resp == nil || len(resp.Addresses) == 0 {
		return fmt.Errorf("no matching Elastic IP found")
	}

	for _, eip := range resp.Addresses {
		log.Printf("[DEBUG] Reading EIP TEST: %s", eip)
		log.Printf("[DEBUG] Reading EIP TEST: %s", *eip.PublicIp)
		if aws.StringValue(eip.Domain) == ec2.DomainTypeVpc {
			rawIds = append(rawIds, aws.StringValue(eip.AllocationId))
		} else {
			log.Printf("[DEBUG] Reading EIP, has no AllocationId, this means we have a Classic EIP, the id will also be the public ip : %s", req)
			rawIds = append(rawIds, aws.StringValue(eip.PublicIp))
		}
		rawPublicIps = append(rawPublicIps, aws.StringValue(eip.PublicIp))
	}
	d.Set("public_ips", rawPublicIps)
	d.Set("ids", rawIds)

	return nil
}
