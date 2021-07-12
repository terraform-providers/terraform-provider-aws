package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDxConnectionAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxConnectionAccepterCreate,
		Read:   resourceAwsDxConnectionAccepterRead,
		Update: resourceAwsDxConnectionAccepterUpdate,
		Delete: resourceAwsDxConnectionAccepterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsDxConnectionAccepterImport,
		},

		Schema: map[string]*schema.Schema{
			"dx_connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsDxConnectionAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	dxConId := d.Get("dx_connection_id").(string)
	req := &directconnect.ConfirmConnectionInput{
		ConnectionId: aws.String(dxConId),
	}

	log.Printf("[DEBUG] Accepting Direct Connection: %#v", req)
	_, err := conn.ConfirmConnection(req)
	if err != nil {
		return fmt.Errorf("Error accepting Direct Connection: %s", err.Error())
	}

	d.SetId(dxConId)
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "directconnect",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("dxcon/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	if err := dxConnectionAccepterWaitUntilAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return err
	}

	return resourceAwsDxConnectionAccepterUpdate(d, meta)
}

func resourceAwsDxConnectionAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	dxCon, err := dxConnectionRead(d.Id(), conn)
	if err != nil {
		return err
	}
	if dxCon == nil {
		log.Printf("[WARN] Direct Connect virtual interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	dxConState := aws.StringValue(dxCon.ConnectionState)
	if dxConState != directconnect.ConnectionStateAvailable &&
		dxConState != directconnect.ConnectionStateDown {
		log.Printf("[WARN] Direct Connect virtual interface (%s) is '%s', removing from state", dxConState, d.Id())
		d.SetId("")
		return nil
	}

	d.Set("dx_connection_id", dxCon.ConnectionId)
	err1 := getTagsDX(conn, d, d.Get("arn").(string))
	return err1
}

func resourceAwsDxConnectionAccepterUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := dxConnectionUpdate(d, meta); err != nil {
		return err
	}

	return resourceAwsDxConnectionAccepterRead(d, meta)
}

func resourceAwsDxConnectionAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Will not delete Direct Connect virtual interface. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}

func resourceAwsDxConnectionAccepterImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "directconnect",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("dxdxCon/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return []*schema.ResourceData{d}, nil
}

func dxConnectionAccepterWaitUntilAvailable(conn *directconnect.DirectConnect, conId string, timeout time.Duration) error {
	return dxConnectionWaitUntilAvailable(
		conn,
		conId,
		timeout,
		[]string{
			directconnect.ConnectionStateRequested,
			directconnect.ConnectionStateOrdering,
			directconnect.ConnectionStatePending,
		},
		[]string{
			directconnect.ConnectionStateAvailable,
			directconnect.ConnectionStateDown,
		})
}
