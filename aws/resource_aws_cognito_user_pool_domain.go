package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cognitoidentityprovider/waiter"
)

func resourceAwsCognitoUserPoolDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoUserPoolDomainCreate,
		Read:   resourceAwsCognitoUserPoolDomainRead,
		Delete: resourceAwsCognitoUserPoolDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},
			"certificate_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"aws_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudfront_distribution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"s3_bucket": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCognitoUserPoolDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	domain := d.Get("domain").(string)

	timeout := 1 * time.Minute //Default timeout for a basic domain

	params := &cognitoidentityprovider.CreateUserPoolDomainInput{
		Domain:     aws.String(domain),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	if v, ok := d.GetOk("certificate_arn"); ok {
		customDomainConfig := &cognitoidentityprovider.CustomDomainConfigType{
			CertificateArn: aws.String(v.(string)),
		}
		params.CustomDomainConfig = customDomainConfig
		timeout = 60 * time.Minute //Custom domains take more time to become active
	}

	log.Printf("[DEBUG] Creating Cognito User Pool Domain: %s", params)

	_, err := conn.CreateUserPoolDomain(params)
	if err != nil {
		return fmt.Errorf("Error creating Cognito User Pool Domain: %w", err)
	}

	d.SetId(domain)

	if _, err := waiter.UserPoolDomainCreated(conn, d.Id(), timeout); err != nil {
		return fmt.Errorf("error waiting for User Pool Domain (%s) creation: %w", d.Id(), err)
	}

	return resourceAwsCognitoUserPoolDomainRead(d, meta)
}

func resourceAwsCognitoUserPoolDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn
	log.Printf("[DEBUG] Reading Cognito User Pool Domain: %s", d.Id())

	domain, err := conn.DescribeUserPoolDomain(&cognitoidentityprovider.DescribeUserPoolDomainInput{
		Domain: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, cognitoidentityprovider.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Cognito User Pool Domain %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	desc := domain.DomainDescription

	if desc.Status == nil {
		log.Printf("[WARN] Cognito User Pool Domain %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("domain", d.Id())
	d.Set("certificate_arn", "")
	if desc.CustomDomainConfig != nil {
		d.Set("certificate_arn", desc.CustomDomainConfig.CertificateArn)
	}
	d.Set("aws_account_id", desc.AWSAccountId)
	d.Set("cloudfront_distribution_arn", desc.CloudFrontDistribution)
	d.Set("s3_bucket", desc.S3Bucket)
	d.Set("user_pool_id", desc.UserPoolId)
	d.Set("version", desc.Version)

	return nil
}

func resourceAwsCognitoUserPoolDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn
	log.Printf("[DEBUG] Deleting Cognito User Pool Domain: %s", d.Id())

	_, err := conn.DeleteUserPoolDomain(&cognitoidentityprovider.DeleteUserPoolDomainInput{
		Domain:     aws.String(d.Id()),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	})
	if err != nil {
		return fmt.Errorf("Error deleting User Pool Domain: %w", err)
	}

	if _, err := waiter.UserPoolDomainDeleted(conn, d.Id()); err != nil {
		if isAWSErr(err, cognitoidentityprovider.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error waiting for User Pool Domain (%s) deletion: %w", d.Id(), err)
	}

	return nil

}
