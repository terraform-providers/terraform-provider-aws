package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsCloudFrontCachePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudFrontCachePolicyCreate,
		Read:   resourceAwsCloudFrontCachePolicyRead,
		Update: resourceAwsCloudFrontCachePolicyUpdate,
		Delete: resourceAwsCloudFrontCachePolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"default_ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  86400,
			},
			"etag": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"parameters_in_cache_key_and_forwarded_to_origin": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cookies_config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cookie_behavior": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"none", "whitelist", "allExcept", "all"}, false),
									},
									"cookies": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
								},
							},
						},
						"enable_accept_encoding_brotli": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"enable_accept_encoding_gzip": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"headers_config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"header_behavior": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"none", "whitelist"}, false),
									},
									"headers": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},
								},
							},
						},
						"query_strings_config": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"query_string_behavior": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"none", "whitelist", "allExcept", "all"}, false),
									},
									"query_strings": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
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
}

func resourceAwsCloudFrontCachePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	request := &cloudfront.CreateCachePolicyInput{
		CachePolicyConfig: expandCloudFrontCachePolicyConfig(d),
	}

	resp, err := conn.CreateCachePolicy(request)

	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(resp.CachePolicy.Id))

	return resourceAwsCloudFrontCachePolicyRead(d, meta)
}

func resourceAwsCloudFrontCachePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn
	request := &cloudfront.GetCachePolicyInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetCachePolicy(request)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "ResourceNotFoundException") {
		log.Printf("[WARN] CloudFront Cache Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}
	d.Set("etag", aws.StringValue(resp.ETag))

	setCloudFrontCachePolicy(d, resp.CachePolicy.CachePolicyConfig)

	return nil
}

func resourceAwsCloudFrontCachePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	request := &cloudfront.UpdateCachePolicyInput{
		CachePolicyConfig: expandCloudFrontCachePolicyConfig(d),
		Id:                aws.String(d.Id()),
		IfMatch:           aws.String(d.Get("etag").(string)),
	}

	_, err := conn.UpdateCachePolicy(request)
	if err != nil {
		return err
	}

	return resourceAwsCloudFrontCachePolicyRead(d, meta)
}

func resourceAwsCloudFrontCachePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	request := &cloudfront.DeleteCachePolicyInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	}

	_, err := conn.DeleteCachePolicy(request)
	if err != nil {
		if isAWSErr(err, cloudfront.ErrCodeNoSuchCachePolicy, "") {
			return nil
		}
		return err
	}

	return nil
}
