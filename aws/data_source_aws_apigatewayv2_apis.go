package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apigatewayv2/finder"
)

func dataSourceAwsApiGatewayV2Apis() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAwsApiGatewayV2ApisRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"protocol_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsAwsApiGatewayV2ApisRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	tagsToMatch := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	apis, err := finder.Apis(conn, &apigatewayv2.GetApisInput{})

	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 APIs: %w", err)
	}

	var ids []*string

	for _, api := range apis {
		if v, ok := d.GetOk("name"); ok && v.(string) != aws.StringValue(api.Name) {
			continue
		}

		if v, ok := d.GetOk("protocol_type"); ok && v.(string) != aws.StringValue(api.ProtocolType) {
			continue
		}

		if len(tagsToMatch) > 0 && !keyvaluetags.Apigatewayv2KeyValueTags(api.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).ContainsAll(tagsToMatch) {
			continue
		}

		ids = append(ids, api.ApiId)
	}

	d.SetId(meta.(*AWSClient).region)

	if err := d.Set("ids", flattenStringSet(ids)); err != nil {
		return fmt.Errorf("error setting ids: %w", err)
	}

	return nil
}
