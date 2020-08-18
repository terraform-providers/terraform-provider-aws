package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func generateQueryFilterSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"delta_time": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"offset_seconds": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"time_expression": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func generateSqlQueryDatasetActionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"sql_query": {
				Type:     schema.TypeString,
				Required: true,
			},
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     generateQueryFilterSchema(),
			},
		},
	}
}

func generateDatasetActionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"query_action": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateSqlQueryDatasetActionSchema(),
			},
		},
	}
}

func generateS3DestinationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"glue_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"table_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func generateDatasetContentDeliveryDestinationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"iotevents_destination": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"input_name": {
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
			"s3_destination": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateS3DestinationSchema(),
			},
		},
	}
}

func generateDatasetContentDeliveryRuleSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"entry_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem:     generateDatasetContentDeliveryDestinationSchema(),
			},
		},
	}
}

func generateRetentionPeriodSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"number_of_days": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"retention_period.0.unlimited"},
				ValidateFunc:  validation.IntAtLeast(1),
			},
			"unlimited": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"retention_period.0.number_of_days"},
			},
		},
	}
}

func generateDatasetTriggerSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"schedule": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"expression": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func generateVersioningConfigurationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"max_versions": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"versioning_configuration.0.unlimited"},
				ValidateFunc:  validation.IntAtLeast(1),
			},
			"unlimited": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"versioning_configuration.0.max_versions"},
			},
		},
	}
}

func resourceAwsIotAnalyticsDataset() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotAnalyticsDatasetCreate,
		Read:   resourceAwsIotAnalyticsDatasetRead,
		Update: resourceAwsIotAnalyticsDatasetUpdate,
		Delete: resourceAwsIotAnalyticsDatasetDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"action": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     generateDatasetActionSchema(),
			},
			"content_delivery_rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     generateDatasetContentDeliveryRuleSchema(),
			},
			"retention_period": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateRetentionPeriodSchema(),
			},
			"trigger": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 5,
				Elem:     generateDatasetTriggerSchema(),
			},
			"versioning_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateVersioningConfigurationSchema(),
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func parseQueryFilter(rawQueryFilter map[string]interface{}) *iotanalytics.QueryFilter {
	rawDeltaTime := rawQueryFilter["delta_time"].(*schema.Set).List()[0].(map[string]interface{})
	deltaTime := &iotanalytics.DeltaTime{
		OffsetSeconds:  aws.Int64(int64(rawDeltaTime["offset_seconds"].(int))),
		TimeExpression: aws.String(rawDeltaTime["time_expression"].(string)),
	}
	queryFilter := &iotanalytics.QueryFilter{
		DeltaTime: deltaTime,
	}
	return queryFilter
}

func parseSqlQueryAction(rawSqlQueryAction map[string]interface{}) *iotanalytics.SqlQueryDatasetAction {
	sqlQueryAction := &iotanalytics.SqlQueryDatasetAction{
		SqlQuery: aws.String(rawSqlQueryAction["sql_query"].(string)),
	}

	filters := make([]*iotanalytics.QueryFilter, 0)
	rawFilters := rawSqlQueryAction["filter"].(*schema.Set).List()
	for _, rawFilter := range rawFilters {
		filter := parseQueryFilter(rawFilter.(map[string]interface{}))
		filters = append(filters, filter)
	}
	sqlQueryAction.Filters = filters
	return sqlQueryAction

}

func parseAction(rawAction map[string]interface{}) *iotanalytics.DatasetAction {
	action := &iotanalytics.DatasetAction{
		ActionName: aws.String(rawAction["name"].(string)),
	}

	rawQueryActionSet := rawAction["query_action"].(*schema.Set).List()
	if len(rawQueryActionSet) > 0 {
		rawQueryAction := rawQueryActionSet[0].(map[string]interface{})
		action.QueryAction = parseSqlQueryAction(rawQueryAction)
	}

	return action
}

