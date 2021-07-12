package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

var dataPipelineScheduleTypeList = []string{
	"cron",
	"ondemand",
	"timeseries",
}

var dataPipelineFailureAndRerunModeList = []string{
	"cascade",
	"none",
}

func resourceAwsDataPipelineDefinition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDataPipelineDefinitionPut,
		Read:   resourceAwsDataPipelineDefinitionRead,
		Update: resourceAwsDataPipelineDefinitionPut,
		Delete: schema.Noop,
		Importer: &schema.ResourceImporter{
			State: resourceAwsDataPipelineDefinitionImport,
		},

		Schema: map[string]*schema.Schema{
			"pipeline_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"default": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_role": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},

						"role": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},

						"schedule_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(dataPipelineScheduleTypeList, false),
						},

						"failure_and_rerun_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(dataPipelineFailureAndRerunModeList, false),
							Default:      "none",
						},

						"pipeline_log_uri": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsDataPipelineDefinitionPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datapipelineconn

	pipelineID := d.Get("pipeline_id").(string)
	input := &datapipeline.PutPipelineDefinitionInput{
		PipelineId: aws.String(pipelineID),
	}

	pipelineObjectsConfigs := []*datapipeline.PipelineObject{}

	if v, ok := d.GetOk("default"); ok {
		pipelineObject, err := expandDataPipelineDefaultPipelineObject(v.([]interface{}))
		if err != nil {
			return err
		}
		pipelineObjectsConfigs = append(pipelineObjectsConfigs, pipelineObject)
	}

	if len(pipelineObjectsConfigs) > 0 {
		input.PipelineObjects = pipelineObjectsConfigs
	}

	_, err := conn.PutPipelineDefinition(input)
	if isAWSErr(err, datapipeline.ErrCodePipelineNotFoundException, "") || isAWSErr(err, datapipeline.ErrCodePipelineDeletedException, "") {
		log.Printf("[WARN] DataPipeline (%s) not found, removing from state", pipelineID)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error putting DataPipeline (%s) Definition: %s", pipelineID, err)
	}

	d.SetId(pipelineID)
	return resourceAwsDataPipelineDefinitionRead(d, meta)
}

func resourceAwsDataPipelineDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datapipelineconn

	v, err := resourceAwsDataPipelineDefinitionRetrieve(d.Id(), conn)
	if isAWSErr(err, datapipeline.ErrCodePipelineNotFoundException, "") || isAWSErr(err, datapipeline.ErrCodePipelineDeletedException, "") {
		log.Printf("[WARN] DataPipeline (%s) Definition not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error getting DataPipeline (%s) Definition: %s", d.Id(), err)
	}

	if err := flattenDataPipelinePipelineObjects(d, v.PipelineObjects); err != nil {
		return err
	}

	return nil
}

func resourceAwsDataPipelineDefinitionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("pipeline_id", d.Id())
	return []*schema.ResourceData{d}, nil
}

func resourceAwsDataPipelineDefinitionRetrieve(id string, conn *datapipeline.DataPipeline) (*datapipeline.GetPipelineDefinitionOutput, error) {
	opts := datapipeline.GetPipelineDefinitionInput{
		PipelineId: aws.String(id),
	}
	log.Printf("[DEBUG] Data Pipeline describe configuration: %#v", opts)

	resp, err := conn.GetPipelineDefinition(&opts)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func flattenDataPipelinePipelineObjects(d *schema.ResourceData, pipelineObjects []*datapipeline.PipelineObject) error {

	for _, pipelineObject := range pipelineObjects {
		for _, field := range pipelineObject.Fields {
			if aws.StringValue(field.Key) == "type" {
				switch aws.StringValue(field.StringValue) {
				case "Default":
					if err := d.Set("default", flattenDataPipelineDefaultPipelineObject(pipelineObject)); err != nil {
						return fmt.Errorf("Error setting default object: %s", err)
					}
				}
			}
		}
	}

	return nil
}
