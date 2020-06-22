package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicediscovery/waiter"
)

func resourceAwsServiceDiscoveryService() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceDiscoveryServiceCreate,
		Read:   resourceAwsServiceDiscoveryServiceRead,
		Update: resourceAwsServiceDiscoveryServiceUpdate,
		Delete: resourceAwsServiceDiscoveryServiceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"namespace_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"dns_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"dns_records": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ttl": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										ValidateFunc: validation.StringInSlice([]string{
											servicediscovery.RecordTypeSrv,
											servicediscovery.RecordTypeA,
											servicediscovery.RecordTypeAaaa,
											servicediscovery.RecordTypeCname,
										}, false),
									},
								},
							},
						},
						"routing_policy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  servicediscovery.RoutingPolicyMultivalue,
							ValidateFunc: validation.StringInSlice([]string{
								servicediscovery.RoutingPolicyMultivalue,
								servicediscovery.RoutingPolicyWeighted,
							}, false),
						},
					},
				},
			},
			"health_check_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"failure_threshold": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"resource_path": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								servicediscovery.HealthCheckTypeHttp,
								servicediscovery.HealthCheckTypeHttps,
								servicediscovery.HealthCheckTypeTcp,
							}, false),
						},
					},
				},
			},
			"health_check_custom_config": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"failure_threshold": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsServiceDiscoveryServiceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sdconn

	input := &servicediscovery.CreateServiceInput{
		Name: aws.String(d.Get("name").(string)),
		Tags: keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().ServicediscoveryTags(),
	}

	dnsConfig := d.Get("dns_config").([]interface{})
	if len(dnsConfig) > 0 {
		input.DnsConfig = expandServiceDiscoveryDnsConfig(dnsConfig[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("namespace_id"); ok {
		input.NamespaceId = aws.String(v.(string))
	}

	hcconfig := d.Get("health_check_config").([]interface{})
	if len(hcconfig) > 0 {
		input.HealthCheckConfig = expandServiceDiscoveryHealthCheckConfig(hcconfig[0].(map[string]interface{}))
	}

	healthCustomConfig := d.Get("health_check_custom_config").([]interface{})
	if len(healthCustomConfig) > 0 {
		input.HealthCheckCustomConfig = expandServiceDiscoveryHealthCheckCustomConfig(healthCustomConfig[0].(map[string]interface{}))
	}

	resp, err := conn.CreateService(input)
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(resp.Service.Id))

	return resourceAwsServiceDiscoveryServiceRead(d, meta)
}

func resourceAwsServiceDiscoveryServiceRead(d *schema.ResourceData, meta interface{}) error {
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

func resourceAwsServiceDiscoveryServiceUpdate(d *schema.ResourceData, meta interface{}) error {
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

func resourceAwsServiceDiscoveryServiceDelete(d *schema.ResourceData, meta interface{}) error {
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

func expandServiceDiscoveryDnsConfig(configured map[string]interface{}) *servicediscovery.DnsConfig {
	result := &servicediscovery.DnsConfig{}

	result.NamespaceId = aws.String(configured["namespace_id"].(string))
	dnsRecords := configured["dns_records"].([]interface{})
	drs := make([]*servicediscovery.DnsRecord, len(dnsRecords))
	for i := range drs {
		raw := dnsRecords[i].(map[string]interface{})
		dr := &servicediscovery.DnsRecord{
			TTL:  aws.Int64(int64(raw["ttl"].(int))),
			Type: aws.String(raw["type"].(string)),
		}
		drs[i] = dr
	}
	result.DnsRecords = drs
	if v, ok := configured["routing_policy"]; ok && v != "" {
		result.RoutingPolicy = aws.String(v.(string))
	}

	return result
}

func flattenServiceDiscoveryDnsConfig(config *servicediscovery.DnsConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{}

	if config.NamespaceId != nil {
		result["namespace_id"] = *config.NamespaceId
	}
	if config.RoutingPolicy != nil {
		result["routing_policy"] = *config.RoutingPolicy
	}
	if config.DnsRecords != nil {
		drs := make([]map[string]interface{}, 0)
		for _, v := range config.DnsRecords {
			dr := map[string]interface{}{}
			dr["ttl"] = *v.TTL
			dr["type"] = *v.Type
			drs = append(drs, dr)
		}
		result["dns_records"] = drs
	}

	if len(result) < 1 {
		return nil
	}

	return []map[string]interface{}{result}
}

func expandServiceDiscoveryDnsConfigChange(configured map[string]interface{}) *servicediscovery.DnsConfigChange {
	result := &servicediscovery.DnsConfigChange{}

	dnsRecords := configured["dns_records"].([]interface{})
	drs := make([]*servicediscovery.DnsRecord, len(dnsRecords))
	for i := range drs {
		raw := dnsRecords[i].(map[string]interface{})
		dr := &servicediscovery.DnsRecord{
			TTL:  aws.Int64(int64(raw["ttl"].(int))),
			Type: aws.String(raw["type"].(string)),
		}
		drs[i] = dr
	}
	result.DnsRecords = drs

	return result
}

func expandServiceDiscoveryHealthCheckConfig(configured map[string]interface{}) *servicediscovery.HealthCheckConfig {
	if len(configured) < 1 {
		return nil
	}
	result := &servicediscovery.HealthCheckConfig{}

	if v, ok := configured["failure_threshold"]; ok && v.(int) != 0 {
		result.FailureThreshold = aws.Int64(int64(v.(int)))
	}
	if v, ok := configured["resource_path"]; ok && v.(string) != "" {
		result.ResourcePath = aws.String(v.(string))
	}
	if v, ok := configured["type"]; ok && v.(string) != "" {
		result.Type = aws.String(v.(string))
	}

	return result
}

func flattenServiceDiscoveryHealthCheckConfig(config *servicediscovery.HealthCheckConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}
	result := map[string]interface{}{}

	if config.FailureThreshold != nil {
		result["failure_threshold"] = *config.FailureThreshold
	}
	if config.ResourcePath != nil {
		result["resource_path"] = *config.ResourcePath
	}
	if config.Type != nil {
		result["type"] = *config.Type
	}

	if len(result) < 1 {
		return nil
	}

	return []map[string]interface{}{result}
}

func expandServiceDiscoveryHealthCheckCustomConfig(configured map[string]interface{}) *servicediscovery.HealthCheckCustomConfig {
	if len(configured) < 1 {
		return nil
	}
	result := &servicediscovery.HealthCheckCustomConfig{}

	if v, ok := configured["failure_threshold"]; ok && v.(int) != 0 {
		result.FailureThreshold = aws.Int64(int64(v.(int)))
	}

	return result
}

func flattenServiceDiscoveryHealthCheckCustomConfig(config *servicediscovery.HealthCheckCustomConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}
	result := map[string]interface{}{}

	if config.FailureThreshold != nil {
		result["failure_threshold"] = *config.FailureThreshold
	}

	if len(result) < 1 {
		return nil
	}

	return []map[string]interface{}{result}
}
