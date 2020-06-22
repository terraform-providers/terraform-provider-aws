package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsOutpostsSites() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOutpostsSitesRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsOutpostsSitesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).outpostsconn

	input := &outposts.ListSitesInput{}

	var ids []string

	err := conn.ListSitesPages(input, func(page *outposts.ListSitesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, site := range page.Sites {
			if site == nil {
				continue
			}

			ids = append(ids, aws.StringValue(site.SiteId))
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing Outposts Sites: %w", err)
	}

	if err := d.Set("ids", ids); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	d.SetId(resource.UniqueId())

	return nil
}
