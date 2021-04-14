package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/networkfirewall/finder"
)

func resourceAwsNetworkFirewallLoggingConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsNetworkFirewallLoggingConfigurationCreate,
		ReadContext:   resourceAwsNetworkFirewallLoggingConfigurationRead,
		UpdateContext: resourceAwsNetworkFirewallLoggingConfigurationUpdate,
		DeleteContext: resourceAwsNetworkFirewallLoggingConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"firewall_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"logging_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_destination_config": {
							// At most 2 configurations can exist,
							// with 1 destination for FLOW logs and 1 for ALERT logs
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 2,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_destination": {
										Type:     schema.TypeMap,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"log_destination_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(networkfirewall.LogDestinationType_Values(), false),
									},
									"log_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(networkfirewall.LogType_Values(), false),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceAwsNetworkFirewallLoggingConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn
	firewallArn := d.Get("firewall_arn").(string)

	log.Printf("[DEBUG] Adding Logging Configuration to NetworkFirewall Firewall: %s", firewallArn)

	loggingConfigs := expandNetworkFirewallLoggingConfiguration(d.Get("logging_configuration").([]interface{}))
	// cumulatively add the configured "log_destination_config" in "logging_configuration"
	err := putNetworkFirewallLoggingConfiguration(ctx, conn, firewallArn, loggingConfigs)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(firewallArn)

	return resourceAwsNetworkFirewallLoggingConfigurationRead(ctx, d, meta)
}

func resourceAwsNetworkFirewallLoggingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn

	log.Printf("[DEBUG] Reading Logging Configuration for NetworkFirewall Firewall: %s", d.Id())

	output, err := finder.LoggingConfiguration(ctx, conn, d.Id())
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Logging Configuration for NetworkFirewall Firewall (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Logging Configuration for NetworkFirewall Firewall: %s: %w", d.Id(), err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error reading Logging Configuration for NetworkFirewall Firewall: %s: empty output", d.Id()))
	}

	d.Set("firewall_arn", output.FirewallArn)

	if err := d.Set("logging_configuration", flattenNetworkFirewallLoggingConfiguration(output.LoggingConfiguration)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting logging_configuration: %w", err))
	}

	return nil
}

func resourceAwsNetworkFirewallLoggingConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn

	log.Printf("[DEBUG] Updating Logging Configuration for NetworkFirewall Firewall: %s", d.Id())

	o, n := d.GetChange("logging_configuration")
	// Remove destination configs one by one, if any
	if oldConfig := o.([]interface{}); len(oldConfig) != 0 && oldConfig[0] != nil {
		loggingConfig := expandNetworkFirewallLoggingConfigurationOnUpdate(oldConfig)
		if loggingConfig != nil {
			err := removeNetworkFirewallLoggingConfiguration(ctx, conn, d.Id(), loggingConfig)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	// Only send new LoggingConfiguration with content
	if newConfig := n.([]interface{}); len(newConfig) != 0 && newConfig[0] != nil {
		loggingConfigs := expandNetworkFirewallLoggingConfiguration(d.Get("logging_configuration").([]interface{}))
		// cumulatively add the configured "log_destination_config" in "logging_configuration"
		err := putNetworkFirewallLoggingConfiguration(ctx, conn, d.Id(), loggingConfigs)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceAwsNetworkFirewallLoggingConfigurationRead(ctx, d, meta)
}

func resourceAwsNetworkFirewallLoggingConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).networkfirewallconn

	log.Printf("[DEBUG] Deleting Logging Configuration for NetworkFirewall Firewall: %s", d.Id())

	output, err := finder.LoggingConfiguration(ctx, conn, d.Id())
	if err != nil {
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Logging Configuration for NetworkFirewall Firewall: %s: %w", d.Id(), err))
	}

	if output != nil && output.LoggingConfiguration != nil {
		err := removeNetworkFirewallLoggingConfiguration(ctx, conn, aws.StringValue(output.FirewallArn), output.LoggingConfiguration)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func putNetworkFirewallLoggingConfiguration(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string, l []*networkfirewall.LoggingConfiguration) error {
	var errors *multierror.Error
	for _, config := range l {
		input := &networkfirewall.UpdateLoggingConfigurationInput{
			FirewallArn:          aws.String(arn),
			LoggingConfiguration: config,
		}
		_, err := conn.UpdateLoggingConfigurationWithContext(ctx, input)
		if err != nil {
			errors = multierror.Append(errors, fmt.Errorf("error adding Logging Configuration to NetworkFirewall Firewall (%s): %w", arn, err))
		}
	}
	return errors.ErrorOrNil()
}

func removeNetworkFirewallLoggingConfiguration(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string, l *networkfirewall.LoggingConfiguration) error {
	if l == nil {
		return nil
	}
	var errors *multierror.Error
	// Must delete destination configs one at a time
	for i, config := range l.LogDestinationConfigs {
		input := &networkfirewall.UpdateLoggingConfigurationInput{
			FirewallArn: aws.String(arn),
		}
		if i == 0 && len(l.LogDestinationConfigs) == 2 {
			loggingConfig := &networkfirewall.LoggingConfiguration{
				LogDestinationConfigs: []*networkfirewall.LogDestinationConfig{config},
			}
			input.LoggingConfiguration = loggingConfig
		}
		_, err := conn.UpdateLoggingConfigurationWithContext(ctx, input)
		if err != nil {
			errors = multierror.Append(errors, fmt.Errorf("error removing Logging Configuration LogDestinationConfig (%v) from NetworkFirewall Firewall: %s: %w", config, arn, err))
		}
	}

	return errors.ErrorOrNil()
}

func expandNetworkFirewallLoggingConfiguration(l []interface{}) []*networkfirewall.LoggingConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	loggingConfigs := make([]*networkfirewall.LoggingConfiguration, 0)
	if tfSet, ok := tfMap["log_destination_config"].(*schema.Set); ok && tfSet.Len() > 0 {
		tfList := tfSet.List()
		for _, tfMapRaw := range tfList {
			tfMap, ok := tfMapRaw.(map[string]interface{})
			if !ok {
				continue
			}
			config := &networkfirewall.LogDestinationConfig{}
			if v, ok := tfMap["log_destination"].(map[string]interface{}); ok && len(v) > 0 {
				config.LogDestination = aws.StringMap(expandNetworkFirewallLogDestinationConfigLogDestination(v))
			}
			if v, ok := tfMap["log_destination_type"].(string); ok && v != "" {
				config.LogDestinationType = aws.String(v)
			}
			if v, ok := tfMap["log_type"].(string); ok && v != "" {
				config.LogType = aws.String(v)
			}
			// exclude empty LogDestinationConfig due to TypeMap in TypeSet behavior
			// Related: https://github.com/hashicorp/terraform-plugin-sdk/issues/588
			if config.LogDestination == nil && config.LogDestinationType == nil && config.LogType == nil {
				continue
			}
			loggingConfig := &networkfirewall.LoggingConfiguration{}
			// include all (max 2) "log_destination_config" i.e. prepend the already-expanded loggingConfig
			if len(loggingConfigs) == 1 && len(loggingConfigs[0].LogDestinationConfigs) == 1 {
				loggingConfig.LogDestinationConfigs = append(loggingConfig.LogDestinationConfigs, loggingConfigs[0].LogDestinationConfigs[0])
			}
			loggingConfig.LogDestinationConfigs = append(loggingConfig.LogDestinationConfigs, config)
			loggingConfigs = append(loggingConfigs, loggingConfig)
		}
	}
	return loggingConfigs
}

func expandNetworkFirewallLoggingConfigurationOnUpdate(l []interface{}) *networkfirewall.LoggingConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	loggingConfig := &networkfirewall.LoggingConfiguration{}
	if tfSet, ok := tfMap["log_destination_config"].(*schema.Set); ok && tfSet.Len() > 0 {
		tfList := tfSet.List()
		destConfigs := make([]*networkfirewall.LogDestinationConfig, 0, len(tfList))
		for _, tfMapRaw := range tfList {
			tfMap, ok := tfMapRaw.(map[string]interface{})
			if !ok {
				continue
			}
			config := &networkfirewall.LogDestinationConfig{}
			if v, ok := tfMap["log_destination"].(map[string]interface{}); ok && len(v) > 0 {
				config.LogDestination = aws.StringMap(expandNetworkFirewallLogDestinationConfigLogDestination(v))
			}
			if v, ok := tfMap["log_destination_type"].(string); ok && v != "" {
				config.LogDestinationType = aws.String(v)
			}
			if v, ok := tfMap["log_type"].(string); ok && v != "" {
				config.LogType = aws.String(v)
			}
			// exclude empty LogDestinationConfig due to TypeMap in TypeSet behavior
			// Related: https://github.com/hashicorp/terraform-plugin-sdk/issues/588
			if config.LogDestination == nil && config.LogDestinationType == nil && config.LogType == nil {
				continue
			}
			destConfigs = append(destConfigs, config)
		}
		loggingConfig.LogDestinationConfigs = destConfigs
	}
	return loggingConfig
}

func expandNetworkFirewallLogDestinationConfigLogDestination(dst map[string]interface{}) map[string]string {
	m := map[string]string{}
	for k, v := range dst {
		m[k] = v.(string)
	}
	return m
}

func flattenNetworkFirewallLoggingConfiguration(lc *networkfirewall.LoggingConfiguration) []interface{} {
	if lc == nil || lc.LogDestinationConfigs == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"log_destination_config": flattenNetworkFirewallLoggingConfigurationLogDestinationConfigs(lc.LogDestinationConfigs),
	}
	return []interface{}{m}
}

func flattenNetworkFirewallLoggingConfigurationLogDestinationConfigs(configs []*networkfirewall.LogDestinationConfig) []interface{} {
	l := make([]interface{}, 0, len(configs))
	for _, config := range configs {
		m := map[string]interface{}{
			"log_destination":      aws.StringValueMap(config.LogDestination),
			"log_destination_type": aws.StringValue(config.LogDestinationType),
			"log_type":             aws.StringValue(config.LogType),
		}
		l = append(l, m)
	}
	return l
}
