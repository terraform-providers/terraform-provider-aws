package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsDxConnectionConfirmation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxConnectionConfirmationCreate,
		Read:   resourceAwsDxConnectionConfirmationRead,
		Delete: resourceAwsDxConnectionConfirmationDelete,

		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsDxConnectionConfirmationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	id := d.Get("connection_id").(string)

	log.Printf("[DEBUG] Confirming Direct Connect connection: %s", id)
	_, err := conn.ConfirmConnection(&directconnect.ConfirmConnectionInput{
		ConnectionId: aws.String(id),
	})
	if err != nil {
		return err
	}

	availableStateConf := &resource.StateChangeConf{
		Pending:    []string{directconnect.ConnectionStatePending, directconnect.ConnectionStateOrdering, directconnect.ConnectionStateRequested},
		Target:     []string{directconnect.ConnectionStateAvailable},
		Refresh:    dxConnectionRefreshStateFunc(dxConnectionDescribe(conn, id)),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = availableStateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Direct Connect connection (%s) to be available: %s", id, err)
	}

	d.SetId(id)

	return nil
}

func resourceAwsDxConnectionConfirmationRead(d *schema.ResourceData, meta interface{}) error {
	dxconn := meta.(*AWSClient).dxconn

	conn, err := dxConnectionDescribe(dxconn, d.Id())()
	if err != nil {
		return err
	}
	if conn == nil {
		log.Printf("[WARN] Direct Connect connection (%s) not found, removing confirmation from state", d.Id())
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsDxConnectionConfirmationDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Will not delete Direct Connect connection. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}
