package aws

import (
	"fmt"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAwsServiceDiscoveryDnsNamespace() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsServiceDiscoveryDnsNamespaceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"public",
					"private",
				}, false),
			},
			"most_recent": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsServiceDiscoveryDnsNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sdconn

	name := d.Get("name").(string)
	nsType := d.Get("type").(string)

	nsTypeMapping := map[string]string{
		"public":  servicediscovery.NamespaceTypeDnsPublic,
		"private": servicediscovery.NamespaceTypeDnsPrivate,
	}

	input := &servicediscovery.ListNamespacesInput{
		Filters: []*servicediscovery.NamespaceFilter{
			{
				Name:      aws.String("TYPE"),
				Condition: aws.String("EQ"),
				Values:    aws.StringSlice([]string{nsTypeMapping[nsType]}),
			},
		}}

	var namespacesFound []*servicediscovery.NamespaceSummary
	err := conn.ListNamespacesPages(input, func(page *servicediscovery.ListNamespacesOutput, lastPage bool) bool {
		for _, ns := range page.Namespaces {
			if aws.StringValue(ns.Name) == name {
				namespacesFound = append(namespacesFound, ns)
			}
		}

		return !lastPage
	})
	if err != nil {
		return err
	}

	var namespace *servicediscovery.NamespaceSummary
	if len(namespacesFound) < 1 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	if len(namespacesFound) > 1 {
		recent := d.Get("most_recent").(bool)
		log.Printf("[DEBUG] aws_service_discovery_dns_namespace - multiple results found "+
			"and `most_recent` is set to: %t", recent)
		if !recent {
			return fmt.Errorf("Your query returned more than one result. " +
				"You can set `most_recent` attribute to true.")
		}

		namespace = mostRecentNS(namespacesFound)
	} else {
		// Query returned single result.
		namespace = namespacesFound[0]
	}

	d.SetId(aws.StringValue(namespace.Id))
	d.Set("arn", namespace.Arn)
	if namespace.Properties != nil {
		d.Set("hosted_zone", namespace.Properties.DnsProperties.HostedZoneId)
	}

	log.Printf("[DEBUG] aws_service_discovery_dns_namespace - Single DNS Namespace found: %s", *namespace.Id)
	return nil
}

type nsSort []*servicediscovery.NamespaceSummary

func (n nsSort) Len() int      { return len(n) }
func (n nsSort) Swap(i, j int) { n[i], n[j] = n[j], n[i] }
func (n nsSort) Less(i, j int) bool {
	itime := *n[i].CreateDate
	jtime := *n[j].CreateDate
	return jtime.Unix() < itime.Unix()
}

func mostRecentNS(nses []*servicediscovery.NamespaceSummary) *servicediscovery.NamespaceSummary {
	sort.Sort(nsSort(nses))
	return nses[0]
}
