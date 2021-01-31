package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsServiceRead,

		Schema: map[string]*schema.Schema{
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"partition": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"reverse_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"reverse_dns_prefix": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"service_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsServiceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient)

	if v, ok := d.GetOk("reverse_dns_name"); ok {
		serviceParts := strings.Split(v.(string), ".")
		if len(serviceParts) < 4 {
			return fmt.Errorf("reverse service DNS names must have at least 4 parts (%s has %d)", v.(string), len(serviceParts))
		}

		d.Set("service_id", serviceParts[len(serviceParts)-1])
		d.Set("region", serviceParts[len(serviceParts)-2])
		d.Set("reverse_dns_prefix", strings.Join(serviceParts[0:len(serviceParts)-2], "."))
	}

	if v, ok := d.GetOk("dns_name"); ok {
		serviceParts := invertStringSlice(strings.Split(v.(string), "."))
		if len(serviceParts) < 4 {
			return fmt.Errorf("service DNS names must have at least 4 parts (%s has %d)", v.(string), len(serviceParts))
		}

		d.Set("service_id", serviceParts[len(serviceParts)-1])
		d.Set("region", serviceParts[len(serviceParts)-2])
		d.Set("reverse_dns_prefix", strings.Join(serviceParts[0:len(serviceParts)-2], "."))
	}

	if _, ok := d.GetOk("region"); !ok {
		d.Set("region", client.region)
	}

	if _, ok := d.GetOk("service_id"); !ok {
		d.Set("service_id", endpoints.Ec2ServiceID)
	}

	if _, ok := d.GetOk("reverse_dns_prefix"); !ok {
		dnsParts := strings.Split(meta.(*AWSClient).dnsSuffix, ".")
		d.Set("reverse_dns_prefix", strings.Join(invertStringSlice(dnsParts), "."))
	}

	reverseDNS := fmt.Sprintf("%s.%s.%s", d.Get("reverse_dns_prefix").(string), d.Get("region").(string), d.Get("service_id").(string))
	d.Set("reverse_dns_name", reverseDNS)
	d.Set("dns_name", strings.ToLower(strings.Join(invertStringSlice(strings.Split(reverseDNS, ".")), ".")))

	d.Set("supported", true)
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), d.Get("region").(string)); ok {
		d.Set("partition", partition.ID())
		if _, ok := partition.Services()[d.Get("service_id").(string)]; !ok {
			d.Set("supported", false)
		}
	}

	d.SetId(reverseDNS)

	return nil
}
