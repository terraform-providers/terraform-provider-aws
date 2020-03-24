package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsS3AccessPoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3AccessPointCreate,
		Read:   resourceAwsS3AccessPointRead,
		Update: resourceAwsS3AccessPointUpdate,
		Delete: resourceAwsS3AccessPointDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"has_public_access_policy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"network_origin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
			"public_access_block_configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				ForceNew:         true,
				MinItems:         0,
				MaxItems:         1,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"block_public_acls": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"block_public_policy": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"ignore_public_acls": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
						"restrict_public_buckets": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},
					},
				},
			},
			"vpc_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vpc_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsS3AccessPointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3controlconn

	accountId := meta.(*AWSClient).accountid
	if v, ok := d.GetOk("account_id"); ok {
		accountId = v.(string)
	}
	name := d.Get("name").(string)

	input := &s3control.CreateAccessPointInput{
		AccountId:                      aws.String(accountId),
		Bucket:                         aws.String(d.Get("bucket").(string)),
		Name:                           aws.String(name),
		PublicAccessBlockConfiguration: expandS3AccessPointPublicAccessBlockConfiguration(d.Get("public_access_block_configuration").([]interface{})),
		VpcConfiguration:               expandS3AccessPointVpcConfiguration(d.Get("vpc_configuration").([]interface{})),
	}

	log.Printf("[DEBUG] Creating S3 Access Point: %s", input)
	_, err := conn.CreateAccessPoint(input)
	if err != nil {
		return fmt.Errorf("error creating S3 Access Point: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", accountId, name))

	if v, ok := d.GetOk("policy"); ok {
		log.Printf("[DEBUG] Putting S3 Access Point policy: %s", d.Id())
		_, err := conn.PutAccessPointPolicy(&s3control.PutAccessPointPolicyInput{
			AccountId: aws.String(accountId),
			Name:      aws.String(name),
			Policy:    aws.String(v.(string)),
		})

		if err != nil {
			return fmt.Errorf("error putting S3 Access Point (%s) policy: %s", d.Id(), err)
		}
	}

	return resourceAwsS3AccessPointRead(d, meta)
}

func resourceAwsS3AccessPointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3controlconn

	accountId, name, err := s3AccessPointParseId(d.Id())
	if err != nil {
		return err
	}

	output, err := conn.GetAccessPoint(&s3control.GetAccessPointInput{
		AccountId: aws.String(accountId),
		Name:      aws.String(name),
	})

	if isAWSErr(err, "NoSuchAccessPoint", "") {
		log.Printf("[WARN] S3 Access Point (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Access Point (%s): %s", d.Id(), err)
	}

	name = aws.StringValue(output.Name)
	arn := arn.ARN{
		AccountID: accountId,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("accesspoint/%s", name),
		Service:   "s3",
	}
	d.Set("account_id", accountId)
	d.Set("arn", arn.String())
	d.Set("bucket", output.Bucket)
	d.Set("domain_name", meta.(*AWSClient).RegionalHostname(fmt.Sprintf("%s-%s.s3-accesspoint", name, accountId)))
	d.Set("name", name)
	d.Set("network_origin", output.NetworkOrigin)
	if err := d.Set("public_access_block_configuration", flattenS3AccessPointPublicAccessBlockConfiguration(output.PublicAccessBlockConfiguration)); err != nil {
		return fmt.Errorf("error setting public_access_block_configuration: %s", err)
	}
	if err := d.Set("vpc_configuration", flattenS3AccessPointVpcConfiguration(output.VpcConfiguration)); err != nil {
		return fmt.Errorf("error setting vpc_configuration: %s", err)
	}

	policyOutput, err := conn.GetAccessPointPolicy(&s3control.GetAccessPointPolicyInput{
		AccountId: aws.String(accountId),
		Name:      aws.String(name),
	})

	if isAWSErr(err, "NoSuchAccessPointPolicy", "") {
		d.Set("policy", "")
	} else {
		if err != nil {
			return fmt.Errorf("error reading S3 Access Point (%s) policy: %s", d.Id(), err)
		}

		d.Set("policy", policyOutput.Policy)
	}

	policyStatusOutput, err := conn.GetAccessPointPolicyStatus(&s3control.GetAccessPointPolicyStatusInput{
		AccountId: aws.String(accountId),
		Name:      aws.String(name),
	})

	if isAWSErr(err, "NoSuchAccessPointPolicy", "") {
		d.Set("has_public_access_policy", false)
	} else {
		if err != nil {
			return fmt.Errorf("error reading S3 Access Point (%s) policy status: %s", d.Id(), err)
		}

		d.Set("has_public_access_policy", policyStatusOutput.PolicyStatus.IsPublic)
	}

	return nil
}

