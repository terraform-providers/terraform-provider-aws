package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	gatewayAssociationStateDeleted = "deleted"
)

func resourceAwsDxGatewayAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxGatewayAssociationCreate,
		Read:   resourceAwsDxGatewayAssociationRead,
		Update: resourceAwsDxGatewayAssociationUpdate,
		Delete: resourceAwsDxGatewayAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsDxGatewayAssociationImport,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceAwsDxGatewayAssociationResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceAwsDxGatewayAssociationStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"allowed_prefixes": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"associated_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"associated_gateway_owner_account_id", "proposal_id"},
			},

			"associated_gateway_owner_account_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validateAwsAccountId,
				ConflictsWith: []string{"associated_gateway_id"},
			},

			"associated_gateway_type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dx_gateway_association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dx_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"dx_gateway_owner_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"proposal_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"associated_gateway_id", "vpn_gateway_id"},
			},

			"vpn_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"associated_gateway_id", "associated_gateway_owner_account_id", "proposal_id"},
				Deprecated:    "use 'associated_gateway_id' argument instead",
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceAwsDxGatewayAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	dxgwId := d.Get("dx_gateway_id").(string)
	gwIdRaw, gwIdOk := d.GetOk("associated_gateway_id")
	gwAcctIdRaw, gwAcctIdOk := d.GetOk("associated_gateway_owner_account_id")
	proposalIdRaw, proposalIdOk := d.GetOk("proposal_id")

	if gwAcctIdOk || proposalIdOk {
		// Cross-account association.
		if !(gwAcctIdOk && proposalIdOk) {
			return fmt.Errorf("associated_gateway_owner_account_id and proposal_id must be configured")
		}
	} else if !gwIdOk {
		return fmt.Errorf("either associated_gateway_owner_account_id and proposal_id, or associated_gateway_id, must be configured")
	}

	associationId := ""
	if gwAcctIdOk {
		req := &directconnect.AcceptDirectConnectGatewayAssociationProposalInput{
			AssociatedGatewayOwnerAccount:                 aws.String(gwAcctIdRaw.(string)),
			DirectConnectGatewayId:                        aws.String(dxgwId),
			OverrideAllowedPrefixesToDirectConnectGateway: expandDxRouteFilterPrefixes(d.Get("allowed_prefixes").(*schema.Set)),
			ProposalId: aws.String(proposalIdRaw.(string)),
		}

		log.Printf("[DEBUG] Accepting Direct Connect gateway association proposal: %#v", req)
		resp, err := conn.AcceptDirectConnectGatewayAssociationProposal(req)
		if err != nil {
			return fmt.Errorf("error accepting Direct Connect gateway association proposal: %s", err)
		}

		// For historical reasons the resource ID isn't set to the association ID returned from the API.
		associationId = aws.StringValue(resp.DirectConnectGatewayAssociation.AssociationId)
		d.SetId(dxGatewayAssociationId(dxgwId, aws.StringValue(resp.DirectConnectGatewayAssociation.AssociatedGateway.Id)))
	} else {
		gwId := gwIdRaw.(string)

		req := &directconnect.CreateDirectConnectGatewayAssociationInput{
			AddAllowedPrefixesToDirectConnectGateway: expandDxRouteFilterPrefixes(d.Get("allowed_prefixes").(*schema.Set)),
			DirectConnectGatewayId:                   aws.String(dxgwId),
			GatewayId:                                aws.String(gwId),
		}

		log.Printf("[DEBUG] Creating Direct Connect gateway association: %#v", req)
		resp, err := conn.CreateDirectConnectGatewayAssociation(req)
		if err != nil {
			return fmt.Errorf("error creating Direct Connect gateway association: %s", err)
		}

		// For historical reasons the resource ID isn't set to the association ID returned from the API.
		associationId = aws.StringValue(resp.DirectConnectGatewayAssociation.AssociationId)
		d.SetId(dxGatewayAssociationId(dxgwId, gwId))
	}
	d.Set("dx_gateway_association_id", associationId)

	if err := waitForDirectConnectGatewayAssociationAvailabilityOnCreate(conn, associationId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Direct Connect gateway association (%s) to become available: %s", d.Id(), err)
	}

	return resourceAwsDxGatewayAssociationRead(d, meta)
}

func resourceAwsDxGatewayAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	associationId := d.Get("dx_gateway_association_id").(string)
	assocRaw, state, err := dxGatewayAssociationStateRefresh(conn, associationId)()
	if err != nil {
		return fmt.Errorf("error reading Direct Connect gateway association (%s): %s", d.Id(), err)
	}
	if state == gatewayAssociationStateDeleted {
		log.Printf("[WARN] Direct Connect gateway association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	assoc := assocRaw.(*directconnect.GatewayAssociation)

	err = d.Set("allowed_prefixes", flattenDxRouteFilterPrefixes(assoc.AllowedPrefixesToDirectConnectGateway))
	if err != nil {
		return fmt.Errorf("error setting allowed_prefixes: %s", err)
	}

	d.Set("associated_gateway_id", assoc.AssociatedGateway.Id)
	d.Set("associated_gateway_owner_account_id", assoc.AssociatedGateway.OwnerAccount)
	d.Set("associated_gateway_type", assoc.AssociatedGateway.Type)
	d.Set("dx_gateway_association_id", assoc.AssociationId)
	d.Set("dx_gateway_id", assoc.DirectConnectGatewayId)
	d.Set("dx_gateway_owner_account_id", assoc.DirectConnectGatewayOwnerAccount)

	return nil
}

func resourceAwsDxGatewayAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	if d.HasChange("allowed_prefixes") {
		associationId := d.Get("dx_gateway_association_id").(string)

		oraw, nraw := d.GetChange("allowed_prefixes")
		o := oraw.(*schema.Set)
		n := nraw.(*schema.Set)
		del := o.Difference(n)
		add := n.Difference(o)

		req := &directconnect.UpdateDirectConnectGatewayAssociationInput{
			AddAllowedPrefixesToDirectConnectGateway:    expandDxRouteFilterPrefixes(add),
			AssociationId:                               aws.String(associationId),
			RemoveAllowedPrefixesToDirectConnectGateway: expandDxRouteFilterPrefixes(del),
		}

		log.Printf("[DEBUG] Updating Direct Connect gateway association: %#v", req)
		_, err := conn.UpdateDirectConnectGatewayAssociation(req)
		if err != nil {
			return fmt.Errorf("error updating Direct Connect gateway association (%s): %s", d.Id(), err)
		}

		if err := waitForDirectConnectGatewayAssociationAvailabilityOnUpdate(conn, associationId, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for Direct Connect gateway association (%s) to become available: %s", d.Id(), err)
		}
	}

	return resourceAwsDxGatewayAssociationRead(d, meta)
}

func resourceAwsDxGatewayAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	associationId := d.Get("dx_gateway_association_id").(string)

	log.Printf("[DEBUG] Deleting Direct Connect gateway association: %s", d.Id())
	_, err := conn.DeleteDirectConnectGatewayAssociation(&directconnect.DeleteDirectConnectGatewayAssociationInput{
		AssociationId: aws.String(associationId),
	})
	if isAWSErr(err, directconnect.ErrCodeClientException, "No association exists") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting Direct Connect gateway association: %s", err)
	}

	if err := waitForDirectConnectGatewayAssociationDeletion(conn, associationId, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for Direct Connect gateway association (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsDxGatewayAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'dx-gw-id/gw-id'", d.Id())
	}

	dxgwId := parts[0]
	gwId := parts[1]
	id := dxGatewayAssociationId(dxgwId, gwId)
	log.Printf("[DEBUG] Importing Direct Connect gateway association %s/%s", dxgwId, gwId)

	conn := meta.(*AWSClient).dxconn

	resp, err := conn.DescribeDirectConnectGatewayAssociations(&directconnect.DescribeDirectConnectGatewayAssociationsInput{
		AssociatedGatewayId:    aws.String(gwId),
		DirectConnectGatewayId: aws.String(dxgwId),
	})
	if err != nil {
		return nil, err
	}
	if n := len(resp.DirectConnectGatewayAssociations); n != 1 {
		return nil, fmt.Errorf("Found %d Direct Connect gateway associations for %s, expected 1", n, id)
	}

	d.SetId(id)
	d.Set("dx_gateway_id", resp.DirectConnectGatewayAssociations[0].DirectConnectGatewayId)
	d.Set("dx_gateway_association_id", resp.DirectConnectGatewayAssociations[0].AssociationId)

	return []*schema.ResourceData{d}, nil
}

func dxGatewayAssociationStateRefresh(conn *directconnect.DirectConnect, associationId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeDirectConnectGatewayAssociations(&directconnect.DescribeDirectConnectGatewayAssociationsInput{
			AssociationId: aws.String(associationId),
		})
		if err != nil {
			return nil, "", err
		}

		n := len(resp.DirectConnectGatewayAssociations)
		switch n {
		case 0:
			return "", gatewayAssociationStateDeleted, nil

		case 1:
			assoc := resp.DirectConnectGatewayAssociations[0]

			if stateChangeError := aws.StringValue(assoc.StateChangeError); stateChangeError != "" {
				id := dxGatewayAssociationId(
					aws.StringValue(resp.DirectConnectGatewayAssociations[0].DirectConnectGatewayId),
					aws.StringValue(resp.DirectConnectGatewayAssociations[0].AssociatedGateway.Id))
				log.Printf("[INFO] Direct Connect gateway association (%s) state change error: %s", id, stateChangeError)
			}

			return assoc, aws.StringValue(assoc.AssociationState), nil

		default:
			return nil, "", fmt.Errorf("Found %d Direct Connect gateway associations for %s, expected 1", n, associationId)
		}
	}
}

// Terraform resource ID.
func dxGatewayAssociationId(dxgwId, gwId string) string {
	return fmt.Sprintf("ga-%s%s", dxgwId, gwId)
}

func waitForDirectConnectGatewayAssociationAvailabilityOnCreate(conn *directconnect.DirectConnect, associationId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{directconnect.GatewayAssociationStateAssociating},
		Target:     []string{directconnect.GatewayAssociationStateAssociated},
		Refresh:    dxGatewayAssociationStateRefresh(conn, associationId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitForDirectConnectGatewayAssociationAvailabilityOnUpdate(conn *directconnect.DirectConnect, associationId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{directconnect.GatewayAssociationStateUpdating},
		Target:     []string{directconnect.GatewayAssociationStateAssociated},
		Refresh:    dxGatewayAssociationStateRefresh(conn, associationId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitForDirectConnectGatewayAssociationDeletion(conn *directconnect.DirectConnect, associationId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{directconnect.GatewayAssociationStateDisassociating},
		Target:     []string{directconnect.GatewayAssociationStateDisassociated, gatewayAssociationStateDeleted},
		Refresh:    dxGatewayAssociationStateRefresh(conn, associationId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}
