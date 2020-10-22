package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsDedicatedHost() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAwsDedicatedHostRead,

		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"tags":   tagsSchemaComputed(),
			"host_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_recovery": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_family": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cores": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"total_vcpus": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sockets": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_placement": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsAwsDedicatedHostRead(d *schema.ResourceData, meta interface{}) error {
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	conn := meta.(*AWSClient).ec2conn
	hostID, hostIDOk := d.GetOk("host_id")

	filters, filtersOk := d.GetOk("filter")
	tags, tagsOk := d.GetOk("tags")

	params := &ec2.DescribeHostsInput{}
	if hostIDOk {
		params.HostIds = []*string{aws.String(hostID.(string))}
	}
	if filtersOk {
		params.Filter = buildAwsDataSourceFilters(filters.(*schema.Set))
	}
	resp, err := conn.DescribeHosts(params)
	if err != nil {
		return err
	}
	// If no hosts were returned, return
	if resp.Hosts == nil || len(resp.Hosts) == 0 {
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
	if tagsOk {
		params.Filter = append(params.Filter, ec2TagFiltersFromMap(tags.(map[string]interface{}))...)
	}
	log.Printf("[DEBUG] aws_dedicated_host - Single host ID found: %s", *host.HostId)
	if err := hostDescriptionAttributes(d, host, ignoreTagsConfig); err != nil {
		return err
	}
	return nil
}
func hostDescriptionAttributes(d *schema.ResourceData, host *ec2.Host, ignoreTagsConfig *keyvaluetags.IgnoreConfig) error {

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
	if host.HostProperties.InstanceType != nil {
		d.Set("instance_type", host.HostProperties.InstanceType)
	}
	if host.HostProperties.InstanceFamily != nil {
		d.Set("instance_family", host.HostProperties.InstanceFamily)
	}
	if host.HostProperties.Cores != nil {
		d.Set("cores", host.HostProperties.Cores)
	}
	if host.HostProperties.Sockets != nil {
		d.Set("sockets", host.HostProperties.Sockets)
	}
	if host.HostProperties.TotalVCpus != nil {
		d.Set("total_vcpus", host.HostProperties.TotalVCpus)
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(host.Tags).IgnoreConfig(ignoreTagsConfig).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}
	return nil

}
