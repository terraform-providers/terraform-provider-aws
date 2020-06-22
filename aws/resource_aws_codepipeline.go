package aws

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsCodePipeline() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodePipelineCreate,
		Read:   resourceAwsCodePipelineRead,
		Update: resourceAwsCodePipelineUpdate,
		Delete: resourceAwsCodePipelineDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"artifact_store": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"location": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								codepipeline.ArtifactStoreTypeS3,
							}, false),
						},
						"encryption_key": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											codepipeline.EncryptionKeyTypeKms,
										}, false),
									},
								},
							},
						},
						"region": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"stage": {
				Type:     schema.TypeList,
				MinItems: 2,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"action": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"configuration": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"category": {
										Type:     schema.TypeString,
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
									"owner": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											codepipeline.ActionOwnerAws,
											codepipeline.ActionOwnerThirdParty,
											codepipeline.ActionOwnerCustom,
										}, false),
									},
									"provider": {
										Type:     schema.TypeString,
										Required: true,
									},
									"version": {
										Type:     schema.TypeString,
										Required: true,
									},
									"input_artifacts": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"output_artifacts": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"role_arn": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"run_order": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"region": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"namespace": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsCodePipelineCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codepipelineconn

	pipeline, err := expandAwsCodePipeline(d)
	if err != nil {
		return err
	}
	params := &codepipeline.CreatePipelineInput{
		Pipeline: pipeline,
		Tags:     keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().CodepipelineTags(),
	}

	var resp *codepipeline.CreatePipelineOutput
	err = resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error

		resp, err = conn.CreatePipeline(params)

		if err != nil {
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err)
	})
	if isResourceTimeoutError(err) {
		resp, err = conn.CreatePipeline(params)
	}
	if err != nil {
		return fmt.Errorf("Error creating CodePipeline: %s", err)
	}
	if resp.Pipeline == nil {
		return fmt.Errorf("Error creating CodePipeline: invalid response from AWS")
	}

	d.SetId(aws.StringValue(resp.Pipeline.Name))

	return resourceAwsCodePipelineRead(d, meta)
}

func expandAwsCodePipeline(d *schema.ResourceData) (*codepipeline.PipelineDeclaration, error) {
	pipeline := codepipeline.PipelineDeclaration{
		Name:    aws.String(d.Get("name").(string)),
		RoleArn: aws.String(d.Get("role_arn").(string)),
		Stages:  expandAwsCodePipelineStages(d),
	}

	pipelineArtifactStores, err := expandAwsCodePipelineArtifactStores(d.Get("artifact_store").(*schema.Set).List())
	if err != nil {
		return nil, err
	}
	if len(pipelineArtifactStores) == 1 {
		for _, v := range pipelineArtifactStores {
			pipeline.ArtifactStore = v
		}
	} else {
		pipeline.ArtifactStores = pipelineArtifactStores
	}

	return &pipeline, nil
}

func expandAwsCodePipelineArtifactStores(configs []interface{}) (map[string]*codepipeline.ArtifactStore, error) {
	if len(configs) == 0 {
		return nil, nil
	}

	regions := make([]string, 0, len(configs))
	pipelineArtifactStores := make(map[string]*codepipeline.ArtifactStore)
	for _, config := range configs {
		region, store := expandAwsCodePipelineArtifactStoreData(config.(map[string]interface{}))
		regions = append(regions, region)
		pipelineArtifactStores[region] = store
	}

	if len(regions) == 1 {
		if regions[0] != "" {
			return nil, errors.New("region cannot be set for a single-region CodePipeline")
		}
	} else {
		for _, v := range regions {
			if v == "" {
				return nil, errors.New("region must be set for a cross-region CodePipeline")
			}
		}
		if len(configs) != len(pipelineArtifactStores) {
			return nil, errors.New("only one Artifact Store can be defined per region for a cross-region CodePipeline")
		}
	}

	return pipelineArtifactStores, nil
}

func expandAwsCodePipelineArtifactStoreData(data map[string]interface{}) (string, *codepipeline.ArtifactStore) {
	pipelineArtifactStore := codepipeline.ArtifactStore{
		Location: aws.String(data["location"].(string)),
		Type:     aws.String(data["type"].(string)),
	}
	tek := data["encryption_key"].([]interface{})
	if len(tek) > 0 {
		vk := tek[0].(map[string]interface{})
		ek := codepipeline.EncryptionKey{
			Type: aws.String(vk["type"].(string)),
			Id:   aws.String(vk["id"].(string)),
		}
		pipelineArtifactStore.EncryptionKey = &ek
	}

	return data["region"].(string), &pipelineArtifactStore
}

