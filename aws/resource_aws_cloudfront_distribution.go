package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsCloudFrontDistribution() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAwsCloudFrontDistributionCreate,
		Read:   resourceAwsCloudFrontDistributionRead,
		Update: resourceAwsCloudFrontDistributionUpdate,
		Delete: resourceAwsCloudFrontDistributionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsCloudFrontDistributionImport,
		},
		MigrateState:  resourceAwsCloudFrontDistributionMigrateState,
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aliases": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      aliasesHash,
			},
			"cache_behavior": {
				Type:     schema.TypeSet,
				Optional: true,
				Removed:  "Use `ordered_cache_behavior` configuration block(s) instead",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_methods": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"cached_methods": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"compress": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"default_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  86400,
						},
						"field_level_encryption_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"forwarded_values": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cookies": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"forward": {
													Type:     schema.TypeString,
													Required: true,
												},
												"whitelisted_names": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"headers": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"query_string": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"query_string_cache_keys": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"lambda_function_association": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 4,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"event_type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"lambda_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
									"include_body": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},
						"max_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  31536000,
						},
						"min_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"path_pattern": {
							Type:     schema.TypeString,
							Required: true,
						},
						"smooth_streaming": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"target_origin_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"trusted_signers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"viewer_protocol_policy": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"ordered_cache_behavior": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_methods": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"cached_methods": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"compress": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"default_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  86400,
						},
						"field_level_encryption_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"forwarded_values": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cookies": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"forward": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.StringInSlice([]string{
														cloudfront.ItemSelectionAll,
														cloudfront.ItemSelectionNone,
														cloudfront.ItemSelectionWhitelist,
													}, false),
												},
												"whitelisted_names": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"headers": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"query_string": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"query_string_cache_keys": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"lambda_function_association": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 4,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"event_type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"lambda_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
									"include_body": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
							Set: lambdaFunctionAssociationHash,
						},
						"max_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  31536000,
						},
						"min_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"path_pattern": {
							Type:     schema.TypeString,
							Required: true,
						},
						"smooth_streaming": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"target_origin_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"trusted_signers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"viewer_protocol_policy": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"custom_error_response": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      customErrorResponseHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"error_caching_min_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"error_code": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"response_code": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"response_page_path": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"default_cache_behavior": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_methods": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"cached_methods": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"compress": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"default_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  86400,
						},
						"field_level_encryption_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"forwarded_values": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cookies": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"forward": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.StringInSlice([]string{
														cloudfront.ItemSelectionAll,
														cloudfront.ItemSelectionNone,
														cloudfront.ItemSelectionWhitelist,
													}, false),
												},
												"whitelisted_names": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
									"headers": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"query_string": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"query_string_cache_keys": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"lambda_function_association": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 4,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"event_type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"lambda_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
									"include_body": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
							Set: lambdaFunctionAssociationHash,
						},
						"max_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  31536000,
						},
						"min_ttl": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"smooth_streaming": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"target_origin_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"trusted_signers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"viewer_protocol_policy": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"default_root_object": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"http_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "http2",
				ValidateFunc: validation.StringInSlice([]string{"http1.1", "http2"}, false),
			},
			"logging_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"include_cookies": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
					},
				},
			},
			"origin_group": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      originGroupHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"origin_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"failover_criteria": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status_codes": {
										Type:     schema.TypeSet,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeInt},
									},
								},
							},
						},
						"member": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 2,
							MaxItems: 2,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"origin_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"origin": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      originHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_origin_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"http_port": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"https_port": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"origin_keepalive_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  5,
									},
									"origin_read_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  30,
									},
									"origin_protocol_policy": {
										Type:     schema.TypeString,
										Required: true,
									},
									"origin_ssl_protocols": {
										Type:     schema.TypeSet,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"domain_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"custom_header": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      originCustomHeaderHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"origin_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"origin_path": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"s3_origin_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"origin_access_identity": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"price_class": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "PriceClass_All",
			},
			"restrictions": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"geo_restriction": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"locations": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"restriction_type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											cloudfront.GeoRestrictionTypeNone,
											cloudfront.GeoRestrictionTypeBlacklist,
											cloudfront.GeoRestrictionTypeWhitelist,
										}, false),
									},
								},
							},
						},
					},
				},
			},
			"viewer_certificate": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"acm_certificate_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cloudfront_default_certificate": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"iam_certificate_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"minimum_protocol_version": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "TLSv1",
						},
						"ssl_support_method": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"web_acl_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"caller_reference": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"active_trusted_signers": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"in_progress_validation_batches": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// retain_on_delete is a non-API attribute that may help facilitate speedy
			// deletion of a resoruce. It's mainly here for testing purposes, so
			// enable at your own risk.
			"retain_on_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"wait_for_deployment": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"is_ipv6_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsCloudFrontDistributionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	params := &cloudfront.CreateDistributionWithTagsInput{
		DistributionConfigWithTags: &cloudfront.DistributionConfigWithTags{
			DistributionConfig: expandDistributionConfig(d),
			Tags:               &cloudfront.Tags{Items: keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().CloudfrontTags()},
		},
	}

	var resp *cloudfront.CreateDistributionWithTagsOutput
	// Handle eventual consistency issues
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.CreateDistributionWithTags(params)

		// ACM and IAM certificate eventual consistency
		// InvalidViewerCertificate: The specified SSL certificate doesn't exist, isn't in us-east-1 region, isn't valid, or doesn't include a valid certificate chain.
		if isAWSErr(err, cloudfront.ErrCodeInvalidViewerCertificate, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	// Propagate AWS Go SDK retried error, if any
	if isResourceTimeoutError(err) {
		resp, err = conn.CreateDistributionWithTags(params)
	}

	if err != nil {
		return fmt.Errorf("error creating CloudFront Distribution: %s", err)
	}

	d.SetId(*resp.Distribution.Id)

	if d.Get("wait_for_deployment").(bool) {
		log.Printf("[DEBUG] Waiting until CloudFront Distribution (%s) is deployed", d.Id())
		if err := resourceAwsCloudFrontDistributionWaitUntilDeployed(d.Id(), meta); err != nil {
			return fmt.Errorf("error waiting until CloudFront Distribution (%s) is deployed: %s", d.Id(), err)
		}
	}

	return resourceAwsCloudFrontDistributionRead(d, meta)
}

func resourceAwsCloudFrontDistributionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	params := &cloudfront.GetDistributionInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetDistribution(params)
	if err != nil {
		if errcode, ok := err.(awserr.Error); ok && errcode.Code() == "NoSuchDistribution" {
			log.Printf("[WARN] No Distribution found: %s", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	// Update attributes from DistributionConfig
	err = flattenDistributionConfig(d, resp.Distribution.DistributionConfig)
	if err != nil {
		return err
	}

	// Update other attributes outside of DistributionConfig
	err = d.Set("active_trusted_signers", flattenActiveTrustedSigners(resp.Distribution.ActiveTrustedSigners))
	if err != nil {
		return err
	}
	d.Set("status", resp.Distribution.Status)
	d.Set("domain_name", resp.Distribution.DomainName)
	d.Set("last_modified_time", aws.String(resp.Distribution.LastModifiedTime.String()))
	d.Set("in_progress_validation_batches", resp.Distribution.InProgressInvalidationBatches)
	d.Set("etag", resp.ETag)
	d.Set("arn", resp.Distribution.ARN)

	tags, err := keyvaluetags.CloudfrontListTags(conn, d.Get("arn").(string))
	if err != nil {
		return fmt.Errorf("error listing tags for CloudFront Distribution (%s): %s", d.Id(), err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsCloudFrontDistributionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn
	params := &cloudfront.UpdateDistributionInput{
		Id:                 aws.String(d.Id()),
		DistributionConfig: expandDistributionConfig(d),
		IfMatch:            aws.String(d.Get("etag").(string)),
	}

	// Handle eventual consistency issues
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.UpdateDistribution(params)

		// ACM and IAM certificate eventual consistency
		// InvalidViewerCertificate: The specified SSL certificate doesn't exist, isn't in us-east-1 region, isn't valid, or doesn't include a valid certificate chain.
		if isAWSErr(err, cloudfront.ErrCodeInvalidViewerCertificate, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	// Propagate AWS Go SDK retried error, if any
	if isResourceTimeoutError(err) {
		_, err = conn.UpdateDistribution(params)
	}

	if err != nil {
		return fmt.Errorf("error updating CloudFront Distribution (%s): %s", d.Id(), err)
	}

	if d.Get("wait_for_deployment").(bool) {
		log.Printf("[DEBUG] Waiting until CloudFront Distribution (%s) is deployed", d.Id())
		if err := resourceAwsCloudFrontDistributionWaitUntilDeployed(d.Id(), meta); err != nil {
			return fmt.Errorf("error waiting until CloudFront Distribution (%s) is deployed: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.CloudfrontUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags for CloudFront Distribution (%s): %s", d.Id(), err)
		}
	}

	return resourceAwsCloudFrontDistributionRead(d, meta)
}

func resourceAwsCloudFrontDistributionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	if d.Get("retain_on_delete").(bool) {
		// Check if we need to disable first
		getDistributionInput := &cloudfront.GetDistributionInput{
			Id: aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Refreshing CloudFront Distribution (%s) to check if disable is necessary", d.Id())
		getDistributionOutput, err := conn.GetDistribution(getDistributionInput)

		if err != nil {
			return fmt.Errorf("error refreshing CloudFront Distribution (%s) to check if disable is necessary: %s", d.Id(), err)
		}

		if getDistributionOutput == nil || getDistributionOutput.Distribution == nil || getDistributionOutput.Distribution.DistributionConfig == nil {
			return fmt.Errorf("error refreshing CloudFront Distribution (%s) to check if disable is necessary: empty response", d.Id())
		}

		if !aws.BoolValue(getDistributionOutput.Distribution.DistributionConfig.Enabled) {
			log.Printf("[WARN] Removing CloudFront Distribution ID %q with `retain_on_delete` set. Please delete this distribution manually.", d.Id())
			return nil
		}

		updateDistributionInput := &cloudfront.UpdateDistributionInput{
			DistributionConfig: getDistributionOutput.Distribution.DistributionConfig,
			Id:                 getDistributionInput.Id,
			IfMatch:            getDistributionOutput.ETag,
		}
		updateDistributionInput.DistributionConfig.Enabled = aws.Bool(false)

		log.Printf("[DEBUG] Disabling CloudFront Distribution: %s", d.Id())
		_, err = conn.UpdateDistribution(updateDistributionInput)

		if err != nil {
			return fmt.Errorf("error disabling CloudFront Distribution (%s): %s", d.Id(), err)
		}

		log.Printf("[WARN] Removing CloudFront Distribution ID %q with `retain_on_delete` set. Please delete this distribution manually.", d.Id())
		return nil
	}

	deleteDistributionInput := &cloudfront.DeleteDistributionInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	}

	log.Printf("[DEBUG] Deleting CloudFront Distribution: %s", d.Id())
	_, err := conn.DeleteDistribution(deleteDistributionInput)

	if err == nil || isAWSErr(err, cloudfront.ErrCodeNoSuchDistribution, "") {
		return nil
	}

	// Refresh our ETag if it is out of date and attempt deletion again.
	if isAWSErr(err, cloudfront.ErrCodeInvalidIfMatchVersion, "") {
		getDistributionInput := &cloudfront.GetDistributionInput{
			Id: aws.String(d.Id()),
		}
		var getDistributionOutput *cloudfront.GetDistributionOutput

		log.Printf("[DEBUG] Refreshing CloudFront Distribution (%s) ETag", d.Id())
		getDistributionOutput, err = conn.GetDistribution(getDistributionInput)

		if err != nil {
			return fmt.Errorf("error refreshing CloudFront Distribution (%s) ETag: %s", d.Id(), err)
		}

		if getDistributionOutput == nil {
			return fmt.Errorf("error refreshing CloudFront Distribution (%s) ETag: empty response", d.Id())
		}

		deleteDistributionInput.IfMatch = getDistributionOutput.ETag

		_, err = conn.DeleteDistribution(deleteDistributionInput)
	}

	// Disable distribution if it is not yet disabled and attempt deletion again.
	// Here we update via the deployed configuration to ensure we are not submitting an out of date
	// configuration from the Terraform configuration, should other changes have occurred manually.
	if isAWSErr(err, cloudfront.ErrCodeDistributionNotDisabled, "") {
		getDistributionInput := &cloudfront.GetDistributionInput{
			Id: aws.String(d.Id()),
		}
		var getDistributionOutput *cloudfront.GetDistributionOutput

		log.Printf("[DEBUG] Refreshing CloudFront Distribution (%s) to disable", d.Id())
		getDistributionOutput, err = conn.GetDistribution(getDistributionInput)

		if err != nil {
			return fmt.Errorf("error refreshing CloudFront Distribution (%s) to disable: %s", d.Id(), err)
		}

		if getDistributionOutput == nil || getDistributionOutput.Distribution == nil {
			return fmt.Errorf("error refreshing CloudFront Distribution (%s) to disable: empty response", d.Id())
		}

		updateDistributionInput := &cloudfront.UpdateDistributionInput{
			DistributionConfig: getDistributionOutput.Distribution.DistributionConfig,
			Id:                 deleteDistributionInput.Id,
			IfMatch:            getDistributionOutput.ETag,
		}
		updateDistributionInput.DistributionConfig.Enabled = aws.Bool(false)
		var updateDistributionOutput *cloudfront.UpdateDistributionOutput

		log.Printf("[DEBUG] Disabling CloudFront Distribution: %s", d.Id())
		updateDistributionOutput, err = conn.UpdateDistribution(updateDistributionInput)

		if err != nil {
			return fmt.Errorf("error disabling CloudFront Distribution (%s): %s", d.Id(), err)
		}

		log.Printf("[DEBUG] Waiting until CloudFront Distribution (%s) is deployed", d.Id())
		if err := resourceAwsCloudFrontDistributionWaitUntilDeployed(d.Id(), meta); err != nil {
			return fmt.Errorf("error waiting until CloudFront Distribution (%s) is deployed: %s", d.Id(), err)
		}

		deleteDistributionInput.IfMatch = updateDistributionOutput.ETag

		_, err = conn.DeleteDistribution(deleteDistributionInput)

		// CloudFront has eventual consistency issues even for "deployed" state.
		// Occasionally the DeleteDistribution call will return this error as well, in which retries will succeed:
		//   * PreconditionFailed: The request failed because it didn't meet the preconditions in one or more request-header fields
		if isAWSErr(err, cloudfront.ErrCodeDistributionNotDisabled, "") || isAWSErr(err, cloudfront.ErrCodePreconditionFailed, "") {
			err = resource.Retry(2*time.Minute, func() *resource.RetryError {
				_, err := conn.DeleteDistribution(deleteDistributionInput)

				if isAWSErr(err, cloudfront.ErrCodeDistributionNotDisabled, "") {
					return resource.RetryableError(err)
				}

				if isAWSErr(err, cloudfront.ErrCodeNoSuchDistribution, "") {
					return nil
				}

				if isAWSErr(err, cloudfront.ErrCodePreconditionFailed, "") {
					return resource.RetryableError(err)
				}

				if err != nil {
					return resource.NonRetryableError(err)
				}

				return nil
			})

			// Propagate AWS Go SDK retried error, if any
			if isResourceTimeoutError(err) {
				_, err = conn.DeleteDistribution(deleteDistributionInput)
			}
		}
	}

	if isAWSErr(err, cloudfront.ErrCodeNoSuchDistribution, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("CloudFront Distribution %s cannot be deleted: %s", d.Id(), err)
	}

	return nil
}

// resourceAwsCloudFrontWebDistributionWaitUntilDeployed blocks until the
// distribution is deployed. It currently takes exactly 15 minutes to deploy
// but that might change in the future.
func resourceAwsCloudFrontDistributionWaitUntilDeployed(id string, meta interface{}) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"InProgress"},
		Target:     []string{"Deployed"},
		Refresh:    resourceAwsCloudFrontWebDistributionStateRefreshFunc(id, meta),
		Timeout:    90 * time.Minute,
		MinTimeout: 15 * time.Second,
		Delay:      1 * time.Minute,
	}

	_, err := stateConf.WaitForState()
	return err
}

// The refresh function for resourceAwsCloudFrontWebDistributionWaitUntilDeployed.
func resourceAwsCloudFrontWebDistributionStateRefreshFunc(id string, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		conn := meta.(*AWSClient).cloudfrontconn
		params := &cloudfront.GetDistributionInput{
			Id: aws.String(id),
		}

		resp, err := conn.GetDistribution(params)
		if err != nil {
			log.Printf("[WARN] Error retrieving CloudFront Distribution %q details: %s", id, err)
			return nil, "", err
		}

		if resp == nil {
			return nil, "", nil
		}

		return resp.Distribution, *resp.Distribution.Status, nil
	}
}
