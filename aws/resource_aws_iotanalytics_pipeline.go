package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func generateAddAttributesActivity() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"attributes": {
				Type:     schema.TypeMap,
				Required: true,
				Elem:     schema.TypeString,
			},
		},
	}
}

func generateRemoveAttributesActivity() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"attributes": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 50,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func generateSelectAttributesActivity() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"attributes": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 50,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func generateChannelActivity() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"channel_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func generateDatastoreActivity() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"datastore_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func generateDeviceRegistryEnrichActivity() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"thing_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"attribute": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func generateDeviceShadowEnrichActivity() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"thing_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"attribute": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func generateFilterActivity() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"filter": {
				Type:     schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func generateLambdaActivity() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"lambda_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"batch_size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"next_activity": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func generateMathActivity() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"math": {
				Type:     schema.TypeString,
				Required: true,
			},
			"attribute": {
				Type:     schema.TypeString,
				Required: true,
			},
			"next_activity": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func generatePipelineActivitySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"add_attributes": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     generateAddAttributesActivity(),
			},
			"remove_attributes": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     generateRemoveAttributesActivity(),
			},
			"select_attributes": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     generateSelectAttributesActivity(),
			},
			"channel": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     generateChannelActivity(),
			},
			"datastore": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     generateDatastoreActivity(),
			},
			"device_registry_enrich": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     generateDeviceRegistryEnrichActivity(),
			},
			"device_shadow_enrich": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     generateDeviceShadowEnrichActivity(),
			},
			"filter": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     generateFilterActivity(),
			},
			"lambda": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     generateLambdaActivity(),
			},
			"math": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem:     generateMathActivity(),
			},
		},
	}
}

func resourceAwsIotAnalyticsPipeline() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotAnalyticsPipelineCreate,
		Read:   resourceAwsIotAnalyticsPipelineRead,
		Update: resourceAwsIotAnalyticsPipelineUpdate,
		Delete: resourceAwsIotAnalyticsPipelineDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"pipeline_activity": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 25,
				Elem:     generatePipelineActivitySchema(),
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func parseAddAttributesActivity(rawAddAttributesActivity map[string]interface{}) *iotanalytics.AddAttributesActivity {
	attributes := make(map[string]*string)
	rawAttributes := rawAddAttributesActivity["attributes"].(map[string]interface{})
	for k, v := range rawAttributes {
		attributes[k] = aws.String(v.(string))
	}

	name := rawAddAttributesActivity["name"].(string)
	addAttrActivity := &iotanalytics.AddAttributesActivity{
		Name:       aws.String(name),
		Attributes: attributes,
	}

	if v, ok := rawAddAttributesActivity["next_activity"].(string); ok && v != "" {
		addAttrActivity.Next = aws.String(v)
	}

	return addAttrActivity
}

func parseChannelActivity(rawChannelActivity map[string]interface{}) *iotanalytics.ChannelActivity {
	name := rawChannelActivity["name"].(string)
	channelName := rawChannelActivity["channel_name"].(string)

	channelActivity := &iotanalytics.ChannelActivity{
		Name:        aws.String(name),
		ChannelName: aws.String(channelName),
	}

	if v, ok := rawChannelActivity["next_activity"].(string); ok && v != "" {
		channelActivity.Next = aws.String(v)
	}

	return channelActivity
}

func parseDatastoreActivity(rawDatastoreActivity map[string]interface{}) *iotanalytics.DatastoreActivity {
	name := rawDatastoreActivity["name"].(string)
	datastoreName := rawDatastoreActivity["datastore_name"].(string)

	datastore := &iotanalytics.DatastoreActivity{
		Name:          aws.String(name),
		DatastoreName: aws.String(datastoreName),
	}

	return datastore
}

func parseDeviceRegistryEnrichActivity(rawDeviceRegistryEnrichActivity map[string]interface{}) *iotanalytics.DeviceRegistryEnrichActivity {
	name := rawDeviceRegistryEnrichActivity["name"].(string)
	attribute := rawDeviceRegistryEnrichActivity["attribute"].(string)
	roleArn := rawDeviceRegistryEnrichActivity["role_arn"].(string)
	thingName := rawDeviceRegistryEnrichActivity["thing_name"].(string)

	deviceRegistryEnrich := &iotanalytics.DeviceRegistryEnrichActivity{
		Name:      aws.String(name),
		Attribute: aws.String(attribute),
		RoleArn:   aws.String(roleArn),
		ThingName: aws.String(thingName),
	}

	if v, ok := rawDeviceRegistryEnrichActivity["next_activity"].(string); ok && v != "" {
		deviceRegistryEnrich.Next = aws.String(v)
	}

	return deviceRegistryEnrich
}

