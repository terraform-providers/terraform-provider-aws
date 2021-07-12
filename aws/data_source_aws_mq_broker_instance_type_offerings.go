package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsMqBrokerInstanceTypeOfferings() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsMqBrokerInstanceTypeOfferingsRead,

		Schema: map[string]*schema.Schema{
			"host_instance_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"engine_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					mq.EngineTypeActivemq,
					mq.EngineTypeRabbitmq,
				}, false),
			},
			"storage_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					mq.BrokerStorageTypeEbs,
					mq.BrokerStorageTypeEfs,
				}, false),
			},
			"availability_zones": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsMqBrokerInstanceTypeOfferingsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mqconn

	input := &mq.DescribeBrokerInstanceOptionsInput{}

	if v, ok := d.GetOk("host_instance_type"); ok {
		input.HostInstanceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("engine_type"); ok {
		input.EngineType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("storage_type"); ok {
		input.StorageType = aws.String(v.(string))
	}

	var availabilityZones []string

	for {
		output, err := conn.DescribeBrokerInstanceOptions(input)

		if err != nil {
			return fmt.Errorf("error reading MQ Instance Type Offerings: %w", err)
		}

		if output == nil {
			break
		}

		for _, instanceTypeOffering := range output.BrokerInstanceOptions {
			if instanceTypeOffering == nil {
				continue
			}
			for _, availabilityZone := range instanceTypeOffering.AvailabilityZones {
				if availabilityZone == nil {
					continue
				}
				availabilityZones = append(availabilityZones, aws.StringValue(availabilityZone.Name))
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if len(availabilityZones) == 0 {
		return fmt.Errorf("no availability_zones can support criteria supplied in aws_mq_broker_instance_type_offerings; try different criteria")
	}

	if err := d.Set("availability_zones", availabilityZones); err != nil {
		return fmt.Errorf("error setting instance_types: %w", err)
	}

	d.SetId(meta.(*AWSClient).region)

	return nil
}