func parseS3Destination(rawS3Destination map[string]interface{}) *iotanalytics.S3DestinationConfiguration {
	s3Destination := &iotanalytics.S3DestinationConfiguration{
		Bucket:  aws.String(rawS3Destination["bucket"].(string)),
		Key:     aws.String(rawS3Destination["key"].(string)),
		RoleArn: aws.String(rawS3Destination["role_arn"].(string)),
	}

	rawGlueConfigurationSet := rawS3Destination["glue_configuration"].(*schema.Set).List()
	if len(rawGlueConfigurationSet) > 0 {
		rawGlueConfiguration := rawGlueConfigurationSet[0].(map[string]interface{})
		s3Destination.GlueConfiguration = &iotanalytics.GlueConfiguration{
			DatabaseName: aws.String(rawGlueConfiguration["database_name"].(string)),
			TableName:    aws.String(rawGlueConfiguration["table_name"].(string)),
		}
	}

	return s3Destination
}

func parseIotEventsDestination(rawIotEventsDestination map[string]interface{}) *iotanalytics.IotEventsDestinationConfiguration {
	return &iotanalytics.IotEventsDestinationConfiguration{
		InputName: aws.String(rawIotEventsDestination["input_name"].(string)),
		RoleArn:   aws.String(rawIotEventsDestination["role_arn"].(string)),
	}
}

func parseDestination(rawDestination map[string]interface{}) *iotanalytics.DatasetContentDeliveryDestination {
	destination := &iotanalytics.DatasetContentDeliveryDestination{}

	rawIotEventsDestinationSet := rawDestination["iotevents_destination"].(*schema.Set).List()
	if len(rawIotEventsDestinationSet) > 0 {
		rawIotEventsDestination := rawIotEventsDestinationSet[0].(map[string]interface{})
		destination.IotEventsDestinationConfiguration = parseIotEventsDestination(rawIotEventsDestination)
	}

	rawS3DestinationSet := rawDestination["s3_destination"].(*schema.Set).List()
	if len(rawS3DestinationSet) > 0 {
		rawS3Destination := rawS3DestinationSet[0].(map[string]interface{})
		destination.S3DestinationConfiguration = parseS3Destination(rawS3Destination)
	}

	return destination
}

func parseContentDeliveryRule(rawContentDeliveryRule map[string]interface{}) *iotanalytics.DatasetContentDeliveryRule {
	rawDestination := rawContentDeliveryRule["destination"].(*schema.Set).List()[0].(map[string]interface{})
	datasetContentDeliveryRule := &iotanalytics.DatasetContentDeliveryRule{
		Destination: parseDestination(rawDestination),
	}

	if rawEntryName, ok := rawContentDeliveryRule["entry_name"]; ok {
		datasetContentDeliveryRule.EntryName = aws.String(rawEntryName.(string))
	}

	return datasetContentDeliveryRule
}

func parseRetentionPeriod(rawRetentionPeriod map[string]interface{}) *iotanalytics.RetentionPeriod {

	var numberOfDays *int64
	if v, ok := rawRetentionPeriod["number_of_days"]; ok && int64(v.(int)) >= 1 {
		numberOfDays = aws.Int64(int64(v.(int)))
	}
	var unlimited *bool
	if v, ok := rawRetentionPeriod["unlimited"]; ok {
		unlimited = aws.Bool(v.(bool))
	}
	return &iotanalytics.RetentionPeriod{
		NumberOfDays: numberOfDays,
		Unlimited:    unlimited,
	}
}

func parseTrigger(rawTrigger map[string]interface{}) *iotanalytics.DatasetTrigger {
	trigger := &iotanalytics.DatasetTrigger{}

	rawScheduleSet := rawTrigger["schedule"].(*schema.Set).List()
	if len(rawScheduleSet) > 0 {
		rawSchedule := rawScheduleSet[0].(map[string]interface{})
		trigger.Schedule = &iotanalytics.Schedule{
			Expression: aws.String(rawSchedule["expression"].(string)),
		}
	}

	return trigger
}

func parseVersioningConfiguration(rawVersioningConfiguration map[string]interface{}) *iotanalytics.VersioningConfiguration {
	var maxVersion *int64
	if v, ok := rawVersioningConfiguration["max_versions"]; ok && int64(v.(int)) > 1 {
		maxVersion = aws.Int64(int64(v.(int)))
	}
	var unlimited *bool
	if v, ok := rawVersioningConfiguration["unlimited"]; ok {
		unlimited = aws.Bool(v.(bool))
	}
	return &iotanalytics.VersioningConfiguration{
		MaxVersions: maxVersion,
		Unlimited:   unlimited,
	}
}

func resourceAwsIotAnalyticsDatasetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	name := d.Get("name").(string)
	params := &iotanalytics.CreateDatasetInput{
		DatasetName: aws.String(name),
	}

	if tags := d.Get("tags").(map[string]interface{}); len(tags) > 0 {
		params.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().IotanalyticsTags()
	}

	rawActions := d.Get("action").(*schema.Set).List()
	actions := make([]*iotanalytics.DatasetAction, 0)
	for _, rawAction := range rawActions {
		action := parseAction(rawAction.(map[string]interface{}))
		actions = append(actions, action)
	}
	params.Actions = actions

	rawContentDeliveryRules := d.Get("content_delivery_rule").(*schema.Set).List()
	contentDeliveryRules := make([]*iotanalytics.DatasetContentDeliveryRule, 0)
	for _, rawRule := range rawContentDeliveryRules {
		rule := parseContentDeliveryRule(rawRule.(map[string]interface{}))
		contentDeliveryRules = append(contentDeliveryRules, rule)
	}
	params.ContentDeliveryRules = contentDeliveryRules

	rawTriggers := d.Get("trigger").(*schema.Set).List()
	triggers := make([]*iotanalytics.DatasetTrigger, 0)
	for _, rawTrigger := range rawTriggers {
		trigger := parseTrigger(rawTrigger.(map[string]interface{}))
		triggers = append(triggers, trigger)
	}
	params.Triggers = triggers

	rawRetentionPeriodSet := d.Get("retention_period").(*schema.Set).List()
	if len(rawRetentionPeriodSet) > 0 {
		rawRetentionPeriod := rawRetentionPeriodSet[0].(map[string]interface{})
		params.RetentionPeriod = parseRetentionPeriod(rawRetentionPeriod)
	}

	rawVersioningConfigurationSet := d.Get("versioning_configuration").(*schema.Set).List()
	if len(rawVersioningConfigurationSet) > 0 {
		rawVersioningConfiguration := rawVersioningConfigurationSet[0].(map[string]interface{})
		params.VersioningConfiguration = parseVersioningConfiguration(rawVersioningConfiguration)
	}

	log.Printf("[DEBUG] Creating IoT Analytics Dataset: %s", params)
	_, err := conn.CreateDataset(params)

	if err != nil {
		return err
	}

	d.SetId(name)

	return resourceAwsIotAnalyticsDatasetRead(d, meta)
}

func wrapMapInList(mapping map[string]interface{}) []map[string]interface{} {
	// We should use TypeList or TypeSet with MaxItems in case we want single object.
	// So, schema.ResourceData.Set requires list as type of argument for such fields,
	// as a result code becomes a little bit messy with such instructions : []map[string]interface{}{someObject}.
	// This helper function wrap mapping in list, and makes code more readable and intuitive.
	if mapping == nil {
		return make([]map[string]interface{}, 0)
	} else {
		return []map[string]interface{}{mapping}
	}
}

func flattenQueryFilter(queryFilter *iotanalytics.QueryFilter) map[string]interface{} {
	rawDeltaTime := map[string]interface{}{
		"offset_seconds":  aws.Int64Value(queryFilter.DeltaTime.OffsetSeconds),
		"time_expression": aws.StringValue(queryFilter.DeltaTime.TimeExpression),
	}
	rawQueryFilter := make(map[string]interface{})
	rawQueryFilter["delta_time"] = wrapMapInList(rawDeltaTime)
	return rawQueryFilter
}

func flattenSqlQueryAction(sqlQueryAction *iotanalytics.SqlQueryDatasetAction) map[string]interface{} {
	rawSqlQueryAction := make(map[string]interface{})
	rawSqlQueryAction["sql_query"] = aws.StringValue(sqlQueryAction.SqlQuery)

	rawFilters := make([]map[string]interface{}, 0)
	for _, filter := range sqlQueryAction.Filters {
		rawFilters = append(rawFilters, flattenQueryFilter(filter))
	}
	rawSqlQueryAction["filter"] = rawFilters
	return rawSqlQueryAction
}

