package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsApiGatewayClientCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayClientCertificateCreate,
		Read:   resourceAwsApiGatewayClientCertificateRead,
		Update: resourceAwsApiGatewayClientCertificateUpdate,
		Delete: resourceAwsApiGatewayClientCertificateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expiration_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pem_encoded_certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsApiGatewayClientCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := apigateway.GenerateClientCertificateInput{}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().ApigatewayTags()
	}
	log.Printf("[DEBUG] Generating API Gateway Client Certificate: %s", input)
	out, err := conn.GenerateClientCertificate(&input)
	if err != nil {
		return fmt.Errorf("Failed to generate client certificate: %s", err)
	}

	d.SetId(aws.StringValue(out.ClientCertificateId))

	return resourceAwsApiGatewayClientCertificateRead(d, meta)
}

func resourceAwsApiGatewayClientCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := apigateway.GetClientCertificateInput{
		ClientCertificateId: aws.String(d.Id()),
	}
	out, err := conn.GetClientCertificate(&input)
	if err != nil {
		if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] API Gateway Client Certificate %s not found, removing", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	log.Printf("[DEBUG] Received API Gateway Client Certificate: %s", out)

	tags := keyvaluetags.ApigatewayKeyValueTags(out.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "apigateway",
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("/clientcertificates/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	d.Set("description", out.Description)
	d.Set("created_date", out.CreatedDate.String())
	d.Set("expiration_date", out.ExpirationDate.String())
	d.Set("pem_encoded_certificate", out.PemEncodedCertificate)

	return nil
}

func resourceAwsApiGatewayClientCertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn

	operations := make([]*apigateway.PatchOperation, 0)
	if d.HasChange("description") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/description"),
			Value: aws.String(d.Get("description").(string)),
		})
	}

	input := apigateway.UpdateClientCertificateInput{
		ClientCertificateId: aws.String(d.Id()),
		PatchOperations:     operations,
	}

	log.Printf("[DEBUG] Updating API Gateway Client Certificate: %s", input)
	_, err := conn.UpdateClientCertificate(&input)
	if err != nil {
		return fmt.Errorf("Updating API Gateway Client Certificate failed: %s", err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.ApigatewayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsApiGatewayClientCertificateRead(d, meta)
}

func resourceAwsApiGatewayClientCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayconn
	log.Printf("[DEBUG] Deleting API Gateway Client Certificate: %s", d.Id())
	input := apigateway.DeleteClientCertificateInput{
		ClientCertificateId: aws.String(d.Id()),
	}
	_, err := conn.DeleteClientCertificate(&input)
	if err != nil {
		return fmt.Errorf("Deleting API Gateway Client Certificate failed: %s", err)
	}

	return nil
}
