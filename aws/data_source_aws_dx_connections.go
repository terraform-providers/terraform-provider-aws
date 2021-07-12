package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/directconnect/finder"
)

func dataSourceAwsDxConnections() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDxConnectionsRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsDxConnectionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	tagsToMatch := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	connections, err := finder.Connections(conn, &directconnect.DescribeConnectionsInput{})
	if err != nil {
		return fmt.Errorf("error reading DirectConnect connections: %w", err)
	}

	var ids []*string

	for _, connection := range connections {
		if v, ok := d.GetOk("name"); ok && v.(string) != aws.StringValue(connection.ConnectionName) {
			continue
		}

		if len(tagsToMatch) > 0 && !keyvaluetags.DirectconnectKeyValueTags(connection.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).ContainsAll(tagsToMatch) {
			continue
		}

		ids = append(ids, connection.ConnectionId)
	}

	d.SetId(meta.(*AWSClient).region)

	if err := d.Set("ids", flattenStringSet(ids)); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	return nil
}
