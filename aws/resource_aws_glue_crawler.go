package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsGlueCrawler() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueCrawlerCreate,
		Read:   resourceAwsGlueCrawlerRead,
		Update: resourceAwsGlueCrawlerUpdate,
		Delete: resourceAwsGlueCrawlerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
				// Glue API always returns name
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					newARN, err := arn.Parse(new)

					if err != nil {
						return false
					}

					return old == strings.TrimPrefix(newARN.Resource, "role/")
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"schedule": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"classifiers": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"schema_change_policy": {
				Type:     schema.TypeList,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "1" && new == "0" {
						return true
					}
					return false
				},
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_behavior": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  glue.DeleteBehaviorDeprecateInDatabase,
							ValidateFunc: validation.StringInSlice([]string{
								glue.DeleteBehaviorDeleteFromDatabase,
								glue.DeleteBehaviorDeprecateInDatabase,
								glue.DeleteBehaviorLog,
							}, false),
						},
						"update_behavior": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  glue.UpdateBehaviorUpdateInDatabase,
							ValidateFunc: validation.StringInSlice([]string{
								glue.UpdateBehaviorLog,
								glue.UpdateBehaviorUpdateInDatabase,
							}, false),
						},
					},
				},
			},
			"table_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"s3_target": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"exclusions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"dynamodb_target": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"jdbc_target": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"exclusions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"catalog_target": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"tables": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"configuration": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				ValidateFunc: validation.StringIsJSON,
			},
			"security_configuration": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsGlueCrawlerCreate(d *schema.ResourceData, meta interface{}) error {
	glueConn := meta.(*AWSClient).glueconn
	name := d.Get("name").(string)

	crawlerInput, err := createCrawlerInput(name, d)
	if err != nil {
		return err
	}

	// Retry for IAM eventual consistency
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err = glueConn.CreateCrawler(crawlerInput)
		if err != nil {
			if isAWSErr(err, glue.ErrCodeInvalidInputException, "Service is unable to assume role") {
				return resource.RetryableError(err)
			}
			// InvalidInputException: Unable to retrieve connection tf-acc-test-8656357591012534997: User: arn:aws:sts::*******:assumed-role/tf-acc-test-8656357591012534997/AWS-Crawler is not authorized to perform: glue:GetConnection on resource: * (Service: AmazonDataCatalog; Status Code: 400; Error Code: AccessDeniedException; Request ID: 4d72b66f-9c75-11e8-9faf-5b526c7be968)
			if isAWSErr(err, glue.ErrCodeInvalidInputException, "is not authorized") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = glueConn.CreateCrawler(crawlerInput)
	}
	if err != nil {
		return fmt.Errorf("error creating Glue crawler: %s", err)
	}
	d.SetId(name)

	return resourceAwsGlueCrawlerRead(d, meta)
}

func createCrawlerInput(crawlerName string, d *schema.ResourceData) (*glue.CreateCrawlerInput, error) {
	crawlerTargets, err := expandGlueCrawlerTargets(d)
	if err != nil {
		return nil, err
	}
	crawlerInput := &glue.CreateCrawlerInput{
		Name:         aws.String(crawlerName),
		DatabaseName: aws.String(d.Get("database_name").(string)),
		Role:         aws.String(d.Get("role").(string)),
		Tags:         keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GlueTags(),
		Targets:      crawlerTargets,
	}
	if description, ok := d.GetOk("description"); ok {
		crawlerInput.Description = aws.String(description.(string))
	}
	if schedule, ok := d.GetOk("schedule"); ok {
		crawlerInput.Schedule = aws.String(schedule.(string))
	}
	if classifiers, ok := d.GetOk("classifiers"); ok {
		crawlerInput.Classifiers = expandStringList(classifiers.([]interface{}))
	}

	crawlerInput.SchemaChangePolicy = expandGlueSchemaChangePolicy(d.Get("schema_change_policy").([]interface{}))

	if tablePrefix, ok := d.GetOk("table_prefix"); ok {
		crawlerInput.TablePrefix = aws.String(tablePrefix.(string))
	}
	if configuration, ok := d.GetOk("configuration"); ok {
		crawlerInput.Configuration = aws.String(configuration.(string))
	}

	if v, ok := d.GetOk("configuration"); ok {
		configuration, err := structure.NormalizeJsonString(v)
		if err != nil {
			return nil, fmt.Errorf("Configuration contains an invalid JSON: %v", err)
		}
		crawlerInput.Configuration = aws.String(configuration)
	}

	if securityConfiguration, ok := d.GetOk("security_configuration"); ok {
		crawlerInput.CrawlerSecurityConfiguration = aws.String(securityConfiguration.(string))
	}

	return crawlerInput, nil
}

