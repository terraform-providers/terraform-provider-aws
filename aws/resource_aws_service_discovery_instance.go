package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicediscovery/waiter"
)

func resourceAwsServiceDiscoveryInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceDiscoveryInstanceCreate,
		Read:   resourceAwsServiceDiscoveryInstanceRead,
		Update: resourceAwsServiceDiscoveryInstanceUpdate,
		Delete: resourceAwsServiceDiscoveryInstanceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"attributes": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 64),
								validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9!-~]+$`), "See https://docs.aws.amazon.com/cloud-map/latest/api/API_RegisterInstance.html#API_RegisterInstance_RequestSyntax"),
							),
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 1024),
								validation.StringMatch(regexp.MustCompile(`^([a-zA-Z0-9!-~][ \ta-zA-Z0-9!-~]*){0,1}[a-zA-Z0-9!-~]{0,1}$`), "See https://docs.aws.amazon.com/cloud-map/latest/api/API_RegisterInstance.html#API_RegisterInstance_RequestSyntax"),
							),
						},
					},
				},
			},
			"creator_request_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

func resourceAwsServiceDiscoveryInstanceCreate(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*AWSClient).sdconn

	input := &servicediscovery.RegisterInstanceInput{
		ServiceId: aws.String(d.Get("service_id").(string)),
		InstanceId: aws.String(d.Get("instance_id").(string)),
	}

	for _, attr := range d.Get("attributes").([]interface{}) {
		v := attr.(map[string]interface{})
		input.Attributes[v["key"].(string)] = aws.String(v["value"].(string))
	}


	if v, ok := d.GetOk("creator_request_id"); ok {
		input.CreatorRequestId = aws.String(v.(string))
	}

	resp, err := conn.RegisterInstance(input)
	if err != nil {
		return err
	}

	fmt.Println(resp.OperationId)

	return nil
	//d.SetId(aws.StringValue(resp.Service.Id))
	//
	//return resourceAwsServiceDiscoveryInstanceRead(d, meta)
}

func resourceAwsServiceDiscoveryInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sdconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &servicediscovery.GetServiceInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetService(input)
	if err != nil {
		if isAWSErr(err, servicediscovery.ErrCodeServiceNotFound, "") {
			log.Printf("[WARN] Service Discovery Service (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	service := resp.Service
	arn := aws.StringValue(service.Arn)
	d.Set("arn", arn)
	d.Set("name", service.Name)
	d.Set("description", service.Description)
	d.Set("namespace_id", service.NamespaceId)
	d.Set("dns_config", flattenServiceDiscoveryDnsConfig(service.DnsConfig))
	d.Set("health_check_config", flattenServiceDiscoveryHealthCheckConfig(service.HealthCheckConfig))
	d.Set("health_check_custom_config", flattenServiceDiscoveryHealthCheckCustomConfig(service.HealthCheckCustomConfig))

	tags, err := keyvaluetags.ServicediscoveryListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsServiceDiscoveryInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sdconn

	if d.HasChanges("description", "dns_config", "health_check_config") {
		input := &servicediscovery.UpdateServiceInput{
			Id: aws.String(d.Id()),
			Service: &servicediscovery.ServiceChange{
				Description: aws.String(d.Get("description").(string)),
				DnsConfig:   expandServiceDiscoveryDnsConfigChange(d.Get("dns_config").([]interface{})[0].(map[string]interface{})),
			},
		}

		hcconfig := d.Get("health_check_config").([]interface{})
		if len(hcconfig) > 0 {
			input.Service.HealthCheckConfig = expandServiceDiscoveryHealthCheckConfig(hcconfig[0].(map[string]interface{}))
		}

		output, err := conn.UpdateService(input)

		if err != nil {
			return fmt.Errorf("error updating Service Discovery Service (%s): %w", d.Id(), err)
		}

		if output != nil && output.OperationId != nil {
			if _, err := waiter.OperationSuccess(conn, aws.StringValue(output.OperationId)); err != nil {
				return fmt.Errorf("error waiting for Service Discovery Service (%s) update: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.ServicediscoveryUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Service Discovery Private DNS Namespace (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsServiceDiscoveryServiceRead(d, meta)
}

func resourceAwsServiceDiscoveryInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sdconn

	input := &servicediscovery.DeleteServiceInput{
		Id: aws.String(d.Id()),
	}

	_, err := conn.DeleteService(input)

	if isAWSErr(err, servicediscovery.ErrCodeServiceNotFound, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Service Discovery Service (%s): %w", d.Id(), err)
	}

	return nil
}

//func expandServiceDiscoveryDnsConfig(configured map[string]interface{}) *servicediscovery.DnsConfig {
//	result := &servicediscovery.DnsConfig{}
//
//	result.NamespaceId = aws.String(configured["namespace_id"].(string))
//	dnsRecords := configured["dns_records"].([]interface{})
//	drs := make([]*servicediscovery.DnsRecord, len(dnsRecords))
//	for i := range drs {
//		raw := dnsRecords[i].(map[string]interface{})
//		dr := &servicediscovery.DnsRecord{
//			TTL:  aws.Int64(int64(raw["ttl"].(int))),
//			Type: aws.String(raw["type"].(string)),
//		}
//		drs[i] = dr
//	}
//	result.DnsRecords = drs
//	if v, ok := configured["routing_policy"]; ok && v != "" {
//		result.RoutingPolicy = aws.String(v.(string))
//	}
//
//	return result
//}
//
//func flattenServiceDiscoveryDnsConfig(config *servicediscovery.DnsConfig) []map[string]interface{} {
//	if config == nil {
//		return nil
//	}
//
//	result := map[string]interface{}{}
//
//	if config.NamespaceId != nil {
//		result["namespace_id"] = *config.NamespaceId
//	}
//	if config.RoutingPolicy != nil {
//		result["routing_policy"] = *config.RoutingPolicy
//	}
//	if config.DnsRecords != nil {
//		drs := make([]map[string]interface{}, 0)
//		for _, v := range config.DnsRecords {
//			dr := map[string]interface{}{}
//			dr["ttl"] = *v.TTL
//			dr["type"] = *v.Type
//			drs = append(drs, dr)
//		}
//		result["dns_records"] = drs
//	}
//
//	if len(result) < 1 {
//		return nil
//	}
//
//	return []map[string]interface{}{result}
//}
//
//func expandServiceDiscoveryDnsConfigChange(configured map[string]interface{}) *servicediscovery.DnsConfigChange {
//	result := &servicediscovery.DnsConfigChange{}
//
//	dnsRecords := configured["dns_records"].([]interface{})
//	drs := make([]*servicediscovery.DnsRecord, len(dnsRecords))
//	for i := range drs {
//		raw := dnsRecords[i].(map[string]interface{})
//		dr := &servicediscovery.DnsRecord{
//			TTL:  aws.Int64(int64(raw["ttl"].(int))),
//			Type: aws.String(raw["type"].(string)),
//		}
//		drs[i] = dr
//	}
//	result.DnsRecords = drs
//
//	return result
//}
//
//func expandServiceDiscoveryHealthCheckConfig(configured map[string]interface{}) *servicediscovery.HealthCheckConfig {
//	if len(configured) < 1 {
//		return nil
//	}
//	result := &servicediscovery.HealthCheckConfig{}
//
//	if v, ok := configured["failure_threshold"]; ok && v.(int) != 0 {
//		result.FailureThreshold = aws.Int64(int64(v.(int)))
//	}
//	if v, ok := configured["resource_path"]; ok && v.(string) != "" {
//		result.ResourcePath = aws.String(v.(string))
//	}
//	if v, ok := configured["type"]; ok && v.(string) != "" {
//		result.Type = aws.String(v.(string))
//	}
//
//	return result
//}
//
//func flattenServiceDiscoveryHealthCheckConfig(config *servicediscovery.HealthCheckConfig) []map[string]interface{} {
//	if config == nil {
//		return nil
//	}
//	result := map[string]interface{}{}
//
//	if config.FailureThreshold != nil {
//		result["failure_threshold"] = *config.FailureThreshold
//	}
//	if config.ResourcePath != nil {
//		result["resource_path"] = *config.ResourcePath
//	}
//	if config.Type != nil {
//		result["type"] = *config.Type
//	}
//
//	if len(result) < 1 {
//		return nil
//	}
//
//	return []map[string]interface{}{result}
//}
//
//func expandServiceDiscoveryHealthCheckCustomConfig(configured map[string]interface{}) *servicediscovery.HealthCheckCustomConfig {
//	if len(configured) < 1 {
//		return nil
//	}
//	result := &servicediscovery.HealthCheckCustomConfig{}
//
//	if v, ok := configured["failure_threshold"]; ok && v.(int) != 0 {
//		result.FailureThreshold = aws.Int64(int64(v.(int)))
//	}
//
//	return result
//}
//
//func flattenServiceDiscoveryHealthCheckCustomConfig(config *servicediscovery.HealthCheckCustomConfig) []map[string]interface{} {
//	if config == nil {
//		return nil
//	}
//	result := map[string]interface{}{}
//
//	if config.FailureThreshold != nil {
//		result["failure_threshold"] = *config.FailureThreshold
//	}
//
//	if len(result) < 1 {
//		return nil
//	}
//
//	return []map[string]interface{}{result}
//}
