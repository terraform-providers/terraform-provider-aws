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

func dxConnectionDelete(d *schema.ResourceData, conn *directconnect.DirectConnect, describeFunc func() (*directconnect.Connection, error)) error {
	log.Printf("[DEBUG] Deleting Direct Connect connection: %s", d.Id())
	_, err := conn.DeleteConnection(&directconnect.DeleteConnectionInput{
		ConnectionId: aws.String(d.Id()),
	})
	if err != nil {
		if isNoSuchDxConnectionErr(err) {
			return nil
		}
		return err
	}

	deleteStateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.ConnectionStatePending, directconnect.ConnectionStateOrdering, directconnect.ConnectionStateAvailable, directconnect.ConnectionStateRequested, directconnect.ConnectionStateDeleting},
		// ConnectionStateRejected occurs when you delete a connection before it has been confirmed
		Target:     []string{directconnect.ConnectionStateDeleted, directconnect.ConnectionStateRejected},
		Refresh:    dxConnectionRefreshStateFunc(describeFunc),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = deleteStateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Direct Connect connection (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func dxConnectionRefreshStateFunc(describeFunc func() (*directconnect.Connection, error)) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		conn, err := describeFunc()
		if err != nil {
			return nil, "failed", err
		}
		if conn == nil {
			// Terraform treats the absence of the resource as an error and does not even check the state
			// so we create a dummy one here
			conn = &directconnect.Connection{
				ConnectionState: aws.String(directconnect.ConnectionStateDeleted),
			}
		}
		return conn, *conn.ConnectionState, nil
	}
}

func isNoSuchDxConnectionErr(err error) bool {
	return isAWSErr(err, "DirectConnectClientException", "Could not find Connection with ID")
}
