package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsRedshiftOrderableCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRedshiftOrderableClusterRead,
		Schema: map[string]*schema.Schema{
			"availability_zones": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"cluster_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cluster_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"node_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"preferred_node_types": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsRedshiftOrderableClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn

	input := &redshift.DescribeOrderableClusterOptionsInput{}

	if v, ok := d.GetOk("cluster_version"); ok {
		input.ClusterVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("node_type"); ok {
		input.NodeType = aws.String(v.(string))
	}

	var orderableClusterOptions []*redshift.OrderableClusterOption

	err := conn.DescribeOrderableClusterOptionsPages(input, func(page *redshift.DescribeOrderableClusterOptionsOutput, lastPage bool) bool {
		for _, orderableClusterOption := range page.OrderableClusterOptions {
			if orderableClusterOption == nil {
				continue
			}

			if v, ok := d.GetOk("cluster_type"); ok {
				if aws.StringValue(orderableClusterOption.ClusterType) != v.(string) {
					continue
				}
			}

			orderableClusterOptions = append(orderableClusterOptions, orderableClusterOption)
		}
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error reading Redshift Orderable Cluster Options: %w", err)
	}

	if len(orderableClusterOptions) == 0 {
		return fmt.Errorf("no Redshift Orderable Cluster Options found matching criteria; try different search")
	}

	var orderableClusterOption *redshift.OrderableClusterOption
	preferredNodeTypes := d.Get("preferred_node_types").([]interface{})
	if len(preferredNodeTypes) > 0 {
		for _, preferredNodeTypeRaw := range preferredNodeTypes {
			preferredNodeType, ok := preferredNodeTypeRaw.(string)

			if !ok {
				continue
			}

			for _, option := range orderableClusterOptions {
				if preferredNodeType == aws.StringValue(option.NodeType) {
					orderableClusterOption = option
					break
				}
			}

			if orderableClusterOption != nil {
				break
			}
		}
	}

	if orderableClusterOption == nil && len(orderableClusterOptions) > 1 {
		return fmt.Errorf("multiple Redshift Orderable Cluster Options (%v) match the criteria; try a different search", orderableClusterOptions)
	}

	if orderableClusterOption == nil && len(orderableClusterOptions) == 1 {
		orderableClusterOption = orderableClusterOptions[0]
	}

	if orderableClusterOption == nil {
		return fmt.Errorf("no Redshift Orderable Cluster Options match the criteria; try a different search")
	}

	d.SetId(aws.StringValue(orderableClusterOption.NodeType))

	var availabilityZones []string
	for _, az := range orderableClusterOption.AvailabilityZones {
		availabilityZones = append(availabilityZones, aws.StringValue(az.Name))
	}
	d.Set("availability_zones", availabilityZones)

	d.Set("cluster_type", orderableClusterOption.ClusterType)
	d.Set("cluster_version", orderableClusterOption.ClusterVersion)
	d.Set("node_type", orderableClusterOption.NodeType)

	return nil
}
