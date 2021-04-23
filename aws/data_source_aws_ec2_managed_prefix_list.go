package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEc2ManagedPrefixList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsEc2ManagedPrefixListRead,
		Schema: map[string]*schema.Schema{
			"address_family": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"entries": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"filter": dataSourceFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"max_entries": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEc2ManagedPrefixListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := ec2.DescribeManagedPrefixListsInput{}

	if filters, ok := d.GetOk("filter"); ok {
		input.Filters = buildAwsDataSourceFilters(filters.(*schema.Set))
	}

	if prefixListId, ok := d.GetOk("id"); ok {
		input.PrefixListIds = aws.StringSlice([]string{prefixListId.(string)})
	}

	if prefixListName, ok := d.GetOk("name"); ok {
		input.Filters = append(input.Filters, &ec2.Filter{
			Name:   aws.String("prefix-list-name"),
			Values: aws.StringSlice([]string{prefixListName.(string)}),
		})
	}

	out, err := conn.DescribeManagedPrefixListsWithContext(ctx, &input)

	if err != nil {
		return diag.Errorf("error describing EC2 Managed Prefix Lists: %s", err)
	}

	if len(out.PrefixLists) < 1 {
		return diag.Errorf("no managed prefix lists matched the given criteria")
	}

	if len(out.PrefixLists) > 1 {
		return diag.Errorf("more than 1 prefix list matched the given criteria")
	}

	pl := *out.PrefixLists[0]

	d.SetId(aws.StringValue(pl.PrefixListId))
	d.Set("name", pl.PrefixListName)
	d.Set("owner_id", pl.OwnerId)
	d.Set("address_family", pl.AddressFamily)
	d.Set("arn", pl.PrefixListArn)
	d.Set("max_entries", pl.MaxEntries)
	d.Set("version", pl.Version)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(pl.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags attribute: %s", err)
	}

	var entries []interface{}

	err = conn.GetManagedPrefixListEntriesPages(
		&ec2.GetManagedPrefixListEntriesInput{
			PrefixListId: pl.PrefixListId,
		},
		func(output *ec2.GetManagedPrefixListEntriesOutput, last bool) bool {
			for _, entry := range output.Entries {
				entries = append(entries, map[string]interface{}{
					"cidr":        aws.StringValue(entry.Cidr),
					"description": aws.StringValue(entry.Description),
				})
			}

			return true
		},
	)

	if err != nil {
		return diag.Errorf("error listing EC2 Managed Prefix List (%s) entries: %s", d.Id(), err)
	}

	if err := d.Set("entries", entries); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
