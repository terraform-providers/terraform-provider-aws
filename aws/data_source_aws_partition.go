package aws

import (
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsPartition() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsPartitionRead,

		Schema: map[string]*schema.Schema{
			"partition": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dns_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"reverse_dns_prefix": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsPartitionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient)

	log.Printf("[DEBUG] Reading Partition.")
	d.SetId(meta.(*AWSClient).partition)

	log.Printf("[DEBUG] Setting AWS Partition to %s.", client.partition)
	d.Set("partition", meta.(*AWSClient).partition)

	log.Printf("[DEBUG] Setting AWS URL Suffix to %s.", client.dnsSuffix)
	d.Set("dns_suffix", meta.(*AWSClient).dnsSuffix)

	dnsParts := strings.Split(meta.(*AWSClient).dnsSuffix, ".")
	sort.Sort(sort.Reverse(sort.StringSlice(dnsParts)))
	servicePrefix := strings.Join(dnsParts, ".")
	d.Set("reverse_dns_prefix", servicePrefix)
	log.Printf("[DEBUG] Setting service prefix to %s.", servicePrefix)

	return nil
}
