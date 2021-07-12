package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsRoute53Zones() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRoute53ZonesRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsRoute53ZonesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn
	req := &route53.ListHostedZonesInput{}
	res, err := conn.ListHostedZones(req)
	if err != nil {
		return err
	}

	if res == nil || len(res.HostedZones) == 0 {
		return fmt.Errorf("no matching hosted zone found")
	}

	zoneIds := make([]string, 0)
	for _, zone := range res.HostedZones {
		zoneIdToSet := cleanZoneID(aws.StringValue(zone.Id))
		zoneIds = append(zoneIds, zoneIdToSet)
	}

	d.SetId(meta.(*AWSClient).region)
	if err := d.Set("ids", zoneIds); err != nil {
		return fmt.Errorf("error setting vpc ids: %s", err)
	}

	return nil
}
