package aws

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsCodeBuildProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeBuildProjectCreate,
		Read:   resourceAwsCodeBuildProjectRead,
		Update: resourceAwsCodeBuildProjectUpdate,
		Delete: resourceAwsCodeBuildProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"artifacts": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"artifact_identifier": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if old == d.Get("name") && new == "" {
									return true
								}
								return false
							},
						},
						"encryption_disabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"location": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"namespace_type": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if d.Get("artifacts.0.type") == codebuild.ArtifactsTypeS3 {
									return old == codebuild.ArtifactNamespaceNone && new == ""
								}
								return false
							},
							ValidateFunc: validation.StringInSlice([]string{
								codebuild.ArtifactNamespaceNone,
								codebuild.ArtifactNamespaceBuildId,
							}, false),
						},
						"packaging": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								switch d.Get("artifacts.0.type") {
								case codebuild.ArtifactsTypeCodepipeline:
									return new == ""
								case codebuild.ArtifactsTypeS3:
									return old == codebuild.ArtifactPackagingNone && new == ""
								}
								return false
							},
							ValidateFunc: validation.StringInSlice([]string{
								codebuild.ArtifactPackagingNone,
								codebuild.ArtifactPackagingZip,
							}, false),
						},
						"path": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								codebuild.ArtifactsTypeCodepipeline,
								codebuild.ArtifactsTypeS3,
								codebuild.ArtifactsTypeNoArtifacts,
							}, false),
						},
						"override_artifact_name": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"cache": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "1" && new == "0" {
						return true
					}
					return false
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  codebuild.CacheTypeNoCache,
							ValidateFunc: validation.StringInSlice([]string{
								codebuild.CacheTypeNoCache,
								codebuild.CacheTypeS3,
								codebuild.CacheTypeLocal,
							}, false),
						},
						"location": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"modes": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"encryption_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"environment": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compute_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								codebuild.ComputeTypeBuildGeneral1Small,
								codebuild.ComputeTypeBuildGeneral1Medium,
								codebuild.ComputeTypeBuildGeneral1Large,
								codebuild.ComputeTypeBuildGeneral12xlarge,
							}, false),
						},
						"environment_variable": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
									"type": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.StringInSlice([]string{
											codebuild.EnvironmentVariableTypePlaintext,
											codebuild.EnvironmentVariableTypeParameterStore,
											codebuild.EnvironmentVariableTypeSecretsManager,
										}, false),
										Default: codebuild.EnvironmentVariableTypePlaintext,
									},
								},
							},
						},
						"image": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								codebuild.EnvironmentTypeLinuxContainer,
								codebuild.EnvironmentTypeLinuxGpuContainer,
								codebuild.EnvironmentTypeWindowsContainer,
								codebuild.EnvironmentTypeArmContainer,
							}, false),
						},
						"image_pull_credentials_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  codebuild.ImagePullCredentialsTypeCodebuild,
							ValidateFunc: validation.StringInSlice([]string{
								codebuild.ImagePullCredentialsTypeCodebuild,
								codebuild.ImagePullCredentialsTypeServiceRole,
							}, false),
						},
						"privileged_mode": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"certificate": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`\.(pem|zip)$`), "must end in .pem or .zip"),
						},
						"registry_credential": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"credential": {
										Type:     schema.TypeString,
										Required: true,
									},
									"credential_provider": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											codebuild.CredentialProviderTypeSecretsManager,
										}, false),
									},
								},
							},
						},
					},
				},
			},
			"logs_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_logs": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  codebuild.LogsConfigStatusTypeEnabled,
										ValidateFunc: validation.StringInSlice([]string{
											codebuild.LogsConfigStatusTypeDisabled,
											codebuild.LogsConfigStatusTypeEnabled,
										}, false),
									},
									"group_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"stream_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
							DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
						},
						"s3_logs": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  codebuild.LogsConfigStatusTypeDisabled,
										ValidateFunc: validation.StringInSlice([]string{
											codebuild.LogsConfigStatusTypeDisabled,
											codebuild.LogsConfigStatusTypeEnabled,
										}, false),
									},
									"location": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateAwsCodeBuildProjectS3LogsLocation,
									},
									"encryption_disabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
							DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
						},
					},
				},
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsCodeBuildProjectName,
			},
			"secondary_artifacts": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      resourceAwsCodeBuildProjectArtifactsHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"encryption_disabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"location": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"namespace_type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								codebuild.ArtifactNamespaceNone,
								codebuild.ArtifactNamespaceBuildId,
							}, false),
							Default: codebuild.ArtifactNamespaceNone,
						},
						"override_artifact_name": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"packaging": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								codebuild.ArtifactPackagingNone,
								codebuild.ArtifactPackagingZip,
							}, false),
							Default: codebuild.ArtifactPackagingNone,
						},
						"path": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"artifact_identifier": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								codebuild.ArtifactsTypeS3,
							}, false),
						},
					},
				},
			},
			"secondary_sources": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource": {
										Type:      schema.TypeString,
										Sensitive: true,
										Optional:  true,
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											codebuild.SourceAuthTypeOauth,
										}, false),
									},
								},
							},
						},
						"buildspec": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"location": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								codebuild.SourceTypeCodecommit,
								codebuild.SourceTypeCodepipeline,
								codebuild.SourceTypeGithub,
								codebuild.SourceTypeS3,
								codebuild.SourceTypeBitbucket,
								codebuild.SourceTypeGithubEnterprise,
							}, false),
						},
						"git_clone_depth": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"git_submodules_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"fetch_submodules": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"insecure_ssl": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"report_build_status": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"source_identifier": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"service_role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"source": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource": {
										Type:      schema.TypeString,
										Sensitive: true,
										Optional:  true,
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											codebuild.SourceAuthTypeOauth,
										}, false),
									},
								},
							},
						},
						"buildspec": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"location": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								codebuild.SourceTypeCodecommit,
								codebuild.SourceTypeCodepipeline,
								codebuild.SourceTypeGithub,
								codebuild.SourceTypeS3,
								codebuild.SourceTypeBitbucket,
								codebuild.SourceTypeGithubEnterprise,
								codebuild.SourceTypeNoSource,
							}, false),
						},
						"git_clone_depth": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"git_submodules_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"fetch_submodules": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"insecure_ssl": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"report_build_status": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"source_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"build_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      "60",
				ValidateFunc: validation.IntBetween(5, 480),
			},
			"queued_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      "480",
				ValidateFunc: validation.IntBetween(5, 480),
			},
			"badge_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"badge_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
			"vpc_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vpc_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"subnets": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							// Set:      schema.HashString,
							MaxItems: 16,
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							// Set:      schema.HashString,
							MaxItems: 5,
						},
					},
				},
			},
		},

		CustomizeDiff: customdiff.Sequence(
			func(diff *schema.ResourceDiff, v interface{}) error {
				// Plan time validation for cache location
				cacheType, cacheTypeOk := diff.GetOk("cache.0.type")
				if !cacheTypeOk || cacheType.(string) == codebuild.CacheTypeNoCache || cacheType.(string) == codebuild.CacheTypeLocal {
					return nil
				}
				if v, ok := diff.GetOk("cache.0.location"); ok && v.(string) != "" {
					return nil
				}
				return fmt.Errorf(`cache location is required when cache type is %q`, cacheType.(string))
			},
		),
	}
}

func resourceAwsCodeBuildProjectCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	projectEnv := expandProjectEnvironment(d)
	projectSource := expandProjectSource(d)
	projectArtifacts := expandProjectArtifacts(d)
	projectSecondaryArtifacts := expandProjectSecondaryArtifacts(d)
	projectSecondarySources := expandProjectSecondarySources(d)
	projectLogsConfig := expandProjectLogsConfig(d)

	if aws.StringValue(projectSource.Type) == codebuild.SourceTypeNoSource {
		if aws.StringValue(projectSource.Buildspec) == "" {
			return fmt.Errorf("`buildspec` must be set when source's `type` is `NO_SOURCE`")
		}

		if aws.StringValue(projectSource.Location) != "" {
			return fmt.Errorf("`location` must be empty when source's `type` is `NO_SOURCE`")
		}
	}

	params := &codebuild.CreateProjectInput{
		Environment:        projectEnv,
		Name:               aws.String(d.Get("name").(string)),
		Source:             &projectSource,
		Artifacts:          &projectArtifacts,
		SecondaryArtifacts: projectSecondaryArtifacts,
		SecondarySources:   projectSecondarySources,
		LogsConfig:         projectLogsConfig,
		Tags:               keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().CodebuildTags(),
	}

	if v, ok := d.GetOk("cache"); ok {
		params.Cache = expandProjectCache(v.([]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("encryption_key"); ok {
		params.EncryptionKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("service_role"); ok {
		params.ServiceRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_version"); ok {
		params.SourceVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("build_timeout"); ok {
		params.TimeoutInMinutes = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("queued_timeout"); ok {
		params.QueuedTimeoutInMinutes = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("vpc_config"); ok {
		params.VpcConfig = expandCodeBuildVpcConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("badge_enabled"); ok {
		params.BadgeEnabled = aws.Bool(v.(bool))
	}

	var resp *codebuild.CreateProjectOutput
	// Handle IAM eventual consistency
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		var err error

		resp, err = conn.CreateProject(params)
		if err != nil {
			// InvalidInputException: CodeBuild is not authorized to perform
			// InvalidInputException: Not authorized to perform DescribeSecurityGroups
			if isAWSErr(err, codebuild.ErrCodeInvalidInputException, "ot authorized to perform") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		resp, err = conn.CreateProject(params)
	}
	if err != nil {
		return fmt.Errorf("Error creating CodeBuild project: %s", err)
	}

	d.SetId(aws.StringValue(resp.Project.Arn))

	return resourceAwsCodeBuildProjectRead(d, meta)
}

func expandProjectSecondaryArtifacts(d *schema.ResourceData) []*codebuild.ProjectArtifacts {
	artifacts := make([]*codebuild.ProjectArtifacts, 0)

	configsList := d.Get("secondary_artifacts").(*schema.Set).List()

	if len(configsList) == 0 {
		return nil
	}

	for _, config := range configsList {
		art := expandProjectArtifactData(config.(map[string]interface{}))
		artifacts = append(artifacts, &art)
	}

	return artifacts
}

func expandProjectArtifacts(d *schema.ResourceData) codebuild.ProjectArtifacts {
	configs := d.Get("artifacts").([]interface{})
	data := configs[0].(map[string]interface{})

	return expandProjectArtifactData(data)
}

func expandProjectArtifactData(data map[string]interface{}) codebuild.ProjectArtifacts {
	artifactType := data["type"].(string)

	projectArtifacts := codebuild.ProjectArtifacts{
		Type: aws.String(artifactType),
	}

	// Only valid for S3 and CODEPIPELINE artifacts types
	// InvalidInputException: Invalid artifacts: artifact type NO_ARTIFACTS should have null encryptionDisabled
	if artifactType == codebuild.ArtifactsTypeS3 || artifactType == codebuild.ArtifactsTypeCodepipeline {
		projectArtifacts.EncryptionDisabled = aws.Bool(data["encryption_disabled"].(bool))
	}

	if data["artifact_identifier"] != nil && data["artifact_identifier"].(string) != "" {
		projectArtifacts.ArtifactIdentifier = aws.String(data["artifact_identifier"].(string))
	}

	if data["location"].(string) != "" {
		projectArtifacts.Location = aws.String(data["location"].(string))
	}

	if data["name"].(string) != "" {
		projectArtifacts.Name = aws.String(data["name"].(string))
	}

	if data["namespace_type"].(string) != "" {
		projectArtifacts.NamespaceType = aws.String(data["namespace_type"].(string))
	}

	if v, ok := data["override_artifact_name"]; ok {
		projectArtifacts.OverrideArtifactName = aws.Bool(v.(bool))
	}

	if data["packaging"].(string) != "" {
		projectArtifacts.Packaging = aws.String(data["packaging"].(string))
	}

	if data["path"].(string) != "" {
		projectArtifacts.Path = aws.String(data["path"].(string))
	}

	return projectArtifacts
}

func expandProjectCache(s []interface{}) *codebuild.ProjectCache {
	var projectCache *codebuild.ProjectCache

	data := s[0].(map[string]interface{})

	projectCache = &codebuild.ProjectCache{
		Type: aws.String(data["type"].(string)),
	}

	if v, ok := data["location"]; ok {
		projectCache.Location = aws.String(v.(string))
	}

	if cacheType := data["type"]; cacheType == codebuild.CacheTypeLocal {
		if modes, modesOk := data["modes"]; modesOk {
			modesStrings := modes.([]interface{})
			projectCache.Modes = expandStringList(modesStrings)
		}
	}

	return projectCache
}

func expandProjectEnvironment(d *schema.ResourceData) *codebuild.ProjectEnvironment {
	configs := d.Get("environment").([]interface{})

	envConfig := configs[0].(map[string]interface{})

	projectEnv := &codebuild.ProjectEnvironment{
		PrivilegedMode: aws.Bool(envConfig["privileged_mode"].(bool)),
	}

	if v := envConfig["compute_type"]; v != nil {
		projectEnv.ComputeType = aws.String(v.(string))
	}

	if v := envConfig["image"]; v != nil {
		projectEnv.Image = aws.String(v.(string))
	}

	if v := envConfig["type"]; v != nil {
		projectEnv.Type = aws.String(v.(string))
	}

	if v, ok := envConfig["certificate"]; ok && v.(string) != "" {
		projectEnv.Certificate = aws.String(v.(string))
	}

	if v := envConfig["image_pull_credentials_type"]; v != nil {
		projectEnv.ImagePullCredentialsType = aws.String(v.(string))
	}

	if v, ok := envConfig["registry_credential"]; ok && len(v.([]interface{})) > 0 {
		config := v.([]interface{})[0].(map[string]interface{})

		projectRegistryCredential := &codebuild.RegistryCredential{}

		if v, ok := config["credential"]; ok && v.(string) != "" {
			projectRegistryCredential.Credential = aws.String(v.(string))
		}

		if v, ok := config["credential_provider"]; ok && v.(string) != "" {
			projectRegistryCredential.CredentialProvider = aws.String(v.(string))
		}

		projectEnv.RegistryCredential = projectRegistryCredential
	}

	if v := envConfig["environment_variable"]; v != nil {
		envVariables := v.([]interface{})
		if len(envVariables) > 0 {
			projectEnvironmentVariables := make([]*codebuild.EnvironmentVariable, 0, len(envVariables))

			for _, envVariablesConfig := range envVariables {
				config := envVariablesConfig.(map[string]interface{})

				projectEnvironmentVar := &codebuild.EnvironmentVariable{}

				if v := config["name"].(string); v != "" {
					projectEnvironmentVar.Name = &v
				}

				if v, ok := config["value"].(string); ok {
					projectEnvironmentVar.Value = &v
				}

				if v := config["type"].(string); v != "" {
					projectEnvironmentVar.Type = &v
				}

				projectEnvironmentVariables = append(projectEnvironmentVariables, projectEnvironmentVar)
			}

			projectEnv.EnvironmentVariables = projectEnvironmentVariables
		}
	}

	return projectEnv
}

func expandProjectLogsConfig(d *schema.ResourceData) *codebuild.LogsConfig {
	logsConfig := &codebuild.LogsConfig{}

	if v, ok := d.GetOk("logs_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		configList := v.([]interface{})
		data := configList[0].(map[string]interface{})

		if v, ok := data["cloudwatch_logs"]; ok {
			logsConfig.CloudWatchLogs = expandCodeBuildCloudWatchLogsConfig(v.([]interface{}))
		}

		if v, ok := data["s3_logs"]; ok {
			logsConfig.S3Logs = expandCodeBuildS3LogsConfig(v.([]interface{}))
		}
	}

	if logsConfig.CloudWatchLogs == nil {
		logsConfig.CloudWatchLogs = &codebuild.CloudWatchLogsConfig{
			Status: aws.String(codebuild.LogsConfigStatusTypeEnabled),
		}
	}

	if logsConfig.S3Logs == nil {
		logsConfig.S3Logs = &codebuild.S3LogsConfig{
			Status: aws.String(codebuild.LogsConfigStatusTypeDisabled),
		}
	}

	return logsConfig
}

func expandCodeBuildCloudWatchLogsConfig(configList []interface{}) *codebuild.CloudWatchLogsConfig {
	if len(configList) == 0 || configList[0] == nil {
		return nil
	}

	data := configList[0].(map[string]interface{})

	status := data["status"].(string)

	cloudWatchLogsConfig := &codebuild.CloudWatchLogsConfig{
		Status: aws.String(status),
	}

	if v, ok := data["group_name"]; ok {
		groupName := v.(string)
		if len(groupName) > 0 {
			cloudWatchLogsConfig.GroupName = aws.String(groupName)
		}
	}

	if v, ok := data["stream_name"]; ok {
		streamName := v.(string)
		if len(streamName) > 0 {
			cloudWatchLogsConfig.StreamName = aws.String(streamName)
		}
	}

	return cloudWatchLogsConfig
}

func expandCodeBuildS3LogsConfig(configList []interface{}) *codebuild.S3LogsConfig {
	if len(configList) == 0 || configList[0] == nil {
		return nil
	}

	data := configList[0].(map[string]interface{})

	status := data["status"].(string)

	s3LogsConfig := &codebuild.S3LogsConfig{
		Status: aws.String(status),
	}

	if v, ok := data["location"]; ok {
		location := v.(string)
		if len(location) > 0 {
			s3LogsConfig.Location = aws.String(location)
		}
	}

	s3LogsConfig.EncryptionDisabled = aws.Bool(data["encryption_disabled"].(bool))

	return s3LogsConfig
}

func expandCodeBuildVpcConfig(rawVpcConfig []interface{}) *codebuild.VpcConfig {
	vpcConfig := codebuild.VpcConfig{}
	if len(rawVpcConfig) == 0 || rawVpcConfig[0] == nil {
		return &vpcConfig
	}

	data := rawVpcConfig[0].(map[string]interface{})
	vpcConfig.VpcId = aws.String(data["vpc_id"].(string))
	vpcConfig.Subnets = expandStringList(data["subnets"].(*schema.Set).List())
	vpcConfig.SecurityGroupIds = expandStringList(data["security_group_ids"].(*schema.Set).List())

	return &vpcConfig
}

func expandProjectSecondarySources(d *schema.ResourceData) []*codebuild.ProjectSource {
	configs := d.Get("secondary_sources").(*schema.Set).List()

	if len(configs) == 0 {
		return nil
	}

	sources := make([]*codebuild.ProjectSource, 0)

	for _, config := range configs {
		source := expandProjectSourceData(config.(map[string]interface{}))
		sources = append(sources, &source)
	}

	return sources
}

func expandProjectSource(d *schema.ResourceData) codebuild.ProjectSource {
	configs := d.Get("source").([]interface{})

	data := configs[0].(map[string]interface{})
	return expandProjectSourceData(data)
}

func expandProjectSourceData(data map[string]interface{}) codebuild.ProjectSource {
	sourceType := data["type"].(string)

	projectSource := codebuild.ProjectSource{
		Buildspec:     aws.String(data["buildspec"].(string)),
		GitCloneDepth: aws.Int64(int64(data["git_clone_depth"].(int))),
		InsecureSsl:   aws.Bool(data["insecure_ssl"].(bool)),
		Type:          aws.String(sourceType),
	}

	if data["source_identifier"] != nil {
		projectSource.SourceIdentifier = aws.String(data["source_identifier"].(string))
	}

	if data["location"].(string) != "" {
		projectSource.Location = aws.String(data["location"].(string))
	}

	// Only valid for BITBUCKET, GITHUB, and GITHUB_ENTERPRISE source types, e.g.
	// InvalidInputException: Source type NO_SOURCE does not support ReportBuildStatus
	if sourceType == codebuild.SourceTypeBitbucket || sourceType == codebuild.SourceTypeGithub || sourceType == codebuild.SourceTypeGithubEnterprise {
		projectSource.ReportBuildStatus = aws.Bool(data["report_build_status"].(bool))
	}

	// Probe data for auth details (max of 1 auth per ProjectSource object)
	if v, ok := data["auth"]; ok && len(v.([]interface{})) > 0 {
		if auths := v.([]interface{}); auths[0] != nil {
			auth := auths[0].(map[string]interface{})
			projectSource.Auth = &codebuild.SourceAuth{
				Type:     aws.String(auth["type"].(string)),
				Resource: aws.String(auth["resource"].(string)),
			}
		}
	}

	// Only valid for CODECOMMIT, GITHUB, GITHUB_ENTERPRISE source types.
	if sourceType == codebuild.SourceTypeCodecommit || sourceType == codebuild.SourceTypeGithub || sourceType == codebuild.SourceTypeGithubEnterprise {
		if v, ok := data["git_submodules_config"]; ok && len(v.([]interface{})) > 0 {
			config := v.([]interface{})[0].(map[string]interface{})

			gitSubmodulesConfig := &codebuild.GitSubmodulesConfig{}

			if v, ok := config["fetch_submodules"]; ok {
				gitSubmodulesConfig.FetchSubmodules = aws.Bool(v.(bool))
			}

			projectSource.GitSubmodulesConfig = gitSubmodulesConfig
		}
	}

	return projectSource
}

func resourceAwsCodeBuildProjectRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.BatchGetProjects(&codebuild.BatchGetProjectsInput{
		Names: []*string{
			aws.String(d.Id()),
		},
	})

	if err != nil {
		return fmt.Errorf("Error retreiving Projects: %q", err)
	}

	// if nothing was found, then return no state
	if len(resp.Projects) == 0 {
		log.Printf("[INFO]: No projects were found, removing from state")
		d.SetId("")
		return nil
	}

	project := resp.Projects[0]

	if err := d.Set("artifacts", flattenAwsCodeBuildProjectArtifacts(project.Artifacts)); err != nil {
		return fmt.Errorf("error setting artifacts: %s", err)
	}

	if err := d.Set("environment", flattenAwsCodeBuildProjectEnvironment(project.Environment)); err != nil {
		return fmt.Errorf("error setting environment: %s", err)
	}

	if err := d.Set("cache", flattenAwsCodebuildProjectCache(project.Cache)); err != nil {
		return fmt.Errorf("error setting cache: %s", err)
	}

	if err := d.Set("logs_config", flattenAwsCodeBuildLogsConfig(project.LogsConfig)); err != nil {
		return fmt.Errorf("error setting logs_config: %s", err)
	}

	if err := d.Set("secondary_artifacts", flattenAwsCodeBuildProjectSecondaryArtifacts(project.SecondaryArtifacts)); err != nil {
		return fmt.Errorf("error setting secondary_artifacts: %s", err)
	}

	if err := d.Set("secondary_sources", flattenAwsCodeBuildProjectSecondarySources(project.SecondarySources)); err != nil {
		return fmt.Errorf("error setting secondary_sources: %s", err)
	}

	if err := d.Set("source", flattenAwsCodeBuildProjectSource(project.Source)); err != nil {
		return fmt.Errorf("error setting source: %s", err)
	}

	if err := d.Set("vpc_config", flattenAwsCodeBuildVpcConfig(project.VpcConfig)); err != nil {
		return fmt.Errorf("error setting vpc_config: %s", err)
	}

	d.Set("arn", project.Arn)
	d.Set("description", project.Description)
	d.Set("encryption_key", project.EncryptionKey)
	d.Set("name", project.Name)
	d.Set("service_role", project.ServiceRole)
	d.Set("source_version", project.SourceVersion)
	d.Set("build_timeout", project.TimeoutInMinutes)
	d.Set("queued_timeout", project.QueuedTimeoutInMinutes)
	if project.Badge != nil {
		d.Set("badge_enabled", project.Badge.BadgeEnabled)
		d.Set("badge_url", project.Badge.BadgeRequestUrl)
	} else {
		d.Set("badge_enabled", false)
		d.Set("badge_url", "")
	}

	if err := d.Set("tags", keyvaluetags.CodebuildKeyValueTags(project.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsCodeBuildProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	params := &codebuild.UpdateProjectInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if d.HasChange("environment") {
		projectEnv := expandProjectEnvironment(d)
		params.Environment = projectEnv
	}

	if d.HasChange("source") {
		projectSource := expandProjectSource(d)
		params.Source = &projectSource
	}

	if d.HasChange("artifacts") {
		projectArtifacts := expandProjectArtifacts(d)
		params.Artifacts = &projectArtifacts
	}

	if d.HasChange("secondary_sources") {
		projectSecondarySources := expandProjectSecondarySources(d)
		params.SecondarySources = projectSecondarySources
	}

	if d.HasChange("secondary_artifacts") {
		projectSecondaryArtifacts := expandProjectSecondaryArtifacts(d)
		params.SecondaryArtifacts = projectSecondaryArtifacts
	}

	if d.HasChange("vpc_config") {
		params.VpcConfig = expandCodeBuildVpcConfig(d.Get("vpc_config").([]interface{}))
	}

	if d.HasChange("logs_config") {
		logsConfig := expandProjectLogsConfig(d)
		params.LogsConfig = logsConfig
	}

	if d.HasChange("cache") {
		if v, ok := d.GetOk("cache"); ok {
			params.Cache = expandProjectCache(v.([]interface{}))
		} else {
			params.Cache = &codebuild.ProjectCache{
				Type: aws.String("NO_CACHE"),
			}
		}
	}

	if d.HasChange("description") {
		params.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("encryption_key") {
		params.EncryptionKey = aws.String(d.Get("encryption_key").(string))
	}

	if d.HasChange("service_role") {
		params.ServiceRole = aws.String(d.Get("service_role").(string))
	}

	if d.HasChange("source_version") {
		params.SourceVersion = aws.String(d.Get("source_version").(string))
	}

	if d.HasChange("build_timeout") {
		params.TimeoutInMinutes = aws.Int64(int64(d.Get("build_timeout").(int)))
	}

	if d.HasChange("queued_timeout") {
		params.QueuedTimeoutInMinutes = aws.Int64(int64(d.Get("queued_timeout").(int)))
	}

	if d.HasChange("badge_enabled") {
		params.BadgeEnabled = aws.Bool(d.Get("badge_enabled").(bool))
	}

	// The documentation clearly says "The replacement set of tags for this build project."
	// But its a slice of pointers so if not set for every update, they get removed.
	params.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().CodebuildTags()

	// Handle IAM eventual consistency
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		_, err = conn.UpdateProject(params)
		if err != nil {
			// InvalidInputException: CodeBuild is not authorized to perform
			// InvalidInputException: Not authorized to perform DescribeSecurityGroups
			if isAWSErr(err, codebuild.ErrCodeInvalidInputException, "ot authorized to perform") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.UpdateProject(params)
	}
	if err != nil {
		return fmt.Errorf(
			"[ERROR] Error updating CodeBuild project (%s): %s",
			d.Id(), err)
	}

	return resourceAwsCodeBuildProjectRead(d, meta)
}

func resourceAwsCodeBuildProjectDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codebuildconn

	_, err := conn.DeleteProject(&codebuild.DeleteProjectInput{
		Name: aws.String(d.Id()),
	})
	return err
}

func flattenAwsCodeBuildLogsConfig(logsConfig *codebuild.LogsConfig) []interface{} {
	if logsConfig == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if v := logsConfig.CloudWatchLogs; v != nil {
		values["cloudwatch_logs"] = flattenAwsCodeBuildCloudWatchLogs(v)
	}

	if v := logsConfig.S3Logs; v != nil {
		values["s3_logs"] = flattenAwsCodeBuildS3Logs(v)
	}

	return []interface{}{values}
}

func flattenAwsCodeBuildCloudWatchLogs(cloudWatchLogsConfig *codebuild.CloudWatchLogsConfig) []interface{} {
	values := map[string]interface{}{}

	if cloudWatchLogsConfig == nil {
		values["status"] = codebuild.LogsConfigStatusTypeDisabled
	} else {
		values["status"] = aws.StringValue(cloudWatchLogsConfig.Status)
		values["group_name"] = aws.StringValue(cloudWatchLogsConfig.GroupName)
		values["stream_name"] = aws.StringValue(cloudWatchLogsConfig.StreamName)
	}

	return []interface{}{values}
}

func flattenAwsCodeBuildS3Logs(s3LogsConfig *codebuild.S3LogsConfig) []interface{} {
	values := map[string]interface{}{}

	if s3LogsConfig == nil {
		values["status"] = codebuild.LogsConfigStatusTypeDisabled
	} else {
		values["status"] = aws.StringValue(s3LogsConfig.Status)
		values["location"] = aws.StringValue(s3LogsConfig.Location)
		values["encryption_disabled"] = aws.BoolValue(s3LogsConfig.EncryptionDisabled)
	}

	return []interface{}{values}
}

func flattenAwsCodeBuildProjectSecondaryArtifacts(artifactsList []*codebuild.ProjectArtifacts) *schema.Set {
	artifactSet := schema.Set{
		F: resourceAwsCodeBuildProjectArtifactsHash,
	}

	for _, artifacts := range artifactsList {
		artifactSet.Add(flattenAwsCodeBuildProjectArtifactsData(*artifacts))
	}
	return &artifactSet
}

func flattenAwsCodeBuildProjectArtifacts(artifacts *codebuild.ProjectArtifacts) []interface{} {
	return []interface{}{flattenAwsCodeBuildProjectArtifactsData(*artifacts)}
}

func flattenAwsCodeBuildProjectArtifactsData(artifacts codebuild.ProjectArtifacts) map[string]interface{} {
	values := map[string]interface{}{}

	values["type"] = aws.StringValue(artifacts.Type)

	if artifacts.ArtifactIdentifier != nil {
		values["artifact_identifier"] = aws.StringValue(artifacts.ArtifactIdentifier)
	}

	if artifacts.EncryptionDisabled != nil {
		values["encryption_disabled"] = aws.BoolValue(artifacts.EncryptionDisabled)
	}

	if artifacts.OverrideArtifactName != nil {
		values["override_artifact_name"] = aws.BoolValue(artifacts.OverrideArtifactName)
	}

	if artifacts.Location != nil {
		values["location"] = aws.StringValue(artifacts.Location)
	}

	if artifacts.Name != nil {
		values["name"] = aws.StringValue(artifacts.Name)
	}

	if artifacts.NamespaceType != nil {
		values["namespace_type"] = aws.StringValue(artifacts.NamespaceType)
	}

	if artifacts.Packaging != nil {
		values["packaging"] = aws.StringValue(artifacts.Packaging)
	}

	if artifacts.Path != nil {
		values["path"] = aws.StringValue(artifacts.Path)
	}
	return values
}

func flattenAwsCodebuildProjectCache(cache *codebuild.ProjectCache) []interface{} {
	if cache == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"location": aws.StringValue(cache.Location),
		"type":     aws.StringValue(cache.Type),
		"modes":    aws.StringValueSlice(cache.Modes),
	}

	return []interface{}{values}
}

func flattenAwsCodeBuildProjectEnvironment(environment *codebuild.ProjectEnvironment) []interface{} {
	envConfig := map[string]interface{}{}

	envConfig["type"] = aws.StringValue(environment.Type)
	envConfig["compute_type"] = aws.StringValue(environment.ComputeType)
	envConfig["image"] = aws.StringValue(environment.Image)
	envConfig["certificate"] = aws.StringValue(environment.Certificate)
	envConfig["privileged_mode"] = aws.BoolValue(environment.PrivilegedMode)
	envConfig["image_pull_credentials_type"] = aws.StringValue(environment.ImagePullCredentialsType)

	envConfig["registry_credential"] = flattenAwsCodebuildRegistryCredential(environment.RegistryCredential)

	if environment.EnvironmentVariables != nil {
		envConfig["environment_variable"] = environmentVariablesToMap(environment.EnvironmentVariables)
	}

	return []interface{}{envConfig}
}

func flattenAwsCodebuildRegistryCredential(registryCredential *codebuild.RegistryCredential) []interface{} {
	if registryCredential == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"credential":          aws.StringValue(registryCredential.Credential),
		"credential_provider": aws.StringValue(registryCredential.CredentialProvider),
	}

	return []interface{}{values}
}

func flattenAwsCodeBuildProjectSecondarySources(sourceList []*codebuild.ProjectSource) []interface{} {
	l := make([]interface{}, 0)

	for _, source := range sourceList {
		l = append(l, flattenAwsCodeBuildProjectSourceData(source))
	}

	return l
}

func flattenAwsCodeBuildProjectSource(source *codebuild.ProjectSource) []interface{} {
	l := make([]interface{}, 1)

	l[0] = flattenAwsCodeBuildProjectSourceData(source)

	return l
}

func flattenAwsCodeBuildProjectSourceData(source *codebuild.ProjectSource) interface{} {
	m := map[string]interface{}{
		"buildspec":           aws.StringValue(source.Buildspec),
		"location":            aws.StringValue(source.Location),
		"git_clone_depth":     int(aws.Int64Value(source.GitCloneDepth)),
		"insecure_ssl":        aws.BoolValue(source.InsecureSsl),
		"report_build_status": aws.BoolValue(source.ReportBuildStatus),
		"type":                aws.StringValue(source.Type),
	}

	m["git_submodules_config"] = flattenAwsCodebuildProjectGitSubmodulesConfig(source.GitSubmodulesConfig)

	if source.Auth != nil {
		m["auth"] = []interface{}{sourceAuthToMap(source.Auth)}
	}
	if source.SourceIdentifier != nil {
		m["source_identifier"] = aws.StringValue(source.SourceIdentifier)
	}

	return m
}

func flattenAwsCodebuildProjectGitSubmodulesConfig(config *codebuild.GitSubmodulesConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"fetch_submodules": aws.BoolValue(config.FetchSubmodules),
	}

	return []interface{}{values}
}

func flattenAwsCodeBuildVpcConfig(vpcConfig *codebuild.VpcConfig) []interface{} {
	if vpcConfig != nil {
		values := map[string]interface{}{}

		values["vpc_id"] = aws.StringValue(vpcConfig.VpcId)
		values["subnets"] = schema.NewSet(schema.HashString, flattenStringList(vpcConfig.Subnets))
		values["security_group_ids"] = schema.NewSet(schema.HashString, flattenStringList(vpcConfig.SecurityGroupIds))

		return []interface{}{values}
	}
	return nil
}

func resourceAwsCodeBuildProjectArtifactsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["artifact_identifier"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["encryption_disabled"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
	}

	if v, ok := m["location"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["namespace_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["override_artifact_name"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
	}

	if v, ok := m["packaging"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["path"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return hashcode.String(buf.String())
}

func environmentVariablesToMap(environmentVariables []*codebuild.EnvironmentVariable) []interface{} {

	envVariables := []interface{}{}
	if len(environmentVariables) > 0 {
		for _, env := range environmentVariables {
			item := map[string]interface{}{}
			item["name"] = aws.StringValue(env.Name)
			item["value"] = aws.StringValue(env.Value)
			if env.Type != nil {
				item["type"] = aws.StringValue(env.Type)
			}
			envVariables = append(envVariables, item)
		}
	}

	return envVariables
}

func sourceAuthToMap(sourceAuth *codebuild.SourceAuth) map[string]interface{} {

	auth := map[string]interface{}{}
	auth["type"] = aws.StringValue(sourceAuth.Type)

	if sourceAuth.Resource != nil {
		auth["resource"] = aws.StringValue(sourceAuth.Resource)
	}

	return auth
}

func validateAwsCodeBuildProjectName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[A-Za-z0-9]`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"first character of %q must be a letter or number", value))
	}

	if !regexp.MustCompile(`^[A-Za-z0-9\-_]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters, hyphens and underscores allowed in %q", value))
	}

	if len(value) > 255 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than 255 characters", value))
	}

	return
}

func validateAwsCodeBuildProjectS3LogsLocation(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if _, errs := validateArn(v, k); len(errs) == 0 {
		errors = append(errors, errs...)
		return
	}

	simplePattern := `^[a-z0-9][^/]*\/(.+)$`
	if !regexp.MustCompile(simplePattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q does not match pattern (%q): %q",
			k, simplePattern, value))
	}

	return
}
