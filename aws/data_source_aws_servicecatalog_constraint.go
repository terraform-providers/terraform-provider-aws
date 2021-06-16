package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfservicecatalog "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
)

func dataSourceAwsServiceCatalogConstraint() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsServiceCatalogConstraintRead,

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      tfservicecatalog.AcceptLanguageEnglish,
				ValidateFunc: validation.StringInSlice(tfservicecatalog.AcceptLanguage_Values(), false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parameters": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"portfolio_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsServiceCatalogConstraintRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	output, err := waiter.ConstraintReady(conn, d.Get("accept_language").(string), d.Get("id").(string))

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Constraint: %w", err)
	}

	if output == nil || output.ConstraintDetail == nil {
		return fmt.Errorf("error getting Service Catalog Constraint: empty response")
	}

	d.Set("accept_language", d.Get("accept_language").(string))

	d.Set("parameters", output.ConstraintParameters)
	d.Set("status", output.Status)

	detail := output.ConstraintDetail

	d.Set("description", detail.Description)
	d.Set("owner", detail.Owner)
	d.Set("portfolio_id", detail.PortfolioId)
	d.Set("product_id", detail.ProductId)
	d.Set("type", detail.Type)

	d.SetId(aws.StringValue(detail.ConstraintId))

	return nil
}
