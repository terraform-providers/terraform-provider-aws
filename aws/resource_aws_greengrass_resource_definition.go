package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func generateLocalDeviceResourceDataSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"source_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"group_owner_setting": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_add_group_owner": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"group_owner": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func generateLocalVolumeResourceDataSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"source_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"group_owner_setting": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_add_group_owner": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"group_owner": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func generateS3MachineLearningModelResourceDataSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"s3_uri": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func generateSageMakerMachineLearningModelResourceDataSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"sagemaker_job_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"destination_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func generateSecretsManagerSecretResourceDataSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"secret_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"additional_staging_labels_to_download": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceAwsGreengrassResourceDefinition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGreengrassResourceDefinitionCreate,
		Read:   resourceAwsGreengrassResourceDefinitionRead,
		Update: resourceAwsGreengrassResourceDefinitionUpdate,
		Delete: resourceAwsGreengrassResourceDefinitionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"amzn_client_token": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"tags": tagsSchema(),
			"latest_definition_version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_definition_version": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"data_container": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"local_device_resource_data": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem:     generateLocalDeviceResourceDataSchema(),
												},
												"local_volume_resource_data": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem:     generateLocalVolumeResourceDataSchema(),
												},
												"s3_machine_learning_model_resource_data": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem:     generateS3MachineLearningModelResourceDataSchema(),
												},
												"sagemaker_machine_learning_model_resource_data": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem:     generateSageMakerMachineLearningModelResourceDataSchema(),
												},
												"secrets_manager_secret_resource_data": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem:     generateSecretsManagerSecretResourceDataSchema(),
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
		},
	}
}

func parseGroupOwnerSettings(rawData map[string]interface{}) *greengrass.GroupOwnerSetting {
	groupOwnerSetting := &greengrass.GroupOwnerSetting{
		AutoAddGroupOwner: aws.Bool(rawData["auto_add_group_owner"].(bool)),
	}

	if v, ok := rawData["group_owner"]; ok {
		groupOwnerSetting.GroupOwner = aws.String(v.(string))
	}

	return groupOwnerSetting
}

func parseLocalDeviceResourceData(rawData map[string]interface{}) *greengrass.LocalDeviceResourceData {
	localDeviceResourceData := &greengrass.LocalDeviceResourceData{}
	if v, ok := rawData["source_path"]; ok {
		localDeviceResourceData.SourcePath = aws.String(v.(string))
	}

	if v := rawData["group_owner_setting"].([]interface{}); len(v) > 0 {
		localDeviceResourceData.GroupOwnerSetting = parseGroupOwnerSettings(v[0].(map[string]interface{}))
	}

	return localDeviceResourceData
}

func parseLocalVolumeResourceData(rawData map[string]interface{}) *greengrass.LocalVolumeResourceData {
	localVolumeResourceData := &greengrass.LocalVolumeResourceData{}
	if v, ok := rawData["source_path"]; ok {
		localVolumeResourceData.SourcePath = aws.String(v.(string))
	}

	if v, ok := rawData["destination_path"]; ok {
		localVolumeResourceData.DestinationPath = aws.String(v.(string))
	}

	if v := rawData["group_owner_setting"].([]interface{}); len(v) > 0 {
		localVolumeResourceData.GroupOwnerSetting = parseGroupOwnerSettings(v[0].(map[string]interface{}))
	}

	return localVolumeResourceData

}

func parseS3MachineLearningModelResourceData(rawData map[string]interface{}) *greengrass.S3MachineLearningModelResourceData {
	s3MachineLearningModelResourceData := &greengrass.S3MachineLearningModelResourceData{}
	if v, ok := rawData["s3_uri"]; ok {
		s3MachineLearningModelResourceData.S3Uri = aws.String(v.(string))
	}
	if v, ok := rawData["destination_path"]; ok {
		s3MachineLearningModelResourceData.DestinationPath = aws.String(v.(string))
	}

	return s3MachineLearningModelResourceData
}