func updateCrawlerInput(crawlerName string, d *schema.ResourceData) (*glue.UpdateCrawlerInput, error) {
	crawlerTargets, err := expandGlueCrawlerTargets(d)
	if err != nil {
		return nil, err
	}
	crawlerInput := &glue.UpdateCrawlerInput{
		Name:         aws.String(crawlerName),
		DatabaseName: aws.String(d.Get("database_name").(string)),
		Role:         aws.String(d.Get("role").(string)),
		Targets:      crawlerTargets,
	}
	if description, ok := d.GetOk("description"); ok {
		crawlerInput.Description = aws.String(description.(string))
	}
	if schedule, ok := d.GetOk("schedule"); ok {
		crawlerInput.Schedule = aws.String(schedule.(string))
	}
	if classifiers, ok := d.GetOk("classifiers"); ok {
		crawlerInput.Classifiers = expandStringList(classifiers.([]interface{}))
	}

	crawlerInput.SchemaChangePolicy = expandGlueSchemaChangePolicy(d.Get("schema_change_policy").([]interface{}))

	if tablePrefix, ok := d.GetOk("table_prefix"); ok {
		crawlerInput.TablePrefix = aws.String(tablePrefix.(string))
	}
	if configuration, ok := d.GetOk("configuration"); ok {
		crawlerInput.Configuration = aws.String(configuration.(string))
	}

	if v, ok := d.GetOk("configuration"); ok {
		configuration, err := structure.NormalizeJsonString(v)
		if err != nil {
			return nil, fmt.Errorf("Configuration contains an invalid JSON: %v", err)
		}
		crawlerInput.Configuration = aws.String(configuration)
	}

	if securityConfiguration, ok := d.GetOk("security_configuration"); ok {
		crawlerInput.CrawlerSecurityConfiguration = aws.String(securityConfiguration.(string))
	}

	return crawlerInput, nil
}

func expandGlueSchemaChangePolicy(v []interface{}) *glue.SchemaChangePolicy {
	if len(v) == 0 {
		return nil
	}

	schemaPolicy := &glue.SchemaChangePolicy{}

	member := v[0].(map[string]interface{})

	if updateBehavior, ok := member["update_behavior"]; ok && updateBehavior.(string) != "" {
		schemaPolicy.UpdateBehavior = aws.String(updateBehavior.(string))
	}

	if deleteBehavior, ok := member["delete_behavior"]; ok && deleteBehavior.(string) != "" {
		schemaPolicy.DeleteBehavior = aws.String(deleteBehavior.(string))
	}
	return schemaPolicy
}

func expandGlueCrawlerTargets(d *schema.ResourceData) (*glue.CrawlerTargets, error) {
	crawlerTargets := &glue.CrawlerTargets{}

	dynamodbTargets, dynamodbTargetsOk := d.GetOk("dynamodb_target")
	jdbcTargets, jdbcTargetsOk := d.GetOk("jdbc_target")
	s3Targets, s3TargetsOk := d.GetOk("s3_target")
	catalogTargets, catalogTargetsOk := d.GetOk("catalog_target")
	if !dynamodbTargetsOk && !jdbcTargetsOk && !s3TargetsOk && !catalogTargetsOk {
		return nil, fmt.Errorf("One of the following configurations is required: dynamodb_target, jdbc_target, s3_target, catalog_target")
	}

	log.Print("[DEBUG] Creating crawler target")
	crawlerTargets.DynamoDBTargets = expandGlueDynamoDBTargets(dynamodbTargets.([]interface{}))
	crawlerTargets.JdbcTargets = expandGlueJdbcTargets(jdbcTargets.([]interface{}))
	crawlerTargets.S3Targets = expandGlueS3Targets(s3Targets.([]interface{}))
	crawlerTargets.CatalogTargets = expandGlueCatalogTargets(catalogTargets.([]interface{}))

	return crawlerTargets, nil
}

func expandGlueDynamoDBTargets(targets []interface{}) []*glue.DynamoDBTarget {
	if len(targets) < 1 {
		return []*glue.DynamoDBTarget{}
	}

	perms := make([]*glue.DynamoDBTarget, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandGlueDynamoDBTarget(cfg)
	}
	return perms
}

