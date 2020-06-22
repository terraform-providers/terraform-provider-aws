package aws

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsApiGatewayRestApi() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayRestApiCreate,
		Read:   resourceAwsApiGatewayRestApiRead,
		Update: resourceAwsApiGatewayRestApiUpdate,
		Delete: resourceAwsApiGatewayRestApiDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"api_key_source": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					apigateway.ApiKeySourceTypeAuthorizer,
					apigateway.ApiKeySourceTypeHeader,
				}, true),
				Default: apigateway.ApiKeySourceTypeHeader,
			},

			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},

			"binary_media_types": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"body": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"minimum_compression_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      -1,
				ValidateFunc: validation.IntBetween(-1, 10485760),
			},

			"root_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"execution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"endpoint_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"types": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									apigateway.EndpointTypeEdge,
									apigateway.EndpointTypeRegional,
									apigateway.EndpointTypePrivate,
								}, false),
							},
						},
						"vpc_endpoint_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsApiGatewayRestApiCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	log.Printf("[DEBUG] Creating API Gateway")

	var description *string
	if d.Get("description").(string) != "" {
		description = aws.String(d.Get("description").(string))
	}

	params := &apigateway.CreateRestApiInput{
		Name:        aws.String(d.Get("name").(string)),
		Description: description,
	}

	if v, ok := d.GetOk("endpoint_configuration"); ok {
		params.EndpointConfiguration = expandApiGatewayEndpointConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("api_key_source"); ok && v.(string) != "" {
		params.ApiKeySource = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy"); ok && v.(string) != "" {
		params.Policy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		params.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().ApigatewayTags()
	}

	binaryMediaTypes, binaryMediaTypesOk := d.GetOk("binary_media_types")
	if binaryMediaTypesOk {
		params.BinaryMediaTypes = expandStringList(binaryMediaTypes.([]interface{}))
	}

	minimumCompressionSize := d.Get("minimum_compression_size").(int)
	if minimumCompressionSize > -1 {
		params.MinimumCompressionSize = aws.Int64(int64(minimumCompressionSize))
	}

	gateway, err := conn.CreateRestApi(params)
	if err != nil {
		return fmt.Errorf("Error creating API Gateway: %s", err)
	}

	d.SetId(*gateway.Id)

	if body, ok := d.GetOk("body"); ok {
		log.Printf("[DEBUG] Initializing API Gateway from OpenAPI spec %s", d.Id())
		_, err := conn.PutRestApi(&apigateway.PutRestApiInput{
			RestApiId: gateway.Id,
			Mode:      aws.String(apigateway.PutModeOverwrite),
			Body:      []byte(body.(string)),
		})
		if err != nil {
			return fmt.Errorf("error creating API Gateway specification: %s", err)
		}
	}

	return resourceAwsApiGatewayRestApiRead(d, meta)
}

func resourceAwsApiGatewayRestApiRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading API Gateway %s", d.Id())

	api, err := conn.GetRestApi(&apigateway.GetRestApiInput{
		RestApiId: aws.String(d.Id()),
	})
	if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] API Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway REST API (%s): %s", d.Id(), err)
	}

	getResourcesInput := &apigateway.GetResourcesInput{
		RestApiId: aws.String(d.Id()),
	}
	err = conn.GetResourcesPages(getResourcesInput, func(page *apigateway.GetResourcesOutput, lastPage bool) bool {
		for _, item := range page.Items {
			if aws.StringValue(item.Path) == "/" {
				d.Set("root_resource_id", item.Id)
				return false
			}
		}
		return !lastPage
	})
	if err != nil {
		return fmt.Errorf("error reading API Gateway REST API (%s) resources: %s", d.Id(), err)
	}

	d.Set("name", api.Name)
	d.Set("description", api.Description)
	d.Set("api_key_source", api.ApiKeySource)

	// The API returns policy as an escaped JSON string
	// {\\\"Version\\\":\\\"2012-10-17\\\",...}
	// The string must be normalized before unquoting as it may contain escaped
	// forward slashes in CIDR blocks, which will break strconv.Unquote

	// I'm not sure why it needs to be wrapped with double quotes first, but it does
	normalized_policy, err := structure.NormalizeJsonString(`"` + aws.StringValue(api.Policy) + `"`)
	if err != nil {
		fmt.Printf("error normalizing policy JSON: %s\n", err)
	}
	policy, err := strconv.Unquote(normalized_policy)
	if err != nil {
		return fmt.Errorf("error unescaping policy: %s", err)
	}
	d.Set("policy", policy)

	d.Set("binary_media_types", api.BinaryMediaTypes)

	execution_arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "execute-api",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  d.Id(),
	}.String()
	d.Set("execution_arn", execution_arn)

	if api.MinimumCompressionSize == nil {
		d.Set("minimum_compression_size", -1)
	} else {
		d.Set("minimum_compression_size", api.MinimumCompressionSize)
	}
	if err := d.Set("created_date", api.CreatedDate.Format(time.RFC3339)); err != nil {
		log.Printf("[DEBUG] Error setting created_date: %s", err)
	}

	if err := d.Set("endpoint_configuration", flattenApiGatewayEndpointConfiguration(api.EndpointConfiguration)); err != nil {
		return fmt.Errorf("error setting endpoint_configuration: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.ApigatewayKeyValueTags(api.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	rest_api_arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "apigateway",
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("/restapis/%s", d.Id()),
	}.String()
	d.Set("arn", rest_api_arn)

	return nil
}

