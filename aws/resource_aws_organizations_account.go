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

func resourceAwsOrganizationsAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsOrganizationsAccountCreate,
		Read:   resourceAwsOrganizationsAccountRead,
		Update: resourceAwsOrganizationsAccountUpdate,
		Delete: resourceAwsOrganizationsAccountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
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
				ForceNew: true,
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(6, 64),
					validation.StringMatch(regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`), "must be a valid email address"),
				),
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
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]{1,64}$`), "must consist of uppercase letters, lowercase letters, digits with no spaces, and any of the following characters"),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsOrganizationsAccountCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	// Create the account
	createOpts := &organizations.CreateAccountInput{
		AccountName: aws.String(d.Get("name").(string)),
		Email:       aws.String(d.Get("email").(string)),
	}
	if role, ok := d.GetOk("role_name"); ok {
		createOpts.RoleName = aws.String(role.(string))
	}

	if iam_user, ok := d.GetOk("iam_user_access_to_billing"); ok {
		createOpts.IamUserAccessToBilling = aws.String(iam_user.(string))
	}

	log.Printf("[DEBUG] Creating AWS Organizations Account: %s", createOpts)

	var resp *organizations.CreateAccountOutput
	err := resource.Retry(4*time.Minute, func() *resource.RetryError {
		var err error

		resp, err = conn.CreateAccount(createOpts)

		if isAWSErr(err, organizations.ErrCodeFinalizingOrganizationException, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		resp, err = conn.CreateAccount(createOpts)
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

	// Store the ID
	accountId := stateResp.(*organizations.CreateAccountStatus).AccountId
	d.SetId(*accountId)

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

func resourceAwsOrganizationsAccountRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	describeOpts := &organizations.DescribeAccountInput{
		AccountId: aws.String(d.Id()),
	}
	resp, err := conn.DescribeAccount(describeOpts)

	if isAWSErr(err, organizations.ErrCodeAccountNotFoundException, "") {
		log.Printf("[WARN] Account does not exist, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing AWS Organizations Account (%s): %s", d.Id(), err)
	}

	account := resp.Account
	if account == nil {
		log.Printf("[WARN] Account does not exist, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	parentId, err := resourceAwsOrganizationsAccountGetParentId(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error getting AWS Organizations Account (%s) parent: %s", d.Id(), err)
	}

	d.Set("arn", account.Arn)
	d.Set("email", account.Email)
	d.Set("joined_method", account.JoinedMethod)
	d.Set("joined_timestamp", aws.TimeValue(account.JoinedTimestamp).Format(time.RFC3339))
	d.Set("name", account.Name)
	d.Set("parent_id", parentId)
	d.Set("status", account.Status)

	tags, err := keyvaluetags.OrganizationsListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for AWS Organizations Account (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsOrganizationsAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	if d.HasChange("parent_id") {
		o, n := d.GetChange("parent_id")

		input := &organizations.MoveAccountInput{
			AccountId:           aws.String(d.Id()),
			SourceParentId:      aws.String(o.(string)),
			DestinationParentId: aws.String(n.(string)),
		}

		if _, err := conn.MoveAccount(input); err != nil {
			return fmt.Errorf("error moving AWS Organizations Account (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.OrganizationsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating AWS Organizations Account (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsOrganizationsAccountRead(d, meta)
}

func resourceAwsOrganizationsAccountDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	input := &organizations.RemoveAccountFromOrganizationInput{
		AccountId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Removing AWS account from organization: %s", input)
	_, err := conn.RemoveAccountFromOrganization(input)
	if err != nil {
		if isAWSErr(err, organizations.ErrCodeAccountNotFoundException, "") {
			return nil
		}
		return err
	}
	return nil
}

// resourceAwsOrganizationsAccountStateRefreshFunc returns a resource.StateRefreshFunc
// that is used to watch a CreateAccount request
func resourceAwsOrganizationsAccountStateRefreshFunc(conn *organizations.Organizations, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		opts := &organizations.DescribeCreateAccountStatusInput{
			CreateAccountRequestId: aws.String(id),
		}
		resp, err := conn.DescribeCreateAccountStatus(opts)
		if err != nil {
			if isAWSErr(err, organizations.ErrCodeCreateAccountStatusNotFoundException, "") {
				resp = nil
			} else {
				log.Printf("Error on OrganizationAccountStateRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our account yet. Return an empty state.
			return nil, "", nil
		}

		accountStatus := resp.CreateAccountStatus
		if *accountStatus.State == organizations.CreateAccountStateFailed {
			return nil, *accountStatus.State, fmt.Errorf(*accountStatus.FailureReason)
		}
		return accountStatus, *accountStatus.State, nil
	}
}

func resourceAwsOrganizationsAccountGetParentId(conn *organizations.Organizations, childId string) (string, error) {
	input := &organizations.ListParentsInput{
		ChildId: aws.String(childId),
	}
	var parents []*organizations.Parent

	err := conn.ListParentsPages(input, func(page *organizations.ListParentsOutput, lastPage bool) bool {
		parents = append(parents, page.Parents...)

		return !lastPage
	})

	if err != nil {
		return "", err
	}

	if len(parents) == 0 {
		return "", nil
	}

	// assume there is only a single parent
	// https://docs.aws.amazon.com/organizations/latest/APIReference/API_ListParents.html
	parent := parents[0]
	return aws.StringValue(parent.Id), nil
}
