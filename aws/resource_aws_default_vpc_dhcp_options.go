package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsDefaultVpcDhcpOptions() *schema.Resource {
	// reuse aws_vpc_dhcp_options schema, and methods for READ, UPDATE
	dvpc := resourceAwsVpcDhcpOptions()
	dvpc.Create = resourceAwsDefaultVpcDhcpOptionsCreate
	dvpc.Delete = resourceAwsDefaultVpcDhcpOptionsDelete

	// domain_name is a computed value for Default Default DHCP Options Sets
	dvpc.Schema["domain_name"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
	// domain_name_servers is a computed value for Default Default DHCP Options Sets
	dvpc.Schema["domain_name_servers"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}
	// ntp_servers is a computed value for Default Default DHCP Options Sets
	dvpc.Schema["ntp_servers"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	}

	return dvpc
}

func resourceAwsDefaultVpcDhcpOptionsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	var domainName string
	awsRegion := meta.(*AWSClient).region
	if awsRegion == "us-east-1" {
		domainName = "ec2.internal"
	} else {
		domainName = awsRegion + ".compute.internal"
	}
	req := &ec2.DescribeDhcpOptionsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("key"),
				Values: aws.StringSlice([]string{"domain-name"}),
			},
			{
				Name:   aws.String("value"),
				Values: aws.StringSlice([]string{domainName}),
			},
			{
				Name:   aws.String("key"),
				Values: aws.StringSlice([]string{"domain-name-servers"}),
			},
			{
				Name:   aws.String("value"),
				Values: aws.StringSlice([]string{"AmazonProvidedDNS"}),
			},
		},
	}

	var dhcpOptions []*ec2.DhcpOptions
	err := conn.DescribeDhcpOptionsPages(req, func(page *ec2.DescribeDhcpOptionsOutput, lastPage bool) bool {
		dhcpOptions = append(dhcpOptions, page.DhcpOptions...)
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("Error describing DHCP options: %s", err)
	}

	if len(dhcpOptions) == 0 {
		return fmt.Errorf("Default DHCP Options Set not found")
	}

	if len(dhcpOptions) > 1 {
		return fmt.Errorf("Multiple default DHCP Options Sets found")
	}

	if dhcpOptions[0] == nil {
		return fmt.Errorf("Default DHCP Options Set is empty")
	}
	d.SetId(aws.StringValue(dhcpOptions[0].DhcpOptionsId))

	return resourceAwsVpcDhcpOptionsUpdate(d, meta)
}

func resourceAwsDefaultVpcDhcpOptionsDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Cannot destroy Default DHCP Options Set. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}
