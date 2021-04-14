package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEcrRepository() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEcrRepositoryRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"kms_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"image_scanning_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"scan_on_push": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"image_tag_mutability": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"repository_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsEcrRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	params := &ecr.DescribeRepositoriesInput{
		RepositoryNames: aws.StringSlice([]string{name}),
	}

	if v, ok := d.GetOk("registry_id"); ok {
		params.RegistryId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Reading ECR repository: %#v", params)
	out, err := conn.DescribeRepositories(params)
	if err != nil {
		if isAWSErr(err, ecr.ErrCodeRepositoryNotFoundException, "") {
			return fmt.Errorf("ECR Repository (%s) not found", name)
		}
		return fmt.Errorf("error reading ECR repository: %w", err)
	}

	repository := out.Repositories[0]
	arn := aws.StringValue(repository.RepositoryArn)

	d.SetId(aws.StringValue(repository.RepositoryName))
	d.Set("arn", arn)
	d.Set("name", repository.RepositoryName)
	d.Set("registry_id", repository.RegistryId)
	d.Set("repository_url", repository.RepositoryUri)
	d.Set("image_tag_mutability", repository.ImageTagMutability)

	tags, err := keyvaluetags.EcrListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for ECR Repository (%s): %w", arn, err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags for ECR Repository (%s): %w", arn, err)
	}

	if err := d.Set("image_scanning_configuration", flattenImageScanningConfiguration(repository.ImageScanningConfiguration)); err != nil {
		return fmt.Errorf("error setting image_scanning_configuration for ECR Repository (%s): %w", arn, err)
	}

	if err := d.Set("encryption_configuration", flattenEcrRepositoryEncryptionConfiguration(repository.EncryptionConfiguration)); err != nil {
		return fmt.Errorf("error setting encryption_configuration for ECR Repository (%s): %w", arn, err)
	}

	return nil
}
