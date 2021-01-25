package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicediscovery/waiter"
	"log"
)

func resourceAwsServiceDiscoveryInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsServiceDiscoveryInstanceCreate,
		ReadContext:   resourceAwsServiceDiscoveryInstanceRead,
		Update:        resourceAwsServiceDiscoveryInstanceUpdate,
		DeleteContext: resourceAwsServiceDiscoveryInstanceDelete,

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
				Type:     schema.TypeMap,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"creator_request_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

func resourceAwsServiceDiscoveryInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*AWSClient).sdconn

	input := &servicediscovery.RegisterInstanceInput{
		ServiceId:  aws.String(d.Get("service_id").(string)),
		InstanceId: aws.String(d.Get("instance_id").(string)),
		Attributes: stringMapToPointers(d.Get("attributes").(map[string]interface{})),
	}

	if v, ok := d.GetOk("creator_request_id"); ok {
		input.CreatorRequestId = aws.String(v.(string))
	}

	resp, err := conn.RegisterInstance(input)
	if err != nil {
		return diag.FromErr(err)
	}

	if resp != nil && resp.OperationId != nil {
		if _, err := waiter.OperationSuccess(conn, aws.StringValue(resp.OperationId)); err != nil {
			return diag.FromErr(fmt.Errorf("error waiting for Service Discovery Service Instance (%s) create: %w", d.Id(), err))
		}
	}

	d.SetId(d.Get("instance_id").(string))

	return resourceAwsServiceDiscoveryInstanceRead(ctx, d, meta)
}

func resourceAwsServiceDiscoveryInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).sdconn

	input := &servicediscovery.GetInstanceInput{
		ServiceId:  aws.String(d.Get("service_id").(string)),
		InstanceId: aws.String(d.Get("instance_id").(string)),
	}

	resp, err := conn.GetInstanceWithContext(ctx, input)
	if err != nil {
		if isAWSErr(err, servicediscovery.ErrCodeInstanceNotFound, "") {
			log.Printf("[WARN] Service Discovery Instance (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("attributes", aws.StringValueMap(resp.Instance.Attributes))

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

func resourceAwsServiceDiscoveryInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).sdconn

	input := &servicediscovery.DeregisterInstanceInput{
		ServiceId:  aws.String(d.Get("service_id").(string)),
		InstanceId: aws.String(d.Get("instance_id").(string)),
	}

	_, err := conn.DeregisterInstanceWithContext(ctx, input)

	if isAWSErr(err, servicediscovery.ErrCodeInstanceNotFound, "") {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deregistering Service Discovery Instance (%s): %w", d.Id(), err))
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
