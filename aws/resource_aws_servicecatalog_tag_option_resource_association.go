package aws

import (
	"fmt"
	"log"
	"time"

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

func resourceAwsServiceCatalogTagOptionResourceAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogTagOptionResourceAssociationCreate,
		Read:   resourceAwsServiceCatalogTagOptionResourceAssociationRead,
		Delete: resourceAwsServiceCatalogTagOptionResourceAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tag_option_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsServiceCatalogTagOptionResourceAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := &servicecatalog.AssociateTagOptionWithResourceInput{
		ResourceId:  aws.String(d.Get("resource_id").(string)),
		TagOptionId: aws.String(d.Get("tag_option_id").(string)),
	}

	var output *servicecatalog.AssociateTagOptionWithResourceOutput
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.AssociateTagOptionWithResource(input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.AssociateTagOptionWithResource(input)
	}

	if err != nil {
		return fmt.Errorf("error associating Service Catalog Tag Option with Resource: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Service Catalog Tag Option Resource Association: empty response")
	}

	d.SetId(tfservicecatalog.TagOptionResourceAssociationID(d.Get("tag_option_id").(string), d.Get("resource_id").(string)))

	return resourceAwsServiceCatalogTagOptionResourceAssociationRead(d, meta)
}

func resourceAwsServiceCatalogTagOptionResourceAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	tagOptionID, resourceID, err := tfservicecatalog.TagOptionResourceAssociationParseID(d.Id())

	if err != nil {
		return fmt.Errorf("could not parse ID (%s): %w", d.Id(), err)
	}

	output, err := waiter.TagOptionResourceAssociationReady(conn, tagOptionID, resourceID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Catalog Tag Option Resource Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Tag Option Resource Association (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting Service Catalog Tag Option Resource Association (%s): empty response", d.Id())
	}

	if output.CreatedTime != nil {
		d.Set("resource_created_time", output.CreatedTime.Format(time.RFC3339))
	}

	d.Set("resource_arn", output.ARN)
	d.Set("resource_description", output.Description)
	d.Set("resource_id", output.Id)
	d.Set("resource_name", output.Name)
	d.Set("tag_option_id", tagOptionID)

	return nil
}

func resourceAwsServiceCatalogTagOptionResourceAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	tagOptionID, resourceID, err := tfservicecatalog.TagOptionResourceAssociationParseID(d.Id())

	if err != nil {
		return fmt.Errorf("could not parse ID (%s): %w", d.Id(), err)
	}

	input := &servicecatalog.DisassociateTagOptionFromResourceInput{
		ResourceId:  aws.String(resourceID),
		TagOptionId: aws.String(tagOptionID),
	}

	_, err = conn.DisassociateTagOptionFromResource(input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating Service Catalog Tag Option from Resource (%s): %w", d.Id(), err)
	}

	err = waiter.TagOptionResourceAssociationDeleted(conn, tagOptionID, resourceID)

	if err != nil && !tfresource.NotFound(err) {
		return fmt.Errorf("error waiting for Service Catalog Tag Option Resource Disassociation (%s): %w", d.Id(), err)
	}

	return nil
}
