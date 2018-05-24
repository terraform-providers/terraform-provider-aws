package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsNetworkAclAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNetworkAclAssociationCreate,
		Read:   resourceAwsNetworkAclAssociationRead,
		Update: resourceAwsNetworkAclAssociationUpdate,
		Delete: resourceAwsNetworkAclAssociationDelete,

		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"network_acl_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsNetworkAclAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	naclId := d.Get("network_acl_id").(string)
	subnetId := d.Get("subnet_id").(string)

	association, errAssociation := findNetworkAclAssociation(subnetId, conn)
	if errAssociation != nil {
		return fmt.Errorf("Failed to find association for subnet %s: %s", subnetId, errAssociation)
	}

	associationOpts := ec2.ReplaceNetworkAclAssociationInput{
		AssociationId: association.NetworkAclAssociationId,
		NetworkAclId:  aws.String(naclId),
	}

	log.Printf("[DEBUG] Creating Network ACL association: %#v", associationOpts)

	var err error
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err = conn.ReplaceNetworkAclAssociation(&associationOpts)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr != nil {
					return resource.RetryableError(awsErr)
				}
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return resourceAwsNetworkAclAssociationRead(d, meta)
}

func resourceAwsNetworkAclAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	// Inspect that the association exists
	subnetId := d.Get("subnet_id").(string)
	_, errAssociation := findNetworkAclAssociation(subnetId, conn)
	if errAssociation != nil {
		log.Printf("[WARN] Association for subnet %s was not found, removing from state", subnetId)
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsNetworkAclAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	naclId := d.Get("network_acl_id").(string)
	subnetId := d.Get("subnet_id").(string)

	association, errAssociation := findNetworkAclAssociation(subnetId, conn)
	if errAssociation != nil {
		return fmt.Errorf("Failed to find association for subnet %s: %s", subnetId, errAssociation)
	}

	associationOpts := ec2.ReplaceNetworkAclAssociationInput{
		AssociationId: association.NetworkAclAssociationId,
		NetworkAclId:  aws.String(naclId),
	}

	_, err := conn.ReplaceNetworkAclAssociation(&associationOpts)

	log.Printf("[DEBUG] Updating Network ACL association: %#v", associationOpts)

	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if ok && ec2err.Code() == "InvalidAssociationID.NotFound" {
			// Not found, so just create a new one
			return resourceAwsNetworkAclAssociationCreate(d, meta)
		}

		return err
	}

	return resourceAwsNetworkAclAssociationRead(d, meta)
}

func resourceAwsNetworkAclAssociationDelete(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*AWSClient).ec2conn

	subnetId := d.Get("subnet_id").(string)

	association, errAssociation := findNetworkAclAssociation(subnetId, conn)
	if errAssociation != nil {
		return fmt.Errorf("Failed to find association for subnet %s: %s", subnetId, errAssociation)
	}

	defaultAcl, err := getDefaultNetworkAcl(d.Get("vpc_id").(string), conn)

	if err != nil {
		return fmt.Errorf("Failed to get networkAcl : %s", err)
	}

	associationOpts := ec2.ReplaceNetworkAclAssociationInput{
		AssociationId: association.NetworkAclAssociationId,
		NetworkAclId:  defaultAcl.NetworkAclId,
	}

	log.Printf("[DEBUG] Replacing Network ACL association: %#v", associationOpts)

	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err = conn.ReplaceNetworkAclAssociation(&associationOpts)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr != nil {
					return resource.RetryableError(awsErr)
				}
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
