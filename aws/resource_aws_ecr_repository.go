package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEcrRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEcrRepositoryCreate,
		Read:   resourceAwsEcrRepositoryRead,
		Update: resourceAwsEcrRepositoryUpdate,
		Delete: resourceAwsEcrRepositoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"image_tag_mutability": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ecr.ImageTagMutabilityMutable,
				ValidateFunc: validation.StringInSlice([]string{
					ecr.ImageTagMutabilityMutable,
					ecr.ImageTagMutabilityImmutable,
				}, false),
			},
			"image_scanning_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"scan_on_push": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEcrRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	input := ecr.CreateRepositoryInput{
		ImageTagMutability: aws.String(d.Get("image_tag_mutability").(string)),
		RepositoryName:     aws.String(d.Get("name").(string)),
		Tags:               keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().EcrTags(),
	}

	imageScanningConfigs := d.Get("image_scanning_configuration").([]interface{})
	if len(imageScanningConfigs) > 0 {
		imageScanningConfig := imageScanningConfigs[0]
		if imageScanningConfig != nil {
			configMap := imageScanningConfig.(map[string]interface{})
			input.ImageScanningConfiguration = &ecr.ImageScanningConfiguration{
				ScanOnPush: aws.Bool(configMap["scan_on_push"].(bool)),
			}
		}
	}

	log.Printf("[DEBUG] Creating ECR repository: %#v", input)
	out, err := conn.CreateRepository(&input)
	if err != nil {
		return fmt.Errorf("error creating ECR repository: %s", err)
	}

	repository := *out.Repository

	log.Printf("[DEBUG] ECR repository created: %q", *repository.RepositoryArn)

	d.SetId(aws.StringValue(repository.RepositoryName))

	return resourceAwsEcrRepositoryRead(d, meta)
}

func resourceAwsEcrRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading ECR repository %s", d.Id())
	var out *ecr.DescribeRepositoriesOutput
	input := &ecr.DescribeRepositoriesInput{
		RepositoryNames: aws.StringSlice([]string{d.Id()}),
	}

	var err error
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		out, err = conn.DescribeRepositories(input)
		if d.IsNewResource() && isAWSErr(err, ecr.ErrCodeRepositoryNotFoundException, "") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		out, err = conn.DescribeRepositories(input)
	}

	if isAWSErr(err, ecr.ErrCodeRepositoryNotFoundException, "") {
		log.Printf("[WARN] ECR Repository (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading ECR repository: %s", err)
	}

	repository := out.Repositories[0]
	arn := aws.StringValue(repository.RepositoryArn)

	d.Set("arn", arn)
	d.Set("name", repository.RepositoryName)
	d.Set("registry_id", repository.RegistryId)
	d.Set("repository_url", repository.RepositoryUri)
	d.Set("image_tag_mutability", repository.ImageTagMutability)

	if err := d.Set("image_scanning_configuration", flattenImageScanningConfiguration(repository.ImageScanningConfiguration)); err != nil {
		return fmt.Errorf("error setting image_scanning_configuration: %s", err)
	}

	tags, err := keyvaluetags.EcrListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for ECR Repository (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func flattenImageScanningConfiguration(isc *ecr.ImageScanningConfiguration) []map[string]interface{} {
	if isc == nil {
		return nil
	}

	config := make(map[string]interface{})
	config["scan_on_push"] = aws.BoolValue(isc.ScanOnPush)

	return []map[string]interface{}{
		config,
	}
}

func resourceAwsEcrRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	arn := d.Get("arn").(string)
	conn := meta.(*AWSClient).ecrconn

	if d.HasChange("image_tag_mutability") {
		if err := resourceAwsEcrRepositoryUpdateImageTagMutability(conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("image_scanning_configuration") {
		if err := resourceAwsEcrRepositoryUpdateImageScanningConfiguration(conn, d); err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.EcrUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating ECR Repository (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsEcrRepositoryRead(d, meta)
}

func resourceAwsEcrRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	_, err := conn.DeleteRepository(&ecr.DeleteRepositoryInput{
		RepositoryName: aws.String(d.Id()),
		RegistryId:     aws.String(d.Get("registry_id").(string)),
		Force:          aws.Bool(true),
	})
	if err != nil {
		if isAWSErr(err, ecr.ErrCodeRepositoryNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting ECR repository: %s", err)
	}

	log.Printf("[DEBUG] Waiting for ECR Repository %q to be deleted", d.Id())
	input := &ecr.DescribeRepositoriesInput{
		RepositoryNames: aws.StringSlice([]string{d.Id()}),
	}
	err = resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err = conn.DescribeRepositories(input)
		if err != nil {
			if isAWSErr(err, ecr.ErrCodeRepositoryNotFoundException, "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		return resource.RetryableError(fmt.Errorf("%q: Timeout while waiting for the ECR Repository to be deleted", d.Id()))
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DescribeRepositories(input)
	}

	if isAWSErr(err, ecr.ErrCodeRepositoryNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting ECR repository: %s", err)
	}

	log.Printf("[DEBUG] repository %q deleted.", d.Get("name").(string))

	return nil
}

func resourceAwsEcrRepositoryUpdateImageTagMutability(conn *ecr.ECR, d *schema.ResourceData) error {
	input := &ecr.PutImageTagMutabilityInput{
		ImageTagMutability: aws.String(d.Get("image_tag_mutability").(string)),
		RepositoryName:     aws.String(d.Id()),
		RegistryId:         aws.String(d.Get("registry_id").(string)),
	}

	_, err := conn.PutImageTagMutability(input)
	if err != nil {
		return fmt.Errorf("Error setting image tag mutability: %s", err.Error())
	}

	return nil
}
func resourceAwsEcrRepositoryUpdateImageScanningConfiguration(conn *ecr.ECR, d *schema.ResourceData) error {

	var ecrImageScanningConfig ecr.ImageScanningConfiguration
	imageScanningConfigs := d.Get("image_scanning_configuration").([]interface{})
	if len(imageScanningConfigs) > 0 {
		imageScanningConfig := imageScanningConfigs[0]
		if imageScanningConfig != nil {
			configMap := imageScanningConfig.(map[string]interface{})
			ecrImageScanningConfig.ScanOnPush = aws.Bool(configMap["scan_on_push"].(bool))
		}
	}

	input := &ecr.PutImageScanningConfigurationInput{
		ImageScanningConfiguration: &ecrImageScanningConfig,
		RepositoryName:             aws.String(d.Id()),
		RegistryId:                 aws.String(d.Get("registry_id").(string)),
	}

	_, err := conn.PutImageScanningConfiguration(input)
	if err != nil {
		return fmt.Errorf("Error setting image scanning configuration: %s", err.Error())
	}

	return nil
}
