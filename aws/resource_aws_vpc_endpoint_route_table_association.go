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

func resourceAwsVpcEndpointRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpcEndpointRouteTableAssociationCreate,
		Read:   resourceAwsVpcEndpointRouteTableAssociationRead,
		Delete: resourceAwsVpcEndpointRouteTableAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsVpcEndpointRouteTableAssociationImport,
		},

		Schema: map[string]*schema.Schema{
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpc_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsVpcEndpointRouteTableAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	endpointID := d.Get("vpc_endpoint_id").(string)
	routeTableID := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointID, routeTableID)

	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:    aws.String(endpointID),
		AddRouteTableIds: aws.StringSlice([]string{routeTableID}),
	}

	log.Printf("[DEBUG] Creating VPC Endpoint Route Table Association: %s", input)
	_, err := conn.ModifyVpcEndpoint(input)

	if err != nil {
		return fmt.Errorf("error creating VPC Endpoint Route Table Association (%s): %w", id, err)
	}

	d.SetId(tfec2.VpcEndpointRouteTableAssociationCreateID(endpointID, routeTableID))

	err = waiter.VpcEndpointRouteTableAssociationReady(conn, endpointID, routeTableID)

	if err != nil {
		return fmt.Errorf("error waiting for VPC Endpoint Route Table Association (%s) to become available: %w", id, err)
	}

	return resourceAwsVpcEndpointRouteTableAssociationRead(d, meta)
}

func resourceAwsVpcEndpointRouteTableAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	endpointID := d.Get("vpc_endpoint_id").(string)
	routeTableID := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointID, routeTableID)

	err := finder.VpcEndpointRouteTableAssociationExists(conn, endpointID, routeTableID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint Route Table Association (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading VPC Endpoint Route Table Association (%s): %w", id, err)
	}

	return nil
}

func resourceAwsVpcEndpointRouteTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	endpointID := d.Get("vpc_endpoint_id").(string)
	routeTableID := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointID, routeTableID)

	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:       aws.String(endpointID),
		RemoveRouteTableIds: aws.StringSlice([]string{routeTableID}),
	}

	log.Printf("[DEBUG] Deleting VPC Endpoint Route Table Association: %s", id)
	_, err := conn.ModifyVpcEndpoint(input)

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidVpcEndpointIdNotFound) || tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidRouteTableIdNotFound) || tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidParameter) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting VPC Endpoint Route Table Association (%s): %w", id, err)
	}

	err = waiter.VpcEndpointRouteTableAssociationDeleted(conn, endpointID, routeTableID)

	if err != nil {
		return fmt.Errorf("error waiting for VPC Endpoint Route Table Association (%s) to delete: %w", id, err)
	}

	return nil
}

func resourceAwsVpcEndpointRouteTableAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Wrong format of resource: %s. Please follow 'vpc-endpoint-id/route-table-id'", d.Id())
	}

	endpointID := parts[0]
	routeTableID := parts[1]
	log.Printf("[DEBUG] Importing VPC Endpoint (%s) Route Table (%s) Association", endpointID, routeTableID)

	d.SetId(tfec2.VpcEndpointRouteTableAssociationCreateID(endpointID, routeTableID))
	d.Set("vpc_endpoint_id", endpointID)
	d.Set("route_table_id", routeTableID)

	return []*schema.ResourceData{d}, nil
}
