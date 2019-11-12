package aws

import (
	"errors"
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
	resp, err := conn.DescribeHosts(params)
	if err != nil {
		return err
	}
	// If no hosts were returned, return
	if len(resp.Hosts) == 0 {
		return errors.New("Your query returned no results. Please change your search criteria and try again.")
	}

	var filteredHosts []*ec2.Host

	// loop through reservations, and remove terminated hosts, populate host slice
	for _, host := range resp.Hosts {

		if host.State != nil && *host.State != "terminated" {
			filteredHosts = append(filteredHosts, host)
		}

	}
	var host *ec2.Host
	if len(filteredHosts) < 1 {
		return errors.New("Your query returned no results. Please change your search criteria and try again.")
	}

	if len(filteredHosts) > 1 {
		return errors.New(`Your query returned more than one result. Please try a more 
			specific search criteria.`)
	} else {
		host = filteredHosts[0]
	}

	log.Printf("[DEBUG] aws_dedicated_host - Single host ID found: %s", *host.HostId)
	if err := hostDescriptionAttributes(d, host, conn); err != nil {
		return err
	}
	return nil
}
func hostDescriptionAttributes(d *schema.ResourceData, host *ec2.Host, conn *ec2.EC2) error {

	d.SetId(*host.HostId)
	d.Set("instance_state", host.State)

	if host.AvailabilityZone != nil {
		d.Set("availability_zone", host.AvailabilityZone)
	}
	if host.HostRecovery != nil {
		d.Set("host_recovery", host.HostRecovery)
	}
	if host.AutoPlacement != nil {
		d.Set("auto_placement", host.AutoPlacement)
	}
	// if host.HostProperties !=nil{
	// 	d.Set()
	// }
	d.Set("tags", tagsToMap(host.Tags))
	return nil

}