func flattenAwsCodePipelineArtifactStore(artifactStore *codepipeline.ArtifactStore) []interface{} {
	if artifactStore == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}
	values["type"] = *artifactStore.Type
	values["location"] = *artifactStore.Location
	if artifactStore.EncryptionKey != nil {
		as := map[string]interface{}{
			"id":   *artifactStore.EncryptionKey.Id,
			"type": *artifactStore.EncryptionKey.Type,
		}
		values["encryption_key"] = []interface{}{as}
	}
	return []interface{}{values}
}

func flattenAwsCodePipelineArtifactStores(artifactStores map[string]*codepipeline.ArtifactStore) []interface{} {
	values := []interface{}{}
	for region, artifactStore := range artifactStores {
		store := flattenAwsCodePipelineArtifactStore(artifactStore)[0].(map[string]interface{})
		store["region"] = region
		values = append(values, store)
	}
	return values
}

func expandAwsCodePipelineStages(d *schema.ResourceData) []*codepipeline.StageDeclaration {
	configs := d.Get("stage").([]interface{})
	pipelineStages := []*codepipeline.StageDeclaration{}

	for _, stage := range configs {
		data := stage.(map[string]interface{})
		a := data["action"].([]interface{})
		actions := expandAwsCodePipelineActions(a)
		pipelineStages = append(pipelineStages, &codepipeline.StageDeclaration{
			Name:    aws.String(data["name"].(string)),
			Actions: actions,
		})
	}
	return pipelineStages
}

func flattenAwsCodePipelineStages(stages []*codepipeline.StageDeclaration) []interface{} {
	stagesList := []interface{}{}
	for _, stage := range stages {
		values := map[string]interface{}{}
		values["name"] = *stage.Name
		values["action"] = flattenAwsCodePipelineStageActions(stage.Actions)
		stagesList = append(stagesList, values)
	}
	return stagesList

}

func expandAwsCodePipelineActions(s []interface{}) []*codepipeline.ActionDeclaration {
	actions := []*codepipeline.ActionDeclaration{}
	for _, config := range s {
		data := config.(map[string]interface{})

		conf := expandAwsCodePipelineStageActionConfiguration(data["configuration"].(map[string]interface{}))
		if data["provider"].(string) == "GitHub" {
			githubToken := os.Getenv("GITHUB_TOKEN")
			if githubToken != "" {
				conf["OAuthToken"] = aws.String(githubToken)
			}

		}

		action := codepipeline.ActionDeclaration{
			ActionTypeId: &codepipeline.ActionTypeId{
				Category: aws.String(data["category"].(string)),
				Owner:    aws.String(data["owner"].(string)),

				Provider: aws.String(data["provider"].(string)),
				Version:  aws.String(data["version"].(string)),
			},
			Name:          aws.String(data["name"].(string)),
			Configuration: conf,
		}

		oa := data["output_artifacts"].([]interface{})
		if len(oa) > 0 {
			outputArtifacts := expandAwsCodePipelineActionsOutputArtifacts(oa)
			action.OutputArtifacts = outputArtifacts

		}
		ia := data["input_artifacts"].([]interface{})
		if len(ia) > 0 {
			inputArtifacts := expandAwsCodePipelineActionsInputArtifacts(ia)
			action.InputArtifacts = inputArtifacts

		}
		ra := data["role_arn"].(string)
		if ra != "" {
			action.RoleArn = aws.String(ra)
		}
		ro := data["run_order"].(int)
		if ro > 0 {
			action.RunOrder = aws.Int64(int64(ro))
		}
		r := data["region"].(string)
		if r != "" {
			action.Region = aws.String(r)
		}
		ns := data["namespace"].(string)
		if len(ns) > 0 {
			action.Namespace = aws.String(ns)
		}
		actions = append(actions, &action)
	}
	return actions
}

func flattenAwsCodePipelineStageActions(actions []*codepipeline.ActionDeclaration) []interface{} {
	actionsList := []interface{}{}
	for _, action := range actions {
		values := map[string]interface{}{
			"category": *action.ActionTypeId.Category,
			"owner":    *action.ActionTypeId.Owner,
			"provider": *action.ActionTypeId.Provider,
			"version":  *action.ActionTypeId.Version,
			"name":     *action.Name,
		}
		if action.Configuration != nil {
			config := flattenAwsCodePipelineStageActionConfiguration(action.Configuration)
			_, ok := config["OAuthToken"]
			actionProvider := *action.ActionTypeId.Provider
			if ok && actionProvider == "GitHub" {
				delete(config, "OAuthToken")
			}
			values["configuration"] = config
		}

		if len(action.OutputArtifacts) > 0 {
			values["output_artifacts"] = flattenAwsCodePipelineActionsOutputArtifacts(action.OutputArtifacts)
		}

		if len(action.InputArtifacts) > 0 {
			values["input_artifacts"] = flattenAwsCodePipelineActionsInputArtifacts(action.InputArtifacts)
		}

		if action.RoleArn != nil {
			values["role_arn"] = *action.RoleArn
		}

		if action.RunOrder != nil {
			values["run_order"] = int(*action.RunOrder)
		}

		if action.Region != nil {
			values["region"] = *action.Region
		}

		if action.Namespace != nil {
			values["namespace"] = aws.StringValue(action.Namespace)
		}

		actionsList = append(actionsList, values)
	}
	return actionsList
}

