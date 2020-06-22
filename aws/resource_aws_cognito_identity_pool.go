package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsCognitoIdentityPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoIdentityPoolCreate,
		Read:   resourceAwsCognitoIdentityPoolRead,
		Update: resourceAwsCognitoIdentityPoolUpdate,
		Delete: resourceAwsCognitoIdentityPoolDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"identity_pool_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCognitoIdentityPoolName,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cognito_identity_providers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateCognitoIdentityProvidersClientId,
						},
						"provider_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateCognitoIdentityProvidersProviderName,
						},
						"server_side_token_check": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"developer_provider_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true, // Forcing a new resource since it cannot be edited afterwards
				ValidateFunc: validateCognitoProviderDeveloperName,
			},

			"allow_unauthenticated_identities": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"openid_connect_provider_arns": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
			},

			"saml_provider_arns": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
			},

			"supported_login_providers": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateCognitoSupportedLoginProviders,
				},
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsCognitoIdentityPoolCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoconn
	log.Print("[DEBUG] Creating Cognito Identity Pool")

	params := &cognitoidentity.CreateIdentityPoolInput{
		IdentityPoolName:               aws.String(d.Get("identity_pool_name").(string)),
		AllowUnauthenticatedIdentities: aws.Bool(d.Get("allow_unauthenticated_identities").(bool)),
	}

	if v, ok := d.GetOk("developer_provider_name"); ok {
		params.DeveloperProviderName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("supported_login_providers"); ok {
		params.SupportedLoginProviders = expandCognitoSupportedLoginProviders(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("cognito_identity_providers"); ok {
		params.CognitoIdentityProviders = expandCognitoIdentityProviders(v.(*schema.Set))
	}

	if v, ok := d.GetOk("saml_provider_arns"); ok {
		params.SamlProviderARNs = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("openid_connect_provider_arns"); ok {
		params.OpenIdConnectProviderARNs = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("tags"); ok {
		params.IdentityPoolTags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().CognitoidentityTags()
	}

	entity, err := conn.CreateIdentityPool(params)
	if err != nil {
		return fmt.Errorf("Error creating Cognito Identity Pool: %s", err)
	}

	d.SetId(*entity.IdentityPoolId)

	return resourceAwsCognitoIdentityPoolRead(d, meta)
}

func resourceAwsCognitoIdentityPoolRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading Cognito Identity Pool: %s", d.Id())

	ip, err := conn.DescribeIdentityPool(&cognitoidentity.DescribeIdentityPoolInput{
		IdentityPoolId: aws.String(d.Id()),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == cognitoidentity.ErrCodeResourceNotFoundException {
			d.SetId("")
			return nil
		}
		return err
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "cognito-identity",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("identitypool/%s", d.Id()),
	}
	d.Set("arn", arn.String())
	d.Set("identity_pool_name", ip.IdentityPoolName)
	d.Set("allow_unauthenticated_identities", ip.AllowUnauthenticatedIdentities)
	d.Set("developer_provider_name", ip.DeveloperProviderName)
	if err := d.Set("tags", keyvaluetags.CognitoidentityKeyValueTags(ip.IdentityPoolTags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("cognito_identity_providers", flattenCognitoIdentityProviders(ip.CognitoIdentityProviders)); err != nil {
		return fmt.Errorf("Error setting cognito_identity_providers error: %#v", err)
	}

	if err := d.Set("openid_connect_provider_arns", flattenStringList(ip.OpenIdConnectProviderARNs)); err != nil {
		return fmt.Errorf("Error setting openid_connect_provider_arns error: %#v", err)
	}

	if err := d.Set("saml_provider_arns", flattenStringList(ip.SamlProviderARNs)); err != nil {
		return fmt.Errorf("Error setting saml_provider_arns error: %#v", err)
	}

	if err := d.Set("supported_login_providers", flattenCognitoSupportedLoginProviders(ip.SupportedLoginProviders)); err != nil {
		return fmt.Errorf("Error setting supported_login_providers error: %#v", err)
	}

	return nil
}

func resourceAwsCognitoIdentityPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoconn
	log.Print("[DEBUG] Updating Cognito Identity Pool")

	params := &cognitoidentity.IdentityPool{
		IdentityPoolId:                 aws.String(d.Id()),
		AllowUnauthenticatedIdentities: aws.Bool(d.Get("allow_unauthenticated_identities").(bool)),
		IdentityPoolName:               aws.String(d.Get("identity_pool_name").(string)),
	}

	if v, ok := d.GetOk("developer_provider_name"); ok {
		params.DeveloperProviderName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cognito_identity_providers"); ok {
		params.CognitoIdentityProviders = expandCognitoIdentityProviders(v.(*schema.Set))
	}

	if v, ok := d.GetOk("supported_login_providers"); ok {
		params.SupportedLoginProviders = expandCognitoSupportedLoginProviders(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("openid_connect_provider_arns"); ok {
		params.OpenIdConnectProviderARNs = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("saml_provider_arns"); ok {
		params.SamlProviderARNs = expandStringList(v.([]interface{}))
	}

	log.Printf("[DEBUG] Updating Cognito Identity Pool: %s", params)

	_, err := conn.UpdateIdentityPool(params)
	if err != nil {
		return fmt.Errorf("Error updating Cognito Identity Pool: %s", err)
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.CognitoidentityUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Cognito Identity Pool (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsCognitoIdentityPoolRead(d, meta)
}

func resourceAwsCognitoIdentityPoolDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoconn
	log.Printf("[DEBUG] Deleting Cognito Identity Pool: %s", d.Id())

	_, err := conn.DeleteIdentityPool(&cognitoidentity.DeleteIdentityPoolInput{
		IdentityPoolId: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("Error deleting Cognito identity pool: %s", err)
	}
	return nil
}