func resourceAwsS3AccessPointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3controlconn

	accountId, name, err := s3AccessPointParseId(d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("policy") {
		if v, ok := d.GetOk("policy"); ok {
			log.Printf("[DEBUG] Putting S3 Access Point policy: %s", d.Id())
			_, err := conn.PutAccessPointPolicy(&s3control.PutAccessPointPolicyInput{
				AccountId: aws.String(accountId),
				Name:      aws.String(name),
				Policy:    aws.String(v.(string)),
			})

			if err != nil {
				return fmt.Errorf("error putting S3 Access Point (%s) policy: %s", d.Id(), err)
			}
		} else {
			log.Printf("[DEBUG] Deleting S3 Access Point policy: %s", d.Id())
			_, err := conn.DeleteAccessPointPolicy(&s3control.DeleteAccessPointPolicyInput{
				AccountId: aws.String(accountId),
				Name:      aws.String(name),
			})

			if err != nil {
				return fmt.Errorf("error deleting S3 Access Point (%s) policy: %s", d.Id(), err)
			}
		}
	}

	return resourceAwsS3AccessPointRead(d, meta)
}

func resourceAwsS3AccessPointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3controlconn

	accountId, name, err := s3AccessPointParseId(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting S3 Access Point: %s", d.Id())
	_, err = conn.DeleteAccessPoint(&s3control.DeleteAccessPointInput{
		AccountId: aws.String(accountId),
		Name:      aws.String(name),
	})

	if isAWSErr(err, "NoSuchAccessPoint", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Access Point (%s): %s", d.Id(), err)
	}

	return nil
}

func s3AccessPointParseId(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected ACCOUNT_ID:NAME", id)
	}

	return parts[0], parts[1], nil
}

func expandS3AccessPointVpcConfiguration(vConfig []interface{}) *s3control.VpcConfiguration {
	if len(vConfig) == 0 || vConfig[0] == nil {
		return nil
	}

	mConfig := vConfig[0].(map[string]interface{})

	return &s3control.VpcConfiguration{
		VpcId: aws.String(mConfig["vpc_id"].(string)),
	}
}

func flattenS3AccessPointVpcConfiguration(config *s3control.VpcConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"vpc_id": aws.StringValue(config.VpcId),
	}}
}

func expandS3AccessPointPublicAccessBlockConfiguration(vConfig []interface{}) *s3control.PublicAccessBlockConfiguration {
	if len(vConfig) == 0 || vConfig[0] == nil {
		return nil
	}

	mConfig := vConfig[0].(map[string]interface{})

	return &s3control.PublicAccessBlockConfiguration{
		BlockPublicAcls:       aws.Bool(mConfig["block_public_acls"].(bool)),
		BlockPublicPolicy:     aws.Bool(mConfig["block_public_policy"].(bool)),
		IgnorePublicAcls:      aws.Bool(mConfig["ignore_public_acls"].(bool)),
		RestrictPublicBuckets: aws.Bool(mConfig["restrict_public_buckets"].(bool)),
	}
}

func flattenS3AccessPointPublicAccessBlockConfiguration(config *s3control.PublicAccessBlockConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"block_public_acls":       aws.BoolValue(config.BlockPublicAcls),
		"block_public_policy":     aws.BoolValue(config.BlockPublicPolicy),
		"ignore_public_acls":      aws.BoolValue(config.IgnorePublicAcls),
		"restrict_public_buckets": aws.BoolValue(config.RestrictPublicBuckets),
	}}
}
