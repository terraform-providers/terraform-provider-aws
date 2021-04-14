package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apigatewayv2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func dataSourceAwsApiGatewayV2Api() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAwsApiGatewayV2ApiRead,

		Schema: map[string]*schema.Schema{
			"api_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"api_key_selection_expression": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cors_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_credentials": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"allow_headers": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      hashStringCaseInsensitive,
						},
						"allow_methods": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      hashStringCaseInsensitive,
						},
						"allow_origins": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      hashStringCaseInsensitive,
						},
						"expose_headers": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      hashStringCaseInsensitive,
						},
						"max_age": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disable_execute_api_endpoint": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"execution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"route_selection_expression": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsAwsApiGatewayV2ApiRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	apiID := d.Get("api_id").(string)

	api, err := finder.ApiByID(conn, apiID)

	if tfresource.NotFound(err) {
		return fmt.Errorf("no API Gateway v2 API matched; change the search criteria and try again")
	}

	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 API (%s): %w", apiID, err)
	}

	d.SetId(apiID)

	d.Set("api_endpoint", api.ApiEndpoint)
	d.Set("api_key_selection_expression", api.ApiKeySelectionExpression)
	apiArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "apigateway",
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("/apis/%s", d.Id()),
	}.String()
	d.Set("arn", apiArn)
	if err := d.Set("cors_configuration", flattenApiGateway2CorsConfiguration(api.CorsConfiguration)); err != nil {
		return fmt.Errorf("error setting cors_configuration: %w", err)
	}
	d.Set("description", api.Description)
	d.Set("disable_execute_api_endpoint", api.DisableExecuteApiEndpoint)
	executionArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "execute-api",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  d.Id(),
	}.String()
	d.Set("execution_arn", executionArn)
	d.Set("name", api.Name)
	d.Set("protocol_type", api.ProtocolType)
	d.Set("route_selection_expression", api.RouteSelectionExpression)
	if err := d.Set("tags", keyvaluetags.Apigatewayv2KeyValueTags(api.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}
	d.Set("version", api.Version)

	return nil
}
