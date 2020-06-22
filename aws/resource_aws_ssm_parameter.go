package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsSsmParameter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsmParameterPut,
		Read:   resourceAwsSsmParameterRead,
		Update: resourceAwsSsmParameterPut,
		Delete: resourceAwsSsmParameterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tier": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ssm.ParameterTierStandard,
				ValidateFunc: validation.StringInSlice([]string{
					ssm.ParameterTierStandard,
					ssm.ParameterTierAdvanced,
				}, false),
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					ssm.ParameterTypeString,
					ssm.ParameterTypeStringList,
					ssm.ParameterTypeSecureString,
				}, false),
			},
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"overwrite": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"allowed_pattern": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags": tagsSchema(),
		},

		CustomizeDiff: customdiff.All(
			// Prevent the following error during tier update from Advanced to Standard:
			// ValidationException: This parameter uses the advanced-parameter tier. You can't downgrade a parameter from the advanced-parameter tier to the standard-parameter tier. If necessary, you can delete the advanced parameter and recreate it as a standard parameter.
			customdiff.ForceNewIfChange("tier", func(old, new, meta interface{}) bool {
				return old.(string) == ssm.ParameterTierAdvanced && new.(string) == ssm.ParameterTierStandard
			}),
		),
	}
}

func resourceAwsSsmParameterRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading SSM Parameter: %s", d.Id())

	resp, err := ssmconn.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(d.Id()),
		WithDecryption: aws.Bool(true),
	})
	if isAWSErr(err, ssm.ErrCodeParameterNotFound, "") {
		log.Printf("[WARN] SSM Parameter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading SSM Parameter (%s): %w", d.Id(), err)
	}

	param := resp.Parameter
	name := *param.Name
	d.Set("name", name)
	d.Set("type", param.Type)
	d.Set("value", param.Value)
	d.Set("version", param.Version)

	describeParamsInput := &ssm.DescribeParametersInput{
		ParameterFilters: []*ssm.ParameterStringFilter{
			{
				Key:    aws.String("Name"),
				Option: aws.String("Equals"),
				Values: []*string{aws.String(name)},
			},
		},
	}
	describeResp, err := ssmconn.DescribeParameters(describeParamsInput)
	if err != nil {
		return fmt.Errorf("error describing SSM parameter: %s", err)
	}

	if describeResp == nil || len(describeResp.Parameters) == 0 || describeResp.Parameters[0] == nil {
		log.Printf("[WARN] SSM Parameter %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	detail := describeResp.Parameters[0]
	d.Set("key_id", detail.KeyId)
	d.Set("description", detail.Description)
	d.Set("tier", ssm.ParameterTierStandard)
	if detail.Tier != nil {
		d.Set("tier", detail.Tier)
	}
	d.Set("allowed_pattern", detail.AllowedPattern)

	tags, err := keyvaluetags.SsmListTags(ssmconn, name, ssm.ResourceTypeForTaggingParameter)

	if err != nil {
		return fmt.Errorf("error listing tags for SSM Parameter (%s): %s", name, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "ssm",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("parameter/%s", strings.TrimPrefix(d.Id(), "/")),
	}
	d.Set("arn", arn.String())

	return nil
}

func resourceAwsSsmParameterDelete(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Deleting SSM Parameter: %s", d.Id())

	_, err := ssmconn.DeleteParameter(&ssm.DeleteParameterInput{
		Name: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return fmt.Errorf("error deleting SSM Parameter (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsSsmParameterPut(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Creating SSM Parameter: %s", d.Get("name").(string))

	paramInput := &ssm.PutParameterInput{
		Name:           aws.String(d.Get("name").(string)),
		Type:           aws.String(d.Get("type").(string)),
		Tier:           aws.String(d.Get("tier").(string)),
		Value:          aws.String(d.Get("value").(string)),
		Overwrite:      aws.Bool(shouldUpdateSsmParameter(d)),
		AllowedPattern: aws.String(d.Get("allowed_pattern").(string)),
	}

	if d.HasChange("description") {
		_, n := d.GetChange("description")
		paramInput.Description = aws.String(n.(string))
	}

	if keyID, ok := d.GetOk("key_id"); ok && d.Get("type").(string) == ssm.ParameterTypeSecureString {
		paramInput.SetKeyId(keyID.(string))
	}

	log.Printf("[DEBUG] Waiting for SSM Parameter %v to be updated", d.Get("name"))
	_, err := ssmconn.PutParameter(paramInput)

	if isAWSErr(err, "ValidationException", "Tier is not supported") {
		paramInput.Tier = nil
		_, err = ssmconn.PutParameter(paramInput)
	}

	if err != nil {
		return fmt.Errorf("error creating SSM parameter: %s", err)
	}

	name := d.Get("name").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SsmUpdateTags(ssmconn, name, ssm.ResourceTypeForTaggingParameter, o, n); err != nil {
			return fmt.Errorf("error updating SSM Parameter (%s) tags: %s", name, err)
		}
	}

	d.SetId(d.Get("name").(string))

	return resourceAwsSsmParameterRead(d, meta)
}

func shouldUpdateSsmParameter(d *schema.ResourceData) bool {
	// If the user has specified a preference, return their preference
	if value, ok := d.GetOkExists("overwrite"); ok {
		return value.(bool)
	}

	// Since the user has not specified a preference, obey lifecycle rules
	// if it is not a new resource, otherwise overwrite should be set to false.
	return !d.IsNewResource()
}
