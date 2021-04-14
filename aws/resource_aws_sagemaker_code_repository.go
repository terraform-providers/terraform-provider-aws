package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

func resourceAwsSagemakerCodeRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerCodeRepositoryCreate,
		Read:   resourceAwsSagemakerCodeRepositoryRead,
		Update: resourceAwsSagemakerCodeRepositoryUpdate,
		Delete: resourceAwsSagemakerCodeRepositoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"code_repository_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"git_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repository_url": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"branch": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"secret_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
		},
	}
}

func resourceAwsSagemakerCodeRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	name := d.Get("code_repository_name").(string)

	input := &sagemaker.CreateCodeRepositoryInput{
		CodeRepositoryName: aws.String(name),
		GitConfig:          expandSagemakerCodeRepositoryGitConfig(d.Get("git_config").([]interface{})),
	}

	log.Printf("[DEBUG] sagemaker code repository create config: %#v", *input)
	_, err := conn.CreateCodeRepository(input)
	if err != nil {
		return fmt.Errorf("error creating SageMaker code repository: %w", err)
	}

	d.SetId(name)

	return resourceAwsSagemakerCodeRepositoryRead(d, meta)
}

func resourceAwsSagemakerCodeRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	codeRepository, err := finder.CodeRepositoryByName(conn, d.Id())
	if err != nil {
		if isAWSErr(err, "ValidationException", "Cannot find CodeRepository") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker code repository (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker code repository (%s): %w", d.Id(), err)

	}

	d.Set("code_repository_name", codeRepository.CodeRepositoryName)
	d.Set("arn", codeRepository.CodeRepositoryArn)

	if err := d.Set("git_config", flattenSagemakerCodeRepositoryGitConfig(codeRepository.GitConfig)); err != nil {
		return fmt.Errorf("error setting git_config for sagemaker code repository (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsSagemakerCodeRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	input := &sagemaker.UpdateCodeRepositoryInput{
		CodeRepositoryName: aws.String(d.Id()),
		GitConfig:          expandSagemakerCodeRepositoryUpdateGitConfig(d.Get("git_config").([]interface{})),
	}

	log.Printf("[DEBUG] sagemaker code repository update config: %#v", *input)
	_, err := conn.UpdateCodeRepository(input)
	if err != nil {
		return fmt.Errorf("error updating SageMaker code repository: %w", err)
	}

	return resourceAwsSagemakerCodeRepositoryRead(d, meta)
}

func resourceAwsSagemakerCodeRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	input := &sagemaker.DeleteCodeRepositoryInput{
		CodeRepositoryName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteCodeRepository(input); err != nil {
		if isAWSErr(err, "ValidationException", "Cannot find CodeRepository") {
			return nil
		}
		return fmt.Errorf("error deleting SageMaker code repository (%s): %w", d.Id(), err)
	}

	return nil
}

func expandSagemakerCodeRepositoryGitConfig(l []interface{}) *sagemaker.GitConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.GitConfig{
		RepositoryUrl: aws.String(m["repository_url"].(string)),
	}

	if v, ok := m["branch"].(string); ok && v != "" {
		config.Branch = aws.String(v)
	}

	if v, ok := m["secret_arn"].(string); ok && v != "" {
		config.SecretArn = aws.String(v)
	}

	return config
}

func flattenSagemakerCodeRepositoryGitConfig(config *sagemaker.GitConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"repository_url": aws.StringValue(config.RepositoryUrl),
	}

	if config.Branch != nil {
		m["branch"] = aws.StringValue(config.Branch)
	}

	if config.SecretArn != nil {
		m["secret_arn"] = aws.StringValue(config.SecretArn)
	}

	return []map[string]interface{}{m}
}

func expandSagemakerCodeRepositoryUpdateGitConfig(l []interface{}) *sagemaker.GitConfigForUpdate {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.GitConfigForUpdate{
		SecretArn: aws.String(m["secret_arn"].(string)),
	}

	return config
}
