package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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

		Schema: map[string]*schema.Schema{
			"listener_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"endpoint_group_region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"health_check_interval_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30,
				ValidateFunc: validation.IntBetween(10, 30),
			},
			"health_check_path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
			},
			"health_check_port": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"health_check_protocol": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  globalaccelerator.HealthCheckProtocolTcp,
				ValidateFunc: validation.StringInSlice([]string{
					globalaccelerator.HealthCheckProtocolTcp,
					globalaccelerator.HealthCheckProtocolHttp,
					globalaccelerator.HealthCheckProtocolHttps,
				}, false),
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
			"endpoint_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"weight": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsGlobalAcceleratorEndpointGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn
	region := meta.(*AWSClient).region

	opts := &globalaccelerator.CreateEndpointGroupInput{
		ListenerArn:         aws.String(d.Get("listener_arn").(string)),
		IdempotencyToken:    aws.String(resource.UniqueId()),
		EndpointGroupRegion: aws.String(region),
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

	if v, ok := d.GetOk("threshold_count"); ok {
		opts.ThresholdCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.Get("traffic_dial_percentage").(float64); ok {
		opts.TrafficDialPercentage = aws.Float64(v)
	}

	if v, ok := d.GetOk("endpoint_configuration"); ok {
		opts.EndpointConfigurations = resourceAwsGlobalAcceleratorEndpointGroupExpandEndpointConfigurations(v.(*schema.Set).List())
	}

	log.Printf("[DEBUG] Create Global Accelerator endpoint group: %s", opts)

	resp, err := conn.CreateEndpointGroup(opts)
	if err != nil {
		return fmt.Errorf("Error creating Global Accelerator endpoint group: %s", err)
	}

	d.SetId(*resp.EndpointGroup.EndpointGroupArn)

	acceleratorArn, err := resourceAwsGlobalAcceleratorListenerParseAcceleratorArn(d.Id())

	if err != nil {
		return err
	}

	err = resourceAwsGlobalAcceleratorAcceleratorWaitForDeployedState(conn, acceleratorArn)

	if err != nil {
		return err
	}

	return resourceAwsGlobalAcceleratorEndpointGroupRead(d, meta)
}

func resourceAwsGlobalAcceleratorEndpointGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn

	endpointGroup, err := resourceAwsGlobalAcceleratorEndpointGroupRetrieve(conn, d.Id())

	if err != nil {
		return fmt.Errorf("Error reading Global Accelerator endpoint group: %s", err)
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

	d.Set("listener_arn", listenerArn)
	d.Set("endpoint_group_region", endpointGroup.EndpointGroupRegion)
	d.Set("health_check_interval_seconds", endpointGroup.HealthCheckIntervalSeconds)
	d.Set("health_check_path", endpointGroup.HealthCheckPath)
	d.Set("health_check_port", endpointGroup.HealthCheckPort)
	d.Set("health_check_protocol", endpointGroup.HealthCheckProtocol)
	d.Set("threshold_count", endpointGroup.ThresholdCount)
	d.Set("traffic_dial_percentage", endpointGroup.TrafficDialPercentage)
	if err := d.Set("endpoint_configuration", resourceAwsGlobalAcceleratorEndpointGroupFlattenEndpointDescriptions(endpointGroup.EndpointDescriptions)); err != nil {
		return fmt.Errorf("error setting endpoint_configuration: %s", err)
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

func resourceAwsGlobalAcceleratorEndpointGroupExpandEndpointConfigurations(configurations []interface{}) []*globalaccelerator.EndpointConfiguration {
	out := make([]*globalaccelerator.EndpointConfiguration, len(configurations))

	for i, raw := range configurations {
		configuration := raw.(map[string]interface{})
		m := globalaccelerator.EndpointConfiguration{}

		m.EndpointId = aws.String(configuration["endpoint_id"].(string))
		m.Weight = aws.Int64(int64(configuration["weight"].(int)))

		out[i] = &m
	}

	log.Printf("[DEBUG] Expand endpoint_configuration: %s", out)
	return out
}

func resourceAwsGlobalAcceleratorEndpointGroupFlattenEndpointDescriptions(configurations []*globalaccelerator.EndpointDescription) []interface{} {
	out := make([]interface{}, len(configurations))

	for i, configuration := range configurations {
		m := make(map[string]interface{})

		m["endpoint_id"] = aws.StringValue(configuration.EndpointId)
		m["weight"] = aws.Int64Value(configuration.Weight)

		out[i] = m
	}

	log.Printf("[DEBUG] Flatten endpoint_configuration: %s", out)
	return out
}

func resourceAwsGlobalAcceleratorEndpointGroupRetrieve(conn *globalaccelerator.GlobalAccelerator, endpointGroupArn string) (*globalaccelerator.EndpointGroup, error) {
	resp, err := conn.DescribeEndpointGroup(&globalaccelerator.DescribeEndpointGroupInput{
		EndpointGroupArn: aws.String(endpointGroupArn),
	})

	if err != nil {
		if isAWSErr(err, globalaccelerator.ErrCodeEndpointGroupNotFoundException, "") {
			return nil, nil
		}
		return nil, err
	}

	return resp.EndpointGroup, nil
}

func resourceAwsGlobalAcceleratorEndpointGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn

	opts := &globalaccelerator.UpdateEndpointGroupInput{
		EndpointGroupArn: aws.String(d.Id()),
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

	if v, ok := d.GetOk("threshold_count"); ok {
		opts.ThresholdCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.Get("traffic_dial_percentage").(float64); ok {
		opts.TrafficDialPercentage = aws.Float64(v)
	}

	if v, ok := d.GetOk("endpoint_configuration"); ok {
		opts.EndpointConfigurations = resourceAwsGlobalAcceleratorEndpointGroupExpandEndpointConfigurations(v.(*schema.Set).List())
	} else {
		opts.EndpointConfigurations = []*globalaccelerator.EndpointConfiguration{}
	}

	log.Printf("[DEBUG] Update Global Accelerator endpoint group: %s", opts)

	_, err := conn.UpdateEndpointGroup(opts)
	if err != nil {
		return fmt.Errorf("Error updating Global Accelerator endpoint group: %s", err)
	}

	acceleratorArn, err := resourceAwsGlobalAcceleratorListenerParseAcceleratorArn(d.Id())

	if err != nil {
		return err
	}

	err = resourceAwsGlobalAcceleratorAcceleratorWaitForDeployedState(conn, acceleratorArn)

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
	if err != nil {
		if isAWSErr(err, globalaccelerator.ErrCodeEndpointGroupNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting Global Accelerator endpoint group: %s", err)
	}

	acceleratorArn, err := resourceAwsGlobalAcceleratorListenerParseAcceleratorArn(d.Id())

	if err != nil {
		return err
	}

	err = resourceAwsGlobalAcceleratorAcceleratorWaitForDeployedState(conn, acceleratorArn)

	if err != nil {
		return err
	}

	return nil
}