func parseDeviceShadowEnrichActivity(rawDeviceShadowEnrichActivity map[string]interface{}) *iotanalytics.DeviceShadowEnrichActivity {
	name := rawDeviceShadowEnrichActivity["name"].(string)
	attribute := rawDeviceShadowEnrichActivity["attribute"].(string)
	roleArn := rawDeviceShadowEnrichActivity["role_arn"].(string)
	thingName := rawDeviceShadowEnrichActivity["thing_name"].(string)

	deviceShadowEnrich := &iotanalytics.DeviceShadowEnrichActivity{
		Name:      aws.String(name),
		Attribute: aws.String(attribute),
		RoleArn:   aws.String(roleArn),
		ThingName: aws.String(thingName),
	}

	if v, ok := rawDeviceShadowEnrichActivity["next_activity"].(string); ok && v != "" {
		deviceShadowEnrich.Next = aws.String(v)
	}

	return deviceShadowEnrich
}

func parseFilterActivity(rawFilterActivity map[string]interface{}) *iotanalytics.FilterActivity {
	name := rawFilterActivity["name"].(string)
	filter := rawFilterActivity["filter"].(string)

	filterActivity := &iotanalytics.FilterActivity{
		Name:   aws.String(name),
		Filter: aws.String(filter),
	}

	if v, ok := rawFilterActivity["next_activity"].(string); ok && v != "" {
		filterActivity.Next = aws.String(v)
	}

	return filterActivity
}

func parseLambdaActivity(rawLambdaActivity map[string]interface{}) *iotanalytics.LambdaActivity {
	name := rawLambdaActivity["name"].(string)
	lambdaName := rawLambdaActivity["lambda_name"].(string)
	batchSize := int64(rawLambdaActivity["batch_size"].(int))

	lambdaActivity := &iotanalytics.LambdaActivity{
		Name:       aws.String(name),
		LambdaName: aws.String(lambdaName),
		BatchSize:  aws.Int64(batchSize),
	}

	if v, ok := rawLambdaActivity["next_activity"].(string); ok && v != "" {
		lambdaActivity.Next = aws.String(v)
	}

	return lambdaActivity
}

func parseMathActivity(rawMathActivity map[string]interface{}) *iotanalytics.MathActivity {
	name := rawMathActivity["name"].(string)
	attribute := rawMathActivity["attribute"].(string)
	math := rawMathActivity["math"].(string)

	mathActivity := &iotanalytics.MathActivity{
		Name:      aws.String(name),
		Attribute: aws.String(attribute),
		Math:      aws.String(math),
	}

	if v, ok := rawMathActivity["next_activity"].(string); ok && v != "" {
		mathActivity.Next = aws.String(v)
	}

	return mathActivity
}

func parseRemoveAttributesActivity(rawRemoveAttributesActivity map[string]interface{}) *iotanalytics.RemoveAttributesActivity {
	attributes := make([]*string, 0)
	rawAttributes := rawRemoveAttributesActivity["attributes"].([]interface{})
	for _, attrName := range rawAttributes {
		attributes = append(attributes, aws.String(attrName.(string)))
	}

	name := rawRemoveAttributesActivity["name"].(string)

	removeAttrActivity := &iotanalytics.RemoveAttributesActivity{
		Name:       aws.String(name),
		Attributes: attributes,
	}

	if v, ok := rawRemoveAttributesActivity["next_activity"].(string); ok && v != "" {
		removeAttrActivity.Next = aws.String(v)
	}

	return removeAttrActivity
}

func parseSelectAttributesActivity(rawSelectAttributesActivity map[string]interface{}) *iotanalytics.SelectAttributesActivity {
	attributes := make([]*string, 0)
	rawAttributes := rawSelectAttributesActivity["attributes"].([]interface{})
	for _, attrName := range rawAttributes {
		attributes = append(attributes, aws.String(attrName.(string)))
	}

	name := rawSelectAttributesActivity["name"].(string)
	selectAttrActivity := &iotanalytics.SelectAttributesActivity{
		Name:       aws.String(name),
		Attributes: attributes,
	}

	if v, ok := rawSelectAttributesActivity["next_activity"].(string); ok && v != "" {
		selectAttrActivity.Next = aws.String(v)
	}

	return selectAttrActivity
}

