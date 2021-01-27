package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicediscovery/waiter"
	"log"
	"strings"
)

func resourceAwsServiceDiscoveryInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsServiceDiscoveryInstanceCreate,
		ReadContext:   resourceAwsServiceDiscoveryInstanceRead,
		UpdateContext: resourceAwsServiceDiscoveryInstanceUpdate,
		DeleteContext: resourceAwsServiceDiscoveryInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceAwsServiceDiscoveryInstanceImport,
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

	err = d.Set("attributes", aws.StringValueMap(resp.Instance.Attributes))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAwsServiceDiscoveryInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceAwsServiceDiscoveryInstanceCreate(ctx, d, meta)
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

func resourceAwsServiceDiscoveryInstanceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format (%q), expected <service-id>/<instance-id>", d.Id())
	}

	serviceId := idParts[0]
	instanceId := idParts[1]

	d.Set("service_id", serviceId)
	d.Set("instance_id", instanceId)
	d.SetId(instanceId)

	return []*schema.ResourceData{d}, nil
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
