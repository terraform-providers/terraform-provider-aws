package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsLambdaAlias() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLambdaAliasRead,

		Schema: map[string]*schema.Schema{
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"invoke_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"function_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsLambdaAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	functionName := d.Get("function_name").(string)
	name := d.Get("name").(string)

	params := &lambda.GetAliasInput{
		FunctionName: aws.String(functionName),
		Name:         aws.String(name),
	}

	aliasConfiguration, err := conn.GetAlias(params)
	if err != nil {
		return fmt.Errorf("Error getting Lambda alias: %s", err)
	}

	d.SetId(*aliasConfiguration.AliasArn)

	d.Set("arn", aliasConfiguration.AliasArn)
	d.Set("description", aliasConfiguration.Description)
	d.Set("function_version", aliasConfiguration.FunctionVersion)

	invokeArn := lambdaFunctionInvokeArn(*aliasConfiguration.AliasArn, meta)
	d.Set("invoke_arn", invokeArn)

	return nil
}
