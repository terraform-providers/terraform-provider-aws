package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsLambdaFunction() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLambdaFunctionRead,

		Schema: map[string]*schema.Schema{
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"qualifier": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dead_letter_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"file_system_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"local_mount_path": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"handler": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"layers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"memory_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"reserved_concurrent_executions": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"runtime": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Computed: true,
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"qualified_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invoke_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_code_hash": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_code_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"environment": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"variables": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"tracing_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"kms_key_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsLambdaFunctionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	functionName := d.Get("function_name").(string)

	input := &lambda.GetFunctionInput{
		FunctionName: aws.String(functionName),
	}

	if v, ok := d.GetOk("qualifier"); ok {
		input.Qualifier = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Getting Lambda Function: %s", input)
	output, err := conn.GetFunction(input)

	if err != nil {
		return fmt.Errorf("error getting Lambda Function (%s): %s", functionName, err)
	}

	if output == nil {
		return fmt.Errorf("error getting Lambda Function (%s): empty response", functionName)
	}

	function := output.Configuration

	functionARN := aws.StringValue(function.FunctionArn)
	qualifierSuffix := fmt.Sprintf(":%s", d.Get("qualifier").(string))
	versionSuffix := fmt.Sprintf(":%s", aws.StringValue(function.Version))

	qualifiedARN := functionARN
	if !strings.HasSuffix(functionARN, qualifierSuffix) && !strings.HasSuffix(functionARN, versionSuffix) {
		qualifiedARN = functionARN + versionSuffix
	}

	unqualifiedARN := strings.TrimSuffix(functionARN, qualifierSuffix)

	d.Set("arn", unqualifiedARN)

	deadLetterConfig := []interface{}{}
	if function.DeadLetterConfig != nil {
		deadLetterConfig = []interface{}{
			map[string]interface{}{
				"target_arn": aws.StringValue(function.DeadLetterConfig.TargetArn),
			},
		}
	}
	if err := d.Set("dead_letter_config", deadLetterConfig); err != nil {
		return fmt.Errorf("error setting dead_letter_config: %s", err)
	}

	d.Set("description", function.Description)

	if err := d.Set("environment", flattenLambdaEnvironment(function.Environment)); err != nil {
		return fmt.Errorf("error setting environment: %s", err)
	}

	d.Set("handler", function.Handler)
	d.Set("invoke_arn", lambdaFunctionInvokeArn(aws.StringValue(function.FunctionArn), meta))
	d.Set("kms_key_arn", function.KMSKeyArn)
	d.Set("last_modified", function.LastModified)

	if err := d.Set("layers", flattenLambdaLayers(function.Layers)); err != nil {
		return fmt.Errorf("Error setting layers for Lambda Function (%s): %s", d.Id(), err)
	}

	d.Set("memory_size", function.MemorySize)
	d.Set("qualified_arn", qualifiedARN)

	reservedConcurrentExecutions := int64(-1)
	if output.Concurrency != nil {
		reservedConcurrentExecutions = aws.Int64Value(output.Concurrency.ReservedConcurrentExecutions)
	}
	d.Set("reserved_concurrent_executions", reservedConcurrentExecutions)

	d.Set("role", function.Role)
	d.Set("runtime", function.Runtime)
	d.Set("source_code_hash", function.CodeSha256)
	d.Set("source_code_size", function.CodeSize)

	if err := d.Set("tags", keyvaluetags.LambdaKeyValueTags(output.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	tracingConfig := []map[string]interface{}{
		{
			"mode": lambda.TracingModePassThrough,
		},
	}
	if function.TracingConfig != nil {
		tracingConfig[0]["mode"] = aws.StringValue(function.TracingConfig.Mode)
	}
	if err := d.Set("tracing_config", tracingConfig); err != nil {
		return fmt.Errorf("error setting tracing_config: %s", tracingConfig)
	}

	d.Set("timeout", function.Timeout)
	d.Set("version", function.Version)

	if err := d.Set("vpc_config", flattenLambdaVpcConfigResponse(function.VpcConfig)); err != nil {
		return fmt.Errorf("error setting vpc_config: %s", err)
	}

	if err := d.Set("file_system_config", flattenLambdaFileSystemConfigs(function.FileSystemConfigs)); err != nil {
		return fmt.Errorf("error setting file_system_config: %s", err)
	}

	d.SetId(aws.StringValue(function.FunctionName))

	return nil
}
