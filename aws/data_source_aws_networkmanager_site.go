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

func dataSourceAwsNetworkManagerSite() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsNetworkManagerSiteRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"location": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"latitude": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"longitude": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsNetworkManagerSiteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &networkmanager.GetSitesInput{
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
	}

	if v, ok := d.GetOk("id"); ok {
		input.SiteIds = aws.StringSlice([]string{v.(string)})
	}

	log.Printf("[DEBUG] Reading Network Manager Site: %s", input)
	output, err := conn.GetSites(input)

	if err != nil {
		return fmt.Errorf("error reading Network Manager Site: %s", err)
	}

	// do filtering here
	var filteredSites []*networkmanager.Site
	if tags, ok := d.GetOk("tags"); ok {
		keyValueTags := keyvaluetags.New(tags.(map[string]interface{})).IgnoreAws()
		for _, site := range output.Sites {
			tagsMatch := true
			if len(keyValueTags) > 0 {
				listTags := keyvaluetags.NetworkmanagerKeyValueTags(site.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)
				tagsMatch = listTags.ContainsAll(keyValueTags)
			}
			if tagsMatch {
				filteredSites = append(filteredSites, site)
			}
		}
	} else {
		filteredSites = output.Sites
	}

	if output == nil || len(filteredSites) == 0 {
		return errors.New("error reading Network Manager Site: no results found")
	}

	if len(filteredSites) > 1 {
		return errors.New("error reading Network Manager Site: more than one result found. Please try a more specific search criteria.")
	}

	site := filteredSites[0]

	if site == nil {
		return errors.New("error reading Network Manager Site: empty result")
	}

	d.Set("description", site.Description)
	d.Set("arn", site.SiteArn)
	if err := d.Set("location", flattenNetworkManagerLocation(site.Location)); err != nil {
		return fmt.Errorf("error setting location: %s", err)
	}
	if err := d.Set("tags", keyvaluetags.NetworkmanagerKeyValueTags(site.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.SetId(aws.StringValue(site.SiteId))

	return nil
}
