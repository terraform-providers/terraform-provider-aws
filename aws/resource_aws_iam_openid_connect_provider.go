package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsIamOpenIDConnectProvider() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamOpenIDConnectProviderCreate,
		Read:   resourceAwsIamOpenIDConnectProviderRead,
		Update: resourceAwsIamOpenIDConnectProviderUpdate,
		Delete: resourceAwsIamOpenIDConnectProviderDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     validateOpenIdURL,
				DiffSuppressFunc: suppressOpenIdURL,
			},
			"client_id_list": {
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
			},
			"thumbprint_list": {
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(40, 40),
				},
				Type:     schema.TypeList,
				Required: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsIamOpenIDConnectProviderCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &iam.CreateOpenIDConnectProviderInput{
		Url:            aws.String(d.Get("url").(string)),
		ClientIDList:   expandStringList(d.Get("client_id_list").([]interface{})),
		ThumbprintList: expandStringList(d.Get("thumbprint_list").([]interface{})),
		Tags:           tags.IgnoreAws().IamTags(),
	}

	out, err := conn.CreateOpenIDConnectProvider(input)
	if err != nil {
		return fmt.Errorf("error creating IAM OIDC Provider: %w", err)
	}

	d.SetId(aws.StringValue(out.OpenIDConnectProviderArn))

	return resourceAwsIamOpenIDConnectProviderRead(d, meta)
}

func resourceAwsIamOpenIDConnectProviderRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &iam.GetOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(d.Id()),
	}
	out, err := conn.GetOpenIDConnectProvider(input)
	if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
		log.Printf("[WARN] IAM OIDC Provider (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading IAM OIDC Provider (%s): %w", d.Id(), err)
	}

	d.Set("arn", d.Id())
	d.Set("url", out.Url)
	d.Set("client_id_list", flattenStringList(out.ClientIDList))
	d.Set("thumbprint_list", flattenStringList(out.ThumbprintList))

	tags := keyvaluetags.IamKeyValueTags(out.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsIamOpenIDConnectProviderUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	if d.HasChange("thumbprint_list") {
		input := &iam.UpdateOpenIDConnectProviderThumbprintInput{
			OpenIDConnectProviderArn: aws.String(d.Id()),
			ThumbprintList:           expandStringList(d.Get("thumbprint_list").([]interface{})),
		}

		_, err := conn.UpdateOpenIDConnectProviderThumbprint(input)
		if err != nil {
			return fmt.Errorf("error updating IAM OIDC Provider (%s) thumbprint: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.IamOpenIDConnectProviderUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags for IAM OIDC Provider (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsIamOpenIDConnectProviderRead(d, meta)
}

func resourceAwsIamOpenIDConnectProviderDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	input := &iam.DeleteOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(d.Id()),
	}
	_, err := conn.DeleteOpenIDConnectProvider(input)
	if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting IAM OIDC Provider (%s): %w", d.Id(), err)
	}

	return nil
}
