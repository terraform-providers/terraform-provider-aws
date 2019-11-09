package aws

import (
	"fmt"
	"log"

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
	filters, filtersOk := d.GetOk("filter")

	params := &ec2.DescribeHostsInput{}
	if filtersOk {
		params.Filter = buildAwsDataSourceFilters(filters.(*schema.Set))
	}
	// var hostIDs []string
	resp, err := conn.DescribeHosts(params)
	if err != nil {
		return err
	}
	// If no hosts were returned, return
	if len(resp.Hosts) == 0 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	var filteredHosts []*ec2.Host

	// loop through reservations, and remove terminated hosts, populate host slice
	for _, host := range resp.Hosts {

		if host.State != nil && *host.State != "terminated" {
			filteredHosts = append(filteredHosts, host)
		}

	}
	var instance *ec2.Host
	if len(filteredHosts) < 1 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	if len(filteredHosts) > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more " +
			"specific search criteria.")
	} else {
		instance = filteredHosts[0]
	}

	log.Printf("[DEBUG] aws_dedicated_host - Single Host ID found: %s", *instance.HostId)

	return nil
}