func resourceAwsApiGatewayRestApiUpdateOperations(d *schema.ResourceData) []*apigateway.PatchOperation {
	operations := make([]*apigateway.PatchOperation, 0)

	if d.HasChange("name") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/name"),
			Value: aws.String(d.Get("name").(string)),
		})
	}

	if d.HasChange("description") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/description"),
			Value: aws.String(d.Get("description").(string)),
		})
	}

	if d.HasChange("api_key_source") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/apiKeySource"),
			Value: aws.String(d.Get("api_key_source").(string)),
		})
	}

	if d.HasChange("policy") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/policy"),
			Value: aws.String(d.Get("policy").(string)),
		})
	}

	if d.HasChange("minimum_compression_size") {
		minimumCompressionSize := d.Get("minimum_compression_size").(int)
		var value string
		if minimumCompressionSize > -1 {
			value = strconv.Itoa(minimumCompressionSize)
		}
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/minimumCompressionSize"),
			Value: aws.String(value),
		})
	}

	if d.HasChange("binary_media_types") {
		o, n := d.GetChange("binary_media_types")
		prefix := "binaryMediaTypes"

		old := o.([]interface{})
		new := n.([]interface{})

		// Remove every binary media types. Simpler to remove and add new ones,
		// since there are no replacings.
		for _, v := range old {
			operations = append(operations, &apigateway.PatchOperation{
				Op:   aws.String(apigateway.OpRemove),
				Path: aws.String(fmt.Sprintf("/%s/%s", prefix, escapeJsonPointer(v.(string)))),
			})
		}

		// Handle additions
		if len(new) > 0 {
			for _, v := range new {
				operations = append(operations, &apigateway.PatchOperation{
					Op:   aws.String(apigateway.OpAdd),
					Path: aws.String(fmt.Sprintf("/%s/%s", prefix, escapeJsonPointer(v.(string)))),
				})
			}
		}
	}

	if d.HasChange("endpoint_configuration.0.types") {
		// The REST API must have an endpoint type.
		// If attempting to remove the configuration, do nothing.
		if v, ok := d.GetOk("endpoint_configuration"); ok && len(v.([]interface{})) > 0 {
			m := v.([]interface{})[0].(map[string]interface{})

			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/endpointConfiguration/types/0"),
				Value: aws.String(m["types"].([]interface{})[0].(string)),
			})
		}
	}

	if d.HasChange("endpoint_configuration.0.vpc_endpoint_ids") {
		o, n := d.GetChange("endpoint_configuration.0.vpc_endpoint_ids")
		prefix := "/endpointConfiguration/vpcEndpointIds"

		old := o.(*schema.Set).List()
		new := n.(*schema.Set).List()

		for _, v := range old {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpRemove),
				Path:  aws.String(prefix),
				Value: aws.String(v.(string)),
			})
		}

		for _, v := range new {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpAdd),
				Path:  aws.String(prefix),
				Value: aws.String(v.(string)),
			})
		}
	}

	return operations
}

func resourceAwsApiGatewayRestApiUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	log.Printf("[DEBUG] Updating API Gateway %s", d.Id())

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.ApigatewayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	if d.HasChange("body") {
		if body, ok := d.GetOk("body"); ok {
			log.Printf("[DEBUG] Updating API Gateway from OpenAPI spec: %s", d.Id())
			_, err := conn.PutRestApi(&apigateway.PutRestApiInput{
				RestApiId: aws.String(d.Id()),
				Mode:      aws.String(apigateway.PutModeOverwrite),
				Body:      []byte(body.(string)),
			})
			if err != nil {
				return fmt.Errorf("error updating API Gateway specification: %s", err)
			}
		}
	}

	_, err := conn.UpdateRestApi(&apigateway.UpdateRestApiInput{
		RestApiId:       aws.String(d.Id()),
		PatchOperations: resourceAwsApiGatewayRestApiUpdateOperations(d),
	})

	if err != nil {
		return err
	}
	log.Printf("[DEBUG] Updated API Gateway %s", d.Id())

	return resourceAwsApiGatewayRestApiRead(d, meta)
}

func resourceAwsApiGatewayRestApiDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn

	input := &apigateway.DeleteRestApiInput{
		RestApiId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting API Gateway: %s", input)
	_, err := conn.DeleteRestApi(input)

	if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting API Gateway (%s): %s", d.Id(), err)
	}

	return nil
}

func expandApiGatewayEndpointConfiguration(l []interface{}) *apigateway.EndpointConfiguration {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	endpointConfiguration := &apigateway.EndpointConfiguration{
		Types: expandStringList(m["types"].([]interface{})),
	}

	if endpointIds, ok := m["vpc_endpoint_ids"]; ok {
		endpointConfiguration.VpcEndpointIds = expandStringSet(endpointIds.(*schema.Set))
	}

	return endpointConfiguration
}

func flattenApiGatewayEndpointConfiguration(endpointConfiguration *apigateway.EndpointConfiguration) []interface{} {
	if endpointConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"types": flattenStringList(endpointConfiguration.Types),
	}

	if len(endpointConfiguration.VpcEndpointIds) > 0 {
		m["vpc_endpoint_ids"] = aws.StringValueSlice(endpointConfiguration.VpcEndpointIds)
	}

	return []interface{}{m}
}
