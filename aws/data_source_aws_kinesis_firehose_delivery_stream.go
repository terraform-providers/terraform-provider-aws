package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsKinesisFirehoseDeliveryStream() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsKinesisFirehoseDeliveryStreamRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsKinesisFirehoseDeliveryStreamRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).firehoseconn
	sn := d.Get("name").(string)
	params := &firehose.DescribeDeliveryStreamInput{
		DeliveryStreamName: aws.String(sn),
	}
	log.Printf("[DEBUG] Describing Delivery Stream: %s", params)
	resp, err := conn.DescribeDeliveryStream(params)
	if err != nil {
		return err
	}
	d.SetId(aws.StringValue(resp.DeliveryStreamDescription.DeliveryStreamARN))
	d.Set("arn", aws.StringValue(resp.DeliveryStreamDescription.DeliveryStreamARN))
	d.Set("name", sn)
	return nil
}
