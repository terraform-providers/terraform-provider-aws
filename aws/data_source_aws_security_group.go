package aws

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
)

func dataSourceAwsSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSecurityGroupRead,

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"filter": ec2CustomFiltersSchema(),

			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchemaComputed(),

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeSecurityGroupsInput{}

	if id, ok := d.GetOk("id"); ok {
		req.GroupIds = []*string{aws.String(id.(string))}
	}

	req.Filters = buildEC2AttributeFilterList(
		map[string]string{
			"group-name": d.Get("name").(string),
			"vpc-id":     d.Get("vpc_id").(string),
		},
	)
	req.Filters = append(req.Filters, buildEC2TagFilterList(
		keyvaluetags.New(d.Get("tags").(map[string]interface{})).Ec2Tags(),
	)...)
	req.Filters = append(req.Filters, buildEC2CustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	sg, err := finder.SecurityGroup(conn, req)
	var nfe *resource.NotFoundError
	if errors.As(err, &nfe) {
		if nfe.Message == "empty result" {
			return fmt.Errorf("no matching SecurityGroup found")
		}
		if strings.HasPrefix(nfe.Message, "too many results:") {
			return fmt.Errorf("multiple Security Groups matched; use additional constraints to reduce matches to a single Security Group")
		}
	}
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(sg.GroupId))
	d.Set("name", sg.GroupName)
	d.Set("description", sg.Description)
	d.Set("vpc_id", sg.VpcId)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(sg.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*AWSClient).region,
		AccountID: *sg.OwnerId,
		Resource:  fmt.Sprintf("security-group/%s", *sg.GroupId),
	}.String()
	d.Set("arn", arn)

	return nil
}
