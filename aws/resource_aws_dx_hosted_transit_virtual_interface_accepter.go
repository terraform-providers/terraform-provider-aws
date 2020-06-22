package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDxHostedTransitVirtualInterfaceAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxHostedTransitVirtualInterfaceAccepterCreate,
		Read:   resourceAwsDxHostedTransitVirtualInterfaceAccepterRead,
		Update: resourceAwsDxHostedTransitVirtualInterfaceAccepterUpdate,
		Delete: resourceAwsDxHostedTransitVirtualInterfaceAccepterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsDxHostedTransitVirtualInterfaceAccepterImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dx_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags": tagsSchema(),
			"virtual_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsDxHostedTransitVirtualInterfaceAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	vifId := d.Get("virtual_interface_id").(string)
	req := &directconnect.ConfirmTransitVirtualInterfaceInput{
		DirectConnectGatewayId: aws.String(d.Get("dx_gateway_id").(string)),
		VirtualInterfaceId:     aws.String(vifId),
	}

	log.Printf("[DEBUG] Accepting Direct Connect hosted transit virtual interface: %s", req)
	_, err := conn.ConfirmTransitVirtualInterface(req)
	if err != nil {
		return fmt.Errorf("error accepting Direct Connect hosted transit virtual interface (%s): %s", vifId, err)
	}

	d.SetId(vifId)
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "directconnect",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("dxvif/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	if err := dxHostedTransitVirtualInterfaceAccepterWaitUntilAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return err
	}

	return resourceAwsDxHostedTransitVirtualInterfaceAccepterUpdate(d, meta)
}

func resourceAwsDxHostedTransitVirtualInterfaceAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	vif, err := dxVirtualInterfaceRead(d.Id(), conn)
	if err != nil {
		return err
	}
	if vif == nil {
		log.Printf("[WARN] Direct Connect transit virtual interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	vifState := aws.StringValue(vif.VirtualInterfaceState)
	if vifState != directconnect.VirtualInterfaceStateAvailable && vifState != directconnect.VirtualInterfaceStateDown {
		log.Printf("[WARN] Direct Connect virtual interface (%s) is '%s', removing from state", vifState, d.Id())
		d.SetId("")
		return nil
	}

	d.Set("dx_gateway_id", vif.DirectConnectGatewayId)
	d.Set("virtual_interface_id", vif.VirtualInterfaceId)

	arn := d.Get("arn").(string)
	tags, err := keyvaluetags.DirectconnectListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Direct Connect hosted transit virtual interface (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDxHostedTransitVirtualInterfaceAccepterUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := dxVirtualInterfaceUpdate(d, meta); err != nil {
		return err
	}

	return resourceAwsDxHostedTransitVirtualInterfaceAccepterRead(d, meta)
}

func resourceAwsDxHostedTransitVirtualInterfaceAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Will not delete Direct Connect virtual interface. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}

func resourceAwsDxHostedTransitVirtualInterfaceAccepterImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*AWSClient).dxconn

	vif, err := dxVirtualInterfaceRead(d.Id(), conn)
	if err != nil {
		return nil, err
	}
	if vif == nil {
		return nil, fmt.Errorf("virtual interface (%s) not found", d.Id())
	}

	if vifType := aws.StringValue(vif.VirtualInterfaceType); vifType != "transit" {
		return nil, fmt.Errorf("virtual interface (%s) has incorrect type: %s", d.Id(), vifType)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "directconnect",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("dxvif/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return []*schema.ResourceData{d}, nil
}

func dxHostedTransitVirtualInterfaceAccepterWaitUntilAvailable(conn *directconnect.DirectConnect, vifId string, timeout time.Duration) error {
	return dxVirtualInterfaceWaitUntilAvailable(
		conn,
		vifId,
		timeout,
		[]string{
			directconnect.VirtualInterfaceStateConfirming,
			directconnect.VirtualInterfaceStatePending,
		},
		[]string{
			directconnect.VirtualInterfaceStateAvailable,
			directconnect.VirtualInterfaceStateDown,
		})
}