func parseSageMakerMachineLearningModelResourceData(rawData map[string]interface{}) *greengrass.SageMakerMachineLearningModelResourceData {
	sageMakerMachineLearningModelResourceData := &greengrass.SageMakerMachineLearningModelResourceData{}
	if v, ok := rawData["sagemaker_job_arn"]; ok {
		sageMakerMachineLearningModelResourceData.SageMakerJobArn = aws.String(v.(string))
	}
	if v, ok := rawData["destination_path"]; ok {
		sageMakerMachineLearningModelResourceData.DestinationPath = aws.String(v.(string))
	}

	return sageMakerMachineLearningModelResourceData
}

func parseSecretsManagerSecretResourceData(rawData map[string]interface{}) *greengrass.SecretsManagerSecretResourceData {
	secretsManagerSecretResourceData := &greengrass.SecretsManagerSecretResourceData{}
	if v, ok := rawData["secret_arn"]; ok {
		secretsManagerSecretResourceData.ARN = aws.String(v.(string))
	}

	downloadLabels := make([]*string, 0)
	for _, rawLabel := range rawData["additional_staging_labels_to_download"].([]interface{}) {
		label := rawLabel.(string)
		downloadLabels = append(downloadLabels, &label)
	}
	secretsManagerSecretResourceData.AdditionalStagingLabelsToDownload = downloadLabels

	return secretsManagerSecretResourceData
}

func parseResourceDataContainer(rawData map[string]interface{}) *greengrass.ResourceDataContainer {
	params := &greengrass.ResourceDataContainer{}

	if v := rawData["local_device_resource_data"].([]interface{}); len(v) > 0 {
		params.LocalDeviceResourceData = parseLocalDeviceResourceData(v[0].(map[string]interface{}))
	}

	if v := rawData["local_volume_resource_data"].([]interface{}); len(v) > 0 {
		params.LocalVolumeResourceData = parseLocalVolumeResourceData(v[0].(map[string]interface{}))
	}

	if v := rawData["s3_machine_learning_model_resource_data"].([]interface{}); len(v) > 0 {
		params.S3MachineLearningModelResourceData = parseS3MachineLearningModelResourceData(v[0].(map[string]interface{}))
	}

	if v := rawData["sagemaker_machine_learning_model_resource_data"].([]interface{}); len(v) > 0 {
		params.SageMakerMachineLearningModelResourceData = parseSageMakerMachineLearningModelResourceData(v[0].(map[string]interface{}))
	}

	if v := rawData["secrets_manager_secret_resource_data"].([]interface{}); len(v) > 0 {
		params.SecretsManagerSecretResourceData = parseSecretsManagerSecretResourceData(v[0].(map[string]interface{}))
	}

	return params

}

func createResourceDefinitionVersion(d *schema.ResourceData, conn *greengrass.Greengrass) error {
	var rawData map[string]interface{}
	if v := d.Get("resource_definition_version").(*schema.Set).List(); len(v) == 0 {
		return nil
	} else {
		rawData = v[0].(map[string]interface{})
	}

	params := &greengrass.CreateResourceDefinitionVersionInput{
		ResourceDefinitionId: aws.String(d.Id()),
	}

	if v := d.Get("amzn_client_token").(string); v != "" {
		params.AmznClientToken = aws.String(v)
	}

	resources := make([]*greengrass.Resource, 0)
	for _, resourceToCast := range rawData["resource"].(*schema.Set).List() {
		rawResource := resourceToCast.(map[string]interface{})
		resource := &greengrass.Resource{
			Id:   aws.String(rawResource["id"].(string)),
			Name: aws.String(rawResource["name"].(string)),
		}
		if v := rawResource["data_container"].([]interface{}); len(v) > 0 {
			resource.ResourceDataContainer = parseResourceDataContainer(v[0].(map[string]interface{}))
		}
		resources = append(resources, resource)
	}
	params.Resources = resources

	log.Printf("[DEBUG] Creating Greengrass Resource Definition Version: %s", params)
	_, err := conn.CreateResourceDefinitionVersion(params)

	if err != nil {
		return err
	}

	return nil
}

func resourceAwsGreengrassResourceDefinitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.CreateResourceDefinitionInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if rawTags := d.Get("tags").(map[string]interface{}); len(rawTags) > 0 {
		params.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GreengrassTags()
	}

	log.Printf("[DEBUG] Creating Greengrass Resource Definition: %s", params)
	out, err := conn.CreateResourceDefinition(params)
	if err != nil {
		return err
	}

	d.SetId(*out.Id)

	err = createResourceDefinitionVersion(d, conn)

	if err != nil {
		return err
	}

	return resourceAwsGreengrassResourceDefinitionRead(d, meta)
}

func flattenGroupOwnerSetting(groupOwnerSetting *greengrass.GroupOwnerSetting) map[string]interface{} {
	rawData := make(map[string]interface{})
	if groupOwnerSetting.AutoAddGroupOwner != nil {
		rawData["auto_add_group_owner"] = aws.BoolValue(groupOwnerSetting.AutoAddGroupOwner)
	}

	if groupOwnerSetting.GroupOwner != nil {
		rawData["group_owner"] = aws.StringValue(groupOwnerSetting.GroupOwner)
	}

	return rawData
}

func flattenLocalDeviceResourceData(localDeviceResourceData *greengrass.LocalDeviceResourceData) map[string]interface{} {
	rawData := make(map[string]interface{})
	if localDeviceResourceData.GroupOwnerSetting != nil {
		rawData["group_owner_setting"] = []map[string]interface{}{flattenGroupOwnerSetting(localDeviceResourceData.GroupOwnerSetting)}
	}

	if localDeviceResourceData.SourcePath != nil {
		rawData["source_path"] = aws.StringValue(localDeviceResourceData.SourcePath)
	}

	return rawData
}

func flattenLocalVolumeResourceData(localVolumeResourceData *greengrass.LocalVolumeResourceData) map[string]interface{} {
	rawData := make(map[string]interface{})
	if localVolumeResourceData.GroupOwnerSetting != nil {
		rawData["group_owner_setting"] = []map[string]interface{}{flattenGroupOwnerSetting(localVolumeResourceData.GroupOwnerSetting)}
	}

	if localVolumeResourceData.SourcePath != nil {
		rawData["source_path"] = aws.StringValue(localVolumeResourceData.SourcePath)
	}

	if localVolumeResourceData.DestinationPath != nil {
		rawData["destination_path"] = aws.StringValue(localVolumeResourceData.DestinationPath)
	}

	return rawData
}

func flattenS3MachineLearningModelResourceData(s3MachineLearningModelResourceData *greengrass.S3MachineLearningModelResourceData) map[string]interface{} {
	rawData := make(map[string]interface{})
	if s3MachineLearningModelResourceData.DestinationPath != nil {
		rawData["destination_path"] = aws.StringValue(s3MachineLearningModelResourceData.DestinationPath)
	}

	if s3MachineLearningModelResourceData.S3Uri != nil {
		rawData["s3_uri"] = aws.StringValue(s3MachineLearningModelResourceData.S3Uri)
	}

	return rawData
}

func flattenSageMakerMachineLearningModelResourceData(sageMakerMachineLearningModelResourceData *greengrass.SageMakerMachineLearningModelResourceData) map[string]interface{} {
	rawData := make(map[string]interface{})
	if sageMakerMachineLearningModelResourceData.DestinationPath != nil {
		rawData["destination_path"] = aws.StringValue(sageMakerMachineLearningModelResourceData.DestinationPath)
	}

	if sageMakerMachineLearningModelResourceData.SageMakerJobArn != nil {
		rawData["sagemaker_job_arn"] = aws.StringValue(sageMakerMachineLearningModelResourceData.SageMakerJobArn)
	}

	return rawData
}

func flattenSecretsManagerSecretResourceData(secretsManagerSecretResourceData *greengrass.SecretsManagerSecretResourceData) map[string]interface{} {
	rawData := make(map[string]interface{})
	if secretsManagerSecretResourceData.ARN != nil {
		rawData["secret_arn"] = aws.StringValue(secretsManagerSecretResourceData.ARN)
	}

	if secretsManagerSecretResourceData.AdditionalStagingLabelsToDownload != nil {
		rawDownloadLabels := make([]string, 0)
		for _, label := range secretsManagerSecretResourceData.AdditionalStagingLabelsToDownload {
			rawDownloadLabels = append(rawDownloadLabels, aws.StringValue(label))
		}
		rawData["additional_staging_labels_to_download"] = rawDownloadLabels
	}

	return rawData
}

