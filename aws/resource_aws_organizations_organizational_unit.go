package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsOrganizationsOrganizationalUnit() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsOrganizationsOrganizationalUnitCreate,
		Read:   resourceAwsOrganizationsOrganizationalUnitRead,
		Update: resourceAwsOrganizationsOrganizationalUnitUpdate,
		Delete: resourceAwsOrganizationsOrganizationalUnitDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"accounts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"parent_id": {
				ForceNew:     true,
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^(r-[0-9a-z]{4,32})|(ou-[0-9a-z]{4,32}-[a-z0-9]{8,32})$"), "see https://docs.aws.amazon.com/organizations/latest/APIReference/API_CreateOrganizationalUnit.html#organizations-CreateOrganizationalUnit-request-ParentId"),
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsOrganizationsOrganizationalUnitCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	// Create the organizational unit
	createOpts := &organizations.CreateOrganizationalUnitInput{
		Name:     aws.String(d.Get("name").(string)),
		ParentId: aws.String(d.Get("parent_id").(string)),
		Tags:     tags.IgnoreAws().OrganizationsTags(),
	}

	var err error
	var resp *organizations.CreateOrganizationalUnitOutput
	err = resource.Retry(4*time.Minute, func() *resource.RetryError {
		resp, err = conn.CreateOrganizationalUnit(createOpts)

		if tfawserr.ErrCodeEquals(err, organizations.ErrCodeFinalizingOrganizationException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		resp, err = conn.CreateOrganizationalUnit(createOpts)
	}

	if err != nil {
		return fmt.Errorf("error creating Organizations Organizational Unit: %w", err)
	}

	// Store the ID
	d.SetId(aws.StringValue(resp.OrganizationalUnit.Id))

	return resourceAwsOrganizationsOrganizationalUnitRead(d, meta)
}

func resourceAwsOrganizationsOrganizationalUnitRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	describeOpts := &organizations.DescribeOrganizationalUnitInput{
		OrganizationalUnitId: aws.String(d.Id()),
	}
	resp, err := conn.DescribeOrganizationalUnit(describeOpts)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, organizations.ErrCodeOrganizationalUnitNotFoundException) {
		log.Printf("[WARN] Organizations Organizational Unit (%s) does not exist, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Organizations Organizational Unit (%s): %w", d.Id(), err)
	}

	if resp == nil {
		return fmt.Errorf("error reading Organizations Organizational Unit (%s): empty response", d.Id())
	}

	ou := resp.OrganizationalUnit
	if ou == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading Organizations Organizational Unit (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] Organizations Organizational Unit (%s) does not exist, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	parentId, err := resourceAwsOrganizationsOrganizationalUnitGetParentId(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing Organizations Organizational Unit (%s) parents: %w", d.Id(), err)
	}

	var accounts []*organizations.Account
	input := &organizations.ListAccountsForParentInput{
		ParentId: aws.String(d.Id()),
	}

	err = conn.ListAccountsForParentPages(input, func(page *organizations.ListAccountsForParentOutput, lastPage bool) bool {
		accounts = append(accounts, page.Accounts...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing Organizations Organizational Unit (%s) accounts: %w", d.Id(), err)
	}

	if err := d.Set("accounts", flattenOrganizationsOrganizationalUnitAccounts(accounts)); err != nil {
		return fmt.Errorf("error setting accounts: %w", err)
	}

	d.Set("arn", ou.Arn)
	d.Set("name", ou.Name)
	d.Set("parent_id", parentId)

	tags, err := keyvaluetags.OrganizationsListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for Organizations Organizational Unit (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsOrganizationsOrganizationalUnitUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	if d.HasChange("name") {
		updateOpts := &organizations.UpdateOrganizationalUnitInput{
			Name:                 aws.String(d.Get("name").(string)),
			OrganizationalUnitId: aws.String(d.Id()),
		}

		_, err := conn.UpdateOrganizationalUnit(updateOpts)
		if err != nil {
			return fmt.Errorf("error updating Organizations Organizational Unit (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.OrganizationsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Organizations Organizational Unit (%s) tags: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsOrganizationsOrganizationalUnitDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	input := &organizations.DeleteOrganizationalUnitInput{
		OrganizationalUnitId: aws.String(d.Id()),
	}

	_, err := conn.DeleteOrganizationalUnit(input)

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeOrganizationalUnitNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Organizations Organizational Unit (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsOrganizationsOrganizationalUnitGetParentId(conn *organizations.Organizations, childId string) (string, error) {
	input := &organizations.ListParentsInput{
		ChildId: aws.String(childId),
	}
	var parents []*organizations.Parent

	err := conn.ListParentsPages(input, func(page *organizations.ListParentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

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

func flattenOrganizationsOrganizationalUnitAccounts(accounts []*organizations.Account) []map[string]interface{} {
	if len(accounts) == 0 {
		return nil
	}

	var result []map[string]interface{}

	for _, account := range accounts {
		result = append(result, map[string]interface{}{
			"arn":   aws.StringValue(account.Arn),
			"email": aws.StringValue(account.Email),
			"id":    aws.StringValue(account.Id),
			"name":  aws.StringValue(account.Name),
		})
	}

	return result
}
