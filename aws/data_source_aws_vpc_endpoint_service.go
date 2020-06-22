package aws

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsVpcEndpointService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsVpcEndpointServiceRead,

		Schema: map[string]*schema.Schema{
			"acceptance_required": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"availability_zones": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},
			"base_endpoint_dns_names": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Set:      schema.HashString,
			},
			"manages_vpc_endpoints": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"service_name"},
			},
			"service_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"service"},
			},
			"service_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"vpc_endpoint_policy_supported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"filter": dataSourceFiltersSchema(),
		},
	}
}

func dataSourceAwsVpcEndpointServiceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	filters, filtersOk := d.GetOk("filter")
	tags, tagsOk := d.GetOk("tags")

	var serviceName string
	serviceNameOk := false
	if v, ok := d.GetOk("service_name"); ok {
		serviceName = v.(string)
		serviceNameOk = true
	} else if v, ok := d.GetOk("service"); ok {
		serviceName = fmt.Sprintf("com.amazonaws.%s.%s", meta.(*AWSClient).region, v.(string))
		serviceNameOk = true
	}

	req := &ec2.DescribeVpcEndpointServicesInput{}
	if filtersOk {
		req.Filters = buildAwsDataSourceFilters(filters.(*schema.Set))
	}
	if serviceNameOk {
		req.ServiceNames = aws.StringSlice([]string{serviceName})
	}
	if tagsOk {
		req.Filters = append(req.Filters, ec2TagFiltersFromMap(tags.(map[string]interface{}))...)
	}

	log.Printf("[DEBUG] Reading VPC Endpoint Service: %s", req)
	resp, err := conn.DescribeVpcEndpointServices(req)
	if err != nil {
		return fmt.Errorf("error reading VPC Endpoint Service (%s): %s", serviceName, err)
	}

	if resp == nil || (len(resp.ServiceNames) == 0 && len(resp.ServiceDetails) == 0) {
		return fmt.Errorf("no matching VPC Endpoint Service found")
	}

	// Note: AWS Commercial now returns a response with `ServiceNames` and
	// `ServiceDetails`, but GovCloud responses only include `ServiceNames`
	if len(resp.ServiceDetails) == 0 {
		// GovCloud doesn't respect the filter.
		names := aws.StringValueSlice(resp.ServiceNames)
		for _, name := range names {
			if name == serviceName {
				d.SetId(strconv.Itoa(hashcode.String(name)))
				d.Set("service_name", name)
				return nil
			}
		}

		return fmt.Errorf("no matching VPC Endpoint Service found")
	}

	if len(resp.ServiceDetails) > 1 {
		return fmt.Errorf("multiple VPC Endpoint Services matched; use additional constraints to reduce matches to a single VPC Endpoint Service")
	}

	sd := resp.ServiceDetails[0]
	serviceName = aws.StringValue(sd.ServiceName)
	d.SetId(strconv.Itoa(hashcode.String(serviceName)))
	d.Set("service_name", serviceName)
	d.Set("acceptance_required", sd.AcceptanceRequired)
	err = d.Set("availability_zones", flattenStringSet(sd.AvailabilityZones))
	if err != nil {
		return fmt.Errorf("error setting availability_zones: %s", err)
	}
	err = d.Set("base_endpoint_dns_names", flattenStringSet(sd.BaseEndpointDnsNames))
	if err != nil {
		return fmt.Errorf("error setting base_endpoint_dns_names: %s", err)
	}
	d.Set("manages_vpc_endpoints", sd.ManagesVpcEndpoints)
	d.Set("owner", sd.Owner)
	d.Set("private_dns_name", sd.PrivateDnsName)
	d.Set("service_id", sd.ServiceId)
	d.Set("service_type", sd.ServiceType[0].ServiceType)
	err = d.Set("tags", keyvaluetags.Ec2KeyValueTags(sd.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())
	if err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}
	d.Set("vpc_endpoint_policy_supported", sd.VpcEndpointPolicySupported)

	return nil
}
