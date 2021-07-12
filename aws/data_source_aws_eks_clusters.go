package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEksClusters() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEksClustersRead,

		Schema: map[string]*schema.Schema{
			// Input values
			"filter": dataSourceFiltersSchema(),
			// Computed values
			"clusters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"certificate_authority": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled_cluster_log_types": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"endpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"identity": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"oidc": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"issuer": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"platform_version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tags": tagsSchemaComputed(),
						"vpc_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cluster_security_group_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"endpoint_private_access": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"endpoint_public_access": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"security_group_ids": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"subnet_ids": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"public_access_cidrs": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"vpc_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func tagInList(tagVal *string, filterVals []*string) bool {
	for _, filterVal := range filterVals {
		if *tagVal == *filterVal {
			return true
		}
	}
	return false
}

func shouldFilterCluster(filters []*ec2.Filter, tags keyvaluetags.KeyValueTags) bool {
	for _, filter := range filters {
		tagKey := strings.Split(*filter.Name, "tag:")[1]
		tagVal, ok := tags[tagKey]
		if !ok {
			return false
		}
		if !tagInList(tagVal.Value, filter.Values) {
			return false
		}
	}
	return true
}

func dataSourceAwsEksClustersRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).eksconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	filterParam, filtersOk := d.GetOk("filter")

	if !filtersOk {
		return fmt.Errorf("filter parameter is required")
	}

	filters := buildAwsDataSourceFilters(filterParam.(*schema.Set))

	for _, filter := range filters {
		if !strings.HasPrefix(*filter.Name, "tag:") {
			return fmt.Errorf("only filtering by tag is supported")
		}
	}

	params := &eks.ListClustersInput{}
	log.Printf("[DEBUG] Reading ListClusters: %s", params)
	var clusterNames []*string
	err := conn.ListClustersPages(params, func(resp *eks.ListClustersOutput, isLast bool) bool {
		clusterNames = append(clusterNames, resp.Clusters...)
		return !isLast
	})
	if err != nil {
		return fmt.Errorf("Error listing EKS clusters: %s", err)
	}

	var filteredClusters []map[string]interface{}
	for _, name := range clusterNames {
		describeClusterInput := eks.DescribeClusterInput{
			Name: name,
		}
		log.Printf("[DEBUG] Reading DescribeCluster: %s", describeClusterInput)
		output, err := conn.DescribeCluster(&describeClusterInput)

		if err != nil {
			return fmt.Errorf("error reading EKS Cluster (%s): %s", *name, err)
		}
		cluster := output.Cluster
		if cluster == nil {
			return fmt.Errorf("EKS Cluster (%s) not found", *name)
		}

		if shouldFilterCluster(filters, keyvaluetags.EksKeyValueTags(cluster.Tags)) {
			clusterSchema := make(map[string]interface{})
			clusterSchema["name"] = cluster.Name
			clusterSchema["arn"] = cluster.Arn
			clusterSchema["certificate_authority"] = flattenEksCertificate(cluster.CertificateAuthority)
			clusterSchema["created_at"] = aws.TimeValue(cluster.CreatedAt).String()
			clusterSchema["enabled_cluster_log_types"] = flattenEksEnabledLogTypes(cluster.Logging)
			clusterSchema["endpoint"] = cluster.Endpoint
			clusterSchema["identity"] = flattenEksIdentity(cluster.Identity)
			clusterSchema["platform_version"] = cluster.PlatformVersion
			clusterSchema["role_arn"] = cluster.RoleArn
			clusterSchema["status"] = cluster.Status
			clusterSchema["tags"] = keyvaluetags.EksKeyValueTags(cluster.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()
			clusterSchema["version"] = cluster.Version
			clusterSchema["vpc_config"] = flattenEksVpcConfigResponse(cluster.ResourcesVpcConfig)

			filteredClusters = append(filteredClusters, clusterSchema)
		}
	}

	if len(filteredClusters) == 0 {
		return fmt.Errorf("your query returned no results, please change your search criteria")
	}

	d.SetId(resource.UniqueId())
	if err := d.Set("clusters", filteredClusters); err != nil {
		return fmt.Errorf("error setting cluster: %s", err)
	}

	return err
}
