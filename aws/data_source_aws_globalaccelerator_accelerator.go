package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/globalaccelerator/finder"
)

func dataSourceAwsGlobalAcceleratorAccelerator() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsGlobalAcceleratorAcceleratorRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ip_address_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_sets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_addresses": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ip_family": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"attributes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"flow_logs_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"flow_logs_s3_bucket": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"flow_logs_s3_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsGlobalAcceleratorAcceleratorRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	var results []*globalaccelerator.Accelerator

	err := conn.ListAcceleratorsPages(&globalaccelerator.ListAcceleratorsInput{}, func(page *globalaccelerator.ListAcceleratorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, l := range page.Accelerators {
			if l == nil {
				continue
			}

			if v, ok := d.GetOk("arn"); ok && v.(string) != aws.StringValue(l.AcceleratorArn) {
				continue
			}

			if v, ok := d.GetOk("name"); ok && v.(string) != aws.StringValue(l.Name) {
				continue
			}

			results = append(results, l)
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error reading AWS Global Accelerator: %w", err)
	}

	if len(results) != 1 {
		return fmt.Errorf("Search returned %d results, please revise so only one is returned", len(results))
	}

	accelerator := results[0]
	d.SetId(aws.StringValue(accelerator.AcceleratorArn))
	d.Set("arn", accelerator.AcceleratorArn)
	d.Set("enabled", accelerator.Enabled)
	d.Set("dns_name", accelerator.DnsName)
	d.Set("hosted_zone_id", globalAcceleratorRoute53ZoneID)
	d.Set("name", accelerator.Name)
	d.Set("ip_address_type", accelerator.IpAddressType)
	d.Set("ip_sets", flattenGlobalAcceleratorIpSets(accelerator.IpSets))

	acceleratorAttributes, err := finder.AcceleratorAttributesByARN(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error reading Global Accelerator Accelerator (%s) attributes: %w", d.Id(), err)
	}

	if err := d.Set("attributes", []interface{}{flattenGlobalAcceleratorAcceleratorAttributes(acceleratorAttributes)}); err != nil {
		return fmt.Errorf("error setting attributes: %w", err)
	}

	tags, err := keyvaluetags.GlobalacceleratorListTags(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error listing tags for Global Accelerator Accelerator (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}
	return nil
}
