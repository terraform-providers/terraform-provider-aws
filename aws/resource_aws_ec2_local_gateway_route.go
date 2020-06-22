package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	ec2LocalGatewayRouteEventualConsistencyTimeout = 1 * time.Minute
)

func resourceAwsEc2LocalGatewayRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2LocalGatewayRouteCreate,
		Read:   resourceAwsEc2LocalGatewayRouteRead,
		Delete: resourceAwsEc2LocalGatewayRouteDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"destination_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCIDRNetworkAddress,
			},
			"local_gateway_route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"local_gateway_virtual_interface_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsEc2LocalGatewayRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	destination := d.Get("destination_cidr_block").(string)
	localGatewayRouteTableID := d.Get("local_gateway_route_table_id").(string)

	input := &ec2.CreateLocalGatewayRouteInput{
		DestinationCidrBlock:                aws.String(destination),
		LocalGatewayRouteTableId:            aws.String(localGatewayRouteTableID),
		LocalGatewayVirtualInterfaceGroupId: aws.String(d.Get("local_gateway_virtual_interface_group_id").(string)),
	}

	_, err := conn.CreateLocalGatewayRoute(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Local Gateway Route: %s", err)
	}

	d.SetId(fmt.Sprintf("%s_%s", localGatewayRouteTableID, destination))

	return resourceAwsEc2LocalGatewayRouteRead(d, meta)
}

func resourceAwsEc2LocalGatewayRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	localGatewayRouteTableID, destination, err := decodeEc2LocalGatewayRouteID(d.Id())
	if err != nil {
		return err
	}

	var localGatewayRoute *ec2.LocalGatewayRoute
	err = resource.Retry(ec2LocalGatewayRouteEventualConsistencyTimeout, func() *resource.RetryError {
		var err error
		localGatewayRoute, err = getEc2LocalGatewayRoute(conn, localGatewayRouteTableID, destination)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if d.IsNewResource() && localGatewayRoute == nil {
			return resource.RetryableError(&resource.NotFoundError{})
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		localGatewayRoute, err = getEc2LocalGatewayRoute(conn, localGatewayRouteTableID, destination)
	}

	if isAWSErr(err, "InvalidRouteTableID.NotFound", "") {
		log.Printf("[WARN] EC2 Local Gateway Route Table (%s) not found, removing from state", localGatewayRouteTableID)
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && isResourceNotFoundError(err) {
		log.Printf("[WARN] EC2 Local Gateway Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Local Gateway Route: %s", err)
	}

	if localGatewayRoute == nil {
		log.Printf("[WARN] EC2 Local Gateway Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	state := aws.StringValue(localGatewayRoute.State)
	if state == ec2.LocalGatewayRouteStateDeleted || state == ec2.LocalGatewayRouteStateDeleting {
		log.Printf("[WARN] EC2 Local Gateway Route (%s) deleted, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("destination_cidr_block", localGatewayRoute.DestinationCidrBlock)
	d.Set("local_gateway_virtual_interface_group_id", localGatewayRoute.LocalGatewayVirtualInterfaceGroupId)
	d.Set("local_gateway_route_table_id", localGatewayRoute.LocalGatewayRouteTableId)

	return nil
}

func resourceAwsEc2LocalGatewayRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	localGatewayRouteTableID, destination, err := decodeEc2LocalGatewayRouteID(d.Id())
	if err != nil {
		return err
	}

	input := &ec2.DeleteLocalGatewayRouteInput{
		DestinationCidrBlock:     aws.String(destination),
		LocalGatewayRouteTableId: aws.String(localGatewayRouteTableID),
	}

	log.Printf("[DEBUG] Deleting EC2 Local Gateway Route (%s): %s", d.Id(), input)
	_, err = conn.DeleteLocalGatewayRoute(input)

	if isAWSErr(err, "InvalidRoute.NotFound", "") || isAWSErr(err, "InvalidRouteTableID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Local Gateway Route: %s", err)
	}

	return nil
}

func decodeEc2LocalGatewayRouteID(id string) (string, string, error) {
	parts := strings.Split(id, "_")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected tgw-rtb-ID_DESTINATION", id)
	}

	return parts[0], parts[1], nil
}

func getEc2LocalGatewayRoute(conn *ec2.EC2, localGatewayRouteTableID, destination string) (*ec2.LocalGatewayRoute, error) {
	input := &ec2.SearchLocalGatewayRoutesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("type"),
				Values: aws.StringSlice([]string{"static"}),
			},
		},
		LocalGatewayRouteTableId: aws.String(localGatewayRouteTableID),
	}

	output, err := conn.SearchLocalGatewayRoutes(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Routes) == 0 {
		return nil, nil
	}

	for _, route := range output.Routes {
		if route == nil {
			continue
		}

		if aws.StringValue(route.DestinationCidrBlock) == destination {
			return route, nil
		}
	}

	return nil, nil
}
