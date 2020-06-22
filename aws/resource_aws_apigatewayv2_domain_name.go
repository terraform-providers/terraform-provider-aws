package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	apiGatewayV2DomainNameStatusDeleted = "DELETED"
)

func resourceAwsApiGatewayV2DomainName() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayV2DomainNameCreate,
		Read:   resourceAwsApiGatewayV2DomainNameRead,
		Update: resourceAwsApiGatewayV2DomainNameUpdate,
		Delete: resourceAwsApiGatewayV2DomainNameDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"api_mapping_selection_expression": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 512),
			},
			"domain_name_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"endpoint_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								apigatewayv2.EndpointTypeRegional,
							}, true),
						},
						"hosted_zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"security_policy": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								apigatewayv2.SecurityPolicyTls12,
							}, true),
						},
						"target_domain_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsApiGatewayV2DomainNameCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.CreateDomainNameInput{
		DomainName:               aws.String(d.Get("domain_name").(string)),
		DomainNameConfigurations: expandApiGatewayV2DomainNameConfiguration(d.Get("domain_name_configuration").([]interface{})),
		Tags:                     keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().Apigatewayv2Tags(),
	}

	log.Printf("[DEBUG] Creating API Gateway v2 domain name: %s", req)
	resp, err := conn.CreateDomainName(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 domain name: %s", err)
	}

	d.SetId(aws.StringValue(resp.DomainName))

	return resourceAwsApiGatewayV2DomainNameRead(d, meta)
}

func resourceAwsApiGatewayV2DomainNameRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	respRaw, state, err := apiGatewayV2DomainNameRefresh(conn, d.Id())()
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 domain name (%s): %s", d.Id(), err)
	}

	if state == apiGatewayV2DomainNameStatusDeleted {
		log.Printf("[WARN] API Gateway v2 domain name (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	resp := respRaw.(*apigatewayv2.GetDomainNameOutput)
	d.Set("api_mapping_selection_expression", resp.ApiMappingSelectionExpression)
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "apigateway",
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("/domainnames/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("domain_name", resp.DomainName)
	err = d.Set("domain_name_configuration", flattenApiGatewayV2DomainNameConfiguration(resp.DomainNameConfigurations[0]))
	if err != nil {
		return fmt.Errorf("error setting domain_name_configuration: %s", err)
	}
	if err := d.Set("tags", keyvaluetags.Apigatewayv2KeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsApiGatewayV2DomainNameUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	if d.HasChange("domain_name_configuration") {
		req := &apigatewayv2.UpdateDomainNameInput{
			DomainName:               aws.String(d.Id()),
			DomainNameConfigurations: expandApiGatewayV2DomainNameConfiguration(d.Get("domain_name_configuration").([]interface{})),
		}

		log.Printf("[DEBUG] Updating API Gateway v2 domain name: %s", req)
		_, err := conn.UpdateDomainName(req)
		if err != nil {
			return fmt.Errorf("error updating API Gateway v2 domain name (%s): %s", d.Id(), err)
		}

		if err := waitForApiGatewayV2DomainNameAvailabilityOnUpdate(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for API Gateway v2 domain name (%s) to become available: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Apigatewayv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating API Gateway v2 domain name (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsApiGatewayV2DomainNameRead(d, meta)
}

func resourceAwsApiGatewayV2DomainNameDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Deleting API Gateway v2 domain name (%s)", d.Id())
	_, err := conn.DeleteDomainName(&apigatewayv2.DeleteDomainNameInput{
		DomainName: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 domain name (%s): %s", d.Id(), err)
	}

	return nil
}

func apiGatewayV2DomainNameRefresh(conn *apigatewayv2.ApiGatewayV2, domainName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.GetDomainName(&apigatewayv2.GetDomainNameInput{
			DomainName: aws.String(domainName),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			return "", apiGatewayV2DomainNameStatusDeleted, nil
		}
		if err != nil {
			return nil, "", err
		}

		if n := len(resp.DomainNameConfigurations); n != 1 {
			return nil, "", fmt.Errorf("Found %d domain name configurations for %s, expected 1", n, domainName)
		}

		domainNameConfiguration := resp.DomainNameConfigurations[0]
		if statusMessage := aws.StringValue(domainNameConfiguration.DomainNameStatusMessage); statusMessage != "" {
			log.Printf("[INFO] Domain name (%s) status message: %s", domainName, statusMessage)
		}

		return resp, aws.StringValue(domainNameConfiguration.DomainNameStatus), nil
	}
}

func waitForApiGatewayV2DomainNameAvailabilityOnUpdate(conn *apigatewayv2.ApiGatewayV2, domainName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{apigatewayv2.DomainNameStatusUpdating},
		Target:     []string{apigatewayv2.DomainNameStatusAvailable},
		Refresh:    apiGatewayV2DomainNameRefresh(conn, domainName),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func expandApiGatewayV2DomainNameConfiguration(vDomainNameConfiguration []interface{}) []*apigatewayv2.DomainNameConfiguration {
	if len(vDomainNameConfiguration) == 0 || vDomainNameConfiguration[0] == nil {
		return nil
	}
	mDomainNameConfiguration := vDomainNameConfiguration[0].(map[string]interface{})

	return []*apigatewayv2.DomainNameConfiguration{{
		CertificateArn: aws.String(mDomainNameConfiguration["certificate_arn"].(string)),
		EndpointType:   aws.String(mDomainNameConfiguration["endpoint_type"].(string)),
		SecurityPolicy: aws.String(mDomainNameConfiguration["security_policy"].(string)),
	}}
}

func flattenApiGatewayV2DomainNameConfiguration(domainNameConfiguration *apigatewayv2.DomainNameConfiguration) []interface{} {
	if domainNameConfiguration == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"certificate_arn":    aws.StringValue(domainNameConfiguration.CertificateArn),
		"endpoint_type":      aws.StringValue(domainNameConfiguration.EndpointType),
		"hosted_zone_id":     aws.StringValue(domainNameConfiguration.HostedZoneId),
		"security_policy":    aws.StringValue(domainNameConfiguration.SecurityPolicy),
		"target_domain_name": aws.StringValue(domainNameConfiguration.ApiGatewayDomainName),
	}}
}
