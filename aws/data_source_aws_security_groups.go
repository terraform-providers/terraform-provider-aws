package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsSecurityGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSecurityGroupsRead,

		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"tags":   tagsSchemaComputed(),

			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsSecurityGroupsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	req := &ec2.DescribeSecurityGroupsInput{}

	filters, filtersOk := d.GetOk("filter")
	tags, tagsOk := d.GetOk("tags")

	if !filtersOk && !tagsOk {
		return fmt.Errorf("One of filters or tags must be assigned")
	}

	if filtersOk {
		req.Filters = append(req.Filters,
			buildAwsDataSourceFilters(filters.(*schema.Set))...)
	}
	if tagsOk {
		req.Filters = append(req.Filters, buildEC2TagFilterList(
			keyvaluetags.New(tags.(map[string]interface{})).Ec2Tags(),
		)...)
	}

	log.Printf("[DEBUG] Reading Security Groups with request: %s", req)

	var ids, vpcIds, arns []string
	for {
		resp, err := conn.DescribeSecurityGroups(req)
		if err != nil {
			return fmt.Errorf("error reading security groups: %w", err)
		}

		for _, sg := range resp.SecurityGroups {
			ids = append(ids, aws.StringValue(sg.GroupId))
			vpcIds = append(vpcIds, aws.StringValue(sg.VpcId))

			arn := arn.ARN{
				Partition: meta.(*AWSClient).partition,
				Service:   ec2.ServiceName,
				Region:    meta.(*AWSClient).region,
				AccountID: aws.StringValue(sg.OwnerId),
				Resource:  fmt.Sprintf("security-group/%s", aws.StringValue(sg.GroupId)),
			}.String()

			arns = append(arns, arn)
		}

		if resp.NextToken == nil {
			break
		}
		req.NextToken = resp.NextToken
	}

	if len(ids) < 1 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	log.Printf("[DEBUG] Found %d security groups via given filter: %s", len(ids), req)

	d.SetId(meta.(*AWSClient).region)

	err := d.Set("ids", ids)
	if err != nil {
		return err
	}

	if err = d.Set("vpc_ids", vpcIds); err != nil {
		return fmt.Errorf("error setting vpc_ids: %s", err)
	}

	if err = d.Set("arns", arns); err != nil {
		return fmt.Errorf("error setting arns: %s", err)
	}

	return nil
}
