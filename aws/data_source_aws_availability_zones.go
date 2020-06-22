package aws

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAwsAvailabilityZones() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAvailabilityZonesRead,

		Schema: map[string]*schema.Schema{
			"all_availability_zones": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"blacklisted_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"blacklisted_zone_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filter": ec2CustomFiltersSchema(),
			"group_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.AvailabilityZoneStateAvailable,
					ec2.AvailabilityZoneStateInformation,
					ec2.AvailabilityZoneStateImpaired,
					ec2.AvailabilityZoneStateUnavailable,
				}, false),
			},
			"zone_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsAvailabilityZonesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[DEBUG] Reading Availability Zones.")
	d.SetId(time.Now().UTC().String())

	request := &ec2.DescribeAvailabilityZonesInput{}

	if v, ok := d.GetOk("all_availability_zones"); ok {
		request.AllAvailabilityZones = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("state"); ok {
		request.Filters = []*ec2.Filter{
			{
				Name:   aws.String("state"),
				Values: []*string{aws.String(v.(string))},
			},
		}
	}

	if filters, filtersOk := d.GetOk("filter"); filtersOk {
		request.Filters = append(request.Filters, buildEC2CustomFilterList(
			filters.(*schema.Set),
		)...)
	}

	if len(request.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		request.Filters = nil
	}

	log.Printf("[DEBUG] Reading Availability Zones: %s", request)
	resp, err := conn.DescribeAvailabilityZones(request)
	if err != nil {
		return fmt.Errorf("Error fetching Availability Zones: %s", err)
	}

	sort.Slice(resp.AvailabilityZones, func(i, j int) bool {
		return aws.StringValue(resp.AvailabilityZones[i].ZoneName) < aws.StringValue(resp.AvailabilityZones[j].ZoneName)
	})

	blacklistedNames := d.Get("blacklisted_names").(*schema.Set)
	blacklistedZoneIDs := d.Get("blacklisted_zone_ids").(*schema.Set)
	groupNames := schema.NewSet(schema.HashString, nil)
	names := []string{}
	zoneIds := []string{}
	for _, v := range resp.AvailabilityZones {
		groupName := aws.StringValue(v.GroupName)
		name := aws.StringValue(v.ZoneName)
		zoneID := aws.StringValue(v.ZoneId)

		if blacklistedNames.Contains(name) {
			continue
		}

		if blacklistedZoneIDs.Contains(zoneID) {
			continue
		}

		if !groupNames.Contains(groupName) {
			groupNames.Add(groupName)
		}

		names = append(names, name)
		zoneIds = append(zoneIds, zoneID)
	}

	if err := d.Set("group_names", groupNames); err != nil {
		return fmt.Errorf("error setting group_names: %s", err)
	}
	if err := d.Set("names", names); err != nil {
		return fmt.Errorf("Error setting Availability Zone names: %s", err)
	}
	if err := d.Set("zone_ids", zoneIds); err != nil {
		return fmt.Errorf("Error setting Availability Zone IDs: %s", err)
	}

	return nil
}
