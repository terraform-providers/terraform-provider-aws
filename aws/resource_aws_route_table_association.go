package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
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
			"gateway_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"subnet_id", "gateway_id"},
			},

			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"subnet_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"subnet_id", "gateway_id"},
			},
		},
	}
}

func resourceAwsRouteTableAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(routeTableID),
	}

	if v, ok := d.GetOk("gateway_id"); ok {
		input.GatewayId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		input.SubnetId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Route Table Association: %s", input)
	output, err := tfresource.RetryWhenAwsErrCodeEquals(
		waiter.RouteTableAssociationPropagationTimeout,
		func() (interface{}, error) {
			return conn.AssociateRouteTable(input)
		},
		tfec2.ErrCodeInvalidRouteTableIDNotFound,
	)

	if err != nil {
		return fmt.Errorf("error creating Route Table (%s) Association: %w", routeTableID, err)
	}

	d.SetId(aws.StringValue(output.(*ec2.AssociateRouteTableOutput).AssociationId))

	log.Printf("[DEBUG] Waiting for Route Table Association (%s) creation", d.Id())
	if _, err := waiter.RouteTableAssociationCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Route Table Association (%s) create: %w", d.Id(), err)
	}

	return resourceAwsRouteTableAssociationRead(d, meta)
}

func resourceAwsRouteTableAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	association, err := finder.RouteTableAssociationByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route Table Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Route Table Association (%s): %w", d.Id(), err)
	}

	d.Set("gateway_id", association.GatewayId)
	d.Set("route_table_id", association.RouteTableId)
	d.Set("subnet_id", association.SubnetId)

	return nil
}

func resourceAwsRouteTableAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.ReplaceRouteTableAssociationInput{
		AssociationId: aws.String(d.Id()),
		RouteTableId:  aws.String(d.Get("route_table_id").(string)),
	}

	log.Printf("[DEBUG] Updating Route Table Association: %s", input)
	output, err := conn.ReplaceRouteTableAssociation(input)

	// This whole thing with the resource ID being changed on update seems unsustainable.
	// Keeping it here for backwards compatibility...

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidAssociationIDNotFound) {
		// Not found, so just create a new one
		return resourceAwsRouteTableAssociationCreate(d, meta)
	}

	if err != nil {
		return fmt.Errorf("error updating Route Table Association (%s): %w", d.Id(), err)
	}

	// I don't think we'll ever reach this code for a subnet/gateway route table association.
	// It would only come in to play for a VPC main route table association.

	d.SetId(aws.StringValue(output.NewAssociationId))

	log.Printf("[DEBUG] Waiting for Route Table Association (%s) update", d.Id())
	if _, err := waiter.RouteTableAssociationUpdated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Route Table Association (%s) update: %w", d.Id(), err)
	}

	return resourceAwsRouteTableAssociationRead(d, meta)
}

func resourceAwsRouteTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	return ec2RouteTableAssociationDelete(conn, d.Id())
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

	routeTable, err := finder.RouteTableByID(conn, routeTableID)

	if err != nil {
		return nil, err
	}

	var associationID string

	for _, association := range routeTable.Associations {
		if aws.StringValue(association.SubnetId) == targetID {
			d.Set("subnet_id", targetID)
			associationID = aws.StringValue(association.RouteTableAssociationId)

			break
		}

		if aws.StringValue(association.GatewayId) == targetID {
			d.Set("gateway_id", targetID)
			associationID = aws.StringValue(association.RouteTableAssociationId)

			break
		}
	}

	if associationID == "" {
		return nil, fmt.Errorf("No association found between route table ID %s and target ID %s", routeTableID, targetID)
	}

	d.SetId(associationID)
	d.Set("route_table_id", routeTableID)

	return []*schema.ResourceData{d}, nil
}

// ec2RouteTableAssociationDelete attempts to delete a route table association.
func ec2RouteTableAssociationDelete(conn *ec2.EC2, associationID string) error {
	log.Printf("[INFO] Deleting Route Table Association: %s", associationID)
	_, err := conn.DisassociateRouteTable(&ec2.DisassociateRouteTableInput{
		AssociationId: aws.String(associationID),
	})

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidAssociationIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route Table Association (%s): %w", associationID, err)
	}

	log.Printf("[DEBUG] Waiting for Route Table Association (%s) deletion", associationID)
	if _, err := waiter.RouteTableAssociationDeleted(conn, associationID); err != nil {
		return fmt.Errorf("error waiting for Route Table Association (%s) delete: %w", associationID, err)
	}

	return nil
}