func parsePipelineActivity(rawPipelineActivity map[string]interface{}) *iotanalytics.PipelineActivity {
	pipelineActivity := &iotanalytics.PipelineActivity{}

	if v := rawPipelineActivity["add_attributes"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.AddAttributes = parseAddAttributesActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["select_attributes"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.SelectAttributes = parseSelectAttributesActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["remove_attributes"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.RemoveAttributes = parseRemoveAttributesActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["channel"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.Channel = parseChannelActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["datastore"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.Datastore = parseDatastoreActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["device_registry_enrich"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.DeviceRegistryEnrich = parseDeviceRegistryEnrichActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["device_shadow_enrich"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.DeviceShadowEnrich = parseDeviceShadowEnrichActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["filter"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.Filter = parseFilterActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["lambda"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.Lambda = parseLambdaActivity(v[0].(map[string]interface{}))
	}

	if v := rawPipelineActivity["math"].(*schema.Set).List(); len(v) != 0 {
		pipelineActivity.Math = parseMathActivity(v[0].(map[string]interface{}))
	}

	return pipelineActivity
}

func resourceAwsIotAnalyticsPipelineCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	rawPipelineActivites := d.Get("pipeline_activity").([]interface{})
	pipelineActivities := make([]*iotanalytics.PipelineActivity, 0)
	for _, rawPA := range rawPipelineActivites {
		pA := parsePipelineActivity(rawPA.(map[string]interface{}))
		pipelineActivities = append(pipelineActivities, pA)
	}

	pipelineName := d.Get("name").(string)
	params := &iotanalytics.CreatePipelineInput{
		PipelineName:       aws.String(pipelineName),
		PipelineActivities: pipelineActivities,
	}

	if tags := d.Get("tags").(map[string]interface{}); len(tags) > 0 {
		params.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().IotanalyticsTags()
	}

	_, err := conn.CreatePipeline(params)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Create IoT Analytics Pipeline: %s", params)

	d.SetId(pipelineName)

	return resourceAwsIotAnalyticsPipelineRead(d, meta)
}

func flattenAddAttributesActivity(addAttributesActivity *iotanalytics.AddAttributesActivity) map[string]interface{} {
	rawAddAttrActivity := make(map[string]interface{})
	rawAddAttrActivity["name"] = aws.StringValue(addAttributesActivity.Name)

	rawAttrs := make(map[string]interface{})
	for k, v := range addAttributesActivity.Attributes {
		rawAttrs[k] = aws.StringValue(v)
	}
	rawAddAttrActivity["attributes"] = rawAttrs

	if addAttributesActivity.Next != nil && *addAttributesActivity.Next != "" {
		rawAddAttrActivity["next_activity"] = aws.StringValue(addAttributesActivity.Next)
	}

	return rawAddAttrActivity
}

func flattenChannelActivity(channelActivity *iotanalytics.ChannelActivity) map[string]interface{} {

	rawChannelActivity := make(map[string]interface{})
	rawChannelActivity["name"] = aws.StringValue(channelActivity.Name)
	rawChannelActivity["channel_name"] = aws.StringValue(channelActivity.ChannelName)

	if channelActivity.Next != nil && *channelActivity.Next != "" {
		rawChannelActivity["next_activity"] = aws.StringValue(channelActivity.Next)
	}

	return rawChannelActivity
}

func flattenDatastoreActivity(datastoreActivity *iotanalytics.DatastoreActivity) map[string]interface{} {
	rawDatastoreActivity := make(map[string]interface{})
	rawDatastoreActivity["name"] = aws.StringValue(datastoreActivity.Name)
	rawDatastoreActivity["datastore_name"] = aws.StringValue(datastoreActivity.DatastoreName)
	return rawDatastoreActivity
}

func flattenDeviceRegistryEnrichActivity(deviceRegistryEnrichActivity *iotanalytics.DeviceRegistryEnrichActivity) map[string]interface{} {
	rawDeviceRegistryEnrichActivity := make(map[string]interface{})
	rawDeviceRegistryEnrichActivity["name"] = aws.StringValue(deviceRegistryEnrichActivity.Name)
	rawDeviceRegistryEnrichActivity["attribute"] = aws.StringValue(deviceRegistryEnrichActivity.Attribute)
	rawDeviceRegistryEnrichActivity["role_arn"] = aws.StringValue(deviceRegistryEnrichActivity.RoleArn)
	rawDeviceRegistryEnrichActivity["thing_name"] = aws.StringValue(deviceRegistryEnrichActivity.ThingName)

	if deviceRegistryEnrichActivity.Next != nil && *deviceRegistryEnrichActivity.Next != "" {
		rawDeviceRegistryEnrichActivity["next_activity"] = aws.StringValue(deviceRegistryEnrichActivity.Next)
	}

	return rawDeviceRegistryEnrichActivity
}

func flattenDeviceShadowEnrichActivity(deviceShadowEnrichActivity *iotanalytics.DeviceShadowEnrichActivity) map[string]interface{} {
	rawDeviceShadowEnrichActivity := make(map[string]interface{})
	rawDeviceShadowEnrichActivity["name"] = aws.StringValue(deviceShadowEnrichActivity.Name)
	rawDeviceShadowEnrichActivity["attribute"] = aws.StringValue(deviceShadowEnrichActivity.Attribute)
	rawDeviceShadowEnrichActivity["role_arn"] = aws.StringValue(deviceShadowEnrichActivity.RoleArn)
	rawDeviceShadowEnrichActivity["thing_name"] = aws.StringValue(deviceShadowEnrichActivity.ThingName)

	if deviceShadowEnrichActivity.Next != nil && *deviceShadowEnrichActivity.Next != "" {
		rawDeviceShadowEnrichActivity["next_activity"] = aws.StringValue(deviceShadowEnrichActivity.Next)
	}

	return rawDeviceShadowEnrichActivity
}

func flattenFilterActivity(filterActivity *iotanalytics.FilterActivity) map[string]interface{} {
	rawFilterActivity := make(map[string]interface{})
	rawFilterActivity["name"] = aws.StringValue(filterActivity.Name)
	rawFilterActivity["filter"] = aws.StringValue(filterActivity.Filter)

	if filterActivity.Next != nil && *filterActivity.Next != "" {
		rawFilterActivity["next_activity"] = aws.StringValue(filterActivity.Next)
	}

	return rawFilterActivity
}

func flattenLambdaActivity(lambdaActivity *iotanalytics.LambdaActivity) map[string]interface{} {
	rawLambdaActivity := make(map[string]interface{})
	rawLambdaActivity["name"] = aws.StringValue(lambdaActivity.Name)
	rawLambdaActivity["lambda_name"] = aws.StringValue(lambdaActivity.LambdaName)
	rawLambdaActivity["batch_size"] = aws.Int64Value(lambdaActivity.BatchSize)

	if lambdaActivity.Next != nil && *lambdaActivity.Next != "" {
		rawLambdaActivity["next_activity"] = aws.StringValue(lambdaActivity.Next)
	}

	return rawLambdaActivity
}

func flattenMathActivity(mathActivity *iotanalytics.MathActivity) map[string]interface{} {
	rawMathActivity := make(map[string]interface{})
	rawMathActivity["name"] = aws.StringValue(mathActivity.Name)
	rawMathActivity["attribute"] = aws.StringValue(mathActivity.Attribute)
	rawMathActivity["math"] = aws.StringValue(mathActivity.Math)

	if mathActivity.Next != nil && *mathActivity.Next != "" {
		rawMathActivity["next_activity"] = aws.StringValue(mathActivity.Next)
	}

	return rawMathActivity
}

func flattenRemoveAttributesActivity(removeAttributesActivity *iotanalytics.RemoveAttributesActivity) map[string]interface{} {
	rawRemoveAttrActivity := make(map[string]interface{})
	rawRemoveAttrActivity["name"] = aws.StringValue(removeAttributesActivity.Name)

	rawAttrs := make([]interface{}, 0)
	for _, v := range removeAttributesActivity.Attributes {
		rawAttrs = append(rawAttrs, aws.StringValue(v))
	}
	rawRemoveAttrActivity["attributes"] = rawAttrs

	if removeAttributesActivity.Next != nil && *removeAttributesActivity.Next != "" {
		rawRemoveAttrActivity["next_activity"] = aws.StringValue(removeAttributesActivity.Next)
	}

	return rawRemoveAttrActivity
}

func flattenSelectAttributesActivity(selectAttributesActivity *iotanalytics.SelectAttributesActivity) map[string]interface{} {
	rawSelectAttrActivity := make(map[string]interface{})
	rawSelectAttrActivity["name"] = aws.StringValue(selectAttributesActivity.Name)

	rawAttrs := make([]interface{}, 0)
	for _, v := range selectAttributesActivity.Attributes {
		rawAttrs = append(rawAttrs, aws.StringValue(v))
	}
	rawSelectAttrActivity["attributes"] = rawAttrs

	if selectAttributesActivity.Next != nil && *selectAttributesActivity.Next != "" {
		rawSelectAttrActivity["next_activity"] = aws.StringValue(selectAttributesActivity.Next)
	}

	return rawSelectAttrActivity
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

func flattenPipelineActivity(pipelineActivity *iotanalytics.PipelineActivity) map[string]interface{} {
	rawPipelineActivity := make(map[string]interface{})

	if pipelineActivity.AddAttributes != nil {
		rawActivity := flattenAddAttributesActivity(pipelineActivity.AddAttributes)
		rawPipelineActivity["add_attributes"] = wrapMapInList(rawActivity)
	}

	if pipelineActivity.RemoveAttributes != nil {
		rawActivity := flattenRemoveAttributesActivity(pipelineActivity.RemoveAttributes)
		rawPipelineActivity["remove_attributes"] = wrapMapInList(rawActivity)
	}

	if pipelineActivity.SelectAttributes != nil {
		rawActivity := flattenSelectAttributesActivity(pipelineActivity.SelectAttributes)
		rawPipelineActivity["select_attributes"] = wrapMapInList(rawActivity)
	}

	if pipelineActivity.Channel != nil {
		rawActivity := flattenChannelActivity(pipelineActivity.Channel)
		rawPipelineActivity["channel"] = wrapMapInList(rawActivity)
	}

	if pipelineActivity.Datastore != nil {
		rawActivity := flattenDatastoreActivity(pipelineActivity.Datastore)
		rawPipelineActivity["datastore"] = wrapMapInList(rawActivity)
	}

	if pipelineActivity.DeviceRegistryEnrich != nil {
		rawActivity := flattenDeviceRegistryEnrichActivity(pipelineActivity.DeviceRegistryEnrich)
		rawPipelineActivity["device_registry_enrich"] = wrapMapInList(rawActivity)
	}

	if pipelineActivity.DeviceShadowEnrich != nil {
		rawActivity := flattenDeviceShadowEnrichActivity(pipelineActivity.DeviceShadowEnrich)
		rawPipelineActivity["device_shadow_enrich"] = wrapMapInList(rawActivity)
	}

	if pipelineActivity.Filter != nil {
		rawActivity := flattenFilterActivity(pipelineActivity.Filter)
		rawPipelineActivity["filter"] = wrapMapInList(rawActivity)
	}

	if pipelineActivity.Lambda != nil {
		rawActivity := flattenLambdaActivity(pipelineActivity.Lambda)
		rawPipelineActivity["lambda"] = wrapMapInList(rawActivity)
	}

	if pipelineActivity.Math != nil {
		rawActivity := flattenMathActivity(pipelineActivity.Math)
		rawPipelineActivity["math"] = wrapMapInList(rawActivity)
	}

	return rawPipelineActivity
}

func resourceAwsIotAnalyticsPipelineRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.DescribePipelineInput{
		PipelineName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading IoT Analytics Pipeline: %s", params)
	out, err := conn.DescribePipeline(params)

	if err != nil {
		return err
	}

	d.Set("name", *out.Pipeline.Name)

	rawPipelineActivites := make([]map[string]interface{}, 0)
	for _, pa := range out.Pipeline.Activities {
		rawPipelineActivites = append(rawPipelineActivites, flattenPipelineActivity(pa))
	}
	d.Set("pipeline_activity", rawPipelineActivites)
	d.Set("arn", out.Pipeline.Arn)

	arn := *out.Pipeline.Arn
	tags, err := keyvaluetags.IotanalyticsListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsIotAnalyticsPipelineUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn
	pipelineName := d.Get("name").(string)

	rawPipelineActivites := d.Get("pipeline_activity").([]interface{})
	pipelineActivities := make([]*iotanalytics.PipelineActivity, 0)
	for _, rawPA := range rawPipelineActivites {
		pA := parsePipelineActivity(rawPA.(map[string]interface{}))
		pipelineActivities = append(pipelineActivities, pA)
	}

	params := &iotanalytics.UpdatePipelineInput{
		PipelineName:       aws.String(pipelineName),
		PipelineActivities: pipelineActivities,
	}

	_, err := conn.UpdatePipeline(params)

	if err != nil {
		return err
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.IotanalyticsUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	log.Printf("[DEBUG] Update IoT Analytics Pipeline: %s", params)

	return resourceAwsIotAnalyticsPipelineRead(d, meta)

}

func resourceAwsIotAnalyticsPipelineDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.DeletePipelineInput{
		PipelineName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting IoT Analytics Pipeline: %s", params)
	_, err := conn.DeletePipeline(params)

	return err
}
