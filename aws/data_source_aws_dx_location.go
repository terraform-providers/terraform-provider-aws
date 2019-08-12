package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsDxLocation() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDxLocationRead,

		Schema: map[string]*schema.Schema{
			"available_port_speeds": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"location_code": {
				Type:     schema.TypeString,
				Required: true,
			},
			"location_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsDxLocationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	log.Printf("[DEBUG] Listing Direct Connect locations")
	resp, err := conn.DescribeLocations(&directconnect.DescribeLocationsInput{})
	if err != nil {
		return fmt.Errorf("error listing Direct Connect locations: %s", err)
	}

	found := false
	locationCode := d.Get("location_code").(string)
	for _, location := range resp.Locations {
		if aws.StringValue(location.LocationCode) == locationCode {
			found = true

			d.SetId(locationCode)
			err = d.Set("available_port_speeds", flattenStringList(location.AvailablePortSpeeds))
			if err != nil {
				return fmt.Errorf("error setting available_port_speeds: %s", err)
			}
			d.Set("location_code", location.LocationCode)
			d.Set("location_name", location.LocationName)

			break
		}
	}
	if !found {
		return fmt.Errorf("no Direct Connect location matched")
	}

	return nil
}
