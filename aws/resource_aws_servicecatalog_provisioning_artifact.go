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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	tfservicecatalog "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsServiceCatalogProvisioningArtifact() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogProvisioningArtifactCreate,
		Read:   resourceAwsServiceCatalogProvisioningArtifactRead,
		Update: resourceAwsServiceCatalogProvisioningArtifactUpdate,
		Delete: resourceAwsServiceCatalogProvisioningArtifactDelete,
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
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"disable_template_validation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"guidance": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      servicecatalog.ProvisioningArtifactGuidanceDefault,
				ValidateFunc: validation.StringInSlice(servicecatalog.ProvisioningArtifactGuidance_Values(), false),
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"template_physical_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"template_url",
					"template_physical_id",
				},
			},
			"template_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"template_url",
					"template_physical_id",
				},
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(servicecatalog.ProvisioningArtifactType_Values(), false),
			},
		},
	}
}

func resourceAwsServiceCatalogProvisioningArtifactCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	parameters := make(map[string]interface{})
	parameters["description"] = d.Get("description")
	parameters["disable_template_validation"] = d.Get("disable_template_validation")
	parameters["name"] = d.Get("name")
	parameters["template_physical_id"] = d.Get("template_physical_id")
	parameters["template_url"] = d.Get("template_url")
	parameters["type"] = d.Get("type")

	input := &servicecatalog.CreateProvisioningArtifactInput{
		IdempotencyToken: aws.String(resource.UniqueId()),
		Parameters:       expandProvisioningArtifactParameters(parameters),
		ProductId:        aws.String(d.Get("product_id").(string)),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	var output *servicecatalog.CreateProvisioningArtifactOutput
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.CreateProvisioningArtifact(input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateProvisioningArtifact(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Service Catalog Provisioning Artifact: %w", err)
	}

	if output == nil || output.ProvisioningArtifactDetail == nil || output.ProvisioningArtifactDetail.Id == nil {
		return fmt.Errorf("error creating Service Catalog Provisioning Artifact: empty response")
	}

	d.SetId(tfservicecatalog.ProvisioningArtifactID(aws.StringValue(output.ProvisioningArtifactDetail.Id), d.Get("product_id").(string)))

	// Active and Guidance are not fields of CreateProvisioningArtifact but are fields of UpdateProvisioningArtifact.
	// In order to set these to non-default values, you must create and then update.

	return resourceAwsServiceCatalogProvisioningArtifactUpdate(d, meta)
}

func resourceAwsServiceCatalogProvisioningArtifactRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	artifactID, productID, err := tfservicecatalog.ProvisioningArtifactParseID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Service Catalog Provisioning Artifact ID (%s): %w", d.Id(), err)
	}

	output, err := waiter.ProvisioningArtifactReady(conn, artifactID, productID)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Service Catalog Provisioning Artifact (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Provisioning Artifact (%s): %w", d.Id(), err)
	}

	if output == nil || output.ProvisioningArtifactDetail == nil {
		return fmt.Errorf("error getting Service Catalog Provisioning Artifact (%s): empty response", d.Id())
	}

	if v, ok := output.Info["ImportFromPhysicalId"]; ok {
		d.Set("template_physical_id", v)
	}

	if v, ok := output.Info["LoadTemplateFromURL"]; ok {
		d.Set("template_url", v)
	}

	pad := output.ProvisioningArtifactDetail

	d.Set("active", pad.Active)
	if pad.CreatedTime != nil {
		d.Set("created_time", pad.CreatedTime.Format(time.RFC3339))
	}
	d.Set("description", pad.Description)
	d.Set("guidance", pad.Guidance)
	d.Set("name", pad.Name)
	d.Set("product_id", productID)
	d.Set("type", pad.Type)

	return nil
}

func resourceAwsServiceCatalogProvisioningArtifactUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	if d.HasChanges("accept_language", "active", "description", "guidance", "name", "product_id") {
		artifactID, productID, err := tfservicecatalog.ProvisioningArtifactParseID(d.Id())

		if err != nil {
			return fmt.Errorf("error parsing Service Catalog Provisioning Artifact ID (%s): %w", d.Id(), err)
		}

		input := &servicecatalog.UpdateProvisioningArtifactInput{
			ProductId:              aws.String(productID),
			ProvisioningArtifactId: aws.String(artifactID),
			Active:                 aws.Bool(d.Get("active").(bool)),
		}

		if v, ok := d.GetOk("accept_language"); ok {
			input.AcceptLanguage = aws.String(v.(string))
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("guidance"); ok {
			input.Guidance = aws.String(v.(string))
		}

		if v, ok := d.GetOk("name"); ok {
			input.Name = aws.String(v.(string))
		}

		err = resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
			_, err := conn.UpdateProvisioningArtifact(input)

			if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateProvisioningArtifact(input)
		}

		if err != nil {
			return fmt.Errorf("error updating Service Catalog Provisioning Artifact (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsServiceCatalogProvisioningArtifactRead(d, meta)
}

func resourceAwsServiceCatalogProvisioningArtifactDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	artifactID, productID, err := tfservicecatalog.ProvisioningArtifactParseID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Service Catalog Provisioning Artifact ID (%s): %w", d.Id(), err)
	}

	input := &servicecatalog.DeleteProvisioningArtifactInput{
		ProductId:              aws.String(productID),
		ProvisioningArtifactId: aws.String(artifactID),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	_, err = conn.DeleteProvisioningArtifact(input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Service Catalog Provisioning Artifact (%s): %w", d.Id(), err)
	}

	if err := waiter.ProvisioningArtifactDeleted(conn, artifactID, productID); err != nil {
		return fmt.Errorf("error waiting for Service Catalog Provisioning Artifact (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
