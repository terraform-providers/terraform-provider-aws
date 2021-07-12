package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsVpcEndpointSecurityGroupAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpcEndpointSecurityGroupAssociationCreate,
		Read:   resourceAwsVpcEndpointSecurityGroupAssociationRead,
		Delete: resourceAwsVpcEndpointSecurityGroupAssociationDelete,

		Schema: map[string]*schema.Schema{
			"replace_default_association": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"security_group_id": {
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

func resourceAwsVpcEndpointSecurityGroupAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcEndpointID := d.Get("vpc_endpoint_id").(string)
	securityGroupID := d.Get("security_group_id").(string)
	replaceDefaultAssociation := d.Get("replace_default_association").(bool)

	defaultSecurityGroupID := ""
	if replaceDefaultAssociation {
		vpcEndpoint, err := finder.VpcEndpointByID(conn, vpcEndpointID)

		if err != nil {
			return fmt.Errorf("error reading VPC endpoint (%s): %w", vpcEndpointID, err)
		}

		vpcID := aws.StringValue(vpcEndpoint.VpcId)

		defaultSecurityGroup, err := finder.DefaultSecurityGroup(conn, vpcID)

		if err != nil {
			return fmt.Errorf("error reading default security group for VPC (%s): %w", vpcID, err)
		}

		defaultSecurityGroupID = aws.StringValue(defaultSecurityGroup.GroupId)

		if defaultSecurityGroupID == securityGroupID {
			return fmt.Errorf("%s is the default security group for VPC (%s)", securityGroupID, vpcID)
		}

		foundDefaultAssociation := false

		for _, group := range vpcEndpoint.Groups {
			if aws.StringValue(group.GroupId) == defaultSecurityGroupID {
				foundDefaultAssociation = true
				break
			}
		}

		if !foundDefaultAssociation {
			return fmt.Errorf("no association of default security group (%s) with VPC endpoint (%s)", defaultSecurityGroupID, vpcEndpointID)
		}
	}

	err := createVpcEndpointSecurityGroupAssociation(conn, vpcEndpointID, securityGroupID)

	if err != nil {
		return err
	}

	d.SetId(tfec2.VpcEndpointSecurityGroupAssociationCreateID(vpcEndpointID, securityGroupID))

	if replaceDefaultAssociation {
		// Delete the existing VPC endpoint/default security group association.
		err := deleteVpcEndpointSecurityGroupAssociation(conn, vpcEndpointID, defaultSecurityGroupID)

		if err != nil {
			return err
		}
	}

	return resourceAwsVpcEndpointSecurityGroupAssociationRead(d, meta)
}

func resourceAwsVpcEndpointSecurityGroupAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcEndpointID := d.Get("vpc_endpoint_id").(string)
	securityGroupID := d.Get("security_group_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", vpcEndpointID, securityGroupID)

	err := finder.VpcEndpointSecurityGroupAssociationExists(conn, vpcEndpointID, securityGroupID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint Security Group Association (%s) not found, removing from state", id)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading VPC Security Group Association (%s): %w", id, err)
	}

	return nil
}

func resourceAwsVpcEndpointSecurityGroupAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcEndpointID := d.Get("vpc_endpoint_id").(string)
	securityGroupID := d.Get("security_group_id").(string)
	replaceDefaultAssociation := d.Get("replace_default_association").(bool)

	if replaceDefaultAssociation {
		vpcEndpoint, err := finder.VpcEndpointByID(conn, vpcEndpointID)

		if err != nil {
			return fmt.Errorf("error reading VPC endpoint (%s): %w", vpcEndpointID, err)
		}

		vpcID := aws.StringValue(vpcEndpoint.VpcId)

		defaultSecurityGroup, err := finder.DefaultSecurityGroup(conn, vpcID)

		if err != nil {
			return fmt.Errorf("error reading default security group for VPC (%s): %w", vpcID, err)
		}

		// Add back the VPC endpoint/default security group association.
		err = createVpcEndpointSecurityGroupAssociation(conn, vpcEndpointID, aws.StringValue(defaultSecurityGroup.GroupId))

		if err != nil {
			return err
		}
	}

	return deleteVpcEndpointSecurityGroupAssociation(conn, vpcEndpointID, securityGroupID)
}

// createVpcEndpointSecurityGroupAssociation creates the specified VPC endpoint/security group association.
func createVpcEndpointSecurityGroupAssociation(conn *ec2.EC2, vpcEndpointID, securityGroupID string) error {
	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:       aws.String(vpcEndpointID),
		AddSecurityGroupIds: aws.StringSlice([]string{securityGroupID}),
	}

	log.Printf("[DEBUG] Creating VPC Endpoint Security Group Association: %s", input)

	_, err := conn.ModifyVpcEndpoint(input)

	if err != nil {
		return fmt.Errorf("error creating VPC Endpoint Security Group Association (%s/%s): %w", vpcEndpointID, securityGroupID, err)
	}

	return nil
}

// deleteVpcEndpointSecurityGroupAssociation deletes the specified VPC endpoint/security group association.
func deleteVpcEndpointSecurityGroupAssociation(conn *ec2.EC2, vpcEndpointID, securityGroupID string) error {
	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:          aws.String(vpcEndpointID),
		RemoveSecurityGroupIds: aws.StringSlice([]string{securityGroupID}),
	}

	log.Printf("[DEBUG] Deleting VPC Endpoint Security Group Association: %s", input)

	_, err := conn.ModifyVpcEndpoint(input)

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidVpcEndpointIdNotFound) || tfawserr.ErrCodeEquals(err, tfec2.InvalidGroupNotFound) || tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidParameter) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting VPC Endpoint Security Group Association (%s/%s): %w", vpcEndpointID, securityGroupID, err)
	}

	return nil
}
