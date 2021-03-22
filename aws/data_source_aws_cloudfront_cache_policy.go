package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsCloudFrontCachePolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudFrontCachePolicyRead,

		Schema: map[string]*schema.Schema{
			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_ttl": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:          schema.TypeString,
				ConflictsWith: []string{"name"},
				Optional:      true,
			},
			"max_ttl": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"min_ttl": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:          schema.TypeString,
				ConflictsWith: []string{"id"},
				Optional:      true,
			},
			"parameters_in_cache_key_and_forwarded_to_origin": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cookies_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cookie_behavior": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"cookies": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
													Type:     schema.TypeSet,
													Computed: true,
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
							Computed: true,
						},
						"enable_accept_encoding_gzip": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"headers_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"header_behavior": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"headers": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
													Type:     schema.TypeSet,
													Computed: true,
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
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"query_string_behavior": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"query_strings": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"items": {
													Type:     schema.TypeSet,
													Computed: true,
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
func dataSourceAwsCloudFrontCachePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	if d.Id() == "" {
		if err := dataSourceAwsCloudFrontCachePolicyFindByName(d, conn); err != nil {
			return fmt.Errorf("unable to locate cache policy by name: %s", err.Error())
		}
	}

	if d.Id() != "" {
		d.Set("id", d.Id())
		request := &cloudfront.GetCachePolicyInput{
			Id: aws.String(d.Id()),
		}

		resp, err := conn.GetCachePolicy(request)
		if err != nil {
			return fmt.Errorf("unable to retrieve cache policy with ID %s: %s", d.Id(), err.Error())
		}
		d.Set("etag", aws.StringValue(resp.ETag))

		setCloudFrontCachePolicy(d, resp.CachePolicy.CachePolicyConfig)
	}

	return nil
}

func dataSourceAwsCloudFrontCachePolicyFindByName(d *schema.ResourceData, conn *cloudfront.CloudFront) error {
	var cachePolicy *cloudfront.CachePolicy
	request := &cloudfront.ListCachePoliciesInput{}
	resp, err := conn.ListCachePolicies(request)
	if err != nil {
		return err
	}

	for _, policySummary := range resp.CachePolicyList.Items {
		if aws.StringValue(policySummary.CachePolicy.CachePolicyConfig.Name) == d.Get("name").(string) {
			cachePolicy = policySummary.CachePolicy
			break
		}
	}

	if cachePolicy != nil {
		d.SetId(aws.StringValue(cachePolicy.Id))
	}
	return nil
}
