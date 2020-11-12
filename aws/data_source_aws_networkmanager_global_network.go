package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsNetworkManagerGlobalNetwork() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsNetworkManagerGlobalNetworkRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsNetworkManagerGlobalNetworkRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &networkmanager.DescribeGlobalNetworksInput{}

	if v, ok := d.GetOk("id"); ok {
		input.GlobalNetworkIds = aws.StringSlice([]string{v.(string)})
	}

	log.Printf("[DEBUG] Reading Network Manager Global Network: %s", input)
	output, err := conn.DescribeGlobalNetworks(input)

	if err != nil {
		return fmt.Errorf("error reading Network Manager Global Network: %s", err)
	}

	// do filtering here
	var filteredGlobalNetworks []*networkmanager.GlobalNetwork
	if tags, ok := d.GetOk("tags"); ok {
		keyValueTags := keyvaluetags.New(tags.(map[string]interface{})).IgnoreAws()
		for _, globalNetwork := range output.GlobalNetworks {
			tagsMatch := true
			if len(keyValueTags) > 0 {
				listTags := keyvaluetags.NetworkmanagerKeyValueTags(globalNetwork.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)
				tagsMatch = listTags.ContainsAll(keyValueTags)
			}
			if tagsMatch {
				filteredGlobalNetworks = append(filteredGlobalNetworks, globalNetwork)
			}
		}
	} else {
		filteredGlobalNetworks = output.GlobalNetworks
	}

	if output == nil || len(filteredGlobalNetworks) == 0 {
		return errors.New("error reading Network Manager Global Network: no results found")
	}

	if len(filteredGlobalNetworks) > 1 {
		return errors.New("error reading Network Manager Global Network: more than one result found. Please try a more specific search criteria.")
	}

	globalNetwork := filteredGlobalNetworks[0]

	if globalNetwork == nil {
		return errors.New("error reading Network Manager Global Network: empty result")
	}

	d.Set("description", globalNetwork.Description)
	d.Set("arn", globalNetwork.GlobalNetworkArn)

	if err := d.Set("tags", keyvaluetags.NetworkmanagerKeyValueTags(globalNetwork.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.SetId(aws.StringValue(globalNetwork.GlobalNetworkId))

	return nil
}
