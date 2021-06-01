package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsDxGatewayAssociationResourceV0() *schema.Resource {
	return &schema.Resource{
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
				ConflictsWith: []string{"associated_gateway_owner_account_id", "proposal_id", "vpn_gateway_id"},
			},

			"associated_gateway_owner_account_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validateAwsAccountId,
				ConflictsWith: []string{"associated_gateway_id", "vpn_gateway_id"},
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
				ForceNew:      true,
				ConflictsWith: []string{"associated_gateway_id", "vpn_gateway_id"},
			},

			"vpn_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"associated_gateway_id", "associated_gateway_owner_account_id", "proposal_id"},
			},
		},
	}
}

func resourceAwsDxGatewayAssociationStateUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	conn := meta.(*AWSClient).dxconn

	log.Println("[INFO] Found Direct Connect gateway association state v0; migrating to v1")

	// dx_gateway_association_id was introduced in v2.8.0. Handle the case where it's not yet present.
	if v, ok := rawState["dx_gateway_association_id"]; !ok || v == nil {
		resp, err := conn.DescribeDirectConnectGatewayAssociations(&directconnect.DescribeDirectConnectGatewayAssociationsInput{
			DirectConnectGatewayId: aws.String(rawState["dx_gateway_id"].(string)),
			VirtualGatewayId:       aws.String(rawState["vpn_gateway_id"].(string)),
		})
		if err != nil {
			return nil, err
		}

		if len(resp.DirectConnectGatewayAssociations) == 0 {
			return nil, fmt.Errorf("Direct Connect gateway association not found, remove from state using 'terraform state rm'")
		}

		rawState["dx_gateway_association_id"] = aws.StringValue(resp.DirectConnectGatewayAssociations[0].AssociationId)
	}

	return rawState, nil
}
