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

func resourceAwsServiceCatalogProductPortfolioAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogProductPortfolioAssociationCreate,
		Read:   resourceAwsServiceCatalogProductPortfolioAssociationRead,
		Delete: resourceAwsServiceCatalogProductPortfolioAssociationDelete,
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
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_portfolio_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsServiceCatalogProductPortfolioAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := &servicecatalog.AssociateProductWithPortfolioInput{
		PortfolioId: aws.String(d.Get("portfolio_id").(string)),
		ProductId:   aws.String(d.Get("product_id").(string)),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_portfolio_id"); ok {
		input.SourcePortfolioId = aws.String(v.(string))
	}

	var output *servicecatalog.AssociateProductWithPortfolioOutput
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.AssociateProductWithPortfolio(input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.AssociateProductWithPortfolio(input)
	}

	if err != nil {
		return fmt.Errorf("error associating Service Catalog Product with Portfolio: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Service Catalog Product Portfolio Association: empty response")
	}

	d.SetId(tfservicecatalog.ProductPortfolioAssociationCreateID(d.Get("accept_language").(string), d.Get("portfolio_id").(string), d.Get("product_id").(string)))

	return resourceAwsServiceCatalogProductPortfolioAssociationRead(d, meta)
}

func resourceAwsServiceCatalogProductPortfolioAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	acceptLanguage, portfolioID, productID, err := tfservicecatalog.ProductPortfolioAssociationParseID(d.Id())

	if err != nil {
		return fmt.Errorf("could not parse ID (%s): %w", d.Id(), err)
	}

	output, err := waiter.ProductPortfolioAssociationReady(conn, acceptLanguage, portfolioID, productID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Catalog Product Portfolio Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Product Portfolio Association (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting Service Catalog Product Portfolio Association (%s): empty response", d.Id())
	}

	d.Set("accept_language", acceptLanguage)
	d.Set("portfolio_id", output.Id)
	d.Set("product_id", productID)
	d.Set("source_portfolio_id", d.Get("source_portfolio_id").(string))

	return nil
}

func resourceAwsServiceCatalogProductPortfolioAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	acceptLanguage, portfolioID, productID, err := tfservicecatalog.ProductPortfolioAssociationParseID(d.Id())

	if err != nil {
		return fmt.Errorf("could not parse ID (%s): %w", d.Id(), err)
	}

	input := &servicecatalog.DisassociateProductFromPortfolioInput{
		PortfolioId: aws.String(portfolioID),
		ProductId:   aws.String(productID),
	}

	if acceptLanguage != "" {
		input.AcceptLanguage = aws.String(acceptLanguage)
	}

	_, err = conn.DisassociateProductFromPortfolio(input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating Service Catalog Product from Portfolio (%s): %w", d.Id(), err)
	}

	err = waiter.ProductPortfolioAssociationDeleted(conn, acceptLanguage, portfolioID, productID)

	if err != nil && !tfresource.NotFound(err) {
		return fmt.Errorf("error waiting for Service Catalog Product Portfolio Disassociation (%s): %w", d.Id(), err)
	}

	return nil
}
