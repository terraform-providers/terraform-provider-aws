package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
)

const (
	// Missing constant in AWS Go SDK
	ec2ResourceTypeLocalGatewayRouteTableVpcAssociation = "local-gateway-route-table-vpc-association"
)

func resourceAwsEc2LocalGatewayRouteTableVpcAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2LocalGatewayRouteTableVpcAssociationCreate,
		Read:   resourceAwsEc2LocalGatewayRouteTableVpcAssociationRead,
		Update: resourceAwsEc2LocalGatewayRouteTableVpcAssociationUpdate,
		Delete: resourceAwsEc2LocalGatewayRouteTableVpcAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"local_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"local_gateway_route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags": tagsSchema(),
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsEc2LocalGatewayRouteTableVpcAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.CreateLocalGatewayRouteTableVpcAssociationInput{
		LocalGatewayRouteTableId: aws.String(d.Get("local_gateway_route_table_id").(string)),
		TagSpecifications:        ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), ec2ResourceTypeLocalGatewayRouteTableVpcAssociation),
		VpcId:                    aws.String(d.Get("vpc_id").(string)),
	}

	output, err := conn.CreateLocalGatewayRouteTableVpcAssociation(req)

	if err != nil {
		return fmt.Errorf("error creating EC2 Local Gateway Route Table VPC Association: %w", err)
	}

	d.SetId(aws.StringValue(output.LocalGatewayRouteTableVpcAssociation.LocalGatewayRouteTableVpcAssociationId))

	if _, err := waiter.LocalGatewayRouteTableVpcAssociationAssociated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Local Gateway Route Table VPC Association (%s) to associate: %w", d.Id(), err)
	}

	return resourceAwsEc2LocalGatewayRouteTableVpcAssociationRead(d, meta)
}

func resourceAwsEc2LocalGatewayRouteTableVpcAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	association, err := getEc2LocalGatewayRouteTableVpcAssociation(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading EC2 Local Gateway Route Table VPC Association (%s): %w", d.Id(), err)
	}

	if association == nil {
		log.Printf("[WARN] EC2 Local Gateway Route Table VPC Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(association.State) != ec2.RouteTableAssociationStateCodeAssociated {
		log.Printf("[WARN] EC2 Local Gateway Route Table VPC Association (%s) status (%s), removing from state", d.Id(), aws.StringValue(association.State))
		d.SetId("")
		return nil
	}

	d.Set("local_gateway_id", association.LocalGatewayId)
	d.Set("local_gateway_route_table_id", association.LocalGatewayRouteTableId)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(association.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("vpc_id", association.VpcId)

	return nil
}

func resourceAwsEc2LocalGatewayRouteTableVpcAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Local Gateway Route Table VPC Association (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsEc2LocalGatewayRouteTableVpcAssociationRead(d, meta)
}

func resourceAwsEc2LocalGatewayRouteTableVpcAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.DeleteLocalGatewayRouteTableVpcAssociationInput{
		LocalGatewayRouteTableVpcAssociationId: aws.String(d.Id()),
	}

	_, err := conn.DeleteLocalGatewayRouteTableVpcAssociation(input)

	if isAWSErr(err, "InvalidLocalGatewayRouteTableVpcAssociationID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Local Gateway Route Table VPC Association (%s): %w", d.Id(), err)
	}

	if _, err := waiter.LocalGatewayRouteTableVpcAssociationDisassociated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EC2 Local Gateway Route Table VPC Association (%s) to disassociate: %w", d.Id(), err)
	}

	return nil
}

func getEc2LocalGatewayRouteTableVpcAssociation(conn *ec2.EC2, localGatewayRouteTableVpcAssociationID string) (*ec2.LocalGatewayRouteTableVpcAssociation, error) {
	input := &ec2.DescribeLocalGatewayRouteTableVpcAssociationsInput{
		LocalGatewayRouteTableVpcAssociationIds: aws.StringSlice([]string{localGatewayRouteTableVpcAssociationID}),
	}

	output, err := conn.DescribeLocalGatewayRouteTableVpcAssociations(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, fmt.Errorf("empty response")
	}

	var association *ec2.LocalGatewayRouteTableVpcAssociation

	for _, outputAssociation := range output.LocalGatewayRouteTableVpcAssociations {
		if outputAssociation == nil {
			continue
		}

		if aws.StringValue(outputAssociation.LocalGatewayRouteTableVpcAssociationId) == localGatewayRouteTableVpcAssociationID {
			association = outputAssociation
			break
		}
	}

	return association, nil
}
