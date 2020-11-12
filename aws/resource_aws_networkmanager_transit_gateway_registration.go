package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsNetworkManagerTransitGatewayRegistration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNetworkManagerTransitGatewayRegistrationCreate,
		Read:   resourceAwsNetworkManagerTransitGatewayRegistrationRead,
		Delete: resourceAwsNetworkManagerTransitGatewayRegistrationDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idErr := fmt.Errorf("Expected ID in format of arn:aws:ec2:REGION:ACCOUNTID:transit-gateway/TGWID,GLOBALNETWORKID and provided: %s", d.Id())

				identifiers := strings.Split(d.Id(), ",")
				if len(identifiers) != 2 {
					return nil, idErr
				}
				if arn.IsARN(identifiers[0]) {
					d.Set("transit_gateway_arn", identifiers[0])
				} else {
					return nil, idErr
				}

				d.Set("global_network_id", identifiers[1])

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"transit_gateway_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsNetworkManagerTransitGatewayRegistrationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	globalNetworkId := d.Get("global_network_id").(string)
	transitGatewayArn := d.Get("transit_gateway_arn").(string)

	input := &networkmanager.RegisterTransitGatewayInput{
		GlobalNetworkId:   aws.String(globalNetworkId),
		TransitGatewayArn: aws.String(transitGatewayArn),
	}

	log.Printf("[DEBUG] Creating Network Manager Transit gateway Registration: %s", input)

	output, err := conn.RegisterTransitGateway(input)
	if err != nil {
		return fmt.Errorf("error creating Network Manager Transit Gateway Registration: %s", err)
	}

	d.SetId(fmt.Sprintf("%s,%s", transitGatewayArn, globalNetworkId))

	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.TransitGatewayRegistrationStatePending},
		Target:  []string{networkmanager.TransitGatewayRegistrationStateAvailable},
		Refresh: networkmanagerTransitGatewayRegistrationRefreshFunc(conn, aws.StringValue(output.TransitGatewayRegistration.GlobalNetworkId), aws.StringValue(output.TransitGatewayRegistration.TransitGatewayArn)),
		Timeout: 10 * time.Minute,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for Network Manager Transit Gateway Registration (%s) availability: %s", d.Id(), err)
	}

	return resourceAwsNetworkManagerTransitGatewayRegistrationRead(d, meta)
}

func resourceAwsNetworkManagerTransitGatewayRegistrationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	transitGatewayRegistration, err := networkmanagerDescribeTransitGatewayRegistration(conn, d.Get("global_network_id").(string), d.Get("transit_gateway_arn").(string))

	if isAWSErr(err, "InvalidTransitGatewayRegistrationID.NotFound", "") {
		log.Printf("[WARN] Network Manager Transit Gateway Registration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Network Manager Transit Gateway Registration: %s", err)
	}

	if transitGatewayRegistration == nil {
		log.Printf("[WARN] Network Manager Transit Gateway Registration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(transitGatewayRegistration.State.Code) == networkmanager.TransitGatewayRegistrationStateDeleting || aws.StringValue(transitGatewayRegistration.State.Code) == networkmanager.TransitGatewayRegistrationStateDeleted {
		log.Printf("[WARN] Network Manager Transit Gateway Registration (%s) in deleted state (%s,%s), removing from state", d.Id(), aws.StringValue(transitGatewayRegistration.State.Code), aws.StringValue(transitGatewayRegistration.State.Message))
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsNetworkManagerTransitGatewayRegistrationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	input := &networkmanager.DeregisterTransitGatewayInput{
		GlobalNetworkId:   aws.String(d.Get("global_network_id").(string)),
		TransitGatewayArn: aws.String(d.Get("transit_gateway_arn").(string)),
	}

	log.Printf("[DEBUG] Deleting Network Manager Transit Gateway Registration (%s): %s", d.Id(), input)
	req, _ := conn.DeregisterTransitGatewayRequest(input)
	err := req.Send()

	if isAWSErr(err, "InvalidTransitGatewayRegistrationArn.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Network Manager Transit Gateway Registration: %s", err)
	}

	if err := waitForNetworkManagerTransitGatewayRegistrationDeletion(conn, d.Get("global_network_id").(string), d.Get("transit_gateway_arn").(string)); err != nil {
		return fmt.Errorf("error waiting for Network Manager Transit Gateway Registration (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func networkmanagerTransitGatewayRegistrationRefreshFunc(conn *networkmanager.NetworkManager, globalNetworkID, transitGatewayArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		transitGatewayRegistration, err := networkmanagerDescribeTransitGatewayRegistration(conn, globalNetworkID, transitGatewayArn)

		if isAWSErr(err, "InvalidTransitGatewayRegistrationID.NotFound", "") {
			return nil, "DELETED", nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading Network Manager Transit Gateway Registration (%s,%s): %s", transitGatewayArn, globalNetworkID, err)
		}

		if transitGatewayRegistration == nil {
			return nil, "DELETED", nil
		}

		return transitGatewayRegistration, aws.StringValue(transitGatewayRegistration.State.Code), nil
	}
}

func networkmanagerDescribeTransitGatewayRegistration(conn *networkmanager.NetworkManager, globalNetworkID, transitGatewayArn string) (*networkmanager.TransitGatewayRegistration, error) {
	input := &networkmanager.GetTransitGatewayRegistrationsInput{
		GlobalNetworkId:    aws.String(globalNetworkID),
		TransitGatewayArns: []*string{aws.String(transitGatewayArn)},
	}

	log.Printf("[DEBUG] Reading Network Manager Transit Gateway Registration (%s): %s", transitGatewayArn, input)
	for {
		output, err := conn.GetTransitGatewayRegistrations(input)

		if err != nil {
			return nil, err
		}

		if output == nil || len(output.TransitGatewayRegistrations) == 0 {
			return nil, nil
		}

		for _, transitGatewayRegistration := range output.TransitGatewayRegistrations {
			if transitGatewayRegistration == nil {
				continue
			}

			if aws.StringValue(transitGatewayRegistration.TransitGatewayArn) == transitGatewayArn {
				return transitGatewayRegistration, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func waitForNetworkManagerTransitGatewayRegistrationDeletion(conn *networkmanager.NetworkManager, globalNetworkID, transitGatewayArn string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			networkmanager.TransitGatewayRegistrationStateAvailable,
			networkmanager.TransitGatewayRegistrationStateDeleting,
		},
		Target: []string{
			networkmanager.TransitGatewayRegistrationStateDeleted,
		},
		Refresh:        networkmanagerTransitGatewayRegistrationRefreshFunc(conn, globalNetworkID, transitGatewayArn),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for Network Manager Transit Gateway Registration (%s) deletion", transitGatewayArn)
	_, err := stateConf.WaitForState()

	if isResourceNotFoundError(err) {
		return nil
	}

	return err
}
