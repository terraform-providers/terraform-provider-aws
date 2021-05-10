// https://pkg.go.dev/github.com/aws/aws-sdk-go@v1.38.31/service/s3control?utm_source=gopls
// https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_access_point
package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsS3ObjectLambdaAccessPoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3ObjectLambdaAccessPointCreate,
		Read:   resourceAwsS3ObjectLambdaAccessPointRead,
		Update: resourceAwsS3ObjectLambdaAccessPointUpdate,
		Delete: resourceAwsS3ObjectLambdaAccessPointDelete,

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
			
			"Allowed_features": {
				Type: schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					},
			},

			"cloud_watch_metrics_enabled": {
				Type: schema.TypeBool,
				Optional: true,
			},

			"supporting_access_point": {
				Type:         schema.TypeString,
				Optional:     false,
			},

			"transformation_configurations": {
				Type:             schema.TypeList,
				Optional:         false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"actions": {
							Type:     schema.TypeString,
							Optional: false,
						},
						"Content_transformation": {
							Type:     schema.schema.TypeList,
							Optional: false,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"aws_lambda": {
										Type:     schema.TypeString,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"function_arn": {
													Type:     schema.TypeString,
													Optional: false,
												},
												"function_payload": {
													Type:     schema.schema.TypeString,
													Optional: true,
						
											},
										},
									},
		
								},
							},
						},

					},
				},
			},
		},
	},
}

func resourceAwsS3ObjectLambdaAccessPointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3controlconn

	accountId := meta.(*AWSClient).accountid
	if v, ok := d.GetOk("account_id"); ok {
		accountId = v.(string)
	}
	name := d.Get("name").(string)

	configuration := &s3control.ObjectLambdaConfiguration{
		AllowedFeatures []*string `locationNameList:"AllowedFeature" type:"list"`
		CloudWatchMetricsEnabled *bool `type:"boolean"`
		SupportingAccessPoint *string `min:"1" type:"string" required:"true"`
		TransformationConfigurations []*ObjectLambdaTransformationConfiguration `locationNameList:"TransformationConfiguration" type:"list" required:"true"`
	}

	input := &s3control.CreateAccessPointForObjectLambdaInput{
		AccountId:     aws.String(accountId),
		Configuration: configuration,
		Name:          aws.String(name),
	}

	/*
		log.Printf("[DEBUG] Creating S3 Access Point: %s", input)
		output, err := conn.CreateAccessPoint(input)

		if err != nil {
			return fmt.Errorf("error creating S3 Control Access Point (%s): %w", name, err)
		}

		if output == nil {
			return fmt.Errorf("error creating S3 Control Access Point (%s): empty response", name)
		}

		parsedARN, err := arn.Parse(aws.StringValue(output.AccessPointArn))

		if err == nil && strings.HasPrefix(parsedARN.Resource, "outpost/") {
			d.SetId(aws.StringValue(output.AccessPointArn))
			name = aws.StringValue(output.AccessPointArn)
		} else {
			d.SetId(fmt.Sprintf("%s:%s", accountId, name))
		}

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
	*/
}

func resourceAwsS3ObjectLambdaAccessPointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3controlconn

	accountId, name, err := s3AccessPointParseId(d.Id())
	if err != nil {
		return err
	}

	/*
		output, err := conn.GetAccessPoint(&s3control.GetAccessPointInput{
			AccountId: aws.String(accountId),
			Name:      aws.String(name),
		})

		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, tfs3control.ErrCodeNoSuchAccessPoint) {
			log.Printf("[WARN] S3 Access Point (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		if err != nil {
			return fmt.Errorf("error reading S3 Access Point (%s): %w", d.Id(), err)
		}

		if output == nil {
			return fmt.Errorf("error reading S3 Access Point (%s): empty response", d.Id())
		}

		if strings.HasPrefix(name, "arn:") {
			parsedAccessPointARN, err := arn.Parse(name)

			if err != nil {
				return fmt.Errorf("error parsing S3 Control Access Point ARN (%s): %w", name, err)
			}

			bucketARN := arn.ARN{
				AccountID: parsedAccessPointARN.AccountID,
				Partition: parsedAccessPointARN.Partition,
				Region:    parsedAccessPointARN.Region,
				Resource: strings.Replace(
					parsedAccessPointARN.Resource,
					fmt.Sprintf("accesspoint/%s", aws.StringValue(output.Name)),
					fmt.Sprintf("bucket/%s", aws.StringValue(output.Bucket)),
					1,
				),
				Service: parsedAccessPointARN.Service,
			}

			d.Set("arn", name)
			d.Set("bucket", bucketARN.String())
		} else {
			accessPointARN := arn.ARN{
				AccountID: accountId,
				Partition: meta.(*AWSClient).partition,
				Region:    meta.(*AWSClient).region,
				Resource:  fmt.Sprintf("accesspoint/%s", aws.StringValue(output.Name)),
				Service:   "s3",
			}

			d.Set("arn", accessPointARN.String())
			d.Set("bucket", output.Bucket)
		}

		d.Set("account_id", accountId)
		d.Set("domain_name", meta.(*AWSClient).RegionalHostname(fmt.Sprintf("%s-%s.s3-accesspoint", aws.StringValue(output.Name), accountId)))
		d.Set("name", output.Name)
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

		// Return early since S3 on Outposts cannot have public policies
		if strings.HasPrefix(name, "arn:") {
			d.Set("has_public_access_policy", false)

			return nil
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
	*/

	return nil
}

func resourceAwsS3ObjectLambdaAccessPointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3controlconn

	accountId, name, err := s3AccessPointParseId(d.Id())
	if err != nil {
		return err
	}

	/*
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
	*/

	return resourceAwsS3AccessPointRead(d, meta)
}

func resourceAwsS3ObjectLambdaAccessPointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3controlconn

	accountId, name, err := s3AccessPointParseId(d.Id())
	if err != nil {
		return err
	}

	/*
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
	*/

	return nil
}
