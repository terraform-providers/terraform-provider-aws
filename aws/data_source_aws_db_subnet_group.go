package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsDbSubnetGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDbSubnetGroupRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsDbSubnetGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	name := d.Get("name").(string)

	opts := &rds.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String(name),
	}

	log.Printf("[DEBUG] Reading DB SubnetGroup: %s", name)

	resp, err := conn.DescribeDBSubnetGroups(opts)
	if err != nil {
		return fmt.Errorf("error reading DB SubnetGroup (%s): %w", name, err)
	}

	if resp == nil || resp.DBSubnetGroups == nil {
		return fmt.Errorf("error reading DB SubnetGroup (%s): empty response", name)
	}
	if len(resp.DBSubnetGroups) > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more specific search criteria.")
	}

	dbSubnetGroup := resp.DBSubnetGroups[0]

	d.SetId(aws.StringValue(dbSubnetGroup.DBSubnetGroupName))
	d.Set("arn", dbSubnetGroup.DBSubnetGroupArn)
	d.Set("description", dbSubnetGroup.DBSubnetGroupDescription)
	d.Set("name", dbSubnetGroup.DBSubnetGroupName)
	d.Set("status", dbSubnetGroup.SubnetGroupStatus)

	subnets := make([]string, 0, len(dbSubnetGroup.Subnets))
	for _, v := range dbSubnetGroup.Subnets {
		subnets = append(subnets, aws.StringValue(v.SubnetIdentifier))
	}
	if err := d.Set("subnet_ids", subnets); err != nil {
		return fmt.Errorf("error setting subnet_ids: %w", err)
	}

	d.Set("vpc_id", dbSubnetGroup.VpcId)

	return nil
}
