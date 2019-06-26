package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsSsmParameter() *schema.Resource {
	return &schema.Resource{
		Read: dataAwsSsmParameterRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
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
			"with_decryption": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func dataAwsSsmParameterRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	name := d.Get("name").(string)

	paramInput := &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(d.Get("with_decryption").(bool)),
	}

	log.Printf("[DEBUG] Reading SSM Parameter: %s", paramInput)
	resp, err := ssmconn.GetParameter(paramInput)

	if err != nil {
		return fmt.Errorf("Error describing SSM parameter: %s", err)
	}

	param := resp.Parameter
	if err := parseAwsSsmParameter(d, param, meta); err != nil {
		return err
	}

	return nil
}

func parseAwsSsmParameter(d *schema.ResourceData, param *ssm.Parameter, meta interface{}) error {
	d.SetId(*param.Name)

	awsClient := meta.(*AWSClient)
	arnData := calculateAwsSsmParameterArn(d.Id(), awsClient)

	if err := d.Set("arn", arnData.String()); err != nil {
		return err
	}
	if err := d.Set("name", param.Name); err != nil {
		return err
	}
	if err := d.Set("type", param.Type); err != nil {
		return err
	}
	if err := d.Set("value", param.Value); err != nil {
		return err
	}

	return nil
}

func calculateAwsSsmParameterArn(id string, awsClient *AWSClient) arn.ARN {
	return arn.ARN{
		Partition: awsClient.partition,
		Region:    awsClient.region,
		Service:   "ssm",
		AccountID: awsClient.accountid,
		Resource:  fmt.Sprintf("parameter/%s", strings.TrimPrefix(id, "/")),
	}
}
