package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRouteTableAssociationCreate,
		Read:   resourceAwsRouteTableAssociationRead,
		Update: resourceAwsRouteTableAssociationUpdate,
		Delete: resourceAwsRouteTableAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsRouteTableAssociationImport,
		},

		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"subnet_id", "gateway_id"},
				ForceNew:     true,
			},
			"gateway_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"subnet_id", "gateway_id"},
				ForceNew:     true,
			},

			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsRouteTableAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	associationOpts := ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(d.Get("route_table_id").(string)),
	}
	if len(d.Get("subnet_id").(string)) > 0 {
		log.Printf(
			"[INFO] Creating route table association: %s => %s",
			d.Get("subnet_id").(string),
			d.Get("route_table_id").(string))
		associationOpts.SubnetId = aws.String(d.Get("subnet_id").(string))
	} else if len(d.Get("gateway_id").(string)) > 0 {
		log.Printf(
			"[INFO] Creating route table association: %s => %s",
			d.Get("gateway_id").(string),
			d.Get("route_table_id").(string))
		associationOpts.GatewayId = aws.String(d.Get("gateway_id").(string))
	}

	var associationID string
	var resp *ec2.AssociateRouteTableOutput
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.AssociateRouteTable(&associationOpts)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "InvalidRouteTableID.NotFound" {
					return resource.RetryableError(awsErr)
				}
			}
			return resource.NonRetryableError(err)
		}
		associationID = *resp.AssociationId
		return nil
	})
	if isResourceTimeoutError(err) {
		resp, err = conn.AssociateRouteTable(&associationOpts)
	}
	if err != nil {
		return fmt.Errorf("Error creating route table association: %s", err)
	}

	// Set the ID and return
	d.SetId(associationID)
	log.Printf("[INFO] Association ID: %s", d.Id())

	return nil
}

func resourceAwsRouteTableAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	// Get the routing table that this association belongs to
	rtRaw, _, err := resourceAwsRouteTableStateRefreshFunc(
		conn, d.Get("route_table_id").(string))()
	if err != nil {
		return err
	}
	if rtRaw == nil {
		return nil
	}
	rt := rtRaw.(*ec2.RouteTable)

	// Inspect that the association exists
	found := false
	for _, a := range rt.Associations {
		if *a.RouteTableAssociationId == d.Id() {
			found = true
			if a.SubnetId != nil {
				d.Set("subnet_id", a.SubnetId)
			}
			if a.GatewayId != nil {
				d.Set("gateway_id", a.GatewayId)
			}
			break
		}
	}

	if !found {
		// It seems it doesn't exist anymore, so clear the ID
		d.SetId("")
	}

	return nil
}

func resourceAwsRouteTableAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf(
		"[INFO] Creating route table association: %s => %s",
		d.Get("subnet_id").(string),
		d.Get("route_table_id").(string))

	req := &ec2.ReplaceRouteTableAssociationInput{
		AssociationId: aws.String(d.Id()),
		RouteTableId:  aws.String(d.Get("route_table_id").(string)),
	}
	resp, err := conn.ReplaceRouteTableAssociation(req)

	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if ok && ec2err.Code() == "InvalidAssociationID.NotFound" {
			// Not found, so just create a new one
			return resourceAwsRouteTableAssociationCreate(d, meta)
		}

		return err
	}

	// Update the ID
	d.SetId(*resp.NewAssociationId)
	log.Printf("[INFO] Association ID: %s", d.Id())

	return nil
}

func resourceAwsRouteTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[INFO] Deleting route table association: %s", d.Id())
	_, err := conn.DisassociateRouteTable(&ec2.DisassociateRouteTableInput{
		AssociationId: aws.String(d.Id()),
	})
	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if ok && ec2err.Code() == "InvalidAssociationID.NotFound" {
			return nil
		}

		return fmt.Errorf("Error deleting route table association: %s", err)
	}

	return nil
}

func resourceAwsRouteTableAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Unexpected format for import: %s. Use 'subnet ID/route table ID' or 'gateway ID/route table ID", d.Id())
	}

	targetID := parts[0]
	routeTableID := parts[1]

	log.Printf("[DEBUG] Importing route table association, target: %s, route table: %s", targetID, routeTableID)

	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DescribeRouteTablesInput{}
	input.Filters = buildEC2AttributeFilterList(
		map[string]string{
			"association.route-table-id": routeTableID,
		},
	)

	output, err := conn.DescribeRouteTables(input)
	if err != nil {
		return nil, fmt.Errorf("Error finding route table: %s", err)
	}
	if len(output.RouteTables) == 0 {
		return nil, fmt.Errorf("No route table found with ID %s", routeTableID)
	}

	rt := output.RouteTables[0]

	var targetType string
	var associationID string
	for _, a := range rt.Associations {
		if aws.StringValue(a.SubnetId) == targetID {
			targetType = "subnet"
			associationID = aws.StringValue(a.RouteTableAssociationId)
			break
		}
		if aws.StringValue(a.SubnetId) == targetID || aws.StringValue(a.GatewayId) == targetID {
			targetType = "gateway"
			associationID = aws.StringValue(a.RouteTableAssociationId)
			break
		}
	}
	if associationID == "" {
		return nil, fmt.Errorf("No association found between route table ID %s and target ID %s", routeTableID, targetID)
	}

	d.SetId(associationID)
	if targetType == "subnet" {
		d.Set("subnet_id", targetID)
	} else if targetType == "gateway" {
		d.Set("gateway_id", targetID)
	}
	d.Set("route_table_id", routeTableID)

	return []*schema.ResourceData{d}, nil
}
