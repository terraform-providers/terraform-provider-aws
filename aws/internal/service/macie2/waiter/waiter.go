package waiter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for the MemberRelationshipStatus to be Invited, Enabled, or Paused
	MemberInvitedTimeout = 5 * time.Minute
)

// MemberInvited waits for an AdminAccount to return Invited, Enabled and Paused
func MemberInvited(ctx context.Context, conn *macie2.Macie2, adminAccountID string) (*macie2.Member, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{macie2.RelationshipStatusCreated, macie2.RelationshipStatusEmailVerificationInProgress},
		Target:  []string{macie2.RelationshipStatusInvited, macie2.RelationshipStatusEnabled, macie2.RelationshipStatusPaused},
		Refresh: MemberRelationshipStatus(conn, adminAccountID),
		Timeout: MemberInvitedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*macie2.Member); ok {
		return output, err
	}

	return nil, err
}
