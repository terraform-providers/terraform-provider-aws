package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
)

func expandDataPipelineDefaultPipelineObject(l []interface{}) (*datapipeline.PipelineObject, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	pipelineObject := &datapipeline.PipelineObject{
		Id:   aws.String("Default"),
		Name: aws.String("Default"),
	}
	fieldList := []*datapipeline.Field{}

	typeField := &datapipeline.Field{
		Key:         aws.String("type"),
		StringValue: aws.String("Default"),
	}
	fieldList = append(fieldList, typeField)

	if val, ok := m["schedule_type"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("scheduleType"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := m["failure_and_rerun_mode"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("failureAndRerunMode"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := m["pipeline_log_uri"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("pipelineLogUri"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := m["role"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("role"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := m["resource_role"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("resourceRole"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	pipelineObject.Fields = fieldList

	err := pipelineObject.Validate()
	if err != nil {
		return nil, fmt.Errorf("Error validate default pipeline object: %s", err)
	}

	return pipelineObject, nil
}

func flattenDataPipelineDefaultPipelineObject(object *datapipeline.PipelineObject) []map[string]interface{} {
	pipelineObject := make(map[string]interface{})
	for _, field := range object.Fields {
		switch aws.StringValue(field.Key) {
		case "scheduleType":
			pipelineObject["schedule_type"] = aws.StringValue(field.StringValue)
		case "failureAndRerunMode":
			pipelineObject["failure_and_rerun_mode"] = aws.StringValue(field.StringValue)
		case "pipelineLogUri":
			pipelineObject["pipeline_log_uri"] = aws.StringValue(field.StringValue)
		case "role":
			pipelineObject["role"] = aws.StringValue(field.StringValue)
		case "resourceRole":
			pipelineObject["resource_role"] = aws.StringValue(field.StringValue)
		case "schedule":
			pipelineObject["schedule"] = aws.StringValue(field.RefValue)
		}
	}

	return []map[string]interface{}{pipelineObject}
}
