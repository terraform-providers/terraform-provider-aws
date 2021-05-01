package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
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
			"vpc_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsVpcEndpointRouteTableAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	endpointId := d.Get("vpc_endpoint_id").(string)
	rtId := d.Get("route_table_id").(string)

	_, err := findResourceVpcEndpoint(conn, endpointId)
	if err != nil {
		return err
	}

	_, err = conn.ModifyVpcEndpoint(&ec2.ModifyVpcEndpointInput{
		VpcEndpointId:    aws.String(endpointId),
		AddRouteTableIds: aws.StringSlice([]string{rtId}),
	})
	if err != nil {
		return fmt.Errorf("Error creating VPC Endpoint/Route Table association: %s", err.Error())
	}

	d.SetId(vpcEndpointIdRouteTableIdHash(endpointId, rtId))

	return resourceAwsVpcEndpointRouteTableAssociationRead(d, meta)
}

func resourceAwsVpcEndpointRouteTableAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	endpointId := d.Get("vpc_endpoint_id").(string)
	rtId := d.Get("route_table_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", endpointId, rtId)

	var routeTableID *string

	err := resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		routeTableID, err = finder.VpcEndpointRouteTableAssociation(conn, endpointId, rtId)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidVpcEndpointIdNotFound) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if d.IsNewResource() && routeTableID == nil {
			return resource.RetryableError(&resource.NotFoundError{
				LastError: fmt.Errorf("VPC Endpoint Route Table Association (%s) not found", id),
			})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		routeTableID, err = finder.VpcEndpointRouteTableAssociation(conn, endpointId, rtId)
	}

	if d.IsNewResource() && tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidVpcEndpointIdNotFound) {
		log.Printf("[WARN] VPC Endpoint Route Table Association (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading VPC Endpoint Route Table Association (%s): %w", id, err)
	}

	if routeTableID == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading VPC Endpoint Route Table Association (%s): not found after creation", id)
		}

		log.Printf("[WARN] VPC Endpoint Route Table Association (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsVpcEndpointRouteTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	endpointId := d.Get("vpc_endpoint_id").(string)
	rtId := d.Get("route_table_id").(string)

	_, err := conn.ModifyVpcEndpoint(&ec2.ModifyVpcEndpointInput{
		VpcEndpointId:       aws.String(endpointId),
		RemoveRouteTableIds: aws.StringSlice([]string{rtId}),
	})
	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return fmt.Errorf("Error deleting VPC Endpoint/Route Table association: %s", err.Error())
		}

		switch ec2err.Code() {
		case "InvalidVpcEndpointId.NotFound":
			fallthrough
		case "InvalidRouteTableId.NotFound":
			fallthrough
		case "InvalidParameter":
			log.Printf("[DEBUG] VPC Endpoint/Route Table association is already gone")
		default:
			return fmt.Errorf("Error deleting VPC Endpoint/Route Table association: %s", err.Error())
		}
	}

	return nil
}

func resourceAwsVpcEndpointRouteTableAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Wrong format of resource: %s. Please follow 'vpc-endpoint-id/route-table-id'", d.Id())
	}

	vpceId := parts[0]
	rtId := parts[1]
	log.Printf("[DEBUG] Importing VPC Endpoint (%s) Route Table (%s) association", vpceId, rtId)

	d.SetId(vpcEndpointIdRouteTableIdHash(vpceId, rtId))
	d.Set("vpc_endpoint_id", vpceId)
	d.Set("route_table_id", rtId)

	return []*schema.ResourceData{d}, nil
}

func findResourceVpcEndpoint(conn *ec2.EC2, id string) (*ec2.VpcEndpoint, error) {
	resp, err := conn.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{
		VpcEndpointIds: aws.StringSlice([]string{id}),
	})
	if err != nil {
		return nil, err
	}

	if resp.VpcEndpoints == nil || len(resp.VpcEndpoints) == 0 {
		return nil, fmt.Errorf("No VPC Endpoints were found for %s", id)
	}

	return resp.VpcEndpoints[0], nil
}

func vpcEndpointIdRouteTableIdHash(endpointId, rtId string) string {
	return fmt.Sprintf("a-%s%d", endpointId, hashcode.String(rtId))
}