func expandGlueDynamoDBTarget(cfg map[string]interface{}) *glue.DynamoDBTarget {
	target := &glue.DynamoDBTarget{
		Path: aws.String(cfg["path"].(string)),
	}

	return target
}

func expandGlueS3Targets(targets []interface{}) []*glue.S3Target {
	if len(targets) < 1 {
		return []*glue.S3Target{}
	}

	perms := make([]*glue.S3Target, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandGlueS3Target(cfg)
	}
	return perms
}

func expandGlueS3Target(cfg map[string]interface{}) *glue.S3Target {
	target := &glue.S3Target{
		Path: aws.String(cfg["path"].(string)),
	}

	if exclusions, ok := cfg["exclusions"]; ok {
		target.Exclusions = expandStringList(exclusions.([]interface{}))
	}
	return target
}

func expandGlueJdbcTargets(targets []interface{}) []*glue.JdbcTarget {
	if len(targets) < 1 {
		return []*glue.JdbcTarget{}
	}

	perms := make([]*glue.JdbcTarget, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandGlueJdbcTarget(cfg)
	}
	return perms
}

func expandGlueJdbcTarget(cfg map[string]interface{}) *glue.JdbcTarget {
	target := &glue.JdbcTarget{
		Path:           aws.String(cfg["path"].(string)),
		ConnectionName: aws.String(cfg["connection_name"].(string)),
	}

	if exclusions, ok := cfg["exclusions"]; ok {
		target.Exclusions = expandStringList(exclusions.([]interface{}))
	}
	return target
}

func expandGlueCatalogTargets(targets []interface{}) []*glue.CatalogTarget {
	if len(targets) < 1 {
		return []*glue.CatalogTarget{}
	}

	perms := make([]*glue.CatalogTarget, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandGlueCatalogTarget(cfg)
	}
	return perms
}

func expandGlueCatalogTarget(cfg map[string]interface{}) *glue.CatalogTarget {
	target := &glue.CatalogTarget{
		DatabaseName: aws.String(cfg["database_name"].(string)),
		Tables:       expandStringList(cfg["tables"].([]interface{})),
	}

	return target
}

