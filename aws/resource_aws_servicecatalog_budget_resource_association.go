package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	tfservicecatalog "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsServiceCatalogBudgetResourceAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogBudgetResourceAssociationCreate,
		Read:   resourceAwsServiceCatalogBudgetResourceAssociationRead,
		Delete: resourceAwsServiceCatalogBudgetResourceAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"budget_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsServiceCatalogBudgetResourceAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := &servicecatalog.AssociateBudgetWithResourceInput{
		BudgetName: aws.String(d.Get("budget_name").(string)),
		ResourceId: aws.String(d.Get("resource_id").(string)),
	}

	var output *servicecatalog.AssociateBudgetWithResourceOutput
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.AssociateBudgetWithResource(input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.AssociateBudgetWithResource(input)
	}

	if err != nil {
		return fmt.Errorf("error associating Service Catalog Budget with Resource: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Service Catalog Budget Resource Association: empty response")
	}

	d.SetId(tfservicecatalog.BudgetResourceAssociationID(d.Get("budget_name").(string), d.Get("resource_id").(string)))

	return resourceAwsServiceCatalogBudgetResourceAssociationRead(d, meta)
}

func resourceAwsServiceCatalogBudgetResourceAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	budgetName, resourceID, err := tfservicecatalog.BudgetResourceAssociationParseID(d.Id())

	if err != nil {
		return fmt.Errorf("could not parse ID (%s): %w", d.Id(), err)
	}

	output, err := waiter.BudgetResourceAssociationReady(conn, budgetName, resourceID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Catalog Budget Resource Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Budget Resource Association (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting Service Catalog Budget Resource Association (%s): empty response", d.Id())
	}

	d.Set("resource_id", resourceID)
	d.Set("budget_name", output.BudgetName)

	return nil
}

func resourceAwsServiceCatalogBudgetResourceAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	budgetName, resourceID, err := tfservicecatalog.BudgetResourceAssociationParseID(d.Id())

	if err != nil {
		return fmt.Errorf("could not parse ID (%s): %w", d.Id(), err)
	}

	input := &servicecatalog.DisassociateBudgetFromResourceInput{
		ResourceId: aws.String(resourceID),
		BudgetName: aws.String(budgetName),
	}

	_, err = conn.DisassociateBudgetFromResource(input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating Service Catalog Budget from Resource (%s): %w", d.Id(), err)
	}

	err = waiter.BudgetResourceAssociationDeleted(conn, budgetName, resourceID)

	if err != nil && !tfresource.NotFound(err) {
		return fmt.Errorf("error waiting for Service Catalog Budget Resource Disassociation (%s): %w", d.Id(), err)
	}

	return nil
}
