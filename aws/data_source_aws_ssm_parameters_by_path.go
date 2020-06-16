package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsSsmParametersByPath() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSsmParametersReadByPath,

		Schema: map[string]*schema.Schema{
			"path": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"values": {
				Type:      schema.TypeList,
				Computed:  true,
				Sensitive: true,
				Elem:      &schema.Schema{Type: schema.TypeString},
			},
			"with_decryption": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func dataSourceAwsSsmParametersReadByPath(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	path := d.Get("path").(string)

	arns := make([]string, 0)
	names := make([]string, 0)
	types := make([]string, 0)
	values := make([]string, 0)

	for {
		paramInput := &ssm.GetParametersByPathInput{
			Path:           aws.String(path),
			WithDecryption: aws.Bool(d.Get("with_decryption").(bool)),
		}

		log.Printf("[DEBUG] Reading SSM Parameters by path: %s", paramInput)
		resp, err := ssmconn.GetParametersByPath(paramInput)

		if err != nil {
			return fmt.Errorf("Error reading SSM parameters by path: %s", err)
		}

		params := resp.Parameters

		for _, param := range params {
			arns = append(arns, *param.ARN)
			names = append(names, *param.Name)
			types = append(types, *param.Type)
			values = append(values, *param.Value)
		}

		if resp.NextToken == nil {
			break
		}
		paramInput.NextToken = resp.NextToken
	}

	d.SetId(resource.UniqueId())

	err := d.Set("arns", arns)
	if err != nil {
		return err
	}

	err = d.Set("names", names)
	if err != nil {
		return err
	}

	err = d.Set("types", types)
	if err != nil {
		return err
	}

	err = d.Set("values", values)
	if err != nil {
		return err
	}

	return nil
}
