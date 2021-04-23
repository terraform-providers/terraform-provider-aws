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
)

func resourceAwsGlobalAcceleratorListener() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlobalAcceleratorListenerCreate,
		Read:   resourceAwsGlobalAcceleratorListenerRead,
		Update: resourceAwsGlobalAcceleratorListenerUpdate,
		Delete: resourceAwsGlobalAcceleratorListenerDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"accelerator_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_affinity": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  globalaccelerator.ClientAffinityNone,
				ValidateFunc: validation.StringInSlice([]string{
					globalaccelerator.ClientAffinityNone,
					globalaccelerator.ClientAffinitySourceIp,
				}, false),
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					globalaccelerator.ProtocolTcp,
					globalaccelerator.ProtocolUdp,
				}, false),
			},
			"port_range": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
						},
						"to_port": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 65535),
						},
					},
				},
			},
		},
	}
}

func resourceAwsGlobalAcceleratorListenerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn

	opts := &globalaccelerator.CreateListenerInput{
		AcceleratorArn:   aws.String(d.Get("accelerator_arn").(string)),
		ClientAffinity:   aws.String(d.Get("client_affinity").(string)),
		IdempotencyToken: aws.String(resource.UniqueId()),
		Protocol:         aws.String(d.Get("protocol").(string)),
		PortRanges:       resourceAwsGlobalAcceleratorListenerExpandPortRanges(d.Get("port_range").(*schema.Set).List()),
	}

	log.Printf("[DEBUG] Create Global Accelerator listener: %s", opts)

	resp, err := conn.CreateListener(opts)
	if err != nil {
		return fmt.Errorf("Error creating Global Accelerator listener: %s", err)
	}

	d.SetId(aws.StringValue(resp.Listener.ListenerArn))

	// Creating a listener triggers the accelerator to change status to InPending
	stateConf := &resource.StateChangeConf{
		Pending: []string{globalaccelerator.AcceleratorStatusInProgress},
		Target:  []string{globalaccelerator.AcceleratorStatusDeployed},
		Refresh: resourceAwsGlobalAcceleratorAcceleratorStateRefreshFunc(conn, d.Get("accelerator_arn").(string)),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	log.Printf("[DEBUG] Waiting for Global Accelerator listener (%s) availability", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Global Accelerator listener (%s) availability: %s", d.Id(), err)
	}

	return resourceAwsGlobalAcceleratorListenerRead(d, meta)
}

func resourceAwsGlobalAcceleratorListenerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn

	listener, err := resourceAwsGlobalAcceleratorListenerRetrieve(conn, d.Id())

	if err != nil {
		return fmt.Errorf("Error reading Global Accelerator listener: %s", err)
	}

	if listener == nil {
		log.Printf("[WARN] Global Accelerator listener (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	acceleratorArn, err := resourceAwsGlobalAcceleratorListenerParseAcceleratorArn(d.Id())

	if err != nil {
		return err
	}

	d.Set("accelerator_arn", acceleratorArn)
	d.Set("client_affinity", listener.ClientAffinity)
	d.Set("protocol", listener.Protocol)
	if err := d.Set("port_range", resourceAwsGlobalAcceleratorListenerFlattenPortRanges(listener.PortRanges)); err != nil {
		return fmt.Errorf("error setting port_range: %s", err)
	}

	return nil
}

func resourceAwsGlobalAcceleratorListenerParseAcceleratorArn(listenerArn string) (string, error) {
	parts := strings.Split(listenerArn, "/")
	if len(parts) < 4 {
		return "", fmt.Errorf("Unable to parse accelerator ARN from %s", listenerArn)
	}
	return strings.Join(parts[0:2], "/"), nil
}

func resourceAwsGlobalAcceleratorListenerExpandPortRanges(portRanges []interface{}) []*globalaccelerator.PortRange {
	out := make([]*globalaccelerator.PortRange, len(portRanges))

	for i, raw := range portRanges {
		portRange := raw.(map[string]interface{})
		m := globalaccelerator.PortRange{}

		m.FromPort = aws.Int64(int64(portRange["from_port"].(int)))
		m.ToPort = aws.Int64(int64(portRange["to_port"].(int)))

		out[i] = &m
	}

	return out
}

func resourceAwsGlobalAcceleratorListenerFlattenPortRanges(portRanges []*globalaccelerator.PortRange) []interface{} {
	out := make([]interface{}, len(portRanges))

	for i, portRange := range portRanges {
		m := make(map[string]interface{})

		m["from_port"] = aws.Int64Value(portRange.FromPort)
		m["to_port"] = aws.Int64Value(portRange.ToPort)

		out[i] = m
	}

	return out
}

func resourceAwsGlobalAcceleratorListenerRetrieve(conn *globalaccelerator.GlobalAccelerator, listenerArn string) (*globalaccelerator.Listener, error) {
	resp, err := conn.DescribeListener(&globalaccelerator.DescribeListenerInput{
		ListenerArn: aws.String(listenerArn),
	})

	if err != nil {
		if isAWSErr(err, globalaccelerator.ErrCodeListenerNotFoundException, "") {
			return nil, nil
		}
		return nil, err
	}

	return resp.Listener, nil
}

func resourceAwsGlobalAcceleratorListenerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn

	opts := &globalaccelerator.UpdateListenerInput{
		ClientAffinity: aws.String(d.Get("client_affinity").(string)),
		ListenerArn:    aws.String(d.Id()),
		Protocol:       aws.String(d.Get("protocol").(string)),
		PortRanges:     resourceAwsGlobalAcceleratorListenerExpandPortRanges(d.Get("port_range").(*schema.Set).List()),
	}

	log.Printf("[DEBUG] Update Global Accelerator listener: %s", opts)

	_, err := conn.UpdateListener(opts)
	if err != nil {
		return fmt.Errorf("Error updating Global Accelerator listener: %s", err)
	}

	// Creating a listener triggers the accelerator to change status to InPending
	err = resourceAwsGlobalAcceleratorAcceleratorWaitForDeployedState(conn, d.Get("accelerator_arn").(string), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return err
	}

	return resourceAwsGlobalAcceleratorListenerRead(d, meta)
}

func resourceAwsGlobalAcceleratorListenerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).globalacceleratorconn

	opts := &globalaccelerator.DeleteListenerInput{
		ListenerArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteListener(opts)
	if err != nil {
		if isAWSErr(err, globalaccelerator.ErrCodeListenerNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting Global Accelerator listener: %s", err)
	}

	// Deleting a listener triggers the accelerator to change status to InPending
	// }
	err = resourceAwsGlobalAcceleratorAcceleratorWaitForDeployedState(conn, d.Get("accelerator_arn").(string), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return err
	}

	return nil
}
