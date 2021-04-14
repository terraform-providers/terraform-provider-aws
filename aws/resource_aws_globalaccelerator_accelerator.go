package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/globalaccelerator/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/globalaccelerator/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

// Global Route53 Zone ID for Global Accelerators, exported as a
// convenience attribute for Route53 aliases (see
// https://docs.aws.amazon.com/Route53/latest/APIReference/API_AliasTarget.html).
const globalAcceleratorRoute53ZoneID = "Z2BJ6XQ5FK7U4H"

func resourceAwsGlobalAcceleratorAccelerator() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlobalAcceleratorAcceleratorCreate,
		Read:   resourceAwsGlobalAcceleratorAcceleratorRead,
		Update: resourceAwsGlobalAcceleratorAcceleratorUpdate,
		Delete: resourceAwsGlobalAcceleratorAcceleratorDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^[0-9A-Za-z-]+$`), "only alphanumeric characters and hyphens are allowed"),
					validation.StringDoesNotMatch(regexp.MustCompile(`^-`), "cannot start with a hyphen"),
					validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end with a hyphen"),
				),
			},
			"ip_address_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      globalaccelerator.IpAddressTypeIpv4,
				ValidateFunc: validation.StringInSlice(globalaccelerator.IpAddressType_Values(), false),
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_sets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_addresses": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ip_family": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"attributes": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"flow_logs_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"flow_logs_s3_bucket": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"flow_logs_s3_prefix": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsGlobalAcceleratorAcceleratorCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn

	name := d.Get("name").(string)
	input := &globalaccelerator.CreateAcceleratorInput{
		Name:             aws.String(name),
		IdempotencyToken: aws.String(resource.UniqueId()),
		Enabled:          aws.Bool(d.Get("enabled").(bool)),
		Tags:             keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GlobalacceleratorTags(),
	}

	if v, ok := d.GetOk("ip_address_type"); ok {
		input.IpAddressType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Global Accelerator Accelerator: %s", input)
	output, err := conn.CreateAccelerator(input)

	if err != nil {
		return fmt.Errorf("error creating Global Accelerator Accelerator (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.Accelerator.AcceleratorArn))

	if _, err := waiter.AcceleratorDeployed(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Global Accelerator Accelerator (%s) deployment: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("attributes"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := expandGlobalAcceleratorUpdateAcceleratorAttributesInput(v.([]interface{})[0].(map[string]interface{}))
		input.AcceleratorArn = aws.String(d.Id())

		log.Printf("[DEBUG] Updating Global Accelerator Accelerator attributes: %s", input)
		if _, err := conn.UpdateAcceleratorAttributes(input); err != nil {
			return fmt.Errorf("error updating Global Accelerator Accelerator (%s) attributes: %w", d.Id(), err)
		}

		if _, err := waiter.AcceleratorDeployed(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return fmt.Errorf("error waiting for Global Accelerator Accelerator (%s) deployment: %w", d.Id(), err)
		}
	}

	return resourceAwsGlobalAcceleratorAcceleratorRead(d, meta)
}

func resourceAwsGlobalAcceleratorAcceleratorRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	accelerator, err := finder.AcceleratorByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Global Accelerator Accelerator (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Global Accelerator Accelerator (%s): %w", d.Id(), err)
	}

	d.Set("enabled", accelerator.Enabled)
	d.Set("dns_name", accelerator.DnsName)
	d.Set("hosted_zone_id", globalAcceleratorRoute53ZoneID)
	d.Set("name", accelerator.Name)
	d.Set("ip_address_type", accelerator.IpAddressType)

	if err := d.Set("ip_sets", flattenGlobalAcceleratorIpSets(accelerator.IpSets)); err != nil {
		return fmt.Errorf("error setting ip_sets: %w", err)
	}

	acceleratorAttributes, err := finder.AcceleratorAttributesByARN(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading Global Accelerator Accelerator (%s) attributes: %w", d.Id(), err)
	}

	if err := d.Set("attributes", []interface{}{flattenGlobalAcceleratorAcceleratorAttributes(acceleratorAttributes)}); err != nil {
		return fmt.Errorf("error setting attributes: %w", err)
	}

	tags, err := keyvaluetags.GlobalacceleratorListTags(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error listing tags for Global Accelerator Accelerator (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsGlobalAcceleratorAcceleratorUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn

	if d.HasChanges("name", "ip_address_type", "enabled") {
		input := &globalaccelerator.UpdateAcceleratorInput{
			AcceleratorArn: aws.String(d.Id()),
			Name:           aws.String(d.Get("name").(string)),
			Enabled:        aws.Bool(d.Get("enabled").(bool)),
		}

		if v, ok := d.GetOk("ip_address_type"); ok {
			input.IpAddressType = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating Global Accelerator Accelerator: %s", input)
		if _, err := conn.UpdateAccelerator(input); err != nil {
			return fmt.Errorf("error updating Global Accelerator Accelerator (%s): %w", d.Id(), err)
		}

		if _, err := waiter.AcceleratorDeployed(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for Global Accelerator Accelerator (%s) deployment: %w", d.Id(), err)
		}
	}

	if d.HasChange("attributes") {
		o, n := d.GetChange("attributes")
		if len(o.([]interface{})) > 0 && o.([]interface{})[0] != nil {
			if len(n.([]interface{})) > 0 && n.([]interface{})[0] != nil {
				oInput := expandGlobalAcceleratorUpdateAcceleratorAttributesInput(o.([]interface{})[0].(map[string]interface{}))
				oInput.AcceleratorArn = aws.String(d.Id())
				nInput := expandGlobalAcceleratorUpdateAcceleratorAttributesInput(n.([]interface{})[0].(map[string]interface{}))
				nInput.AcceleratorArn = aws.String(d.Id())

				// To change flow logs bucket and prefix attributes while flows are enabled, first disable flow logs.
				if aws.BoolValue(oInput.FlowLogsEnabled) && aws.BoolValue(nInput.FlowLogsEnabled) {
					oInput.FlowLogsEnabled = aws.Bool(false)

					log.Printf("[DEBUG] Updating Global Accelerator Accelerator attributes: %s", oInput)
					if _, err := conn.UpdateAcceleratorAttributes(oInput); err != nil {
						return fmt.Errorf("error updating Global Accelerator Accelerator (%s) attributes: %w", d.Id(), err)
					}

					if _, err := waiter.AcceleratorDeployed(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
						return fmt.Errorf("error waiting for Global Accelerator Accelerator (%s) deployment: %w", d.Id(), err)
					}
				}

				log.Printf("[DEBUG] Updating Global Accelerator Accelerator attributes: %s", nInput)
				if _, err := conn.UpdateAcceleratorAttributes(nInput); err != nil {
					return fmt.Errorf("error updating Global Accelerator Accelerator (%s) attributes: %w", d.Id(), err)
				}

				if _, err := waiter.AcceleratorDeployed(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return fmt.Errorf("error waiting for Global Accelerator Accelerator (%s) deployment: %w", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.GlobalacceleratorUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Global Accelerator Accelerator (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsGlobalAcceleratorAcceleratorRead(d, meta)
}

func resourceAwsGlobalAcceleratorAcceleratorDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn

	{
		input := &globalaccelerator.UpdateAcceleratorInput{
			AcceleratorArn: aws.String(d.Id()),
			Enabled:        aws.Bool(false),
		}

		log.Printf("[DEBUG] Updating Global Accelerator Accelerator: %s", input)
		_, err := conn.UpdateAccelerator(input)

		if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error disabling Global Accelerator Accelerator (%s): %w", d.Id(), err)
		}

		if _, err := waiter.AcceleratorDeployed(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for Global Accelerator Accelerator (%s) deployment: %w", d.Id(), err)
		}
	}

	{
		input := &globalaccelerator.DeleteAcceleratorInput{
			AcceleratorArn: aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Deleting Global Accelerator Accelerator (%s)", d.Id())
		_, err := conn.DeleteAccelerator(input)

		if tfawserr.ErrCodeEquals(err, globalaccelerator.ErrCodeAcceleratorNotFoundException) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error deleting Global Accelerator Accelerator (%s): %w", d.Id(), err)
		}
	}

	return nil
}

func expandGlobalAcceleratorUpdateAcceleratorAttributesInput(tfMap map[string]interface{}) *globalaccelerator.UpdateAcceleratorAttributesInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &globalaccelerator.UpdateAcceleratorAttributesInput{}

	if v, ok := tfMap["flow_logs_enabled"].(bool); ok {
		apiObject.FlowLogsEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["flow_logs_s3_bucket"].(string); ok && v != "" {
		apiObject.FlowLogsS3Bucket = aws.String(v)
	}

	if v, ok := tfMap["flow_logs_s3_prefix"].(string); ok && v != "" {
		apiObject.FlowLogsS3Prefix = aws.String(v)
	}

	return apiObject
}

func flattenGlobalAcceleratorIpSet(apiObject *globalaccelerator.IpSet) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.IpAddresses; v != nil {
		tfMap["ip_addresses"] = aws.StringValueSlice(v)
	}

	if v := apiObject.IpFamily; v != nil {
		tfMap["ip_family"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenGlobalAcceleratorIpSets(apiObjects []*globalaccelerator.IpSet) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenGlobalAcceleratorIpSet(apiObject))
	}

	return tfList
}

func flattenGlobalAcceleratorAcceleratorAttributes(apiObject *globalaccelerator.AcceleratorAttributes) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.FlowLogsEnabled; v != nil {
		tfMap["flow_logs_enabled"] = aws.BoolValue(v)
	}

	if v := apiObject.FlowLogsS3Bucket; v != nil {
		tfMap["flow_logs_s3_bucket"] = aws.StringValue(v)
	}

	if v := apiObject.FlowLogsS3Prefix; v != nil {
		tfMap["flow_logs_s3_prefix"] = aws.StringValue(v)
	}

	return tfMap
}
