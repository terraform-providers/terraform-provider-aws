package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceServiceDiscoveryDnsNamespace() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServiceDiscoveryDnsNamespaceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					servicediscovery.NamespaceTypeDnsPublic,
					servicediscovery.NamespaceTypeDnsPrivate,
				}, false),
			},
		},
	}
}

func dataSourceServiceDiscoveryDnsNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).sdconn

	name := d.Get("name").(string)

	input := &servicediscovery.ListNamespacesInput{}

	var filters []*servicediscovery.NamespaceFilter

	filter := &servicediscovery.NamespaceFilter{
		Condition: aws.String(servicediscovery.FilterConditionEq),
		Name:      aws.String(servicediscovery.NamespaceFilterNameType),
		Values:    []*string{aws.String(d.Get("type").(string))},
	}

	filters = append(filters, filter)

	input.Filters = filters

	namespaceIds := make([]string, 0)

	err := conn.ListNamespacesPagesWithContext(ctx, input, func(page *servicediscovery.ListNamespacesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, namespace := range page.Namespaces {
			if namespace == nil {
				continue
			}

			if name == aws.StringValue(namespace.Name) {
				namespaceIds = append(namespaceIds, aws.StringValue(namespace.Id))
			}
		}
		return !lastPage
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing Service Discovery DNS Namespaces: %w", err))
	}

	if len(namespaceIds) == 0 {
		return diag.Errorf("no matching Service Discovery DNS Namespace found")
	}

	if len(namespaceIds) != 1 {
		return diag.FromErr(fmt.Errorf("search returned %d Service Discovery DNS Namespaces, please revise so only one is returned", len(namespaceIds)))
	}

	d.SetId(namespaceIds[0])

	req := &servicediscovery.GetNamespaceInput{
		Id: aws.String(d.Id()),
	}

	output, err := conn.GetNamespaceWithContext(ctx, req)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Service Discovery DNS Namespace (%s): %w", d.Id(), err))
	}

	if output == nil || output.Namespace == nil {
		return diag.FromErr(fmt.Errorf("error reading Service Discovery DNS Namespace (%s): empty output", d.Id()))
	}

	namespace := output.Namespace

	d.Set("name", namespace.Name)
	d.Set("description", namespace.Description)
	d.Set("arn", namespace.Arn)
	if namespace.Properties != nil && namespace.Properties.DnsProperties != nil {
		d.Set("hosted_zone", namespace.Properties.DnsProperties.HostedZoneId)
	}

	return nil
}
