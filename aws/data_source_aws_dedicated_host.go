package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsDedicatedHost() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAwsDedicatedHostRead,

		Schema: map[string]*schema.Schema{

			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"host_recovery": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"auto_placement": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsAwsDedicatedHostRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	params := &ec2.DescribeHostsInput{}
	var hostIDs []string
	err := conn.DescribeHostsPages(params, func(resp *ec2.DescribeHostsOutput, isLast bool) bool {
		for _, res := range resp.Hosts {
			hostIDs = append(hostIDs, *res.HostId)
		}

		return !isLast
	})
	if err != nil {
		return err
	}
	err = d.Set("ids", hostIDs)
	if err != nil {
		return err
	}

	return nil
}
