package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsMskConfiguration() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsMskConfigurationRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kafka_versions": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"latest_revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"server_properties": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsMskConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	listConfigurationsInput := &kafka.ListConfigurationsInput{}

	var configuration *kafka.Configuration
	err := conn.ListConfigurationsPages(listConfigurationsInput, func(page *kafka.ListConfigurationsOutput, lastPage bool) bool {
		for _, config := range page.Configurations {
			if aws.StringValue(config.Name) == d.Get("name").(string) {
				configuration = config
				break
			}
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing MSK Configurations: %s", err)
	}

	if configuration == nil {
		return fmt.Errorf("error reading MSK Configuration: no results found")
	}

	if configuration.LatestRevision == nil {
		return fmt.Errorf("error describing MSK Configuration (%s): missing latest revision", d.Id())
	}

	revision := configuration.LatestRevision.Revision
	revisionInput := &kafka.DescribeConfigurationRevisionInput{
		Arn:      configuration.Arn,
		Revision: revision,
	}

	revisionOutput, err := conn.DescribeConfigurationRevision(revisionInput)

	if err != nil {
		return fmt.Errorf("error describing MSK Configuration (%s) Revision (%d): %s", d.Id(), aws.Int64Value(revision), err)
	}

	if revisionOutput == nil {
		return fmt.Errorf("error describing MSK Configuration (%s) Revision (%d): missing result", d.Id(), aws.Int64Value(revision))
	}

	d.Set("arn", aws.StringValue(configuration.Arn))
	d.Set("description", aws.StringValue(configuration.Description))

	if err := d.Set("kafka_versions", aws.StringValueSlice(configuration.KafkaVersions)); err != nil {
		return fmt.Errorf("error setting kafka_versions: %s", err)
	}

	d.Set("latest_revision", aws.Int64Value(revision))
	d.Set("name", aws.StringValue(configuration.Name))
	d.Set("server_properties", string(revisionOutput.ServerProperties))

	d.SetId(aws.StringValue(configuration.Arn))

	return nil
}
