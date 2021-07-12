package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
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
			"optional": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataAwsSsmParameterRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	name := d.Get("name").(string)

	optional := d.Get("optional").(bool)

	paramInput := &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(d.Get("with_decryption").(bool)),
	}

	log.Printf("[DEBUG] Reading SSM Parameter: %s", paramInput)
	resp, err := ssmconn.GetParameter(paramInput)

	if err != nil {
		if optional && isAWSErr(err, ssm.ErrCodeParameterNotFound, "") {
			// return nil value if the ssm parameter set to optional
			d.SetId(name)
			d.Set("name", name)
			d.Set("arn", "null")
			d.Set("value", "null")

			return nil
		} else {
			return fmt.Errorf("Error describing SSM parameter (%s): %w", name, err)
		}
	}

	param := resp.Parameter

	d.SetId(aws.StringValue(param.Name))
	d.Set("arn", param.ARN)
	d.Set("name", param.Name)
	d.Set("type", param.Type)
	d.Set("value", param.Value)
	d.Set("version", param.Version)

	return nil
}
