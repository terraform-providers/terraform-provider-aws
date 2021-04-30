package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsSesConfigurationSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesConfigurationSetCreate,
		Read:   resourceAwsSesConfigurationSetRead,
		Update: resourceAwsSesConfigurationSetUpdate,
		Delete: resourceAwsSesConfigurationSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delivery_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tls_policy": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      ses.TlsPolicyOptional,
							ValidateFunc: validation.StringInSlice(ses.TlsPolicy_Values(), false),
						},
					},
				},
			},
			"last_fresh_start": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reputation_metrics_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"sending_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

func resourceAwsSesConfigurationSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	configurationSetName := d.Get("name").(string)

	createOpts := &ses.CreateConfigurationSetInput{
		ConfigurationSet: &ses.ConfigurationSet{
			Name: aws.String(configurationSetName),
		},
	}

	_, err := conn.CreateConfigurationSet(createOpts)
	if err != nil {
		return fmt.Errorf("error creating SES configuration set (%s): %w", configurationSetName, err)
	}

	d.SetId(configurationSetName)

	if v, ok := d.GetOk("delivery_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := &ses.PutConfigurationSetDeliveryOptionsInput{
			ConfigurationSetName: aws.String(configurationSetName),
			DeliveryOptions:      expandSesConfigurationSetDeliveryOptions(v.([]interface{})),
		}

		_, err := conn.PutConfigurationSetDeliveryOptions(input)
		if err != nil {
			return fmt.Errorf("error adding SES configuration set (%s) delivery options: %w", configurationSetName, err)
		}
	}

	if v := d.Get("reputation_metrics_enabled"); v.(bool) {
		input := &ses.UpdateConfigurationSetReputationMetricsEnabledInput{
			ConfigurationSetName: aws.String(configurationSetName),
			Enabled:              aws.Bool(v.(bool)),
		}

		_, err := conn.UpdateConfigurationSetReputationMetricsEnabled(input)
		if err != nil {
			return fmt.Errorf("error adding SES configuration set (%s) reputation metrics enabled: %w", configurationSetName, err)
		}
	}

	if v := d.Get("sending_enabled"); !v.(bool) {
		input := &ses.UpdateConfigurationSetSendingEnabledInput{
			ConfigurationSetName: aws.String(configurationSetName),
			Enabled:              aws.Bool(v.(bool)),
		}

		_, err := conn.UpdateConfigurationSetSendingEnabled(input)
		if err != nil {
			return fmt.Errorf("error adding SES configuration set (%s) sending enabled: %w", configurationSetName, err)
		}
	}

	return resourceAwsSesConfigurationSetRead(d, meta)
}

func resourceAwsSesConfigurationSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	configSetInput := &ses.DescribeConfigurationSetInput{
		ConfigurationSetName: aws.String(d.Id()),
		ConfigurationSetAttributeNames: aws.StringSlice([]string{
			ses.ConfigurationSetAttributeDeliveryOptions,
			ses.ConfigurationSetAttributeReputationOptions}),
	}

	response, err := conn.DescribeConfigurationSet(configSetInput)

	if err != nil {
		if isAWSErr(err, ses.ErrCodeConfigurationSetDoesNotExistException, "") {
			log.Printf("[WARN] SES Configuration Set (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if err := d.Set("delivery_options", flattenSesConfigurationSetDeliveryOptions(response.DeliveryOptions)); err != nil {
		return fmt.Errorf("error setting delivery_options: %w", err)
	}

	d.Set("name", response.ConfigurationSet.Name)

	repOpts := response.ReputationOptions
	if repOpts != nil {
		d.Set("reputation_metrics_enabled", repOpts.ReputationMetricsEnabled)
		d.Set("sending_enabled", repOpts.SendingEnabled)
		d.Set("last_fresh_start", aws.TimeValue(repOpts.LastFreshStart).Format(time.RFC3339))
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   ses.ServiceName,
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("configuration-set/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return nil
}

func resourceAwsSesConfigurationSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	if d.HasChange("delivery_options") {
		input := &ses.PutConfigurationSetDeliveryOptionsInput{
			ConfigurationSetName: aws.String(d.Id()),
			DeliveryOptions:      expandSesConfigurationSetDeliveryOptions(d.Get("delivery_options").([]interface{})),
		}

		_, err := conn.PutConfigurationSetDeliveryOptions(input)
		if err != nil {
			return fmt.Errorf("error updating SES configuration set (%s) delivery options: %w", d.Id(), err)
		}
	}

	if d.HasChange("reputation_metrics_enabled") {
		input := &ses.UpdateConfigurationSetReputationMetricsEnabledInput{
			ConfigurationSetName: aws.String(d.Id()),
			Enabled:              aws.Bool(d.Get("reputation_metrics_enabled").(bool)),
		}

		_, err := conn.UpdateConfigurationSetReputationMetricsEnabled(input)
		if err != nil {
			return fmt.Errorf("error updating SES configuration set (%s) reputation metrics enabled: %w", d.Id(), err)
		}
	}

	if d.HasChange("sending_enabled") {
		input := &ses.UpdateConfigurationSetSendingEnabledInput{
			ConfigurationSetName: aws.String(d.Id()),
			Enabled:              aws.Bool(d.Get("sending_enabled").(bool)),
		}

		_, err := conn.UpdateConfigurationSetSendingEnabled(input)
		if err != nil {
			return fmt.Errorf("error updating SES configuration set (%s) reputation metrics enabled: %w", d.Id(), err)
		}
	}

	return resourceAwsSesConfigurationSetRead(d, meta)
}

func resourceAwsSesConfigurationSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesconn

	log.Printf("[DEBUG] SES Delete Configuration Rule Set: %s", d.Id())
	input := &ses.DeleteConfigurationSetInput{
		ConfigurationSetName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteConfigurationSet(input); err != nil {
		if !isAWSErr(err, ses.ErrCodeConfigurationSetDoesNotExistException, "") {
			return fmt.Errorf("error deleting SES Configuration Set (%s): %w", d.Id(), err)
		}
	}

	return nil
}

func expandSesConfigurationSetDeliveryOptions(l []interface{}) *ses.DeliveryOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &ses.DeliveryOptions{}

	if v, ok := tfMap["tls_policy"].(string); ok && v != "" {
		options.TlsPolicy = aws.String(v)
	}

	return options
}

func flattenSesConfigurationSetDeliveryOptions(options *ses.DeliveryOptions) []interface{} {
	if options == nil {
		return nil
	}

	m := map[string]interface{}{
		"tls_policy": aws.StringValue(options.TlsPolicy),
	}

	return []interface{}{m}
}
