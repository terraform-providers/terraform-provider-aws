package aws

import (
	"fmt"
	"log"
	//"strings"

	"github.com/aws/aws-sdk-go/aws"
	//"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsSsmParameters() *schema.Resource {
	parameterSchema := map[string]*schema.Schema{
		"arn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"name": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"type": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"value": {
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},
	}

	return &schema.Resource{
		Read: dataAwsSsmParametersRead,
		Schema: map[string]*schema.Schema{
			"path": {
				Type:     schema.TypeString,
				Computed: false,
				Required: true,
			},
			"parameters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: parameterSchema,
				},
			},
		},
	}
}

func dataAwsSsmParametersRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	path := d.Get("path").(string)

	parameters := make([]ssm.Parameter, 0)
	paramInput := &ssm.GetParametersByPathInput{
		Path:           &path,
		MaxResults:     aws.Int64(10),
		Recursive:      aws.Bool(false),
		WithDecryption: aws.Bool(false),
	}

	log.Printf("[DEBUG] Reading SSM Parameters: %s", paramInput)
	err := ssmconn.GetParametersByPathPages(paramInput, func(resp *ssm.GetParametersByPathOutput, isLast bool) bool {
		for _, parameter := range resp.Parameters {
			parameters = append(parameters, *parameter)
		}
		return !isLast
	})

	if err != nil {
		return fmt.Errorf("error describing SSM parameters: %s", err)
	}

	parsedParameters := make([]*map[string]string, len(parameters))
	for i, parameter := range parameters {
		data := awsSsmParameterToMap(&parameter, meta.(*AWSClient))
		parsedParameters[i] = &data
	}

	d.SetId(path)

	if err := d.Set("parameters", &parsedParameters); err != nil {
		return fmt.Errorf("error setting parameters: %s", err)
	}

	return nil
}

func awsSsmParameterToMap(parameter *ssm.Parameter, awsClient *AWSClient) map[string]string {
	arnData := calculateAwsSsmParameterArn(*parameter.Name, awsClient)

	return map[string]string{
		"arn":   arnData.String(),
		"name":  aws.StringValue(parameter.Name),
		"type":  aws.StringValue(parameter.Type),
		"value": aws.StringValue(parameter.Value),
	}
}
