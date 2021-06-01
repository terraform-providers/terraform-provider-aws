package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/globalaccelerator/finder"
)

func resourceAwsGlobalAcceleratorEndpointGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlobalAcceleratorEndpointGroupCreate,
		Read:   resourceAwsGlobalAcceleratorEndpointGroupRead,
		Update: resourceAwsGlobalAcceleratorEndpointGroupUpdate,
		Delete: resourceAwsGlobalAcceleratorEndpointGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"endpoint_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_ip_preservation_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},

						"endpoint_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},

						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 255),
						},
					},
				},
			},

			"endpoint_group_region": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"health_check_interval_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30,
				ValidateFunc: validation.IntBetween(10, 30),
			},

			"health_check_path": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"health_check_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsPortNumber,
			},

			"health_check_protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      globalaccelerator.HealthCheckProtocolTcp,
				ValidateFunc: validation.StringInSlice(globalaccelerator.HealthCheckProtocol_Values(), false),
			},

			"listener_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"port_override": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},

						"listener_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},
					},
				},
			},

			"threshold_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3,
				ValidateFunc: validation.IntBetween(1, 10),
			},

			"traffic_dial_percentage": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Default:      100.0,
				ValidateFunc: validation.FloatBetween(0.0, 100.0),
			},
		},
	}
}

