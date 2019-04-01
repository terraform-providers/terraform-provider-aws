package aws

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsCodePipelineCustomActionType() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodePipelineCustomActionTypeCreate,
		Read:   resourceAwsCodePipelineCustomActionTypeRead,
		Delete: resourceAwsCodePipelineCustomActionTypeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"category": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					codepipeline.ActionCategorySource,
					codepipeline.ActionCategoryBuild,
					codepipeline.ActionCategoryDeploy,
					codepipeline.ActionCategoryTest,
					codepipeline.ActionCategoryInvoke,
					codepipeline.ActionCategoryApproval,
				}, false),
			},
			"configuration_properties": {
				Type:     schema.TypeList,
				MaxItems: 10,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"key": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"queryable": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"required": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"secret": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								codepipeline.ActionConfigurationPropertyTypeString,
								codepipeline.ActionConfigurationPropertyTypeNumber,
								codepipeline.ActionConfigurationPropertyTypeBoolean,
							}, false),
						},
					},
				},
			},
			"input_artifact_details": {
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maximum_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 5),
						},
						"minimum_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 5),
						},
					},
				},
			},
			"output_artifact_details": {
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maximum_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 5),
						},
						"minimum_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 5),
						},
					},
				},
			},
			"provider_name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 25),
			},
			"settings": {
				Type:     schema.TypeList,
				ForceNew: true,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entity_url_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"execution_url_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"revision_url_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"third_party_configuration_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"version": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 9),
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCodePipelineCustomActionTypeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codepipelineconn

	input := &codepipeline.CreateCustomActionTypeInput{
		Category:              aws.String(d.Get("category").(string)),
		InputArtifactDetails:  expandAwsCodePipelineArtifactDetails(d.Get("input_artifact_details").([]interface{})[0].(map[string]interface{})),
		OutputArtifactDetails: expandAwsCodePipelineArtifactDetails(d.Get("output_artifact_details").([]interface{})[0].(map[string]interface{})),
		Provider:              aws.String(d.Get("provider_name").(string)),
		Version:               aws.String(d.Get("version").(string)),
	}

	confProps := d.Get("configuration_properties").([]interface{})
	if len(confProps) > 0 {
		input.ConfigurationProperties = expandAwsCodePipelineActionConfigurationProperty(confProps)
	}

	settings := d.Get("settings").([]interface{})
	if len(settings) > 0 {
		input.Settings = expandAwsCodePipelineActionTypeSettings(settings[0].(map[string]interface{}))
	}

	resp, err := conn.CreateCustomActionType(input)
	if err != nil {
		return fmt.Errorf("Error creating CodePipeline CustomActionType: %s", err)
	}

	if resp.ActionType == nil || resp.ActionType.Id == nil ||
		resp.ActionType.Id.Owner == nil || resp.ActionType.Id.Category == nil || resp.ActionType.Id.Provider == nil || resp.ActionType.Id.Version == nil {
		return errors.New("Error creating CodePipeline CustomActionType: invalid response from AWS")
	}
	d.SetId(fmt.Sprintf("%s:%s:%s:%s", *resp.ActionType.Id.Owner, *resp.ActionType.Id.Category, *resp.ActionType.Id.Provider, *resp.ActionType.Id.Version))
	return resourceAwsCodePipelineCustomActionTypeRead(d, meta)
}

func resourceAwsCodePipelineCustomActionTypeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codepipelineconn
	owner, category, provider, version, err := decodeAwsCodePipelineCustomActionTypeId(d.Id())
	if err != nil {
		return fmt.Errorf("Error reading CodePipeline CustomActionType: %s", err)
	}

	actionType, err := lookAwsCodePipelineCustomActionType(conn, d.Id())
	if err != nil {
		return fmt.Errorf("Error reading CodePipeline CustomActionType: %s", err)
	}
	if actionType == nil {
		log.Printf("[INFO] Codepipeline CustomActionType %q not found", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("owner", owner)
	d.Set("category", category)
	d.Set("provider_name", provider)
	d.Set("version", version)

	if err := d.Set("configuration_properties", flattenAwsCodePipelineActionConfigurationProperty(actionType.ActionConfigurationProperties)); err != nil {
		return fmt.Errorf("error setting configuration_properties: %s", err)
	}

	if err := d.Set("input_artifact_details", flattenAwsCodePipelineArtifactDetails(actionType.InputArtifactDetails)); err != nil {
		return fmt.Errorf("error setting input_artifact_details: %s", err)
	}

	if err := d.Set("output_artifact_details", flattenAwsCodePipelineArtifactDetails(actionType.OutputArtifactDetails)); err != nil {
		return fmt.Errorf("error setting output_artifact_details: %s", err)
	}

	if err := d.Set("settings", flattenAwsCodePipelineActionTypeSettings(actionType.Settings)); err != nil {
		return fmt.Errorf("error setting settings: %s", err)
	}

	return nil
}

func resourceAwsCodePipelineCustomActionTypeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codepipelineconn

	_, category, provider, version, err := decodeAwsCodePipelineCustomActionTypeId(d.Id())
	if err != nil {
		return fmt.Errorf("Error deleting CodePipeline CustomActionType: %s", err)
	}

	input := &codepipeline.DeleteCustomActionTypeInput{
		Category: aws.String(category),
		Provider: aws.String(provider),
		Version:  aws.String(version),
	}

	_, err = conn.DeleteCustomActionType(input)
	if err != nil {
		if isAWSErr(err, codepipeline.ErrCodeActionTypeNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting CodePipeline CustomActionType (%s): %s", d.Id(), err)
	}

	return nil
}

func lookAwsCodePipelineCustomActionType(conn *codepipeline.CodePipeline, id string) (*codepipeline.ActionType, error) {
	owner, category, provider, version, err := decodeAwsCodePipelineCustomActionTypeId(id)
	if err != nil {
		return nil, fmt.Errorf("Error reading CodePipeline CustomActionType: %s", err)
	}

	var actionType *codepipeline.ActionType

	input := &codepipeline.ListActionTypesInput{
		ActionOwnerFilter: aws.String(owner),
	}
	for {
		resp, err := conn.ListActionTypes(input)
		if err != nil {
			return nil, fmt.Errorf("Error reading CodePipeline CustomActionType: %s", err)
		}

		for _, v := range resp.ActionTypes {
			if atid := v.Id; atid != nil {
				if atid.Category == nil || atid.Provider == nil || atid.Version == nil {
					continue
				}
				if *atid.Category == category && *atid.Provider == provider && *atid.Version == version {
					actionType = v
					break
				}
			}
		}

		if actionType != nil {
			break
		}

		if resp.NextToken == nil {
			break
		} else {
			input.NextToken = resp.NextToken
		}
	}

	return actionType, nil
}

func decodeAwsCodePipelineCustomActionTypeId(id string) (owner string, category string, provider string, version string, e error) {
	ss := strings.Split(id, ":")
	if len(ss) != 4 {
		e = fmt.Errorf("invalid AwsCodePipelineCustomActionType ID: %s", id)
		return
	}
	owner, category, provider, version = ss[0], ss[1], ss[2], ss[3]
	return
}

func expandAwsCodePipelineArtifactDetails(d map[string]interface{}) *codepipeline.ArtifactDetails {
	return &codepipeline.ArtifactDetails{
		MaximumCount: aws.Int64(int64(d["maximum_count"].(int))),
		MinimumCount: aws.Int64(int64(d["minimum_count"].(int))),
	}
}

func flattenAwsCodePipelineArtifactDetails(ad *codepipeline.ArtifactDetails) []map[string]interface{} {
	m := make(map[string]interface{})

	m["maximum_count"] = aws.Int64Value(ad.MaximumCount)
	m["minimum_count"] = aws.Int64Value(ad.MinimumCount)

	return []map[string]interface{}{m}
}

func expandAwsCodePipelineActionConfigurationProperty(d []interface{}) []*codepipeline.ActionConfigurationProperty {
	if len(d) == 0 {
		return nil
	}
	result := make([]*codepipeline.ActionConfigurationProperty, 0, len(d))

	for _, v := range d {
		m := v.(map[string]interface{})
		acp := &codepipeline.ActionConfigurationProperty{
			Key:      aws.Bool(m["key"].(bool)),
			Name:     aws.String(m["name"].(string)),
			Required: aws.Bool(m["required"].(bool)),
			Secret:   aws.Bool(m["secret"].(bool)),
		}
		if raw, ok := m["description"]; ok && raw.(string) != "" {
			acp.Description = aws.String(raw.(string))
		}
		if raw, ok := m["queryable"]; ok {
			acp.Queryable = aws.Bool(raw.(bool))
		}
		if raw, ok := m["type"]; ok && raw.(string) != "" {
			acp.Type = aws.String(raw.(string))
		}
		result = append(result, acp)
	}

	return result
}

func flattenAwsCodePipelineActionConfigurationProperty(acps []*codepipeline.ActionConfigurationProperty) []interface{} {
	result := make([]interface{}, 0, len(acps))

	for _, acp := range acps {
		m := map[string]interface{}{}
		m["description"] = aws.StringValue(acp.Description)
		m["key"] = aws.BoolValue(acp.Key)
		m["name"] = aws.StringValue(acp.Name)
		m["queryable"] = aws.BoolValue(acp.Queryable)
		m["required"] = aws.BoolValue(acp.Required)
		m["secret"] = aws.BoolValue(acp.Secret)
		// Currently AWS doesn't return type
		if acp.Type != nil {
			m["type"] = aws.StringValue(acp.Type)
		}
		result = append(result, m)
	}
	return result
}

func expandAwsCodePipelineActionTypeSettings(d map[string]interface{}) *codepipeline.ActionTypeSettings {
	if len(d) == 0 {
		return nil
	}
	result := &codepipeline.ActionTypeSettings{}

	if raw, ok := d["entity_url_template"]; ok && raw.(string) != "" {
		result.EntityUrlTemplate = aws.String(raw.(string))
	}
	if raw, ok := d["execution_url_template"]; ok && raw.(string) != "" {
		result.ExecutionUrlTemplate = aws.String(raw.(string))
	}
	if raw, ok := d["revision_url_template"]; ok && raw.(string) != "" {
		result.RevisionUrlTemplate = aws.String(raw.(string))
	}
	if raw, ok := d["third_party_configuration_url"]; ok && raw.(string) != "" {
		result.ThirdPartyConfigurationUrl = aws.String(raw.(string))
	}

	return result
}

func flattenAwsCodePipelineActionTypeSettings(settings *codepipeline.ActionTypeSettings) []map[string]interface{} {
	m := make(map[string]interface{})

	if settings.EntityUrlTemplate != nil {
		m["entity_url_template"] = aws.StringValue(settings.EntityUrlTemplate)
	}
	if settings.ExecutionUrlTemplate != nil {
		m["execution_url_template"] = aws.StringValue(settings.ExecutionUrlTemplate)
	}
	if settings.RevisionUrlTemplate != nil {
		m["revision_url_template"] = aws.StringValue(settings.RevisionUrlTemplate)
	}
	if settings.ThirdPartyConfigurationUrl != nil {
		m["third_party_configuration_url"] = aws.StringValue(settings.ThirdPartyConfigurationUrl)
	}

	if len(m) == 0 {
		return nil
	}
	return []map[string]interface{}{m}
}
