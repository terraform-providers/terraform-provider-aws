package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/waiter"
)

func resourceAwsSagemakerDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerDomainCreate,
		Read:   resourceAwsSagemakerDomainRead,
		Update: resourceAwsSagemakerDomainUpdate,
		Delete: resourceAwsSagemakerDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"auth_mode": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringInSlice(sagemaker.AuthMode_Values(), false),
			},
			"vpc_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MaxItems: 16,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"app_network_access_type": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Default:      sagemaker.AppNetworkAccessTypePublicInternetOnly,
				ValidateFunc: validation.StringInSlice(sagemaker.AppNetworkAccessType_Values(), false),
			},
			"default_user_settings": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 5,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"execution_role": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"sharing_settings": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"notebook_output_option": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      sagemaker.NotebookOutputOptionDisabled,
										ValidateFunc: validation.StringInSlice(sagemaker.NotebookOutputOption_Values(), false),
									},
									"s3_kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateArn,
									},
									"s3_output_path": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"tensor_board_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_resource_spec": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateArn,
												},
											},
										},
									},
								},
							},
						},
						"jupyter_server_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_resource_spec": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateArn,
												},
											},
										},
									},
								},
							},
						},
						"kernel_gateway_app_settings": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_resource_spec": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
												},
												"sagemaker_image_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateArn,
												},
											},
										},
									},
									"custom_image": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 30,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"app_image_config_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"image_version_number": {
													Type:     schema.TypeInt,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"tags": tagsSchema(),
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"single_sign_on_managed_application_instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"home_efs_file_system_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsSagemakerDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	input := &sagemaker.CreateDomainInput{
		DomainName:           aws.String(d.Get("domain_name").(string)),
		AuthMode:             aws.String(d.Get("auth_mode").(string)),
		VpcId:                aws.String(d.Get("vpc_id").(string)),
		AppNetworkAccessType: aws.String(d.Get("app_network_access_type").(string)),
		SubnetIds:            expandStringSet(d.Get("subnet_ids").(*schema.Set)),
		DefaultUserSettings:  expandSagemakerDomainDefaultUserSettings(d.Get("default_user_settings").([]interface{})),
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SagemakerTags()
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] sagemaker domain create config: %#v", *input)
	output, err := conn.CreateDomain(input)
	if err != nil {
		return fmt.Errorf("error creating SageMaker domain: %w", err)
	}

	domainArn := aws.StringValue(output.DomainArn)
	domainID, err := decodeSagemakerDomainID(domainArn)
	if err != nil {
		return err
	}

	d.SetId(domainID)

	if _, err := waiter.DomainInService(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for SageMaker domain (%s) to create: %w", d.Id(), err)
	}

	return resourceAwsSagemakerDomainRead(d, meta)
}

func resourceAwsSagemakerDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	domain, err := finder.DomainByName(conn, d.Id())
	if err != nil {
		if isAWSErr(err, sagemaker.ErrCodeResourceNotFound, "") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker domain (%s), removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker domain (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(domain.DomainArn)
	d.Set("domain_name", domain.DomainName)
	d.Set("auth_mode", domain.AuthMode)
	d.Set("app_network_access_type", domain.AppNetworkAccessType)
	d.Set("arn", arn)
	d.Set("home_efs_file_system_id", domain.HomeEfsFileSystemId)
	d.Set("single_sign_on_managed_application_instance_id", domain.SingleSignOnManagedApplicationInstanceId)
	d.Set("url", domain.Url)
	d.Set("vpc_id", domain.VpcId)
	d.Set("kms_key_id", domain.KmsKeyId)

	if err := d.Set("subnet_ids", flattenStringSet(domain.SubnetIds)); err != nil {
		return fmt.Errorf("error setting subnet_ids for SageMaker domain (%s): %w", d.Id(), err)
	}

	if err := d.Set("default_user_settings", flattenSagemakerDomainDefaultUserSettings(domain.DefaultUserSettings)); err != nil {
		return fmt.Errorf("error setting default_user_settings for SageMaker domain (%s): %w", d.Id(), err)
	}

	tags, err := keyvaluetags.SagemakerListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker Domain (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsSagemakerDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	if d.HasChange("default_user_settings") {
		input := &sagemaker.UpdateDomainInput{
			DomainId:            aws.String(d.Id()),
			DefaultUserSettings: expandSagemakerDomainDefaultUserSettings(d.Get("default_user_settings").([]interface{})),
		}

		log.Printf("[DEBUG] sagemaker domain update config: %#v", *input)
		_, err := conn.UpdateDomain(input)
		if err != nil {
			return fmt.Errorf("error updating SageMaker domain: %w", err)
		}

		if _, err := waiter.DomainInService(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for SageMaker domain (%s) to update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SagemakerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker domain (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsSagemakerDomainRead(d, meta)
}

func resourceAwsSagemakerDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	input := &sagemaker.DeleteDomainInput{
		DomainId: aws.String(d.Id()),
		RetentionPolicy: &sagemaker.RetentionPolicy{
			HomeEfsFileSystem: aws.String(sagemaker.RetentionTypeDelete),
		},
	}

	if _, err := conn.DeleteDomain(input); err != nil {
		if !isAWSErr(err, sagemaker.ErrCodeResourceNotFound, "") {
			return fmt.Errorf("error deleting SageMaker domain (%s): %w", d.Id(), err)
		}
	}

	if _, err := waiter.DomainDeleted(conn, d.Id()); err != nil {
		if !isAWSErr(err, sagemaker.ErrCodeResourceNotFound, "") {
			return fmt.Errorf("error waiting for SageMaker domain (%s) to delete: %w", d.Id(), err)
		}
	}

	return nil
}

func expandSagemakerDomainDefaultUserSettings(l []interface{}) *sagemaker.UserSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.UserSettings{}

	if v, ok := m["execution_role"].(string); ok && v != "" {
		config.ExecutionRole = aws.String(v)
	}

	if v, ok := m["security_groups"].(*schema.Set); ok && v.Len() > 0 {
		config.SecurityGroups = expandStringSet(v)
	}

	if v, ok := m["tensor_board_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.TensorBoardAppSettings = expandSagemakerDomainTensorBoardAppSettings(v)
	}

	if v, ok := m["kernel_gateway_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.KernelGatewayAppSettings = expandSagemakerDomainKernelGatewayAppSettings(v)
	}

	if v, ok := m["jupyter_server_app_settings"].([]interface{}); ok && len(v) > 0 {
		config.JupyterServerAppSettings = expandSagemakerDomainJupyterServerAppSettings(v)
	}

	if v, ok := m["sharing_settings"].([]interface{}); ok && len(v) > 0 {
		config.SharingSettings = expandSagemakerDomainShareSettings(v)
	}

	return config
}

func expandSagemakerDomainJupyterServerAppSettings(l []interface{}) *sagemaker.JupyterServerAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.JupyterServerAppSettings{}

	if v, ok := m["default_resource_spec"].([]interface{}); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandSagemakerDomainDefaultResourceSpec(v)
	}

	return config
}

func expandSagemakerDomainKernelGatewayAppSettings(l []interface{}) *sagemaker.KernelGatewayAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.KernelGatewayAppSettings{}

	if v, ok := m["default_resource_spec"].([]interface{}); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandSagemakerDomainDefaultResourceSpec(v)
	}

	if v, ok := m["custom_image"].([]interface{}); ok && len(v) > 0 {
		config.CustomImages = expandSagemakerDomainCustomImages(v)
	}

	return config
}

func expandSagemakerDomainTensorBoardAppSettings(l []interface{}) *sagemaker.TensorBoardAppSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.TensorBoardAppSettings{}

	if v, ok := m["default_resource_spec"].([]interface{}); ok && len(v) > 0 {
		config.DefaultResourceSpec = expandSagemakerDomainDefaultResourceSpec(v)
	}

	return config
}

func expandSagemakerDomainDefaultResourceSpec(l []interface{}) *sagemaker.ResourceSpec {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.ResourceSpec{}

	if v, ok := m["instance_type"].(string); ok && v != "" {
		config.InstanceType = aws.String(v)
	}

	if v, ok := m["sagemaker_image_arn"].(string); ok && v != "" {
		config.SageMakerImageArn = aws.String(v)
	}

	return config
}

func expandSagemakerDomainShareSettings(l []interface{}) *sagemaker.SharingSettings {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.SharingSettings{
		NotebookOutputOption: aws.String(m["notebook_output_option"].(string)),
	}

	if v, ok := m["s3_kms_key_id"].(string); ok && v != "" {
		config.S3KmsKeyId = aws.String(v)
	}

	if v, ok := m["s3_output_path"].(string); ok && v != "" {
		config.S3OutputPath = aws.String(v)
	}

	return config
}

