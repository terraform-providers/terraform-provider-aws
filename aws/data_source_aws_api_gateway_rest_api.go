package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsApiGatewayRestApi() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsApiGatewayRestApiRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"root_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"api_key_source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"minimum_compression_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"binary_media_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"endpoint_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"types": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_endpoint_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"execution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsApiGatewayRestApiRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	params := &apigateway.GetRestApisInput{}

	target := d.Get("name")
	var matchedApis []*apigateway.RestApi
	log.Printf("[DEBUG] Reading API Gateway REST APIs: %s", params)
	err := conn.GetRestApisPages(params, func(page *apigateway.GetRestApisOutput, lastPage bool) bool {
		for _, api := range page.Items {
			if aws.StringValue(api.Name) == target {
				matchedApis = append(matchedApis, api)
			}
		}
		return !lastPage
	})
	if err != nil {
		return fmt.Errorf("error describing API Gateway REST APIs: %s", err)
	}

	if len(matchedApis) == 0 {
		return fmt.Errorf("no REST APIs with name %q found in this region", target)
	}
	if len(matchedApis) > 1 {
		return fmt.Errorf("multiple REST APIs with name %q found in this region", target)
	}

	match := matchedApis[0]

	d.SetId(*match.Id)

	restApiArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "apigateway",
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("/restapis/%s", d.Id()),
	}.String()
	d.Set("arn", restApiArn)
	d.Set("description", match.Description)
	d.Set("policy", match.Policy)
	d.Set("api_key_source", match.ApiKeySource)
	d.Set("binary_media_types", match.BinaryMediaTypes)

	if match.MinimumCompressionSize == nil {
		d.Set("minimum_compression_size", -1)
	} else {
		d.Set("minimum_compression_size", match.MinimumCompressionSize)
	}

	if err := d.Set("endpoint_configuration", flattenApiGatewayEndpointConfiguration(match.EndpointConfiguration)); err != nil {
		return fmt.Errorf("error setting endpoint_configuration: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.ApigatewayKeyValueTags(match.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	executionArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "execute-api",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  d.Id(),
	}.String()
	d.Set("execution_arn", executionArn)

	resourceParams := &apigateway.GetResourcesInput{
		RestApiId: aws.String(d.Id()),
	}

	err = conn.GetResourcesPages(resourceParams, func(page *apigateway.GetResourcesOutput, lastPage bool) bool {
		for _, item := range page.Items {
			if *item.Path == "/" {
				d.Set("root_resource_id", item.Id)
				return false
			}
		}
		return !lastPage
	})

	return err
}
