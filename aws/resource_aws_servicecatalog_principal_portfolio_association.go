package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
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

func resourceAwsServiceCatalogPrincipalPortfolioAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogPrincipalPortfolioAssociationCreate,
		Read:   resourceAwsServiceCatalogPrincipalPortfolioAssociationRead,
		Delete: resourceAwsServiceCatalogPrincipalPortfolioAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      tfservicecatalog.AcceptLanguageEnglish,
				ValidateFunc: validation.StringInSlice(tfservicecatalog.AcceptLanguage_Values(), false),
			},
			"portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"principal_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"principal_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      servicecatalog.PrincipalTypeIam,
				ValidateFunc: validation.StringInSlice(servicecatalog.PrincipalType_Values(), false),
			},
		},
	}
}

func resourceAwsServiceCatalogPrincipalPortfolioAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := &servicecatalog.AssociatePrincipalWithPortfolioInput{
		PortfolioId:  aws.String(d.Get("portfolio_id").(string)),
		PrincipalARN: aws.String(d.Get("principal_arn").(string)),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("principal_type"); ok {
		input.PrincipalType = aws.String(v.(string))
	}

	var output *servicecatalog.AssociatePrincipalWithPortfolioOutput
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.AssociatePrincipalWithPortfolio(input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.AssociatePrincipalWithPortfolio(input)
	}

	if err != nil {
		return fmt.Errorf("error associating Service Catalog Principal with Portfolio: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Service Catalog Principal Portfolio Association: empty response")
	}

	d.SetId(tfservicecatalog.PrincipalPortfolioAssociationID(d.Get("accept_language").(string), d.Get("principal_arn").(string), d.Get("portfolio_id").(string)))

	return resourceAwsServiceCatalogPrincipalPortfolioAssociationRead(d, meta)
}

func resourceAwsServiceCatalogPrincipalPortfolioAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	acceptLanguage, principalARN, portfolioID, err := tfservicecatalog.PrincipalPortfolioAssociationParseID(d.Id())

	if err != nil {
		return fmt.Errorf("could not parse ID (%s): %w", d.Id(), err)
	}

	if acceptLanguage == "" {
		acceptLanguage = tfservicecatalog.AcceptLanguageEnglish
	}

	output, err := waiter.PrincipalPortfolioAssociationReady(conn, acceptLanguage, principalARN, portfolioID)

	if !d.IsNewResource() && (tfresource.NotFound(err) || tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException)) {
		log.Printf("[WARN] Service Catalog Principal Portfolio Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Principal Portfolio Association (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting Service Catalog Principal Portfolio Association (%s): empty response", d.Id())
	}

	d.Set("accept_language", acceptLanguage)
	d.Set("portfolio_id", portfolioID)
	d.Set("principal_arn", output.PrincipalARN)
	d.Set("principal_type", output.PrincipalType)

	return nil
}

func resourceAwsServiceCatalogPrincipalPortfolioAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	acceptLanguage, principalARN, portfolioID, err := tfservicecatalog.PrincipalPortfolioAssociationParseID(d.Id())

	if err != nil {
		return fmt.Errorf("could not parse ID (%s): %w", d.Id(), err)
	}

	if acceptLanguage == "" {
		acceptLanguage = tfservicecatalog.AcceptLanguageEnglish
	}

	input := &servicecatalog.DisassociatePrincipalFromPortfolioInput{
		PortfolioId:    aws.String(portfolioID),
		PrincipalARN:   aws.String(principalARN),
		AcceptLanguage: aws.String(acceptLanguage),
	}

	_, err = conn.DisassociatePrincipalFromPortfolio(input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating Service Catalog Principal from Portfolio (%s): %w", d.Id(), err)
	}

	err = waiter.PrincipalPortfolioAssociationDeleted(conn, acceptLanguage, principalARN, portfolioID)

	if tfresource.NotFound(err) || tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error waiting for Service Catalog Principal Portfolio Disassociation (%s): %w", d.Id(), err)
	}

	return nil
}