func expandSagemakerDomainCustomImages(l []interface{}) []*sagemaker.CustomImage {
	images := make([]*sagemaker.CustomImage, 0, len(l))

	for _, eRaw := range l {
		data := eRaw.(map[string]interface{})

		image := &sagemaker.CustomImage{
			AppImageConfigName: aws.String(data["app_image_config_name"].(string)),
			ImageName:          aws.String(data["image_name"].(string)),
		}

		if v, ok := data["image_version_number"].(int); ok {
			image.ImageVersionNumber = aws.Int64(int64(v))
		}

		images = append(images, image)
	}

	return images
}

func flattenSagemakerDomainDefaultUserSettings(config *sagemaker.UserSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.ExecutionRole != nil {
		m["execution_role"] = aws.StringValue(config.ExecutionRole)
	}

	if config.SecurityGroups != nil {
		m["security_groups"] = flattenStringSet(config.SecurityGroups)
	}

	if config.JupyterServerAppSettings != nil {
		m["jupyter_server_app_settings"] = flattenSagemakerDomainJupyterServerAppSettings(config.JupyterServerAppSettings)
	}

	if config.KernelGatewayAppSettings != nil {
		m["kernel_gateway_app_settings"] = flattenSagemakerDomainKernelGatewayAppSettings(config.KernelGatewayAppSettings)
	}

	if config.TensorBoardAppSettings != nil {
		m["tensor_board_app_settings"] = flattenSagemakerDomainTensorBoardAppSettings(config.TensorBoardAppSettings)
	}

	if config.SharingSettings != nil {
		m["sharing_settings"] = flattenSagemakerDomainShareSettings(config.SharingSettings)
	}

	return []map[string]interface{}{m}
}

func flattenSagemakerDomainDefaultResourceSpec(config *sagemaker.ResourceSpec) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.InstanceType != nil {
		m["instance_type"] = aws.StringValue(config.InstanceType)
	}

	if config.SageMakerImageArn != nil {
		m["sagemaker_image_arn"] = aws.StringValue(config.SageMakerImageArn)
	}

	return []map[string]interface{}{m}
}

func flattenSagemakerDomainTensorBoardAppSettings(config *sagemaker.TensorBoardAppSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenSagemakerDomainDefaultResourceSpec(config.DefaultResourceSpec)
	}

	return []map[string]interface{}{m}
}

func flattenSagemakerDomainJupyterServerAppSettings(config *sagemaker.JupyterServerAppSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenSagemakerDomainDefaultResourceSpec(config.DefaultResourceSpec)
	}

	return []map[string]interface{}{m}
}

func flattenSagemakerDomainKernelGatewayAppSettings(config *sagemaker.KernelGatewayAppSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.DefaultResourceSpec != nil {
		m["default_resource_spec"] = flattenSagemakerDomainDefaultResourceSpec(config.DefaultResourceSpec)
	}

	if config.CustomImages != nil {
		m["custom_image"] = flattenSagemakerDomainCustomImages(config.CustomImages)
	}

	return []map[string]interface{}{m}
}

func flattenSagemakerDomainShareSettings(config *sagemaker.SharingSettings) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"notebook_output_option": aws.StringValue(config.NotebookOutputOption),
	}

	if config.S3KmsKeyId != nil {
		m["s3_kms_key_id"] = aws.StringValue(config.S3KmsKeyId)
	}

	if config.S3OutputPath != nil {
		m["s3_output_path"] = aws.StringValue(config.S3OutputPath)
	}

	return []map[string]interface{}{m}
}

func flattenSagemakerDomainCustomImages(config []*sagemaker.CustomImage) []map[string]interface{} {
	images := make([]map[string]interface{}, 0, len(config))

	for _, raw := range config {
		image := make(map[string]interface{})

		image["app_image_config_name"] = aws.StringValue(raw.AppImageConfigName)
		image["image_name"] = aws.StringValue(raw.ImageName)

		if raw.ImageVersionNumber != nil {
			image["image_version_number"] = aws.Int64Value(raw.ImageVersionNumber)
		}

		images = append(images, image)
	}

	return images
}

func decodeSagemakerDomainID(id string) (string, error) {
	domainArn, err := arn.Parse(id)
	if err != nil {
		return "", err
	}

	domainName := strings.TrimPrefix(domainArn.Resource, "domain/")
	return domainName, nil
}
