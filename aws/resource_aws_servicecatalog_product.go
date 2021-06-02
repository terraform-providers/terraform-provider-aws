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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	tfservicecatalog "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsServiceCatalogProduct() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogProductCreate,
		Read:   resourceAwsServiceCatalogProductRead,
		Update: resourceAwsServiceCatalogProductUpdate,
		Delete: resourceAwsServiceCatalogProductDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "en",
				ValidateFunc: validation.StringInSlice(tfservicecatalog.AcceptLanguage_Values(), false),
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
			"distributor": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"has_default_path": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Required: true,
			},
			"provisioning_artifact_parameters": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"disable_template_validation": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"template_physical_id": {
							Type:     schema.TypeString,
							Optional: true,
							ExactlyOneOf: []string{
								"provisioning_artifact_parameters.0.template_url",
								"provisioning_artifact_parameters.0.template_physical_id",
							},
						},
						"template_url": {
							Type:     schema.TypeString,
							Optional: true,
							ExactlyOneOf: []string{
								"provisioning_artifact_parameters.0.template_url",
								"provisioning_artifact_parameters.0.template_physical_id",
							},
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(servicecatalog.ProvisioningArtifactType_Values(), false),
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"support_email": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"support_url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(servicecatalog.ProductType_Values(), false),
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsServiceCatalogProductCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &servicecatalog.CreateProductInput{
		IdempotencyToken: aws.String(resource.UniqueId()),
		Name:             aws.String(d.Get("name").(string)),
		Owner:            aws.String(d.Get("owner").(string)),
		ProductType:      aws.String(d.Get("type").(string)),
		ProvisioningArtifactParameters: expandProvisioningArtifactParameters(
			d.Get("provisioning_artifact_parameters").([]interface{})[0].(map[string]interface{}),
		),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("distributor"); ok {
		input.Distributor = aws.String(v.(string))
	}

	if v, ok := d.GetOk("support_description"); ok {
		input.SupportDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("support_email"); ok {
		input.SupportEmail = aws.String(v.(string))
	}

	if v, ok := d.GetOk("support_url"); ok {
		input.SupportUrl = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().ServicecatalogTags()
	}

	var output *servicecatalog.CreateProductOutput
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.CreateProduct(input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateProduct(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Service Catalog Product: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Service Catalog Product: empty response")
	}

	if output.ProductViewDetail == nil || output.ProductViewDetail.ProductViewSummary == nil {
		return fmt.Errorf("error creating Service Catalog Product: no product view detail or summary")
	}

	if output.ProvisioningArtifactDetail == nil {
		return fmt.Errorf("error creating Service Catalog Product: no provisioning artifact detail")
	}

	d.SetId(aws.StringValue(output.ProductViewDetail.ProductViewSummary.ProductId))

	if _, err := waiter.ProductReady(conn, aws.StringValue(input.AcceptLanguage),
		aws.StringValue(output.ProductViewDetail.ProductViewSummary.ProductId)); err != nil {
		return fmt.Errorf("error waiting for Service Catalog Product (%s) to be ready: %w", d.Id(), err)
	}

	return resourceAwsServiceCatalogProductRead(d, meta)
}

func resourceAwsServiceCatalogProductRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	output, err := waiter.ProductReady(conn, d.Get("accept_language").(string), d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Service Catalog Product (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Product (%s): %w", d.Id(), err)
	}

	if output == nil || output.ProductViewDetail == nil || output.ProductViewDetail.ProductViewSummary == nil {
		return fmt.Errorf("error getting Service Catalog Product (%s): empty response", d.Id())
	}

	pvs := output.ProductViewDetail.ProductViewSummary

	d.Set("arn", output.ProductViewDetail.ProductARN)
	if output.ProductViewDetail.CreatedTime != nil {
		d.Set("created_time", output.ProductViewDetail.CreatedTime.Format(time.RFC3339))
	}
	d.Set("description", pvs.ShortDescription)
	d.Set("distributor", pvs.Distributor)
	d.Set("has_default_path", pvs.HasDefaultPath)
	d.Set("name", pvs.Name)
	d.Set("owner", pvs.Owner)
	d.Set("status", output.ProductViewDetail.Status)
	d.Set("support_description", pvs.SupportDescription)
	d.Set("support_email", pvs.SupportEmail)
	d.Set("support_url", pvs.SupportUrl)
	d.Set("type", pvs.Type)

	tags := keyvaluetags.ServicecatalogKeyValueTags(output.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsServiceCatalogProductUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &servicecatalog.UpdateProductInput{
			Id: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("accept_language"); ok {
			input.AcceptLanguage = aws.String(v.(string))
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("distributor"); ok {
			input.Distributor = aws.String(v.(string))
		}

		if v, ok := d.GetOk("name"); ok {
			input.Name = aws.String(v.(string))
		}

		if v, ok := d.GetOk("owner"); ok {
			input.Owner = aws.String(v.(string))
		}

		if v, ok := d.GetOk("support_description"); ok {
			input.SupportDescription = aws.String(v.(string))
		}

		if v, ok := d.GetOk("support_email"); ok {
			input.SupportEmail = aws.String(v.(string))
		}

		if v, ok := d.GetOk("support_url"); ok {
			input.SupportUrl = aws.String(v.(string))
		}

		err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
			_, err := conn.UpdateProduct(input)

			if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateProduct(input)
		}

		if err != nil {
			return fmt.Errorf("error updating Service Catalog Product (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.ServiceCatalogProductUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags for Service Catalog Product (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsServiceCatalogProductRead(d, meta)
}

func resourceAwsServiceCatalogProductDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := &servicecatalog.DeleteProductInput{
		Id: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	_, err := conn.DeleteProduct(input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Service Catalog Product (%s): %w", d.Id(), err)
	}

	if _, err := waiter.ProductDeleted(conn, d.Get("accept_language").(string), d.Id()); err != nil {
		return fmt.Errorf("error waiting for Service Catalog Product (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func expandProvisioningArtifactParameters(tfMap map[string]interface{}) *servicecatalog.ProvisioningArtifactProperties {
	if tfMap == nil {
		return nil
	}

	apiObject := &servicecatalog.ProvisioningArtifactProperties{}

	if v, ok := tfMap["description"].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap["disable_template_validation"].(bool); ok {
		apiObject.DisableTemplateValidation = aws.Bool(v)
	}

	info := make(map[string]*string)

	// schema will enforce that one of these is present
	if v, ok := tfMap["template_physical_id"].(string); ok && v != "" {
		info["ImportFromPhysicalId"] = aws.String(v)
	}

	if v, ok := tfMap["template_url"].(string); ok && v != "" {
		info["LoadTemplateFromURL"] = aws.String(v)
	}

	apiObject.Info = info

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}
