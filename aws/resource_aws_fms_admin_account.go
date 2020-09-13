package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsFmsAdminAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsFmsAdminAccountCreate,
		Read:   resourceAwsFmsAdminAccountRead,
		Delete: resourceAwsFmsAdminAccountDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
		},
	}
}

func resourceAwsFmsAdminAccountCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fmsconn

	// Ensure there is not an existing FMS Admin Account
	getAdminAccountOutput, err := conn.GetAdminAccount(&fms.GetAdminAccountInput{})

	if err != nil && !isAWSErr(err, fms.ErrCodeResourceNotFoundException, "") {
		return fmt.Errorf("error getting FMS Admin Account: %s", err)
	}

	if getAdminAccountOutput != nil && getAdminAccountOutput.AdminAccount != nil {
		return fmt.Errorf("FMS Admin Account (%s) already associated: import this Terraform resource to manage", aws.StringValue(getAdminAccountOutput.AdminAccount))
	}

	accountID := meta.(*AWSClient).accountid

	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	stateConf := &resource.StateChangeConf{
		Target:  []string{accountID},
		Refresh: associateFmsAdminAccountRefreshFunc(conn, accountID),
		Timeout: 1 * time.Minute,
		Delay:   10 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for FMS Admin Account (%s) association: %s", accountID, err)
	}

	d.SetId(accountID)

	return resourceAwsFmsAdminAccountRead(d, meta)
}

func associateFmsAdminAccountRefreshFunc(conn *fms.FMS, accountId string) resource.StateRefreshFunc {
	// This is all wrapped in a refresh func since AssociateAdminAccount returns
	// success even though it failed if called too quickly after creating an organization
	return func() (interface{}, string, error) {
		req := &fms.AssociateAdminAccountInput{
			AdminAccount: aws.String(accountId),
		}

		_, aserr := conn.AssociateAdminAccount(req)
		if aserr != nil {
			return nil, "", aserr
		}

		res, err := conn.GetAdminAccount(&fms.GetAdminAccountInput{})
		if err != nil {
			// FMS returns an AccessDeniedException if no account is associated,
			// but does not define this in its error codes
			if isAWSErr(err, "AccessDeniedException", "is not currently delegated by AWS FM") {
				return nil, "", nil
			}
			if isAWSErr(err, fms.ErrCodeResourceNotFoundException, "") {
				return nil, "", nil
			}
			return nil, "", err
		}
		return *res, *res.AdminAccount, err
	}
}

func resourceAwsFmsAdminAccountRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fmsconn

	output, err := conn.GetAdminAccount(&fms.GetAdminAccountInput{})

	if isAWSErr(err, fms.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] FMS Admin Account not found, removing from state")
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting FMS Admin Account: %s", err)
	}

	if aws.StringValue(output.RoleStatus) == fms.AccountRoleStatusDeleted {
		log.Printf("[WARN] FMS Admin Account not found, removing from state")
		d.SetId("")
		return nil
	}

	d.Set("account_id", output.AdminAccount)

	return nil
}

func resourceAwsFmsAdminAccountDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).fmsconn

	_, err := conn.DisassociateAdminAccount(&fms.DisassociateAdminAccountInput{})

	if err != nil {
		return fmt.Errorf("error disassociating FMS Admin Account: %s", err)
	}

	if err := waitForFmsAdminAccountDeletion(conn); err != nil {
		return fmt.Errorf("error waiting for FMS Admin Account (%s) disassociation: %s", d.Id(), err)
	}

	return nil
}

func waitForFmsAdminAccountDeletion(conn *fms.FMS) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			fms.AccountRoleStatusDeleting,
			fms.AccountRoleStatusPendingDeletion,
			fms.AccountRoleStatusReady,
		},
		Target: []string{fms.AccountRoleStatusDeleted},
		Refresh: func() (interface{}, string, error) {
			output, err := conn.GetAdminAccount(&fms.GetAdminAccountInput{})

			if isAWSErr(err, fms.ErrCodeResourceNotFoundException, "") {
				return output, fms.AccountRoleStatusDeleted, nil
			}

			if err != nil {
				return nil, "", err
			}

			return output, aws.StringValue(output.RoleStatus), nil
		},
		Timeout: 1 * time.Minute,
		Delay:   10 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}