func expandAwsCodePipelineStageActionConfiguration(config map[string]interface{}) map[string]*string {
	m := map[string]*string{}
	for k, v := range config {
		s := v.(string)
		m[k] = &s
	}
	return m
}

func flattenAwsCodePipelineStageActionConfiguration(config map[string]*string) map[string]string {
	m := map[string]string{}
	for k, v := range config {
		m[k] = *v
	}
	return m
}

func expandAwsCodePipelineActionsOutputArtifacts(s []interface{}) []*codepipeline.OutputArtifact {
	outputArtifacts := []*codepipeline.OutputArtifact{}
	for _, artifact := range s {
		if artifact == nil {
			continue
		}
		outputArtifacts = append(outputArtifacts, &codepipeline.OutputArtifact{
			Name: aws.String(artifact.(string)),
		})
	}
	return outputArtifacts
}

func flattenAwsCodePipelineActionsOutputArtifacts(artifacts []*codepipeline.OutputArtifact) []string {
	values := []string{}
	for _, artifact := range artifacts {
		values = append(values, *artifact.Name)
	}
	return values
}

func expandAwsCodePipelineActionsInputArtifacts(s []interface{}) []*codepipeline.InputArtifact {
	outputArtifacts := []*codepipeline.InputArtifact{}
	for _, artifact := range s {
		if artifact == nil {
			continue
		}
		outputArtifacts = append(outputArtifacts, &codepipeline.InputArtifact{
			Name: aws.String(artifact.(string)),
		})
	}
	return outputArtifacts
}

func flattenAwsCodePipelineActionsInputArtifacts(artifacts []*codepipeline.InputArtifact) []string {
	values := []string{}
	for _, artifact := range artifacts {
		values = append(values, *artifact.Name)
	}
	return values
}

func resourceAwsCodePipelineRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codepipelineconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.GetPipeline(&codepipeline.GetPipelineInput{
		Name: aws.String(d.Id()),
	})

	if isAWSErr(err, codepipeline.ErrCodePipelineNotFoundException, "") {
		log.Printf("[WARN] Codepipeline (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Codepipeline: %s", err)
	}

	metadata := resp.Metadata
	pipeline := resp.Pipeline

	if pipeline.ArtifactStore != nil {
		if err := d.Set("artifact_store", flattenAwsCodePipelineArtifactStore(pipeline.ArtifactStore)); err != nil {
			return err
		}
	} else if pipeline.ArtifactStores != nil {
		if err := d.Set("artifact_store", flattenAwsCodePipelineArtifactStores(pipeline.ArtifactStores)); err != nil {
			return err
		}
	}

	if err := d.Set("stage", flattenAwsCodePipelineStages(pipeline.Stages)); err != nil {
		return err
	}

	arn := aws.StringValue(metadata.PipelineArn)
	d.Set("arn", arn)
	d.Set("name", pipeline.Name)
	d.Set("role_arn", pipeline.RoleArn)

	tags, err := keyvaluetags.CodepipelineListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Codepipeline (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsCodePipelineUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codepipelineconn

	pipeline, err := expandAwsCodePipeline(d)
	if err != nil {
		return err
	}
	params := &codepipeline.UpdatePipelineInput{
		Pipeline: pipeline,
	}
	_, err = conn.UpdatePipeline(params)

	if err != nil {
		return fmt.Errorf(
			"[ERROR] Error updating CodePipeline (%s): %s",
			d.Id(), err)
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.CodepipelineUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Codepipeline (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsCodePipelineRead(d, meta)
}

func resourceAwsCodePipelineDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codepipelineconn

	_, err := conn.DeletePipeline(&codepipeline.DeletePipelineInput{
		Name: aws.String(d.Id()),
	})

	if isAWSErr(err, codepipeline.ErrCodePipelineNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Codepipeline (%s): %s", d.Id(), err)
	}

	return err
}
