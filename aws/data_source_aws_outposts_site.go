package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsOutpostsSite() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOutpostsSiteRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
		},
	}
}

func dataSourceAwsOutpostsSiteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).outpostsconn

	input := &outposts.ListSitesInput{}

	var results []*outposts.Site

	err := conn.ListSitesPages(input, func(page *outposts.ListSitesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, site := range page.Sites {
			if site == nil {
				continue
			}

			if v, ok := d.GetOk("id"); ok && v.(string) != aws.StringValue(site.SiteId) {
				continue
			}

			if v, ok := d.GetOk("name"); ok && v.(string) != aws.StringValue(site.Name) {
				continue
			}

			results = append(results, site)
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing Outposts Sites: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no Outposts Site found matching criteria; try different search")
	}

	if len(results) > 1 {
		return fmt.Errorf("multiple Outposts Sites found matching criteria; try different search")
	}

	site := results[0]

	d.SetId(aws.StringValue(site.SiteId))
	d.Set("account_id", site.AccountId)
	d.Set("description", site.Description)
	d.Set("name", site.Name)

	return nil
}
