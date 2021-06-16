package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	tfservicecatalog "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsServiceCatalogPortfolioShare() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogPortfolioShareCreate,
		Read:   resourceAwsServiceCatalogPortfolioShareRead,
		Update: resourceAwsServiceCatalogPortfolioShareUpdate,
		Delete: resourceAwsServiceCatalogPortfolioShareDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      tfservicecatalog.AcceptLanguageEnglish,
				ValidateFunc: validation.StringInSlice(tfservicecatalog.AcceptLanguage_Values(), false),
			},
			"accepted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// maintaining organization_node as a separate config block makes weird configs with duplicate types
			// also, principal_id is true to API since describe gives "PrincipalId"
			"principal_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateServiceCatalogSharePrincipal,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					newARN, err := arn.Parse(new)

					if err != nil {
						return old == new
					}

					parts := strings.Split(newARN.Resource, "/")

					return old == parts[len(parts)-1]
				},
			},
			"share_tag_options": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(servicecatalog.DescribePortfolioShareType_Values(), false),
			},
			"wait_for_acceptance": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceAwsServiceCatalogPortfolioShareCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := &servicecatalog.CreatePortfolioShareInput{
		PortfolioId: aws.String(d.Get("portfolio_id").(string)),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type"); ok && v.(string) == servicecatalog.DescribePortfolioShareTypeAccount {
		input.AccountId = aws.String(d.Get("principal_id").(string))
	} else {
		orgNode := &servicecatalog.OrganizationNode{}
		orgNode.Value = aws.String(d.Get("principal_id").(string))

		if v.(string) == servicecatalog.DescribePortfolioShareTypeOrganizationMemberAccount {
			// portfolio_share type ORGANIZATION_MEMBER_ACCOUNT = org node type ACCOUNT
			orgNode.Type = aws.String(servicecatalog.OrganizationNodeTypeAccount)
		} else {
			orgNode.Type = aws.String(d.Get("type").(string))
		}

		input.OrganizationNode = orgNode
	}

	if v, ok := d.GetOk("share_tag_options"); ok {
		input.ShareTagOptions = aws.Bool(v.(bool))
	}

	var output *servicecatalog.CreatePortfolioShareOutput
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.CreatePortfolioShare(input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreatePortfolioShare(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Service Catalog Portfolio Share: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Service Catalog Portfolio Share: empty response")
	}

	d.SetId(tfservicecatalog.PortfolioShareCreateResourceID(d.Get("portfolio_id").(string), d.Get("type").(string), d.Get("principal_id").(string)))

	waitForAcceptance := false
	if v, ok := d.GetOk("wait_for_acceptance"); ok {
		waitForAcceptance = v.(bool)
	}

	// only get a token if organization node, otherwise check without token
	if output.PortfolioShareToken != nil {
		if _, err := waiter.PortfolioShareCreatedWithToken(conn, aws.StringValue(output.PortfolioShareToken), waitForAcceptance); err != nil {
			return fmt.Errorf("error waiting for Service Catalog Portfolio Share (%s) to be ready: %w", d.Id(), err)
		}
	} else {
		if _, err := waiter.PortfolioShareReady(conn, d.Get("portfolio_id").(string), d.Get("type").(string), d.Get("principal_id").(string), waitForAcceptance); err != nil {
			return fmt.Errorf("error waiting for Service Catalog Portfolio Share (%s) to be ready: %w", d.Id(), err)
		}
	}

	return resourceAwsServiceCatalogPortfolioShareRead(d, meta)
}

func resourceAwsServiceCatalogPortfolioShareRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	portfolioID, shareType, principalID, err := tfservicecatalog.PortfolioShareParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("could not parse ID (%s): %w", d.Id(), err)
	}

	waitForAcceptance := false
	if v, ok := d.GetOk("wait_for_acceptance"); ok {
		waitForAcceptance = v.(bool)
	}

	output, err := waiter.PortfolioShareReady(conn, portfolioID, shareType, principalID, waitForAcceptance)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Service Catalog Portfolio Share (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Portfolio Share (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting Service Catalog Portfolio Share (%s): empty response", d.Id())
	}

	d.Set("accepted", output.Accepted)
	d.Set("portfolio_id", portfolioID)
	d.Set("principal_id", output.PrincipalId)
	d.Set("share_tag_options", output.ShareTagOptions)
	d.Set("type", output.Type)
	d.Set("wait_for_acceptance", waitForAcceptance)

	return nil
}

func resourceAwsServiceCatalogPortfolioShareUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	if d.HasChanges("accept_language", "share_tag_options") {
		input := &servicecatalog.UpdatePortfolioShareInput{
			PortfolioId: aws.String(d.Get("portfolio_id").(string)),
		}

		if v, ok := d.GetOk("accept_language"); ok {
			input.AcceptLanguage = aws.String(v.(string))
		}

		if v, ok := d.GetOk("share_tag_options"); ok {
			input.ShareTagOptions = aws.Bool(v.(bool))
		}

		err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
			_, err := conn.UpdatePortfolioShare(input)

			if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdatePortfolioShare(input)
		}

		if err != nil {
			return fmt.Errorf("error updating Service Catalog Portfolio Share (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsServiceCatalogPortfolioShareRead(d, meta)
}

func resourceAwsServiceCatalogPortfolioShareDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := &servicecatalog.DeletePortfolioShareInput{
		PortfolioId: aws.String(d.Get("portfolio_id").(string)),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type"); ok && v.(string) == servicecatalog.DescribePortfolioShareTypeAccount {
		input.AccountId = aws.String(d.Get("principal_id").(string))
	} else {
		orgNode := &servicecatalog.OrganizationNode{}
		orgNode.Value = aws.String(d.Get("principal_id").(string))

		if v.(string) == servicecatalog.DescribePortfolioShareTypeOrganizationMemberAccount {
			// portfolio_share type ORGANIZATION_MEMBER_ACCOUNT = org node type ACCOUNT
			orgNode.Type = aws.String(servicecatalog.OrganizationNodeTypeAccount)
		} else {
			orgNode.Type = aws.String(d.Get("type").(string))
		}

		input.OrganizationNode = orgNode
	}

	output, err := conn.DeletePortfolioShare(input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Service Catalog Portfolio Share (%s): %w", d.Id(), err)
	}

	// only get a token if organization node, otherwise check without token
	if output.PortfolioShareToken != nil {
		if _, err := waiter.PortfolioShareDeletedWithToken(conn, aws.StringValue(output.PortfolioShareToken)); err != nil {
			return fmt.Errorf("error waiting for Service Catalog Portfolio Share (%s) to be deleted: %w", d.Id(), err)
		}
	} else {
		if _, err := waiter.PortfolioShareDeleted(conn, d.Get("portfolio_id").(string), d.Get("type").(string), d.Get("principal_id").(string)); err != nil {
			return fmt.Errorf("error waiting for Service Catalog Portfolio Share (%s) to be deleted: %w", d.Id(), err)
		}
	}

	return nil
}