func resourceAwsGlueCrawlerUpdate(d *schema.ResourceData, meta interface{}) error {
	glueConn := meta.(*AWSClient).glueconn
	name := d.Get("name").(string)

	if d.HasChanges(
		"catalog_target", "classifiers", "configuration", "description", "dynamodb_target", "jdbc_target", "role",
		"s3_target", "schedule", "schema_change_policy", "security_configuration", "table_prefix") {
		updateCrawlerInput, err := updateCrawlerInput(name, d)
		if err != nil {
			return err
		}

		// Retry for IAM eventual consistency
		err = resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := glueConn.UpdateCrawler(updateCrawlerInput)
			if err != nil {
				if isAWSErr(err, glue.ErrCodeInvalidInputException, "Service is unable to assume role") {
					return resource.RetryableError(err)
				}
				// InvalidInputException: Unable to retrieve connection tf-acc-test-8656357591012534997: User: arn:aws:sts::*******:assumed-role/tf-acc-test-8656357591012534997/AWS-Crawler is not authorized to perform: glue:GetConnection on resource: * (Service: AmazonDataCatalog; Status Code: 400; Error Code: AccessDeniedException; Request ID: 4d72b66f-9c75-11e8-9faf-5b526c7be968)
				if isAWSErr(err, glue.ErrCodeInvalidInputException, "is not authorized") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("error updating Glue crawler: %s", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.GlueUpdateTags(glueConn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsGlueCrawlerRead(d, meta)
}

func resourceAwsGlueCrawlerRead(d *schema.ResourceData, meta interface{}) error {
	glueConn := meta.(*AWSClient).glueconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &glue.GetCrawlerInput{
		Name: aws.String(d.Id()),
	}

	crawlerOutput, err := glueConn.GetCrawler(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Crawler (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error reading Glue crawler: %s", err.Error())
	}

	if crawlerOutput.Crawler == nil {
		log.Printf("[WARN] Glue Crawler (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	crawlerARN := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "glue",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("crawler/%s", d.Id()),
	}.String()
	d.Set("arn", crawlerARN)
	d.Set("name", crawlerOutput.Crawler.Name)
	d.Set("database_name", crawlerOutput.Crawler.DatabaseName)
	d.Set("role", crawlerOutput.Crawler.Role)
	d.Set("configuration", crawlerOutput.Crawler.Configuration)
	d.Set("description", crawlerOutput.Crawler.Description)
	d.Set("security_configuration", crawlerOutput.Crawler.CrawlerSecurityConfiguration)
	d.Set("schedule", "")
	if crawlerOutput.Crawler.Schedule != nil {
		d.Set("schedule", crawlerOutput.Crawler.Schedule.ScheduleExpression)
	}
	if err := d.Set("classifiers", flattenStringList(crawlerOutput.Crawler.Classifiers)); err != nil {
		return fmt.Errorf("error setting classifiers: %s", err)
	}
	d.Set("table_prefix", crawlerOutput.Crawler.TablePrefix)

	if crawlerOutput.Crawler.SchemaChangePolicy != nil {
		schemaPolicy := map[string]string{
			"delete_behavior": aws.StringValue(crawlerOutput.Crawler.SchemaChangePolicy.DeleteBehavior),
			"update_behavior": aws.StringValue(crawlerOutput.Crawler.SchemaChangePolicy.UpdateBehavior),
		}

		if err := d.Set("schema_change_policy", []map[string]string{schemaPolicy}); err != nil {
			return fmt.Errorf("error setting schema_change_policy: %s", schemaPolicy)
		}
	}

	if crawlerOutput.Crawler.Targets != nil {
		if err := d.Set("dynamodb_target", flattenGlueDynamoDBTargets(crawlerOutput.Crawler.Targets.DynamoDBTargets)); err != nil {
			return fmt.Errorf("error setting dynamodb_target: %s", err)
		}

		if err := d.Set("jdbc_target", flattenGlueJdbcTargets(crawlerOutput.Crawler.Targets.JdbcTargets)); err != nil {
			return fmt.Errorf("error setting jdbc_target: %s", err)
		}

		if err := d.Set("s3_target", flattenGlueS3Targets(crawlerOutput.Crawler.Targets.S3Targets)); err != nil {
			return fmt.Errorf("error setting s3_target: %s", err)
		}

		if err := d.Set("catalog_target", flattenGlueCatalogTargets(crawlerOutput.Crawler.Targets.CatalogTargets)); err != nil {
			return fmt.Errorf("error setting catalog_target: %s", err)
		}
	}

	tags, err := keyvaluetags.GlueListTags(glueConn, crawlerARN)

	if err != nil {
		return fmt.Errorf("error listing tags for Glue Crawler (%s): %s", crawlerARN, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func flattenGlueS3Targets(s3Targets []*glue.S3Target) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, s3Target := range s3Targets {
		attrs := make(map[string]interface{})
		attrs["exclusions"] = flattenStringList(s3Target.Exclusions)
		attrs["path"] = aws.StringValue(s3Target.Path)

		result = append(result, attrs)
	}
	return result
}

func flattenGlueCatalogTargets(CatalogTargets []*glue.CatalogTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, catalogTarget := range CatalogTargets {
		attrs := make(map[string]interface{})
		attrs["tables"] = flattenStringList(catalogTarget.Tables)
		attrs["database_name"] = aws.StringValue(catalogTarget.DatabaseName)

		result = append(result, attrs)
	}
	return result
}

func flattenGlueDynamoDBTargets(dynamodbTargets []*glue.DynamoDBTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, dynamodbTarget := range dynamodbTargets {
		attrs := make(map[string]interface{})
		attrs["path"] = aws.StringValue(dynamodbTarget.Path)

		result = append(result, attrs)
	}
	return result
}

func flattenGlueJdbcTargets(jdbcTargets []*glue.JdbcTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, jdbcTarget := range jdbcTargets {
		attrs := make(map[string]interface{})
		attrs["connection_name"] = aws.StringValue(jdbcTarget.ConnectionName)
		attrs["exclusions"] = flattenStringList(jdbcTarget.Exclusions)
		attrs["path"] = aws.StringValue(jdbcTarget.Path)

		result = append(result, attrs)
	}
	return result
}

func resourceAwsGlueCrawlerDelete(d *schema.ResourceData, meta interface{}) error {
	glueConn := meta.(*AWSClient).glueconn

	log.Printf("[DEBUG] deleting Glue crawler: %s", d.Id())
	_, err := glueConn.DeleteCrawler(&glue.DeleteCrawlerInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Glue crawler: %s", err.Error())
	}
	return nil
}
