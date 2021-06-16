package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsCloudWatchEventSource() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudWatchEventSourceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsCloudWatchEventSourceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	input := &events.ListEventSourcesInput{}
	if v, ok := d.GetOk("name_prefix"); ok {
		input.NamePrefix = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Listing cloudwatch Event sources: %s", input)

	resp, err := conn.ListEventSources(input)
	if err != nil {
		return fmt.Errorf("error listing cloudwatch event sources: %w", err)
	}

	if resp == nil || len(resp.EventSources) == 0 {
		return fmt.Errorf("no matching partner event source")
	}
	if len(resp.EventSources) > 1 {
		return fmt.Errorf("multiple event sources matched; use additional constraints to reduce matches to a single event source")
	}

	es := resp.EventSources[0]

	d.SetId(aws.StringValue(es.Name))
	d.Set("arn", es.Arn)
	d.Set("created_by", es.CreatedBy)
	d.Set("name", es.Name)
	d.Set("state", es.State)

	return nil
}
