package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsCloudFrontOriginRequestPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudFrontOriginRequestPolicyRead,

		Schema: map[string]*schema.Schema{
			"comment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cookies_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cookie_behavior": {
							Computed: true,
							Type:     schema.TypeString,
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
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"headers_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"header_behavior": {
							Computed: true,
							Type:     schema.TypeString,
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
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
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
	}
}

func dataSourceAwsCloudFrontOriginRequestPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudfrontconn

	if d.Get("id").(string) == "" {
		if err := dataSourceAwsCloudFrontOriginRequestPolicyFindByName(d, conn); err != nil {
			return fmt.Errorf("Unable to find origin request policy by name: %w", err)
		}
	}

	if d.Id() != "" {
		request := &cloudfront.GetOriginRequestPolicyInput{
			Id: aws.String(d.Id()),
		}

		resp, err := conn.GetOriginRequestPolicy(request)
		if err != nil {
			return fmt.Errorf("Unable to retrieve origin request policy with ID %s: %w", d.Id(), err)
		}

		if resp == nil || resp.OriginRequestPolicy == nil || resp.OriginRequestPolicy.OriginRequestPolicyConfig == nil {
			return nil
		}

		d.Set("etag", aws.StringValue(resp.ETag))

		originRequestPolicy := resp.OriginRequestPolicy.OriginRequestPolicyConfig
		d.Set("comment", aws.StringValue(originRequestPolicy.Comment))
		d.Set("name", aws.StringValue(originRequestPolicy.Name))
		d.Set("cookies_config", flattenCloudFrontOriginRequestPolicyCookiesConfig(originRequestPolicy.CookiesConfig))
		d.Set("headers_config", flattenCloudFrontOriginRequestPolicyHeadersConfig(originRequestPolicy.HeadersConfig))
		d.Set("query_strings_config", flattenCloudFrontOriginRequestPolicyQueryStringsConfig(originRequestPolicy.QueryStringsConfig))
	}

	return nil
}

func dataSourceAwsCloudFrontOriginRequestPolicyFindByName(d *schema.ResourceData, conn *cloudfront.CloudFront) error {
	var originRequestPolicy *cloudfront.OriginRequestPolicy
	request := &cloudfront.ListOriginRequestPoliciesInput{}
	resp, err := conn.ListOriginRequestPolicies(request)
	if err != nil {
		return err
	}

	for _, policySummary := range resp.OriginRequestPolicyList.Items {
		if aws.StringValue(policySummary.OriginRequestPolicy.OriginRequestPolicyConfig.Name) == d.Get("name").(string) {
			originRequestPolicy = policySummary.OriginRequestPolicy
			break
		}
	}

	if originRequestPolicy != nil {
		d.SetId(aws.StringValue(originRequestPolicy.Id))
	}
	return nil
}
