package aws

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsVpcEndpointConnectionAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpcEndpointConnectionAccepterCreate,
		Read:   resourceAwsVpcEndpointConnectionAccepterRead,
		Delete: resourceAwsVpcEndpointConnectionAccepterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				svcID, vpceID := parseVPCEndpointConnectionAccepterID(d.Id())

				d.Set("service_id", svcID)
				d.Set("endpoint_id", vpceID)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"service_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsVpcEndpointConnectionAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	svcID := d.Get("service_id").(string)
	vpceID := d.Get("endpoint_id").(string)

	input := &ec2.AcceptVpcEndpointConnectionsInput{
		ServiceId:      aws.String(svcID),
		VpcEndpointIds: aws.StringSlice([]string{vpceID}),
	}

	log.Printf("[DEBUG] Accepting VPC Endpoint Connection: %#v", input)
	_, err := conn.AcceptVpcEndpointConnections(input)
	if err != nil {
		return fmt.Errorf("error accepting VPC Endpoint Connection: %s", err.Error())
	}

	d.SetId(makeVPCEndpointConnectionAccepterID(svcID, vpceID))

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pendingAcceptance", "pending"},
		Target:     []string{"available"},
		Refresh:    vpcEndpointConnectionRefresh(conn, svcID, vpceID),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	vpceConn, err := stateConf.WaitForStateContext(context.Background())
	if err != nil {
		return fmt.Errorf("error waiting for VPC Endpoint (%s) to be accepted by VPC Endpoint Service (%s): %s", vpceID, svcID, err)
	}

	d.Set("state", aws.StringValue(vpceConn.(ec2.VpcEndpointConnection).VpcEndpointState))

	return resourceAwsVpcEndpointConnectionAccepterRead(d, meta)
}

func vpcEndpointConnectionRefresh(conn *ec2.EC2, svcID, vpceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[DEBUG] Reading VPC Endpoint Connections for VPC Endpoint Service (%s)", svcID)

		input := &ec2.DescribeVpcEndpointConnectionsInput{
			Filters: buildEC2AttributeFilterList(map[string]string{"service-id": svcID}),
		}

		var vpceConn *ec2.VpcEndpointConnection

		paginator := func(page *ec2.DescribeVpcEndpointConnectionsOutput, lastPage bool) bool {
			for _, c := range page.VpcEndpointConnections {
				if aws.StringValue(c.VpcEndpointId) == vpceID {
					log.Printf("[DEBUG] Found VPC Endpoint Connection from VPC Endpoint Service (%s) to VPC Endpoint (%s): %s", svcID, vpceID, *c)
					vpceConn = c
					return false
				}
			}
			return !lastPage
		}

		if err := conn.DescribeVpcEndpointConnectionsPages(input, paginator); err != nil {
			return nil, "", err
		}

		if vpceConn == nil {
			return nil, "", fmt.Errorf("VPC Endpoint Connection from VPC Endpoint Service (%s) to VPC Endpoint (%s) not found", svcID, vpceID)
		}

		state := aws.StringValue(vpceConn.VpcEndpointState)
		log.Printf("[DEBUG] state %s", state)

		// No point in retrying if the endpoint is in a failed state.
		if state == "failed" {
			return nil, state, fmt.Errorf("VPC Endpoint Connection state %q", state)
		}

		return *vpceConn, state, nil
	}
}

func resourceAwsVpcEndpointConnectionAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	svcID, vpceID := parseVPCEndpointConnectionAccepterID(d.Id())

	vpceConn, state, err := vpcEndpointConnectionRefresh(conn, svcID, vpceID)()
	if err != nil && state != "failed" {
		return fmt.Errorf("error reading VPC Endpoint Connection from VPC Endpoint Service (%s) to VPC Endpoint (%s): %s", svcID, vpceID, err)
	}

	d.Set("state", aws.StringValue(vpceConn.(ec2.VpcEndpointConnection).VpcEndpointState))

	return nil
}

func resourceAwsVpcEndpointConnectionAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	svcID, vpceID := parseVPCEndpointConnectionAccepterID(d.Id())

	input := &ec2.RejectVpcEndpointConnectionsInput{
		ServiceId:      aws.String(svcID),
		VpcEndpointIds: aws.StringSlice([]string{vpceID}),
	}

	if _, err := conn.RejectVpcEndpointConnections(input); err != nil {
		return fmt.Errorf("error rejecting VPC Endpoint Connection from VPC Endpoint (%s) to VPC Endpoint Service (%s): %s", vpceID, svcID, err)
	}

	return nil
}

const vpcEndpointConnectionAccepterIDSeparator = "_"

func makeVPCEndpointConnectionAccepterID(svcID, vpceID string) string {
	return strings.Join([]string{svcID, vpceID}, vpcEndpointConnectionAccepterIDSeparator)
}

func parseVPCEndpointConnectionAccepterID(vpceConnID string) (svcID string, vpceID string) {
	split := strings.Split(vpceConnID, vpcEndpointConnectionAccepterIDSeparator)

	svcID = split[0]
	vpceID = split[1]

	return
}
