package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsApiGatewayApiKey() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsApiGatewayApiKeyRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"value": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsApiGatewayApiKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	apiKey, err := conn.GetApiKey(&apigateway.GetApiKeyInput{
		ApiKey:       aws.String(d.Get("id").(string)),
		IncludeValue: aws.Bool(true),
	})

	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(apiKey.Id))
	d.Set("name", apiKey.Name)
	d.Set("value", apiKey.Value)
	d.Set("created_date", aws.TimeValue(apiKey.CreatedDate).Format(time.RFC3339))
	d.Set("description", apiKey.Description)
	d.Set("enabled", apiKey.Enabled)
	d.Set("last_updated_date", aws.TimeValue(apiKey.LastUpdatedDate).Format(time.RFC3339))

	if err := d.Set("tags", keyvaluetags.ApigatewayKeyValueTags(apiKey.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}
	return nil
}
