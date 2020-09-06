package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsDxGatewayAssociationProposal() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxGatewayAssociationProposalCreate,
		Read:   resourceAwsDxGatewayAssociationProposalRead,
		Delete: resourceAwsDxGatewayAssociationProposalDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.Sequence(
			// Accepting the proposal with overridden prefixes changes the returned RequestedAllowedPrefixesToDirectConnectGateway value (allowed_prefixes attribute).
			// We only want to force a new resource if this value changes and the current proposal state is "requested".
			customdiff.ForceNewIf("allowed_prefixes", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				conn := meta.(*AWSClient).dxconn

				proposal, err := describeDirectConnectGatewayAssociationProposal(conn, d.Id())
				if err != nil {
					log.Printf("[ERROR] Error reading Direct Connect Gateway Association Proposal (%s): %s", d.Id(), err)
					return false
				}

				return proposal != nil && aws.StringValue(proposal.ProposalState) == directconnect.GatewayAssociationProposalStateRequested
			}),
		),

		Schema: map[string]*schema.Schema{
			"allowed_prefixes": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"associated_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"associated_gateway_owner_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"associated_gateway_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dx_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"dx_gateway_owner_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
		},
	}
}

func resourceAwsDxGatewayAssociationProposalCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	allowedPrefixes := expandDirectConnectGatewayAssociationProposalAllowedPrefixes(d.Get("allowed_prefixes").(*schema.Set).List())
	input := &directconnect.CreateDirectConnectGatewayAssociationProposalInput{
		AddAllowedPrefixesToDirectConnectGateway: allowedPrefixes,
		DirectConnectGatewayId:                   aws.String(d.Get("dx_gateway_id").(string)),
		DirectConnectGatewayOwnerAccount:         aws.String(d.Get("dx_gateway_owner_account_id").(string)),
		GatewayId:                                aws.String(d.Get("associated_gateway_id").(string)),
	}

	log.Printf("[DEBUG] Creating Direct Connect Gateway Association Proposal: %s", input)
	output, err := conn.CreateDirectConnectGatewayAssociationProposal(input)

	if err != nil {
		return fmt.Errorf("error creating Direct Connect Gateway Association Proposal: %s", err)
	}

	d.SetId(aws.StringValue(output.DirectConnectGatewayAssociationProposal.ProposalId))

	return resourceAwsDxGatewayAssociationProposalRead(d, meta)
}

func resourceAwsDxGatewayAssociationProposalRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	proposal, err := describeDirectConnectGatewayAssociationProposal(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading Direct Connect Gateway Association Proposal (%s): %s", d.Id(), err)
	}

	if proposal == nil {
		log.Printf("[WARN] Direct Connect Gateway Association Proposal (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(proposal.ProposalState) == directconnect.GatewayAssociationProposalStateDeleted {
		log.Printf("[WARN] Direct Connect Gateway Association Proposal (%s) deleted, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if proposal.AssociatedGateway == nil {
		return fmt.Errorf("error reading Direct Connect Gateway Association Proposal (%s): missing associated gateway information", d.Id())
	}

	if err := d.Set("allowed_prefixes", flattenDirectConnectGatewayAssociationProposalAllowedPrefixes(proposal.RequestedAllowedPrefixesToDirectConnectGateway)); err != nil {
		return fmt.Errorf("error setting allowed_prefixes: %s", err)
	}

	d.Set("associated_gateway_id", aws.StringValue(proposal.AssociatedGateway.Id))
	d.Set("associated_gateway_owner_account_id", proposal.AssociatedGateway.OwnerAccount)
	d.Set("associated_gateway_type", proposal.AssociatedGateway.Type)
	d.Set("dx_gateway_id", aws.StringValue(proposal.DirectConnectGatewayId))
	d.Set("dx_gateway_owner_account_id", aws.StringValue(proposal.DirectConnectGatewayOwnerAccount))

	return nil
}

func resourceAwsDxGatewayAssociationProposalDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	input := &directconnect.DeleteDirectConnectGatewayAssociationProposalInput{
		ProposalId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Direct Connect Gateway Association Proposal: %s", d.Id())

	_, err := conn.DeleteDirectConnectGatewayAssociationProposal(input)

	if err != nil {
		return fmt.Errorf("error deleting Direct Connect Gateway Association Proposal (%s): %s", d.Id(), err)
	}

	return nil
}

func describeDirectConnectGatewayAssociationProposal(conn *directconnect.DirectConnect, proposalID string) (*directconnect.GatewayAssociationProposal, error) {
	input := &directconnect.DescribeDirectConnectGatewayAssociationProposalsInput{
		ProposalId: aws.String(proposalID),
	}

	for {
		output, err := conn.DescribeDirectConnectGatewayAssociationProposals(input)

		if err != nil {
			return nil, err
		}

		if output == nil {
			continue
		}

		for _, proposal := range output.DirectConnectGatewayAssociationProposals {
			if aws.StringValue(proposal.ProposalId) == proposalID {
				return proposal, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func expandDirectConnectGatewayAssociationProposalAllowedPrefixes(allowedPrefixes []interface{}) []*directconnect.RouteFilterPrefix {
	if len(allowedPrefixes) == 0 {
		return nil
	}

	var routeFilterPrefixes []*directconnect.RouteFilterPrefix

	for _, allowedPrefixRaw := range allowedPrefixes {
		if allowedPrefixRaw == nil {
			continue
		}

		routeFilterPrefix := &directconnect.RouteFilterPrefix{
			Cidr: aws.String(allowedPrefixRaw.(string)),
		}

		routeFilterPrefixes = append(routeFilterPrefixes, routeFilterPrefix)
	}

	return routeFilterPrefixes
}

func flattenDirectConnectGatewayAssociationProposalAllowedPrefixes(routeFilterPrefixes []*directconnect.RouteFilterPrefix) []interface{} {
	if len(routeFilterPrefixes) == 0 {
		return []interface{}{}
	}

	var allowedPrefixes []interface{}

	for _, routeFilterPrefix := range routeFilterPrefixes {
		if routeFilterPrefix == nil {
			continue
		}

		allowedPrefix := aws.StringValue(routeFilterPrefix.Cidr)

		allowedPrefixes = append(allowedPrefixes, allowedPrefix)
	}

	return allowedPrefixes
}
