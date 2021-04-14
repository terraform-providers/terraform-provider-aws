package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apigatewayv2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apigatewayv2/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
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
			Create: schema.DefaultTimeout(10 * time.Minute),
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
			"mutual_tls_authentication": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"truststore_uri": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"truststore_version": {
							Type:     schema.TypeString,
							Optional: true,
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
	domainName := d.Get("domain_name").(string)

	input := &apigatewayv2.CreateDomainNameInput{
		DomainName:               aws.String(domainName),
		DomainNameConfigurations: expandApiGatewayV2DomainNameConfiguration(d.Get("domain_name_configuration").([]interface{})),
		MutualTlsAuthentication:  expandApiGatewayV2MutualTlsAuthentication(d.Get("mutual_tls_authentication").([]interface{})),
		Tags:                     keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().Apigatewayv2Tags(),
	}

	log.Printf("[DEBUG] Creating API Gateway v2 domain name: %s", input)
	output, err := conn.CreateDomainName(input)

	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 domain name (%s): %w", domainName, err)
	}

	d.SetId(aws.StringValue(output.DomainName))

	if _, err := waiter.DomainNameAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for API Gateway v2 domain name (%s) to become available: %w", d.Id(), err)
	}

	return resourceAwsApiGatewayV2DomainNameRead(d, meta)
}

func resourceAwsApiGatewayV2DomainNameRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	output, err := finder.DomainNameByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway v2 domain name (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 domain name (%s): %w", d.Id(), err)
	}

	d.Set("api_mapping_selection_expression", output.ApiMappingSelectionExpression)
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "apigateway",
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("/domainnames/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("domain_name", output.DomainName)
	err = d.Set("domain_name_configuration", flattenApiGatewayV2DomainNameConfiguration(output.DomainNameConfigurations[0]))
	if err != nil {
		return fmt.Errorf("error setting domain_name_configuration: %w", err)
	}
	err = d.Set("mutual_tls_authentication", flattenApiGatewayV2MutualTlsAuthentication(output.MutualTlsAuthentication))
	if err != nil {
		return fmt.Errorf("error setting mutual_tls_authentication: %w", err)
	}
	if err := d.Set("tags", keyvaluetags.Apigatewayv2KeyValueTags(output.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsApiGatewayV2DomainNameUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	if d.HasChanges("domain_name_configuration", "mutual_tls_authentication") {
		input := &apigatewayv2.UpdateDomainNameInput{
			DomainName:               aws.String(d.Id()),
			DomainNameConfigurations: expandApiGatewayV2DomainNameConfiguration(d.Get("domain_name_configuration").([]interface{})),
		}

		if d.HasChange("mutual_tls_authentication") {
			vMutualTlsAuthentication := d.Get("mutual_tls_authentication").([]interface{})

			if len(vMutualTlsAuthentication) == 0 || vMutualTlsAuthentication[0] == nil {
				// To disable mutual TLS for a custom domain name, remove the truststore from your custom domain name.
				input.MutualTlsAuthentication = &apigatewayv2.MutualTlsAuthenticationInput{
					TruststoreUri: aws.String(""),
				}
			} else {
				input.MutualTlsAuthentication = &apigatewayv2.MutualTlsAuthenticationInput{
					TruststoreVersion: aws.String(vMutualTlsAuthentication[0].(map[string]interface{})["truststore_version"].(string)),
				}
			}
		}

		log.Printf("[DEBUG] Updating API Gateway v2 domain name: %s", input)
		_, err := conn.UpdateDomainName(input)

		if err != nil {
			return fmt.Errorf("error updating API Gateway v2 domain name (%s): %w", d.Id(), err)
		}

		if _, err := waiter.DomainNameAvailable(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for API Gateway v2 domain name (%s) to become available: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Apigatewayv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating API Gateway v2 domain name (%s) tags: %w", d.Id(), err)
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

	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 domain name (%s): %w", d.Id(), err)
	}

	return nil
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

func expandApiGatewayV2MutualTlsAuthentication(vMutualTlsAuthentication []interface{}) *apigatewayv2.MutualTlsAuthenticationInput {
	if len(vMutualTlsAuthentication) == 0 || vMutualTlsAuthentication[0] == nil {
		return nil
	}
	mMutualTlsAuthentication := vMutualTlsAuthentication[0].(map[string]interface{})

	mutualTlsAuthentication := &apigatewayv2.MutualTlsAuthenticationInput{
		TruststoreUri: aws.String(mMutualTlsAuthentication["truststore_uri"].(string)),
	}

	if vTruststoreVersion, ok := mMutualTlsAuthentication["truststore_version"].(string); ok && vTruststoreVersion != "" {
		mutualTlsAuthentication.TruststoreVersion = aws.String(vTruststoreVersion)
	}

	return mutualTlsAuthentication
}

func flattenApiGatewayV2MutualTlsAuthentication(mutualTlsAuthentication *apigatewayv2.MutualTlsAuthentication) []interface{} {
	if mutualTlsAuthentication == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"truststore_uri":     aws.StringValue(mutualTlsAuthentication.TruststoreUri),
		"truststore_version": aws.StringValue(mutualTlsAuthentication.TruststoreVersion),
	}}
}