func resourceAwsGlobalAcceleratorEndpointGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn
	region := meta.(*AWSClient).region

	opts := &globalaccelerator.CreateEndpointGroupInput{
		EndpointGroupRegion: aws.String(region),
		IdempotencyToken:    aws.String(resource.UniqueId()),
		ListenerArn:         aws.String(d.Get("listener_arn").(string)),
	}

	if v, ok := d.GetOk("endpoint_configuration"); ok {
		opts.EndpointConfigurations = expandGlobalAcceleratorEndpointConfigurations(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("endpoint_group_region"); ok {
		opts.EndpointGroupRegion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("health_check_interval_seconds"); ok {
		opts.HealthCheckIntervalSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("health_check_path"); ok {
		opts.HealthCheckPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("health_check_port"); ok {
		opts.HealthCheckPort = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("health_check_protocol"); ok {
		opts.HealthCheckProtocol = aws.String(v.(string))
	}

	if v, ok := d.GetOk("port_override"); ok {
		opts.PortOverrides = expandGlobalAcceleratorPortOverrides(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("threshold_count"); ok {
		opts.ThresholdCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.Get("traffic_dial_percentage").(float64); ok {
		opts.TrafficDialPercentage = aws.Float64(v)
	}

	log.Printf("[DEBUG] Create Global Accelerator endpoint group: %s", opts)

	resp, err := conn.CreateEndpointGroup(opts)
	if err != nil {
		return fmt.Errorf("error creating Global Accelerator endpoint group: %w", err)
	}

	d.SetId(aws.StringValue(resp.EndpointGroup.EndpointGroupArn))

	acceleratorArn, err := resourceAwsGlobalAcceleratorListenerParseAcceleratorArn(d.Id())

	if err != nil {
		return err
	}

	err = resourceAwsGlobalAcceleratorAcceleratorWaitForDeployedState(conn, acceleratorArn, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return err
	}

	return resourceAwsGlobalAcceleratorEndpointGroupRead(d, meta)
}

func resourceAwsGlobalAcceleratorEndpointGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn

	endpointGroup, err := finder.EndpointGroupByARN(conn, d.Id())

	if isAWSErr(err, globalaccelerator.ErrCodeEndpointGroupNotFoundException, "") {
		log.Printf("[WARN] Global Accelerator endpoint group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Global Accelerator endpoint group (%s): %w", d.Id(), err)
	}

	if endpointGroup == nil {
		log.Printf("[WARN] Global Accelerator endpoint group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	listenerArn, err := resourceAwsGlobalAcceleratorEndpointGroupParseListenerArn(d.Id())

	if err != nil {
		return err
	}

	d.Set("arn", endpointGroup.EndpointGroupArn)
	if err := d.Set("endpoint_configuration", flattenGlobalAcceleratorEndpointDescriptions(endpointGroup.EndpointDescriptions)); err != nil {
		return fmt.Errorf("error setting endpoint_configuration: %w", err)
	}
	d.Set("endpoint_group_region", endpointGroup.EndpointGroupRegion)
	d.Set("health_check_interval_seconds", endpointGroup.HealthCheckIntervalSeconds)
	d.Set("health_check_path", endpointGroup.HealthCheckPath)
	d.Set("health_check_port", endpointGroup.HealthCheckPort)
	d.Set("health_check_protocol", endpointGroup.HealthCheckProtocol)
	d.Set("listener_arn", listenerArn)
	if err := d.Set("port_override", flattenGlobalAcceleratorPortOverrides(endpointGroup.PortOverrides)); err != nil {
		return fmt.Errorf("error setting port_override: %w", err)
	}
	d.Set("threshold_count", endpointGroup.ThresholdCount)
	d.Set("traffic_dial_percentage", endpointGroup.TrafficDialPercentage)

	return nil
}

func resourceAwsGlobalAcceleratorEndpointGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn

	opts := &globalaccelerator.UpdateEndpointGroupInput{
		EndpointGroupArn: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("endpoint_configuration"); ok {
		opts.EndpointConfigurations = expandGlobalAcceleratorEndpointConfigurations(v.(*schema.Set).List())
	} else {
		opts.EndpointConfigurations = []*globalaccelerator.EndpointConfiguration{}
	}

	if v, ok := d.GetOk("health_check_interval_seconds"); ok {
		opts.HealthCheckIntervalSeconds = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("health_check_path"); ok {
		opts.HealthCheckPath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("health_check_port"); ok {
		opts.HealthCheckPort = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("health_check_protocol"); ok {
		opts.HealthCheckProtocol = aws.String(v.(string))
	}

	if v, ok := d.GetOk("port_override"); ok {
		opts.PortOverrides = expandGlobalAcceleratorPortOverrides(v.(*schema.Set).List())
	} else {
		opts.PortOverrides = []*globalaccelerator.PortOverride{}
	}

	if v, ok := d.GetOk("threshold_count"); ok {
		opts.ThresholdCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.Get("traffic_dial_percentage").(float64); ok {
		opts.TrafficDialPercentage = aws.Float64(v)
	}

	log.Printf("[DEBUG] Update Global Accelerator endpoint group: %s", opts)

	_, err := conn.UpdateEndpointGroup(opts)

	if err != nil {
		return fmt.Errorf("error updating Global Accelerator endpoint group (%s): %w", d.Id(), err)
	}

	acceleratorArn, err := resourceAwsGlobalAcceleratorListenerParseAcceleratorArn(d.Id())

	if err != nil {
		return err
	}

	err = resourceAwsGlobalAcceleratorAcceleratorWaitForDeployedState(conn, acceleratorArn, d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		return err
	}

	return resourceAwsGlobalAcceleratorEndpointGroupRead(d, meta)
}

func resourceAwsGlobalAcceleratorEndpointGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn

	opts := &globalaccelerator.DeleteEndpointGroupInput{
		EndpointGroupArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteEndpointGroup(opts)

	if isAWSErr(err, globalaccelerator.ErrCodeEndpointGroupNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Global Accelerator endpoint group (%s): %w", d.Id(), err)
	}

	acceleratorArn, err := resourceAwsGlobalAcceleratorListenerParseAcceleratorArn(d.Id())

	if err != nil {
		return err
	}

	err = resourceAwsGlobalAcceleratorAcceleratorWaitForDeployedState(conn, acceleratorArn, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return err
	}

	return nil
}

func resourceAwsGlobalAcceleratorEndpointGroupParseListenerArn(endpointGroupArn string) (string, error) {
	parts := strings.Split(endpointGroupArn, "/")
	if len(parts) < 6 {
		return "", fmt.Errorf("Unable to parse listener ARN from %s", endpointGroupArn)
	}
	return strings.Join(parts[0:4], "/"), nil
}

func expandGlobalAcceleratorEndpointConfigurations(configurations []interface{}) []*globalaccelerator.EndpointConfiguration {
	out := make([]*globalaccelerator.EndpointConfiguration, len(configurations))

	for i, raw := range configurations {
		configuration := raw.(map[string]interface{})
		m := globalaccelerator.EndpointConfiguration{}

		m.EndpointId = aws.String(configuration["endpoint_id"].(string))
		m.Weight = aws.Int64(int64(configuration["weight"].(int)))
		m.ClientIPPreservationEnabled = aws.Bool(configuration["client_ip_preservation_enabled"].(bool))

		out[i] = &m
	}

	return out
}

func expandGlobalAcceleratorPortOverrides(vPortOverrides []interface{}) []*globalaccelerator.PortOverride {
	portOverrides := []*globalaccelerator.PortOverride{}

	for _, vPortOverride := range vPortOverrides {
		portOverride := &globalaccelerator.PortOverride{}

		mPortOverride := vPortOverride.(map[string]interface{})

		if vEndpointPort, ok := mPortOverride["endpoint_port"].(int); ok && vEndpointPort > 0 {
			portOverride.EndpointPort = aws.Int64(int64(vEndpointPort))
		}
		if vListenerPort, ok := mPortOverride["listener_port"].(int); ok && vListenerPort > 0 {
			portOverride.ListenerPort = aws.Int64(int64(vListenerPort))
		}

		portOverrides = append(portOverrides, portOverride)
	}

	return portOverrides
}

func flattenGlobalAcceleratorEndpointDescriptions(configurations []*globalaccelerator.EndpointDescription) []interface{} {
	out := make([]interface{}, len(configurations))

	for i, configuration := range configurations {
		m := make(map[string]interface{})

		m["endpoint_id"] = aws.StringValue(configuration.EndpointId)
		m["weight"] = aws.Int64Value(configuration.Weight)
		m["client_ip_preservation_enabled"] = aws.BoolValue(configuration.ClientIPPreservationEnabled)

		out[i] = m
	}

	return out
}

func flattenGlobalAcceleratorPortOverrides(portOverrides []*globalaccelerator.PortOverride) []interface{} {
	if len(portOverrides) == 0 || portOverrides[0] == nil {
		return []interface{}{}
	}

	vPortOverrides := []interface{}{}

	for _, portOverride := range portOverrides {
		mPortOverride := map[string]interface{}{
			"endpoint_port": int(aws.Int64Value(portOverride.EndpointPort)),
			"listener_port": int(aws.Int64Value(portOverride.ListenerPort)),
		}

		vPortOverrides = append(vPortOverrides, mPortOverride)
	}

	return vPortOverrides
}
