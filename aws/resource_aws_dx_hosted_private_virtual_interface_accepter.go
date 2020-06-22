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

func resourceAwsDxHostedPrivateVirtualInterfaceAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxHostedPrivateVirtualInterfaceAccepterCreate,
		Read:   resourceAwsDxHostedPrivateVirtualInterfaceAccepterRead,
		Update: resourceAwsDxHostedPrivateVirtualInterfaceAccepterUpdate,
		Delete: resourceAwsDxHostedPrivateVirtualInterfaceAccepterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsDxHostedPrivateVirtualInterfaceAccepterImport,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dx_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"vpn_gateway_id"},
			},
			"tags": tagsSchema(),
			"virtual_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpn_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"dx_gateway_id"},
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsDxHostedPrivateVirtualInterfaceAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	vgwIdRaw, vgwOk := d.GetOk("vpn_gateway_id")
	dxgwIdRaw, dxgwOk := d.GetOk("dx_gateway_id")
	if vgwOk == dxgwOk {
		return fmt.Errorf(
			"One of ['vpn_gateway_id', 'dx_gateway_id'] must be set to create a Direct Connect private virtual interface accepter")
	}

	vifId := d.Get("virtual_interface_id").(string)
	req := &directconnect.ConfirmPrivateVirtualInterfaceInput{
		VirtualInterfaceId: aws.String(vifId),
	}
	if dxgwOk && dxgwIdRaw.(string) != "" {
		req.DirectConnectGatewayId = aws.String(dxgwIdRaw.(string))
	}
	if vgwOk && vgwIdRaw.(string) != "" {
		req.VirtualGatewayId = aws.String(vgwIdRaw.(string))
	}

	log.Printf("[DEBUG] Accepting Direct Connect hosted private virtual interface: %s", req)
	_, err := conn.ConfirmPrivateVirtualInterface(req)
	if err != nil {
		return fmt.Errorf("error accepting Direct Connect hosted private virtual interface: %s", err)
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

	if err := dxHostedPrivateVirtualInterfaceAccepterWaitUntilAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return err
	}

	return resourceAwsDxHostedPrivateVirtualInterfaceAccepterUpdate(d, meta)
}

func resourceAwsDxHostedPrivateVirtualInterfaceAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	vif, err := dxVirtualInterfaceRead(d.Id(), conn)
	if err != nil {
		return err
	}
	if vif == nil {
		log.Printf("[WARN] Direct Connect hosted private virtual interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	vifState := aws.StringValue(vif.VirtualInterfaceState)
	if vifState != directconnect.VirtualInterfaceStateAvailable &&
		vifState != directconnect.VirtualInterfaceStateDown {
		log.Printf("[WARN] Direct Connect hosted private virtual interface (%s) is '%s', removing from state", vifState, d.Id())
		d.SetId("")
		return nil
	}

	d.Set("dx_gateway_id", vif.DirectConnectGatewayId)
	d.Set("virtual_interface_id", vif.VirtualInterfaceId)
	d.Set("vpn_gateway_id", vif.VirtualGatewayId)

	arn := d.Get("arn").(string)
	tags, err := keyvaluetags.DirectconnectListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Direct Connect hosted private virtual interface (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDxHostedPrivateVirtualInterfaceAccepterUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := dxVirtualInterfaceUpdate(d, meta); err != nil {
		return err
	}

	return resourceAwsDxHostedPrivateVirtualInterfaceAccepterRead(d, meta)
}

func resourceAwsDxHostedPrivateVirtualInterfaceAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Will not delete Direct Connect virtual interface. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}

func resourceAwsDxHostedPrivateVirtualInterfaceAccepterImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*AWSClient).dxconn

	vif, err := dxVirtualInterfaceRead(d.Id(), conn)
	if err != nil {
		return nil, err
	}
	if vif == nil {
		return nil, fmt.Errorf("virtual interface (%s) not found", d.Id())
	}

	if vifType := aws.StringValue(vif.VirtualInterfaceType); vifType != "private" {
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

func dxHostedPrivateVirtualInterfaceAccepterWaitUntilAvailable(conn *directconnect.DirectConnect, vifId string, timeout time.Duration) error {
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
