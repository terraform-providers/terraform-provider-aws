package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/codestarconnections/finder"
)

func dataSourceAwsCodeStarConnectionsConnection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCodeStarConnectionsConnectionRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateArn,
			},

			"connection_status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"host_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"provider_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsCodeStarConnectionsConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codestarconnectionsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Getting CodeStar Connection")

	var connection *codestarconnections.Connection
	var err error

	if v, ok := d.GetOk("arn"); ok {
		arn := v.(string)
		connection, err = finder.ConnectionByArn(conn, arn)
		if err != nil {
			return fmt.Errorf("error getting CodeStar Connection (%s): %w", arn, err)
		}
		if connection == nil {
			return fmt.Errorf("Could not find CodeStar connection with arn (%s)", arn)
		}
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		connection, err = finder.ConnectionByName(conn, name)
		if err != nil {
			return fmt.Errorf("error getting CodeStar Connection (%s): %w", name, err)
		}
		if connection == nil {
			return fmt.Errorf("Could not find CodeStar connection with name (%s)", name)
		}
	} else {
		return fmt.Errorf("Either arn or name must be specified")
	}

	log.Printf("[DEBUG] CodeStar Connection: %#v", connection)

	arn := aws.StringValue(connection.ConnectionArn)
	d.SetId(arn)
	d.Set("arn", arn)
	d.Set("connection_status", connection.ConnectionStatus)
	d.Set("host_arn", connection.HostArn)
	d.Set("name", connection.ConnectionName)
	d.Set("provider_type", connection.ProviderType)

	tags, err := keyvaluetags.CodestarconnectionsListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for CodeStar Connection (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags for CodeStar Connection (%s): %w", arn, err)
	}

	return nil
}
