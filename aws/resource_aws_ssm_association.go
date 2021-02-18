package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsSsmAssociation() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAwsSsmAssociationCreate,
		Read:   resourceAwsSsmAssociationRead,
		Update: resourceAwsSsmAssociationUpdate,
		Delete: resourceAwsSsmAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		MigrateState:  resourceAwsSsmAssociationMigrateState,
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"apply_only_at_cron_interval": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"association_name": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]{3,128}$`), "must contain only alphanumeric, underscore, hyphen, or period characters"),
				),
			},
			"association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"document_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([$]LATEST|[$]DEFAULT|^[1-9][0-9]*$)$`), ""),
			},
			"max_concurrency": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([1-9][0-9]*|[1-9][0-9]%|[1-9]%|100%)$`), "must be a valid number (e.g. 10) or percentage including the percent sign (e.g. 10%)"),
			},
			"max_errors": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([1-9][0-9]*|[0]|[1-9][0-9]%|[0-9]%|100%)$`), "must be a valid number (e.g. 10) or percentage including the percent sign (e.g. 10%)"),
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"schedule_expression": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"output_location": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_bucket_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"s3_key_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"targets": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 163),
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"target_location": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 100,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accounts": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 50,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateAwsAccountId,
							},
						},
						"execution_role_name": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "AWS-SystemsManager-AutomationExecutionRole",
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 64),
								validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]+$`), ""),
							),
						},
						"regions": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"target_location_max_concurrency": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([1-9][0-9]*|[1-9][0-9]%|[1-9]%|100%)$`), "must be a valid number (e.g. 10) or percentage including the percent sign (e.g. 10%)"),
						},
						"target_location_max_errors": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([1-9][0-9]*|[0]|[1-9][0-9]%|[0-9]%|100%)$`), "must be a valid number (e.g. 10) or percentage including the percent sign (e.g. 10%)"),
						},
					},
				},
			},
			"compliance_severity": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ssm.ComplianceSeverity_Values(), false),
			},
			"automation_target_parameter_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"sync_compliance": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ssm.AssociationSyncComplianceAuto,
				ValidateFunc: validation.StringInSlice(ssm.AssociationSyncCompliance_Values(), false),
			},
		},
	}
}

func resourceAwsSsmAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[DEBUG] SSM association create: %s", d.Id())

	associationInput := &ssm.CreateAssociationInput{
		Name:           aws.String(d.Get("name").(string)),
		SyncCompliance: aws.String(d.Get("sync_compliance").(string)),
	}

	if v, ok := d.GetOk("apply_only_at_cron_interval"); ok {
		associationInput.ApplyOnlyAtCronInterval = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("association_name"); ok {
		associationInput.AssociationName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_id"); ok {
		associationInput.InstanceId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("document_version"); ok {
		associationInput.DocumentVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedule_expression"); ok {
		associationInput.ScheduleExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters"); ok {
		associationInput.Parameters = expandSSMDocumentParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("targets"); ok {
		associationInput.Targets = expandAwsSsmTargets(v.([]interface{}))
	}

	if v, ok := d.GetOk("target_location"); ok {
		associationInput.TargetLocations = expandAwsSsmTargetLocations(v.([]interface{}))
	}

	if v, ok := d.GetOk("output_location"); ok {
		associationInput.OutputLocation = expandSSMAssociationOutputLocation(v.([]interface{}))
	}

	if v, ok := d.GetOk("compliance_severity"); ok {
		associationInput.ComplianceSeverity = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_concurrency"); ok {
		associationInput.MaxConcurrency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_errors"); ok {
		associationInput.MaxErrors = aws.String(v.(string))
	}

	if v, ok := d.GetOk("automation_target_parameter_name"); ok {
		associationInput.AutomationTargetParameterName = aws.String(v.(string))
	}

	resp, err := ssmconn.CreateAssociation(associationInput)
	if err != nil {
		return fmt.Errorf("Error creating SSM association: %s", err)
	}

	if resp.AssociationDescription == nil {
		return fmt.Errorf("AssociationDescription was nil")
	}

	d.SetId(aws.StringValue(resp.AssociationDescription.AssociationId))
	d.Set("association_id", resp.AssociationDescription.AssociationId)

	return resourceAwsSsmAssociationRead(d, meta)
}

func resourceAwsSsmAssociationRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[DEBUG] Reading SSM Association: %s", d.Id())

	params := &ssm.DescribeAssociationInput{
		AssociationId: aws.String(d.Id()),
	}

	resp, err := ssmconn.DescribeAssociation(params)

	if err != nil {
		if isAWSErr(err, ssm.ErrCodeAssociationDoesNotExist, "") {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading SSM association: %s", err)
	}
	if resp.AssociationDescription == nil {
		return fmt.Errorf("AssociationDescription was nil")
	}

	association := resp.AssociationDescription
	d.Set("apply_only_at_cron_interval", association.ApplyOnlyAtCronInterval)
	d.Set("association_name", association.AssociationName)
	d.Set("instance_id", association.InstanceId)
	d.Set("name", association.Name)
	d.Set("association_id", association.AssociationId)
	d.Set("schedule_expression", association.ScheduleExpression)
	d.Set("document_version", association.DocumentVersion)
	d.Set("compliance_severity", association.ComplianceSeverity)
	d.Set("max_concurrency", association.MaxConcurrency)
	d.Set("max_errors", association.MaxErrors)
	d.Set("automation_target_parameter_name", association.AutomationTargetParameterName)
	d.Set("sync_compliance", association.SyncCompliance)

	if err := d.Set("parameters", flattenAwsSsmParameters(association.Parameters)); err != nil {
		return fmt.Errorf("Error setting parameters: %w", err)
	}

	if err := d.Set("targets", flattenAwsSsmTargets(association.Targets)); err != nil {
		return fmt.Errorf("Error setting targets: %w", err)
	}

	if err := d.Set("target_location", flattenAwsSsmTargetLocations(association.TargetLocations)); err != nil {
		return fmt.Errorf("Error setting target_location: %w", err)
	}

	if err := d.Set("output_location", flattenAwsSsmAssociationOutoutLocation(association.OutputLocation)); err != nil {
		return fmt.Errorf("Error setting output_location: %w", err)
	}

	return nil
}

func resourceAwsSsmAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[DEBUG] SSM Association update: %s", d.Id())

	associationInput := &ssm.UpdateAssociationInput{
		AssociationId: aws.String(d.Get("association_id").(string)),
	}

	if v, ok := d.GetOk("apply_only_at_cron_interval"); ok {
		associationInput.ApplyOnlyAtCronInterval = aws.Bool(v.(bool))
	}

	// AWS creates a new version every time the association is updated, so everything should be passed in the update.
	if v, ok := d.GetOk("association_name"); ok {
		associationInput.AssociationName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("document_version"); ok {
		associationInput.DocumentVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedule_expression"); ok {
		associationInput.ScheduleExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters"); ok {
		associationInput.Parameters = expandSSMDocumentParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("targets"); ok {
		associationInput.Targets = expandAwsSsmTargets(v.([]interface{}))
	}

	if v, ok := d.GetOk("target_location"); ok {
		associationInput.TargetLocations = expandAwsSsmTargetLocations(v.([]interface{}))
	}

	if v, ok := d.GetOk("output_location"); ok {
		associationInput.OutputLocation = expandSSMAssociationOutputLocation(v.([]interface{}))
	}

	if v, ok := d.GetOk("compliance_severity"); ok {
		associationInput.ComplianceSeverity = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_concurrency"); ok {
		associationInput.MaxConcurrency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_errors"); ok {
		associationInput.MaxErrors = aws.String(v.(string))
	}

	if v, ok := d.GetOk("automation_target_parameter_name"); ok {
		associationInput.AutomationTargetParameterName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sync_compliance"); ok {
		associationInput.SyncCompliance = aws.String(v.(string))
	}

	_, err := ssmconn.UpdateAssociation(associationInput)
	if err != nil {
		return fmt.Errorf("Error updating SSM association: %s", err)
	}

	return resourceAwsSsmAssociationRead(d, meta)
}

func resourceAwsSsmAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[DEBUG] Deleting SSM Association: %s", d.Id())

	params := &ssm.DeleteAssociationInput{
		AssociationId: aws.String(d.Get("association_id").(string)),
	}

	_, err := ssmconn.DeleteAssociation(params)

	if err != nil {
		return fmt.Errorf("Error deleting SSM association: %s", err)
	}

	return nil
}

func expandSSMDocumentParameters(params map[string]interface{}) map[string][]*string {
	var docParams = make(map[string][]*string)
	for k, v := range params {
		values := make([]*string, 1)
		values[0] = aws.String(v.(string))
		docParams[k] = values
	}

	return docParams
}

func expandSSMAssociationOutputLocation(config []interface{}) *ssm.InstanceAssociationOutputLocation {
	if config == nil {
		return nil
	}

	//We only allow 1 Item so we can grab the first in the list only
	locationConfig := config[0].(map[string]interface{})

	S3OutputLocation := &ssm.S3OutputLocation{
		OutputS3BucketName: aws.String(locationConfig["s3_bucket_name"].(string)),
	}

	if v, ok := locationConfig["s3_key_prefix"]; ok {
		S3OutputLocation.OutputS3KeyPrefix = aws.String(v.(string))
	}

	return &ssm.InstanceAssociationOutputLocation{
		S3Location: S3OutputLocation,
	}
}

func flattenAwsSsmAssociationOutoutLocation(location *ssm.InstanceAssociationOutputLocation) []map[string]interface{} {
	if location == nil {
		return nil
	}

	result := make([]map[string]interface{}, 0)
	item := make(map[string]interface{})

	item["s3_bucket_name"] = *location.S3Location.OutputS3BucketName

	if location.S3Location.OutputS3KeyPrefix != nil {
		item["s3_key_prefix"] = *location.S3Location.OutputS3KeyPrefix
	}

	result = append(result, item)

	return result
}

func expandAwsSsmTargetLocations(in []interface{}) []*ssm.TargetLocation {
	targets := make([]*ssm.TargetLocation, 0)

	for _, tConfig := range in {
		config := tConfig.(map[string]interface{})

		target := &ssm.TargetLocation{}

		if v, ok := config["execution_role_name"].(string); ok && v != "" {
			target.ExecutionRoleName = aws.String(v)
		}

		if v, ok := config["target_location_max_concurrency"].(string); ok && v != "" {
			target.TargetLocationMaxConcurrency = aws.String(v)
		}

		if v, ok := config["target_location_max_errors"].(string); ok && v != "" {
			target.TargetLocationMaxErrors = aws.String(v)
		}

		if v, ok := config["accounts"].(*schema.Set); ok && v.Len() > 0 {
			target.Accounts = expandStringSet(v)
		}

		if v, ok := config["regions"].(*schema.Set); ok && v.Len() > 0 {
			target.Regions = expandStringSet(v)
		}

		targets = append(targets, target)
	}

	return targets
}

func flattenAwsSsmTargetLocations(targets []*ssm.TargetLocation) []map[string]interface{} {
	if len(targets) == 0 {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(targets))
	for _, target := range targets {
		item := make(map[string]interface{}, 1)

		if target.ExecutionRoleName != nil {
			item["execution_role_name"] = aws.StringValue(target.ExecutionRoleName)
		}

		if target.TargetLocationMaxConcurrency != nil {
			item["target_location_max_concurrency"] = aws.StringValue(target.TargetLocationMaxConcurrency)
		}

		if target.TargetLocationMaxErrors != nil {
			item["target_location_max_errors"] = aws.StringValue(target.TargetLocationMaxErrors)
		}

		if target.Accounts != nil && len(target.Accounts) > 0 {
			item["accounts"] = flattenStringSet(target.Accounts)
		}

		if target.Regions != nil && len(target.Regions) > 0 {
			item["regions"] = flattenStringSet(target.Regions)
		}

		result = append(result, item)
	}

	return result
}
