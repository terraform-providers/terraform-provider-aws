package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsLambdaFunctions() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLambdaFunctionsRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsLambdaFunctionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	input := &lambda.ListFunctionsInput{}

	log.Printf("[DEBUG] Getting list of all Lambda Functions")
	//output, err := conn.ListFunctions(input)

	//var functionIds, functionNames []string
	var lambdaFunctions []*lambda.FunctionConfiguration
	err := conn.ListFunctionsPages(input, func(resp *lambda.ListFunctionsOutput, isLast bool) bool {
		for _, function := range resp.Functions {
			lambdaFunctions = append(lambdaFunctions, function)
		}

		return !isLast
	})

	if err != nil {
		return fmt.Errorf("error getting Lambda Functions: %w", err)
	}

	if lambdaFunctions == nil {
		return fmt.Errorf("error getting Lambda Functions: empty response")
	}

	return nil
}
