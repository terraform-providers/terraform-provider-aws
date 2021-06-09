package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsOrganizationsGovCloudAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsOrganizationsGovCloudAccountCreate,
		Read:   resourceAwsOrganizationsAccountRead,
		Update: resourceAwsOrganizationsAccountUpdate,
		Delete: resourceAwsOrganizationsAccountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"govcloud_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"joined_method": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"joined_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_id": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^(r-[0-9a-z]{4,32})|(ou-[0-9a-z]{4,32}-[a-z0-9]{8,32})$"), "see https://docs.aws.amazon.com/organizations/latest/APIReference/API_MoveAccount.html#organizations-MoveAccount-request-DestinationParentId"),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				ForceNew:     true,
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"email": {
				ForceNew:     true,
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateAwsOrganizationsAccountEmail,
			},
			"iam_user_access_to_billing": {
				ForceNew:     true,
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{organizations.IAMUserAccessToBillingAllow, organizations.IAMUserAccessToBillingDeny}, true),
			},
			"role_name": {
				ForceNew:     true,
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateAwsOrganizationsAccountRoleName,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsOrganizationsGovCloudAccountCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	// Create the account
	createOpts := &organizations.CreateGovCloudAccountInput{
		AccountName: aws.String(d.Get("name").(string)),
		Email:       aws.String(d.Get("email").(string)),
	}
	if role, ok := d.GetOk("role_name"); ok {
		createOpts.RoleName = aws.String(role.(string))
	}

	if iam_user, ok := d.GetOk("iam_user_access_to_billing"); ok {
		createOpts.IamUserAccessToBilling = aws.String(iam_user.(string))
	}

	log.Printf("[DEBUG] Creating AWS Organizations GovCloud Account: %s", createOpts)

	var resp *organizations.CreateGovCloudAccountOutput
	err := resource.Retry(4*time.Minute, func() *resource.RetryError {
		var err error

		resp, err = conn.CreateGovCloudAccount(createOpts)

		if isAWSErr(err, organizations.ErrCodeFinalizingOrganizationException, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		resp, err = conn.CreateGovCloudAccount(createOpts)
	}

	if err != nil {
		return fmt.Errorf("Error creating account: %s", err)
	}

	requestId := *resp.CreateAccountStatus.Id

	// Wait for the account to become available
	log.Printf("[DEBUG] Waiting for account request (%s) to succeed", requestId)

	stateConf := &resource.StateChangeConf{
		Pending:      []string{organizations.CreateAccountStateInProgress},
		Target:       []string{organizations.CreateAccountStateSucceeded},
		Refresh:      resourceAwsOrganizationsAccountStateRefreshFunc(conn, requestId),
		PollInterval: 10 * time.Second,
		Timeout:      5 * time.Minute,
	}
	stateResp, stateErr := stateConf.WaitForState()
	if stateErr != nil {
		return fmt.Errorf(
			"Error waiting for account request (%s) to become available: %s",
			requestId, stateErr)
	}

	// Store the Commercial account ID
	accountId := stateResp.(*organizations.CreateAccountStatus).AccountId
	d.SetId(*accountId)

	// Store the GovCloud account ID
	govCloudAccountId := stateResp.(*organizations.CreateAccountStatus).GovCloudAccountId
	d.Set("govcloud_account_id", *govCloudAccountId)

	if v, ok := d.GetOk("parent_id"); ok {
		newParentID := v.(string)

		existingParentID, err := resourceAwsOrganizationsAccountGetParentId(conn, d.Id())

		if err != nil {
			return fmt.Errorf("error getting AWS Organizations Account (%s) parent: %s", d.Id(), err)
		}

		if newParentID != existingParentID {
			input := &organizations.MoveAccountInput{
				AccountId:           accountId,
				SourceParentId:      aws.String(existingParentID),
				DestinationParentId: aws.String(newParentID),
			}

			if _, err := conn.MoveAccount(input); err != nil {
				return fmt.Errorf("error moving AWS Organizations Account (%s): %s", d.Id(), err)
			}
		}
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		if err := keyvaluetags.OrganizationsUpdateTags(conn, d.Id(), nil, v); err != nil {
			return fmt.Errorf("error adding AWS Organizations Account (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsOrganizationsAccountRead(d, meta)
}
