package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func dxConnectionRead(id string, conn *directconnect.DirectConnect) (*directconnect.Connection, error) {
	resp, state, err := dxConnectionStateRefresh(conn, id)()
	if err != nil {
		return nil, fmt.Errorf("Error reading Direct Connection: %s", err)
	}
	if state == directconnect.ConnectionStateDeleted {
		return nil, nil
	}

	return resp.(*directconnect.Connection), nil
}

func dxConnectionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "directconnect",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("dxcon/%s", d.Id()),
	}.String()
	if err := setTagsDX(conn, d, arn); err != nil {
		return err
	}

	return nil
}

func dxConnectionStateRefresh(conn *directconnect.DirectConnect, dxConId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeConnections(&directconnect.DescribeConnectionsInput{
			ConnectionId: aws.String(dxConId),
		})
		if err != nil {
			return nil, "", err
		}

		n := len(resp.Connections)
		switch n {
		case 0:
			return "", directconnect.ConnectionStateDeleted, nil

		case 1:
			dxCon := resp.Connections[0]
			return dxCon, aws.StringValue(dxCon.ConnectionState), nil

		default:
			return nil, "", fmt.Errorf("Found %d Direct Connection for %s, expected 1", n, dxConId)
		}
	}
}

func dxConnectionWaitUntilAvailable(conn *directconnect.DirectConnect, dxConId string, timeout time.Duration, pending, target []string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    dxConnectionStateRefresh(conn, dxConId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for Direct Connection (%s) to become available: %s", dxConId, err)
	}

	return nil
}