func flattenAction(action *iotanalytics.DatasetAction) map[string]interface{} {
	rawAction := make(map[string]interface{})
	rawAction["name"] = aws.StringValue(action.ActionName)

	if action.QueryAction != nil {
		rawQueryAction := flattenSqlQueryAction(action.QueryAction)
		rawAction["query_action"] = wrapMapInList(rawQueryAction)
	}

	return rawAction
}

func flattenS3Destination(s3Destination *iotanalytics.S3DestinationConfiguration) map[string]interface{} {
	rawS3Destination := make(map[string]interface{})
	rawS3Destination["bucket"] = aws.StringValue(s3Destination.Bucket)
	rawS3Destination["key"] = aws.StringValue(s3Destination.Key)
	rawS3Destination["role_arn"] = aws.StringValue(s3Destination.RoleArn)

	if s3Destination.GlueConfiguration != nil {
		rawGlueConfiguration := map[string]interface{}{
			"database_name": aws.StringValue(s3Destination.GlueConfiguration.DatabaseName),
			"table_name":    aws.StringValue(s3Destination.GlueConfiguration.TableName),
		}
		rawS3Destination["glue_configuration"] = wrapMapInList(rawGlueConfiguration)
	}
	return rawS3Destination
}

func flattenIotEventsDestination(iotEventsDestination *iotanalytics.IotEventsDestinationConfiguration) map[string]interface{} {
	rawIotEventsDestination := map[string]interface{}{
		"input_name": aws.StringValue(iotEventsDestination.InputName),
		"role_arn":   aws.StringValue(iotEventsDestination.RoleArn),
	}
	return rawIotEventsDestination
}

func flattenDestination(destination *iotanalytics.DatasetContentDeliveryDestination) map[string]interface{} {
	rawDestination := make(map[string]interface{})

	if destination.IotEventsDestinationConfiguration != nil {
		rawIotEventsDestination := flattenIotEventsDestination(destination.IotEventsDestinationConfiguration)
		rawDestination["iotevents_destination"] = wrapMapInList(rawIotEventsDestination)
	}

	if destination.S3DestinationConfiguration != nil {
		rawS3Destination := flattenS3Destination(destination.S3DestinationConfiguration)
		rawDestination["s3_destination"] = wrapMapInList(rawS3Destination)
	}

	return rawDestination
}

func flattenContentDeliveryRule(datasetContentDeliveryRule *iotanalytics.DatasetContentDeliveryRule) map[string]interface{} {
	rawContentDeliveryRule := make(map[string]interface{})

	rawDestination := flattenDestination(datasetContentDeliveryRule.Destination)
	rawContentDeliveryRule["destination"] = wrapMapInList(rawDestination)

	if datasetContentDeliveryRule.EntryName != nil {
		rawContentDeliveryRule["entry_name"] = aws.StringValue(datasetContentDeliveryRule.EntryName)
	}

	return rawContentDeliveryRule
}

func flattenRetentionPeriod(retentionPeriod *iotanalytics.RetentionPeriod) map[string]interface{} {
	if retentionPeriod == nil {
		return nil
	}

	rawRetentionPeriod := make(map[string]interface{})

	if retentionPeriod.NumberOfDays != nil {
		rawRetentionPeriod["number_of_days"] = aws.Int64Value(retentionPeriod.NumberOfDays)
	}
	if retentionPeriod.Unlimited != nil {
		rawRetentionPeriod["unlimited"] = aws.BoolValue(retentionPeriod.Unlimited)
	}

	return rawRetentionPeriod
}

func flattenTrigger(trigger *iotanalytics.DatasetTrigger) map[string]interface{} {
	rawTrigger := make(map[string]interface{})

	if trigger.Schedule != nil {
		rawSchedule := map[string]interface{}{
			"expression": aws.StringValue(trigger.Schedule.Expression),
		}
		rawTrigger["schedule"] = wrapMapInList(rawSchedule)
	}

	return rawTrigger
}

func flattenVersioningConfiguration(versioningConfiguration *iotanalytics.VersioningConfiguration) map[string]interface{} {
	if versioningConfiguration == nil {
		return nil
	}

	rawVersioningConfiguration := make(map[string]interface{})

	if versioningConfiguration.MaxVersions != nil {
		rawVersioningConfiguration["max_versions"] = aws.Int64Value(versioningConfiguration.MaxVersions)
	}
	if versioningConfiguration.Unlimited != nil {
		rawVersioningConfiguration["unlimited"] = aws.BoolValue(versioningConfiguration.Unlimited)
	}

	return rawVersioningConfiguration
}

func resourceAwsIotAnalyticsDatasetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.DescribeDatasetInput{
		DatasetName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading IoT Analytics Dataset: %s", params)
	out, err := conn.DescribeDataset(params)

	if err != nil {
		return err
	}

	d.Set("name", out.Dataset.Name)

	rawActions := make([]map[string]interface{}, 0)
	for _, action := range out.Dataset.Actions {
		rawActions = append(rawActions, flattenAction(action))
	}
	d.Set("action", rawActions)

	rawContentDeliveryRules := make([]map[string]interface{}, 0)
	for _, rule := range out.Dataset.ContentDeliveryRules {
		rawContentDeliveryRules = append(rawContentDeliveryRules, flattenContentDeliveryRule(rule))
	}
	d.Set("content_delivery_rule", rawContentDeliveryRules)

	rawRetentionPeriod := flattenRetentionPeriod(out.Dataset.RetentionPeriod)
	d.Set("retention_period", wrapMapInList(rawRetentionPeriod))

	rawTriggers := make([]map[string]interface{}, 0)
	for _, trigger := range out.Dataset.Triggers {
		rawTriggers = append(rawTriggers, flattenTrigger(trigger))
	}
	d.Set("trigger", rawTriggers)

	rawVersioningConfiguration := flattenVersioningConfiguration(out.Dataset.VersioningConfiguration)
	d.Set("versioning_configuration", wrapMapInList(rawVersioningConfiguration))
	d.Set("arn", out.Dataset.Arn)

	arn := *out.Dataset.Arn
	tags, err := keyvaluetags.IotanalyticsListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsIotAnalyticsDatasetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	name := d.Get("name").(string)
	params := &iotanalytics.UpdateDatasetInput{
		DatasetName: aws.String(name),
	}

	rawActions := d.Get("action").(*schema.Set).List()
	actions := make([]*iotanalytics.DatasetAction, 0)
	for _, rawAction := range rawActions {
		action := parseAction(rawAction.(map[string]interface{}))
		actions = append(actions, action)
	}
	params.Actions = actions

	rawContentDeliveryRules := d.Get("content_delivery_rule").(*schema.Set).List()
	contentDeliveryRules := make([]*iotanalytics.DatasetContentDeliveryRule, 0)
	for _, rawRule := range rawContentDeliveryRules {
		rule := parseContentDeliveryRule(rawRule.(map[string]interface{}))
		contentDeliveryRules = append(contentDeliveryRules, rule)
	}
	params.ContentDeliveryRules = contentDeliveryRules

	rawTriggers := d.Get("trigger").(*schema.Set).List()
	triggers := make([]*iotanalytics.DatasetTrigger, 0)
	for _, rawTrigger := range rawTriggers {
		trigger := parseTrigger(rawTrigger.(map[string]interface{}))
		triggers = append(triggers, trigger)
	}
	params.Triggers = triggers

	rawRetentionPeriodSet := d.Get("retention_period").(*schema.Set).List()
	if len(rawRetentionPeriodSet) > 0 {
		rawRetentionPeriod := rawRetentionPeriodSet[0].(map[string]interface{})
		params.RetentionPeriod = parseRetentionPeriod(rawRetentionPeriod)
	}

	rawVersioningConfigurationSet := d.Get("versioning_configuration").(*schema.Set).List()
	if len(rawVersioningConfigurationSet) > 0 {
		rawVersioningConfiguration := rawVersioningConfigurationSet[0].(map[string]interface{})
		params.VersioningConfiguration = parseVersioningConfiguration(rawVersioningConfiguration)
	}

	log.Printf("[DEBUG] Creating IoT Analytics Dataset: %s", params)
	_, err := conn.UpdateDataset(params)

	if err != nil {
		return err
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.IotanalyticsUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsIotAnalyticsDatasetRead(d, meta)

}

func resourceAwsIotAnalyticsDatasetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.DeleteDatasetInput{
		DatasetName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting IoT Analytics Dataset: %s", params)
	_, err := conn.DeleteDataset(params)

	return err
}
