package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsElasticSearchDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsElasticSearchDomainCreate,
		Read:   resourceAwsElasticSearchDomainRead,
		Update: resourceAwsElasticSearchDomainUpdate,
		Delete: resourceAwsElasticSearchDomainDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsElasticSearchDomainImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(60 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIf("elasticsearch_version", func(d *schema.ResourceDiff, meta interface{}) bool {
				newVersion := d.Get("elasticsearch_version").(string)
				domainName := d.Get("domain_name").(string)

				conn := meta.(*AWSClient).esconn
				resp, err := conn.GetCompatibleElasticsearchVersions(&elasticsearch.GetCompatibleElasticsearchVersionsInput{
					DomainName: aws.String(domainName),
				})
				if err != nil {
					log.Printf("[ERROR] Failed to get compatible ElasticSearch versions %s", domainName)
					return false
				}
				if len(resp.CompatibleElasticsearchVersions) != 1 {
					return true
				}
				for _, targetVersion := range resp.CompatibleElasticsearchVersions[0].TargetVersions {
					if aws.StringValue(targetVersion) == newVersion {
						return false
					}
				}
				return true
			}),
		),

		Schema: map[string]*schema.Schema{
			"access_policies": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
			"advanced_options": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][0-9a-z\-]{2,27}$`),
					"must start with a lowercase alphabet and be at least 3 and no more than 28 characters long."+
						" Valid characters are a-z (lowercase letters), 0-9, and - (hyphen)."),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_endpoint_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enforce_https": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"tls_security_policy": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								elasticsearch.TLSSecurityPolicyPolicyMinTls10201907,
								elasticsearch.TLSSecurityPolicyPolicyMinTls12201907,
							}, false),
						},
					},
				},
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kibana_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ebs_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ebs_enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"iops": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"volume_size": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"volume_type": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								elasticsearch.VolumeTypeStandard,
								elasticsearch.VolumeTypeGp2,
								elasticsearch.VolumeTypeIo1,
							}, false),
						},
					},
				},
			},
			"encrypt_at_rest": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
							ForceNew: true,
						},
						"kms_key_id": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ForceNew:         true,
							DiffSuppressFunc: suppressEquivalentKmsKeyIds,
						},
					},
				},
			},
			"node_to_node_encryption": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"cluster_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dedicated_master_count": {
							Type:             schema.TypeInt,
							Optional:         true,
							DiffSuppressFunc: isDedicatedMasterDisabled,
						},
						"dedicated_master_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"dedicated_master_type": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: isDedicatedMasterDisabled,
						},
						"instance_count": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
						"instance_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  elasticsearch.ESPartitionInstanceTypeM3MediumElasticsearch,
						},
						"zone_awareness_config": {
							Type:             schema.TypeList,
							Optional:         true,
							MaxItems:         1,
							DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"availability_zone_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      2,
										ValidateFunc: validation.IntInSlice([]int{2, 3}),
									},
								},
							},
						},
						"zone_awareness_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"warm_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"warm_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(2, 150),
						},
						"warm_type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								elasticsearch.ESWarmPartitionInstanceTypeUltrawarm1MediumElasticsearch,
								elasticsearch.ESWarmPartitionInstanceTypeUltrawarm1LargeElasticsearch,
								"ultrawarm1.xlarge.elasticsearch",
							}, false),
						},
					},
				},
			},
			"snapshot_options": {
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
						"automated_snapshot_start_hour": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"vpc_options": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"log_publishing_options": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								elasticsearch.LogTypeIndexSlowLogs,
								elasticsearch.LogTypeSearchSlowLogs,
								elasticsearch.LogTypeEsApplicationLogs,
							}, false),
						},
						"cloudwatch_log_group_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
			"elasticsearch_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1.5",
			},
			"cognito_options": {
				Type:             schema.TypeList,
				Optional:         true,
				ForceNew:         false,
				MaxItems:         1,
				DiffSuppressFunc: esCognitoOptionsDiffSuppress,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"user_pool_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"identity_pool_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsElasticSearchDomainImport(
	d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("domain_name", d.Id())
	return []*schema.ResourceData{d}, nil
}

func resourceAwsElasticSearchDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).esconn

	// The API doesn't check for duplicate names
	// so w/out this check Create would act as upsert
	// and might cause duplicate domain to appear in state
	resp, err := conn.DescribeElasticsearchDomain(&elasticsearch.DescribeElasticsearchDomainInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	})
	if err == nil {
		return fmt.Errorf("ElasticSearch domain %s already exists", aws.StringValue(resp.DomainStatus.DomainName))
	}

	input := elasticsearch.CreateElasticsearchDomainInput{
		DomainName:           aws.String(d.Get("domain_name").(string)),
		ElasticsearchVersion: aws.String(d.Get("elasticsearch_version").(string)),
	}

	if v, ok := d.GetOk("access_policies"); ok {
		input.AccessPolicies = aws.String(v.(string))
	}

	if v, ok := d.GetOk("advanced_options"); ok {
		input.AdvancedOptions = stringMapToPointers(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("ebs_options"); ok {
		options := v.([]interface{})

		if len(options) == 1 {
			if options[0] == nil {
				return fmt.Errorf("At least one field is expected inside ebs_options")
			}

			s := options[0].(map[string]interface{})
			input.EBSOptions = expandESEBSOptions(s)
		}
	}

	if v, ok := d.GetOk("encrypt_at_rest"); ok {
		options := v.([]interface{})
		if options[0] == nil {
			return fmt.Errorf("At least one field is expected inside encrypt_at_rest")
		}

		s := options[0].(map[string]interface{})
		input.EncryptionAtRestOptions = expandESEncryptAtRestOptions(s)
	}

	if v, ok := d.GetOk("cluster_config"); ok {
		config := v.([]interface{})

		if len(config) == 1 {
			if config[0] == nil {
				return fmt.Errorf("At least one field is expected inside cluster_config")
			}
			m := config[0].(map[string]interface{})
			input.ElasticsearchClusterConfig = expandESClusterConfig(m)
		}
	}

	if v, ok := d.GetOk("node_to_node_encryption"); ok {
		options := v.([]interface{})

		s := options[0].(map[string]interface{})
		input.NodeToNodeEncryptionOptions = expandESNodeToNodeEncryptionOptions(s)
	}

	if v, ok := d.GetOk("snapshot_options"); ok {
		options := v.([]interface{})

		if len(options) == 1 {
			if options[0] == nil {
				return fmt.Errorf("At least one field is expected inside snapshot_options")
			}

			o := options[0].(map[string]interface{})

			snapshotOptions := elasticsearch.SnapshotOptions{
				AutomatedSnapshotStartHour: aws.Int64(int64(o["automated_snapshot_start_hour"].(int))),
			}

			input.SnapshotOptions = &snapshotOptions
		}
	}

	if v, ok := d.GetOk("vpc_options"); ok {
		options := v.([]interface{})
		if options[0] == nil {
			return fmt.Errorf("At least one field is expected inside vpc_options")
		}

		s := options[0].(map[string]interface{})
		input.VPCOptions = expandESVPCOptions(s)
	}

	if v, ok := d.GetOk("log_publishing_options"); ok {
		input.LogPublishingOptions = make(map[string]*elasticsearch.LogPublishingOption)
		options := v.(*schema.Set).List()
		for _, vv := range options {
			lo := vv.(map[string]interface{})
			input.LogPublishingOptions[lo["log_type"].(string)] = &elasticsearch.LogPublishingOption{
				CloudWatchLogsLogGroupArn: aws.String(lo["cloudwatch_log_group_arn"].(string)),
				Enabled:                   aws.Bool(lo["enabled"].(bool)),
			}
		}
	}

	if v, ok := d.GetOk("domain_endpoint_options"); ok {
		input.DomainEndpointOptions = expandESDomainEndpointOptions(v.([]interface{}))
	}

	if v, ok := d.GetOk("cognito_options"); ok {
		input.CognitoOptions = expandESCognitoOptions(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating ElasticSearch domain: %s", input)

	// IAM Roles can take some time to propagate if set in AccessPolicies and created in the same terraform
	var out *elasticsearch.CreateElasticsearchDomainOutput
	err = resource.Retry(30*time.Second, func() *resource.RetryError {
		var err error
		out, err = conn.CreateElasticsearchDomain(&input)
		if err != nil {
			if isAWSErr(err, "InvalidTypeException", "Error setting policy") {
				log.Printf("[DEBUG] Retrying creation of ElasticSearch domain %s", aws.StringValue(input.DomainName))
				return resource.RetryableError(err)
			}
			if isAWSErr(err, "ValidationException", "enable a service-linked role to give Amazon ES permissions") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, "ValidationException", "Domain is still being deleted") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, "ValidationException", "Amazon Elasticsearch must be allowed to use the passed role") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, "ValidationException", "The passed role has not propagated yet") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, "ValidationException", "Authentication error") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, "ValidationException", "Unauthorized Operation: Elasticsearch must be authorised to describe") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, "ValidationException", "The passed role must authorize Amazon Elasticsearch to describe") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		out, err = conn.CreateElasticsearchDomain(&input)
	}
	if err != nil {
		return fmt.Errorf("Error creating ElasticSearch domain: %s", err)
	}

	d.SetId(aws.StringValue(out.DomainStatus.ARN))

	// Whilst the domain is being created, we can initialise the tags.
	// This should mean that if the creation fails (eg because your token expired
	// whilst the operation is being performed), we still get the required tags on
	// the resources.
	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		if err := keyvaluetags.ElasticsearchserviceUpdateTags(conn, d.Id(), nil, v); err != nil {
			return fmt.Errorf("error adding Elasticsearch Cluster (%s) tags: %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Waiting for ElasticSearch domain %q to be created", d.Id())
	err = waitForElasticSearchDomainCreation(conn, d.Get("domain_name").(string), d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] ElasticSearch domain %q created", d.Id())

	return resourceAwsElasticSearchDomainRead(d, meta)
}

func waitForElasticSearchDomainCreation(conn *elasticsearch.ElasticsearchService, domainName, arn string) error {
	input := &elasticsearch.DescribeElasticsearchDomainInput{
		DomainName: aws.String(domainName),
	}
	var out *elasticsearch.DescribeElasticsearchDomainOutput
	err := resource.Retry(60*time.Minute, func() *resource.RetryError {
		var err error
		out, err = conn.DescribeElasticsearchDomain(input)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if !aws.BoolValue(out.DomainStatus.Processing) && (out.DomainStatus.Endpoint != nil || out.DomainStatus.Endpoints != nil) {
			return nil
		}

		return resource.RetryableError(
			fmt.Errorf("%q: Timeout while waiting for the domain to be created", arn))
	})
	if isResourceTimeoutError(err) {
		out, err = conn.DescribeElasticsearchDomain(input)
		if err != nil {
			return fmt.Errorf("Error describing ElasticSearch domain: %s", err)
		}
		if !aws.BoolValue(out.DomainStatus.Processing) && (out.DomainStatus.Endpoint != nil || out.DomainStatus.Endpoints != nil) {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error waiting for ElasticSearch domain to be created: %s", err)
	}
	return nil
}

func resourceAwsElasticSearchDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).esconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	out, err := conn.DescribeElasticsearchDomain(&elasticsearch.DescribeElasticsearchDomainInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	})
	if err != nil {
		if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "ResourceNotFoundException" {
			log.Printf("[INFO] ElasticSearch Domain %q not found", d.Get("domain_name").(string))
			d.SetId("")
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] Received ElasticSearch domain: %s", out)

	ds := out.DomainStatus

	if ds.AccessPolicies != nil && aws.StringValue(ds.AccessPolicies) != "" {
		policies, err := structure.NormalizeJsonString(aws.StringValue(ds.AccessPolicies))
		if err != nil {
			return fmt.Errorf("access policies contain an invalid JSON: %s", err)
		}
		d.Set("access_policies", policies)
	}
	err = d.Set("advanced_options", pointersMapToStringList(ds.AdvancedOptions))
	if err != nil {
		return err
	}
	d.SetId(aws.StringValue(ds.ARN))
	d.Set("domain_id", ds.DomainId)
	d.Set("domain_name", ds.DomainName)
	d.Set("elasticsearch_version", ds.ElasticsearchVersion)

	err = d.Set("ebs_options", flattenESEBSOptions(ds.EBSOptions))
	if err != nil {
		return err
	}
	err = d.Set("encrypt_at_rest", flattenESEncryptAtRestOptions(ds.EncryptionAtRestOptions))
	if err != nil {
		return err
	}
	err = d.Set("cluster_config", flattenESClusterConfig(ds.ElasticsearchClusterConfig))
	if err != nil {
		return err
	}
	err = d.Set("cognito_options", flattenESCognitoOptions(ds.CognitoOptions))
	if err != nil {
		return err
	}
	err = d.Set("node_to_node_encryption", flattenESNodeToNodeEncryptionOptions(ds.NodeToNodeEncryptionOptions))
	if err != nil {
		return err
	}

	if err := d.Set("snapshot_options", flattenESSnapshotOptions(ds.SnapshotOptions)); err != nil {
		return fmt.Errorf("error setting snapshot_options: %s", err)
	}

	if ds.VPCOptions != nil {
		err = d.Set("vpc_options", flattenESVPCDerivedInfo(ds.VPCOptions))
		if err != nil {
			return err
		}
		endpoints := pointersMapToStringList(ds.Endpoints)
		err = d.Set("endpoint", endpoints["vpc"])
		if err != nil {
			return err
		}
		d.Set("kibana_endpoint", getKibanaEndpoint(d))
		if ds.Endpoint != nil {
			return fmt.Errorf("%q: Elasticsearch domain in VPC expected to have null Endpoint value", d.Id())
		}
	} else {
		if ds.Endpoint != nil {
			d.Set("endpoint", aws.StringValue(ds.Endpoint))
			d.Set("kibana_endpoint", getKibanaEndpoint(d))
		}
		if ds.Endpoints != nil {
			return fmt.Errorf("%q: Elasticsearch domain not in VPC expected to have null Endpoints value", d.Id())
		}
	}

	if ds.LogPublishingOptions != nil {
		m := make([]map[string]interface{}, 0)
		for k, val := range ds.LogPublishingOptions {
			mm := map[string]interface{}{}
			mm["log_type"] = k
			if val.CloudWatchLogsLogGroupArn != nil {
				mm["cloudwatch_log_group_arn"] = aws.StringValue(val.CloudWatchLogsLogGroupArn)
			}
			mm["enabled"] = aws.BoolValue(val.Enabled)
			m = append(m, mm)
		}
		d.Set("log_publishing_options", m)
	}

	if err := d.Set("domain_endpoint_options", flattenESDomainEndpointOptions(ds.DomainEndpointOptions)); err != nil {
		return fmt.Errorf("error setting domain_endpoint_options: %s", err)
	}

	d.Set("arn", ds.ARN)

	tags, err := keyvaluetags.ElasticsearchserviceListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for Elasticsearch Cluster (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsElasticSearchDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).esconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.ElasticsearchserviceUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Elasticsearch Cluster (%s) tags: %s", d.Id(), err)
		}
	}

	input := elasticsearch.UpdateElasticsearchDomainConfigInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	if d.HasChange("access_policies") {
		input.AccessPolicies = aws.String(d.Get("access_policies").(string))
	}

	if d.HasChange("advanced_options") {
		input.AdvancedOptions = stringMapToPointers(d.Get("advanced_options").(map[string]interface{}))
	}

	if d.HasChange("domain_endpoint_options") {
		input.DomainEndpointOptions = expandESDomainEndpointOptions(d.Get("domain_endpoint_options").([]interface{}))
	}

	if d.HasChanges("ebs_options", "cluster_config") {
		options := d.Get("ebs_options").([]interface{})

		if len(options) == 1 {
			s := options[0].(map[string]interface{})
			input.EBSOptions = expandESEBSOptions(s)
		}

		if d.HasChange("cluster_config") {
			config := d.Get("cluster_config").([]interface{})

			if len(config) == 1 {
				m := config[0].(map[string]interface{})
				input.ElasticsearchClusterConfig = expandESClusterConfig(m)
			}
		}

	}

	if d.HasChange("snapshot_options") {
		options := d.Get("snapshot_options").([]interface{})

		if len(options) == 1 {
			o := options[0].(map[string]interface{})

			snapshotOptions := elasticsearch.SnapshotOptions{
				AutomatedSnapshotStartHour: aws.Int64(int64(o["automated_snapshot_start_hour"].(int))),
			}

			input.SnapshotOptions = &snapshotOptions
		}
	}

	if d.HasChange("vpc_options") {
		options := d.Get("vpc_options").([]interface{})
		s := options[0].(map[string]interface{})
		input.VPCOptions = expandESVPCOptions(s)
	}

	if d.HasChange("cognito_options") {
		options := d.Get("cognito_options").([]interface{})
		input.CognitoOptions = expandESCognitoOptions(options)
	}

	if d.HasChange("log_publishing_options") {
		input.LogPublishingOptions = make(map[string]*elasticsearch.LogPublishingOption)
		options := d.Get("log_publishing_options").(*schema.Set).List()
		for _, vv := range options {
			lo := vv.(map[string]interface{})
			input.LogPublishingOptions[lo["log_type"].(string)] = &elasticsearch.LogPublishingOption{
				CloudWatchLogsLogGroupArn: aws.String(lo["cloudwatch_log_group_arn"].(string)),
				Enabled:                   aws.Bool(lo["enabled"].(bool)),
			}
		}
	}

	_, err := conn.UpdateElasticsearchDomainConfig(&input)
	if err != nil {
		return err
	}

	descInput := &elasticsearch.DescribeElasticsearchDomainInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	}
	var out *elasticsearch.DescribeElasticsearchDomainOutput
	err = resource.Retry(60*time.Minute, func() *resource.RetryError {
		out, err = conn.DescribeElasticsearchDomain(descInput)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if !aws.BoolValue(out.DomainStatus.Processing) {
			return nil
		}

		return resource.RetryableError(
			fmt.Errorf("%q: Timeout while waiting for changes to be processed", d.Id()))
	})
	if isResourceTimeoutError(err) {
		out, err = conn.DescribeElasticsearchDomain(descInput)
		if err != nil {
			return fmt.Errorf("Error describing ElasticSearch domain: %s", err)
		}
		if !aws.BoolValue(out.DomainStatus.Processing) {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error waiting for ElasticSearch domain changes to be processed: %s", err)
	}

	if d.HasChange("elasticsearch_version") {
		upgradeInput := elasticsearch.UpgradeElasticsearchDomainInput{
			DomainName:    aws.String(d.Get("domain_name").(string)),
			TargetVersion: aws.String(d.Get("elasticsearch_version").(string)),
		}

		_, err := conn.UpgradeElasticsearchDomain(&upgradeInput)
		if err != nil {
			return fmt.Errorf("Failed to upgrade elasticsearch domain: %s", err)
		}

		stateConf := &resource.StateChangeConf{
			Pending: []string{elasticsearch.UpgradeStatusInProgress},
			Target:  []string{elasticsearch.UpgradeStatusSucceeded},
			Refresh: func() (interface{}, string, error) {
				out, err := conn.GetUpgradeStatus(&elasticsearch.GetUpgradeStatusInput{
					DomainName: aws.String(d.Get("domain_name").(string)),
				})
				if err != nil {
					return nil, "", err
				}

				// Elasticsearch upgrades consist of multiple steps:
				// https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-version-migration.html
				// Prevent false positive completion where the UpgradeStep is not the final UPGRADE step.
				if aws.StringValue(out.StepStatus) == elasticsearch.UpgradeStatusSucceeded && aws.StringValue(out.UpgradeStep) != elasticsearch.UpgradeStepUpgrade {
					return out, elasticsearch.UpgradeStatusInProgress, nil
				}

				return out, aws.StringValue(out.StepStatus), nil
			},
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			MinTimeout: 10 * time.Second,
			Delay:      30 * time.Second, // The upgrade status isn't instantly available for the current upgrade so will either be nil or reflect a previous upgrade
		}
		_, waitErr := stateConf.WaitForState()
		if waitErr != nil {
			return waitErr
		}
	}

	return resourceAwsElasticSearchDomainRead(d, meta)
}

func resourceAwsElasticSearchDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).esconn
	domainName := d.Get("domain_name").(string)

	log.Printf("[DEBUG] Deleting ElasticSearch domain: %q", domainName)
	_, err := conn.DeleteElasticsearchDomain(&elasticsearch.DeleteElasticsearchDomainInput{
		DomainName: aws.String(domainName),
	})
	if err != nil {
		if isAWSErr(err, elasticsearch.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] Waiting for ElasticSearch domain %q to be deleted", domainName)
	err = resourceAwsElasticSearchDomainDeleteWaiter(domainName, conn)

	return err
}

func resourceAwsElasticSearchDomainDeleteWaiter(domainName string, conn *elasticsearch.ElasticsearchService) error {
	input := &elasticsearch.DescribeElasticsearchDomainInput{
		DomainName: aws.String(domainName),
	}
	var out *elasticsearch.DescribeElasticsearchDomainOutput
	err := resource.Retry(90*time.Minute, func() *resource.RetryError {
		var err error
		out, err = conn.DescribeElasticsearchDomain(input)

		if err != nil {
			if isAWSErr(err, elasticsearch.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		if out.DomainStatus != nil && !aws.BoolValue(out.DomainStatus.Processing) {
			return nil
		}

		return resource.RetryableError(fmt.Errorf("timeout while waiting for the domain %q to be deleted", domainName))
	})
	if isResourceTimeoutError(err) {
		out, err = conn.DescribeElasticsearchDomain(input)
		if err != nil {
			if isAWSErr(err, elasticsearch.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return fmt.Errorf("Error describing ElasticSearch domain: %s", err)
		}
		if out.DomainStatus != nil && !aws.BoolValue(out.DomainStatus.Processing) {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error waiting for ElasticSearch domain to be deleted: %s", err)
	}
	return nil
}

func suppressEquivalentKmsKeyIds(k, old, new string, d *schema.ResourceData) bool {
	// The Elasticsearch API accepts a short KMS key id but always returns the ARN of the key.
	// The ARN is of the format 'arn:aws:kms:REGION:ACCOUNT_ID:key/KMS_KEY_ID'.
	// These should be treated as equivalent.
	return strings.Contains(old, new)
}

func getKibanaEndpoint(d *schema.ResourceData) string {
	return d.Get("endpoint").(string) + "/_plugin/kibana/"
}

func esCognitoOptionsDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	if old == "1" && new == "0" {
		return true
	}
	return false
}

func isDedicatedMasterDisabled(k, old, new string, d *schema.ResourceData) bool {
	v, ok := d.GetOk("cluster_config")
	if ok {
		clusterConfig := v.([]interface{})[0].(map[string]interface{})
		return !clusterConfig["dedicated_master_enabled"].(bool)
	}
	return false
}

func expandESNodeToNodeEncryptionOptions(s map[string]interface{}) *elasticsearch.NodeToNodeEncryptionOptions {
	options := elasticsearch.NodeToNodeEncryptionOptions{}

	if v, ok := s["enabled"]; ok {
		options.Enabled = aws.Bool(v.(bool))
	}
	return &options
}

func flattenESNodeToNodeEncryptionOptions(o *elasticsearch.NodeToNodeEncryptionOptions) []map[string]interface{} {
	if o == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}
	if o.Enabled != nil {
		m["enabled"] = aws.BoolValue(o.Enabled)
	}

	return []map[string]interface{}{m}
}

func expandESClusterConfig(m map[string]interface{}) *elasticsearch.ElasticsearchClusterConfig {
	config := elasticsearch.ElasticsearchClusterConfig{}

	if v, ok := m["dedicated_master_enabled"]; ok {
		isEnabled := v.(bool)
		config.DedicatedMasterEnabled = aws.Bool(isEnabled)

		if isEnabled {
			if v, ok := m["dedicated_master_count"]; ok && v.(int) > 0 {
				config.DedicatedMasterCount = aws.Int64(int64(v.(int)))
			}
			if v, ok := m["dedicated_master_type"]; ok && v.(string) != "" {
				config.DedicatedMasterType = aws.String(v.(string))
			}
		}
	}

	if v, ok := m["instance_count"]; ok {
		config.InstanceCount = aws.Int64(int64(v.(int)))
	}
	if v, ok := m["instance_type"]; ok {
		config.InstanceType = aws.String(v.(string))
	}

	if v, ok := m["zone_awareness_enabled"]; ok {
		isEnabled := v.(bool)
		config.ZoneAwarenessEnabled = aws.Bool(isEnabled)

		if isEnabled {
			if v, ok := m["zone_awareness_config"]; ok {
				config.ZoneAwarenessConfig = expandElasticsearchZoneAwarenessConfig(v.([]interface{}))
			}
		}
	}

	if v, ok := m["warm_enabled"]; ok {
		isEnabled := v.(bool)
		config.WarmEnabled = aws.Bool(isEnabled)

		if isEnabled {
			if v, ok := m["warm_count"]; ok {
				config.WarmCount = aws.Int64(int64(v.(int)))
			}

			if v, ok := m["warm_type"]; ok {
				config.WarmType = aws.String(v.(string))
			}
		}
	}

	return &config
}

func expandElasticsearchZoneAwarenessConfig(l []interface{}) *elasticsearch.ZoneAwarenessConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	zoneAwarenessConfig := &elasticsearch.ZoneAwarenessConfig{}

	if v, ok := m["availability_zone_count"]; ok && v.(int) > 0 {
		zoneAwarenessConfig.AvailabilityZoneCount = aws.Int64(int64(v.(int)))
	}

	return zoneAwarenessConfig
}

func flattenESClusterConfig(c *elasticsearch.ElasticsearchClusterConfig) []map[string]interface{} {
	m := map[string]interface{}{
		"zone_awareness_config":  flattenElasticsearchZoneAwarenessConfig(c.ZoneAwarenessConfig),
		"zone_awareness_enabled": aws.BoolValue(c.ZoneAwarenessEnabled),
	}

	if c.DedicatedMasterCount != nil {
		m["dedicated_master_count"] = aws.Int64Value(c.DedicatedMasterCount)
	}
	if c.DedicatedMasterEnabled != nil {
		m["dedicated_master_enabled"] = aws.BoolValue(c.DedicatedMasterEnabled)
	}
	if c.DedicatedMasterType != nil {
		m["dedicated_master_type"] = aws.StringValue(c.DedicatedMasterType)
	}
	if c.InstanceCount != nil {
		m["instance_count"] = aws.Int64Value(c.InstanceCount)
	}
	if c.InstanceType != nil {
		m["instance_type"] = aws.StringValue(c.InstanceType)
	}
	if c.WarmEnabled != nil {
		m["warm_enabled"] = aws.BoolValue(c.WarmEnabled)
	}
	if c.WarmCount != nil {
		m["warm_count"] = aws.Int64Value(c.WarmCount)
	}
	if c.WarmType != nil {
		m["warm_type"] = aws.StringValue(c.WarmType)
	}

	return []map[string]interface{}{m}
}

func flattenElasticsearchZoneAwarenessConfig(zoneAwarenessConfig *elasticsearch.ZoneAwarenessConfig) []interface{} {
	if zoneAwarenessConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"availability_zone_count": aws.Int64Value(zoneAwarenessConfig.AvailabilityZoneCount),
	}

	return []interface{}{m}
}
