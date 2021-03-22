package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsConfigConformancePack() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsConfigConformancePackPut,
		Read:   resourceAwsConfigConformancePackRead,
		Update: resourceAwsConfigConformancePackPut,
		Delete: resourceAwsConfigConformancePackDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"delivery_s3_bucket": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
				),
			},
			"delivery_s3_key_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"input_parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 60,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"parameter_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"parameter_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 256),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]+$`), "must contain only alphanumeric and hyphen characters")),
			},
			"template_body": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJsonOrYamlDiffs,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 51200),
					validateStringIsJsonOrYaml,
				),
				AtLeastOneOf: []string{"template_body", "template_s3_uri"},
			},
			"template_s3_uri": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1024),
					validation.StringMatch(regexp.MustCompile(`^s3://`), "must begin with s3://"),
				),
				AtLeastOneOf: []string{"template_s3_uri", "template_body"},
			},
		},
	}
}

func resourceAwsConfigConformancePackPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	name := d.Get("name").(string)

	input := configservice.PutConformancePackInput{
		ConformancePackName: aws.String(name),
	}

	if v, ok := d.GetOk("delivery_s3_bucket"); ok {
		input.DeliveryS3Bucket = aws.String(v.(string))
	}

	if v, ok := d.GetOk("delivery_s3_key_prefix"); ok {
		input.DeliveryS3KeyPrefix = aws.String(v.(string))
	}

	if v, ok := d.GetOk("input_parameter"); ok {
		input.ConformancePackInputParameters = expandConfigConformancePackInputParameters(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("template_body"); ok {
		input.TemplateBody = aws.String(v.(string))
	}

	if v, ok := d.GetOk("template_s3_uri"); ok {
		input.TemplateS3Uri = aws.String(v.(string))
	}

	_, err := conn.PutConformancePack(&input)
	if err != nil {
		return fmt.Errorf("error creating Config Conformance Pack (%s): %w", name, err)
	}

	d.SetId(name)

	if err := configWaitForConformancePackStateCreateComplete(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Config Conformance Pack (%s) to be created: %w", d.Id(), err)
	}

	return resourceAwsConfigConformancePackRead(d, meta)
}

func resourceAwsConfigConformancePackRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	pack, err := configDescribeConformancePack(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConformancePackException) {
		log.Printf("[WARN] Config Conformance Pack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Config Conformance Pack (%s): %w", d.Id(), err)
	}

	if pack == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error describing Config Conformance Pack (%s): not found", d.Id())
		}

		log.Printf("[WARN] Config Conformance Pack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", pack.ConformancePackArn)
	d.Set("delivery_s3_bucket", pack.DeliveryS3Bucket)
	d.Set("delivery_s3_key_prefix", pack.DeliveryS3KeyPrefix)
	d.Set("name", pack.ConformancePackName)

	if err = d.Set("input_parameter", flattenConfigConformancePackInputParameters(pack.ConformancePackInputParameters)); err != nil {
		return fmt.Errorf("error setting input_parameter: %w", err)
	}

	return nil
}

func resourceAwsConfigConformancePackDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	input := &configservice.DeleteConformancePackInput{
		ConformancePackName: aws.String(d.Id()),
	}

	err := resource.Retry(ConfigConformancePackDeleteTimeout, func() *resource.RetryError {
		_, err := conn.DeleteConformancePack(input)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, configservice.ErrCodeResourceInUseException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteConformancePack(input)
	}

	if err != nil {
		if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConformancePackException) {
			return nil
		}

		return fmt.Errorf("erorr deleting Config Conformance Pack (%s): %w", d.Id(), err)
	}

	if err := configWaitForConformancePackStateDeleteComplete(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Config Conformance Pack (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func expandConfigConformancePackInputParameters(l []interface{}) []*configservice.ConformancePackInputParameter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	params := make([]*configservice.ConformancePackInputParameter, 0, len(l))

	for _, v := range l {
		tfMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		param := &configservice.ConformancePackInputParameter{}

		if name, ok := tfMap["parameter_name"].(string); ok && name != "" {
			param.ParameterName = aws.String(name)
		}

		if value, ok := tfMap["parameter_value"].(string); ok && value != "" {
			param.ParameterValue = aws.String(value)
		}

		params = append(params, param)
	}

	return params
}

func flattenConfigConformancePackInputParameters(parameters []*configservice.ConformancePackInputParameter) []interface{} {
	if parameters == nil {
		return nil
	}

	params := make([]interface{}, 0, len(parameters))

	for _, p := range parameters {
		if p == nil {
			continue
		}

		param := map[string]interface{}{
			"parameter_name":  aws.StringValue(p.ParameterName),
			"parameter_value": aws.StringValue(p.ParameterValue),
		}

		params = append(params, param)
	}

	return params
}
