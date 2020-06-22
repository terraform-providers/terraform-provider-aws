package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsApiGatewayApiKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayApiKeyCreate,
		Read:   resourceAwsApiGatewayApiKeyRead,
		Update: resourceAwsApiGatewayApiKeyUpdate,
		Delete: resourceAwsApiGatewayApiKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"stage_key": {
				Type:     schema.TypeSet,
				Optional: true,
				Removed:  "Since the API Gateway usage plans feature was launched on August 11, 2016, usage plans are now required to associate an API key with an API stage",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rest_api_id": {
							Type:     schema.TypeString,
							Required: true,
						},

						"stage_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"value": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(30, 128),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsApiGatewayApiKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	log.Printf("[DEBUG] Creating API Gateway API Key")

	apiKey, err := conn.CreateApiKey(&apigateway.CreateApiKeyInput{
		Name:        aws.String(d.Get("name").(string)),
		Description: aws.String(d.Get("description").(string)),
		Enabled:     aws.Bool(d.Get("enabled").(bool)),
		Value:       aws.String(d.Get("value").(string)),
		Tags:        keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().ApigatewayTags(),
	})
	if err != nil {
		return fmt.Errorf("Error creating API Gateway API Key: %s", err)
	}

	d.SetId(aws.StringValue(apiKey.Id))

	return resourceAwsApiGatewayApiKeyRead(d, meta)
}

func resourceAwsApiGatewayApiKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading API Gateway API Key: %s", d.Id())

	apiKey, err := conn.GetApiKey(&apigateway.GetApiKeyInput{
		ApiKey:       aws.String(d.Id()),
		IncludeValue: aws.Bool(true),
	})
	if err != nil {
		if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] API Gateway API Key (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	if err := d.Set("tags", keyvaluetags.ApigatewayKeyValueTags(apiKey.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "apigateway",
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("/apikeys/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	d.Set("name", apiKey.Name)
	d.Set("description", apiKey.Description)
	d.Set("enabled", apiKey.Enabled)
	d.Set("value", apiKey.Value)

	if err := d.Set("created_date", apiKey.CreatedDate.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting created_date: %s", err)
	}

	if err := d.Set("last_updated_date", apiKey.LastUpdatedDate.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("error setting last_updated_date: %s", err)
	}

	return nil
}

func resourceAwsApiGatewayApiKeyUpdateOperations(d *schema.ResourceData) []*apigateway.PatchOperation {
	operations := make([]*apigateway.PatchOperation, 0)
	if d.HasChange("enabled") {
		isEnabled := "false"
		if d.Get("enabled").(bool) {
			isEnabled = "true"
		}
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/enabled"),
			Value: aws.String(isEnabled),
		})
	}

	if d.HasChange("description") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/description"),
			Value: aws.String(d.Get("description").(string)),
		})
	}

	return operations
}

func resourceAwsApiGatewayApiKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn

	log.Printf("[DEBUG] Updating API Gateway API Key: %s", d.Id())

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.ApigatewayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	_, err := conn.UpdateApiKey(&apigateway.UpdateApiKeyInput{
		ApiKey:          aws.String(d.Id()),
		PatchOperations: resourceAwsApiGatewayApiKeyUpdateOperations(d),
	})
	if err != nil {
		return err
	}

	return resourceAwsApiGatewayApiKeyRead(d, meta)
}

func resourceAwsApiGatewayApiKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	log.Printf("[DEBUG] Deleting API Gateway API Key: %s", d.Id())

	_, err := conn.DeleteApiKey(&apigateway.DeleteApiKeyInput{
		ApiKey: aws.String(d.Id()),
	})

	if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting API Gateway API Key (%s): %s", d.Id(), err)
	}

	return nil
}
