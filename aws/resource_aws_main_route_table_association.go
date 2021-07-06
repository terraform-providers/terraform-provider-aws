package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsMainRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMainRouteTableAssociationCreate,
		Read:   resourceAwsMainRouteTableAssociationRead,
		Update: resourceAwsMainRouteTableAssociationUpdate,
		Delete: resourceAwsMainRouteTableAssociationDelete,

		Schema: map[string]*schema.Schema{
			// We use this field to record the main route table that is automatically
			// created when the VPC is created. We need this to be able to "destroy"
			// our main route table association, which we do by returning this route
			// table to its original place as the Main Route Table for the VPC.
			"original_route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsMainRouteTableAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcID := d.Get("vpc_id").(string)

	association, err := finder.MainRouteTableAssociationByVpcID(conn, vpcID)

	if err != nil {
		return fmt.Errorf("error reading Main Route Table Association (%s): %w", vpcID, err)
	}

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.ReplaceRouteTableAssociationInput{
		AssociationId: association.RouteTableAssociationId,
		RouteTableId:  aws.String(routeTableID),
	}

	log.Printf("[DEBUG] Creating Main Route Table Association: %s", input)
	output, err := conn.ReplaceRouteTableAssociation(input)

	if err != nil {
		return fmt.Errorf("error creating Main Route Table Association (%s): %w", routeTableID, err)
	}

	d.SetId(aws.StringValue(output.NewAssociationId))

	log.Printf("[DEBUG] Waiting for Main Route Table Association (%s) creation", d.Id())
	if _, err := waiter.RouteTableAssociationUpdated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Main Route Table Association (%s) create: %w", d.Id(), err)
	}

	d.Set("original_route_table_id", association.RouteTableId)

	return resourceAwsMainRouteTableAssociationRead(d, meta)
}

func resourceAwsMainRouteTableAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	_, err := finder.MainRouteTableAssociationByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Main Route Table Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Main Route Table Association (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsMainRouteTableAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.ReplaceRouteTableAssociationInput{
		AssociationId: aws.String(d.Id()),
		RouteTableId:  aws.String(routeTableID),
	}

	log.Printf("[DEBUG] Updating Main Route Table Association: %s", input)
	output, err := conn.ReplaceRouteTableAssociation(input)

	if err != nil {
		return fmt.Errorf("error updating Main Route Table Association (%s): %w", routeTableID, err)
	}

	// This whole thing with the resource ID being changed on update seems unsustainable.
	// Keeping it here for backwards compatibility...
	d.SetId(aws.StringValue(output.NewAssociationId))

	log.Printf("[DEBUG] Waiting for Main Route Table Association (%s) update", d.Id())
	if _, err := waiter.RouteTableAssociationUpdated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Main Route Table Association (%s) update: %w", d.Id(), err)
	}

	return resourceAwsMainRouteTableAssociationRead(d, meta)
}

func resourceAwsMainRouteTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.ReplaceRouteTableAssociationInput{
		AssociationId: aws.String(d.Id()),
		RouteTableId:  aws.String(d.Get("original_route_table_id").(string)),
	}

	log.Printf("[DEBUG] Deleting Main Route Table Association: %s", input)
	output, err := conn.ReplaceRouteTableAssociation(input)

	if err != nil {
		return fmt.Errorf("error deleting Main Route Table Association (%s): %w", d.Get("route_table_id").(string), err)
	}

	log.Printf("[DEBUG] Waiting for Main Route Table Association (%s) deletion", d.Id())
	if _, err := waiter.RouteTableAssociationUpdated(conn, aws.StringValue(output.NewAssociationId)); err != nil {
		return fmt.Errorf("error waiting for Main Route Table Association (%s) delete: %w", d.Id(), err)
	}

	return nil
}