func flattenResourceDataContainer(resourceDataContainer *greengrass.ResourceDataContainer) map[string]interface{} {
	rawData := make(map[string]interface{})
	if v := resourceDataContainer.LocalDeviceResourceData; v != nil {
		rawData["local_device_resource_data"] = []map[string]interface{}{flattenLocalDeviceResourceData(v)}
	}

	if v := resourceDataContainer.LocalVolumeResourceData; v != nil {
		rawData["local_volume_resource_data"] = []map[string]interface{}{flattenLocalVolumeResourceData(v)}
	}

	if v := resourceDataContainer.S3MachineLearningModelResourceData; v != nil {
		rawData["s3_machine_learning_model_resource_data"] = []map[string]interface{}{flattenS3MachineLearningModelResourceData(v)}
	}

	if v := resourceDataContainer.SageMakerMachineLearningModelResourceData; v != nil {
		rawData["sagemaker_machine_learning_model_resource_data"] = []map[string]interface{}{flattenSageMakerMachineLearningModelResourceData(v)}
	}

	if v := resourceDataContainer.SecretsManagerSecretResourceData; v != nil {
		rawData["secrets_manager_secret_resource_data"] = []map[string]interface{}{flattenSecretsManagerSecretResourceData(v)}
	}

	return rawData
}

func setResourceDefinitionVersion(latestVersion string, d *schema.ResourceData, conn *greengrass.Greengrass) error {
	params := &greengrass.GetResourceDefinitionVersionInput{
		ResourceDefinitionId:        aws.String(d.Id()),
		ResourceDefinitionVersionId: aws.String(latestVersion),
	}

	out, err := conn.GetResourceDefinitionVersion(params)

	if err != nil {
		return err
	}

	rawVersion := make(map[string]interface{})
	d.Set("latest_definition_version_arn", *out.Arn)

	rawResourceList := make([]map[string]interface{}, 0)
	for _, resource := range out.Definition.Resources {
		rawResource := make(map[string]interface{})
		rawResource["id"] = *resource.Id
		rawResource["name"] = *resource.Name
		rawResource["data_container"] = []map[string]interface{}{flattenResourceDataContainer(resource.ResourceDataContainer)}
		rawResourceList = append(rawResourceList, rawResource)
	}

	rawVersion["resource"] = rawResourceList

	d.Set("resource_definition_version", []map[string]interface{}{rawVersion})

	return nil
}

func resourceAwsGreengrassResourceDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.GetResourceDefinitionInput{
		ResourceDefinitionId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading Greengrass Resource Definition: %s", params)
	out, err := conn.GetResourceDefinition(params)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Received Greengrass Resource Definition: %s", out)

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)

	arn := *out.Arn
	tags, err := keyvaluetags.GreengrassListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}
	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if out.LatestVersion != nil {
		err = setResourceDefinitionVersion(*out.LatestVersion, d, conn)

		if err != nil {
			return err
		}
	}

	return nil
}

func resourceAwsGreengrassResourceDefinitionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.UpdateResourceDefinitionInput{
		Name:                 aws.String(d.Get("name").(string)),
		ResourceDefinitionId: aws.String(d.Id()),
	}

	_, err := conn.UpdateResourceDefinition(params)
	if err != nil {
		return err
	}

	if d.HasChange("resource_definition_version") {
		err = createResourceDefinitionVersion(d, conn)
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.GreengrassUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsGreengrassResourceDefinitionRead(d, meta)
}

func resourceAwsGreengrassResourceDefinitionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.DeleteResourceDefinitionInput{
		ResourceDefinitionId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting Greengrass Resource Definition: %s", params)

	_, err := conn.DeleteResourceDefinition(params)

	if err != nil {
		return err
	}

	return nil
}
