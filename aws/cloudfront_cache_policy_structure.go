package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandCloudFrontCachePolicyCookieNames(tfMap map[string]interface{}) *cloudfront.CookieNames {
	if tfMap == nil {
		return nil
	}

	items := expandStringSet(tfMap["items"].(*schema.Set))

	apiObject := &cloudfront.CookieNames{
		Items:    items,
		Quantity: aws.Int64(int64(len(items))),
	}

	return apiObject
}

func expandCloudFrontCachePolicyCookiesConfig(tfMap map[string]interface{}) *cloudfront.CachePolicyCookiesConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.CachePolicyCookiesConfig{
		CookieBehavior: aws.String(tfMap["cookie_behavior"].(string)),
	}

	if items, ok := tfMap["cookies"].([]interface{}); ok && len(items) == 1 {
		apiObject.Cookies = expandCloudFrontCachePolicyCookieNames(items[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCloudFrontCachePolicyHeaders(tfMap map[string]interface{}) *cloudfront.Headers {
	if tfMap == nil {
		return nil
	}

	items := expandStringSet(tfMap["items"].(*schema.Set))

	apiObject := &cloudfront.Headers{
		Items:    items,
		Quantity: aws.Int64(int64(len(items))),
	}

	return apiObject
}

func expandCloudFrontCachePolicyHeadersConfig(tfMap map[string]interface{}) *cloudfront.CachePolicyHeadersConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.CachePolicyHeadersConfig{
		HeaderBehavior: aws.String(tfMap["header_behavior"].(string)),
	}

	if items, ok := tfMap["headers"].([]interface{}); ok && len(items) == 1 && tfMap["header_behavior"] != "none" {
		apiObject.Headers = expandCloudFrontCachePolicyHeaders(items[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCloudFrontCachePolicyQueryStringNames(tfMap map[string]interface{}) *cloudfront.QueryStringNames {
	if tfMap == nil {
		return nil
	}

	items := expandStringSet(tfMap["items"].(*schema.Set))

	apiObject := &cloudfront.QueryStringNames{
		Items:    items,
		Quantity: aws.Int64(int64(len(items))),
	}

	return apiObject
}

func expandCloudFrontCachePolicyQueryStringConfig(tfMap map[string]interface{}) *cloudfront.CachePolicyQueryStringsConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.CachePolicyQueryStringsConfig{
		QueryStringBehavior: aws.String(tfMap["query_string_behavior"].(string)),
	}

	if items, ok := tfMap["query_strings"].([]interface{}); ok && len(items) == 1 {
		apiObject.QueryStrings = expandCloudFrontCachePolicyQueryStringNames(items[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCloudFrontCachePolicyParametersConfig(tfMap map[string]interface{}) *cloudfront.ParametersInCacheKeyAndForwardedToOrigin {
	if tfMap == nil {
		return nil
	}

	var cookiesConfig *cloudfront.CachePolicyCookiesConfig
	var headersConfig *cloudfront.CachePolicyHeadersConfig
	var queryStringsConfig *cloudfront.CachePolicyQueryStringsConfig

	if cookiesFlat, ok := tfMap["cookies_config"].([]interface{}); ok && len(cookiesFlat) == 1 {
		cookiesConfig = expandCloudFrontCachePolicyCookiesConfig(cookiesFlat[0].(map[string]interface{}))
	}

	if headersFlat, ok := tfMap["headers_config"].([]interface{}); ok && len(headersFlat) == 1 {
		headersConfig = expandCloudFrontCachePolicyHeadersConfig(headersFlat[0].(map[string]interface{}))
	}

	if queryStringsFlat, ok := tfMap["query_strings_config"].([]interface{}); ok && len(queryStringsFlat) == 1 {
		queryStringsConfig = expandCloudFrontCachePolicyQueryStringConfig(queryStringsFlat[0].(map[string]interface{}))
	}

	parametersConfig := &cloudfront.ParametersInCacheKeyAndForwardedToOrigin{
		CookiesConfig:              cookiesConfig,
		EnableAcceptEncodingBrotli: aws.Bool(tfMap["enable_accept_encoding_brotli"].(bool)),
		EnableAcceptEncodingGzip:   aws.Bool(tfMap["enable_accept_encoding_gzip"].(bool)),
		HeadersConfig:              headersConfig,
		QueryStringsConfig:         queryStringsConfig,
	}

	return parametersConfig
}

func expandCloudFrontCachePolicyConfig(d *schema.ResourceData) *cloudfront.CachePolicyConfig {
	parametersConfig := &cloudfront.ParametersInCacheKeyAndForwardedToOrigin{}

	if parametersFlat, ok := d.GetOk("parameters_in_cache_key_and_forwarded_to_origin"); ok {
		parametersConfig = expandCloudFrontCachePolicyParametersConfig(parametersFlat.([]interface{})[0].(map[string]interface{}))
	}
	cachePolicy := &cloudfront.CachePolicyConfig{
		Comment:                                  aws.String(d.Get("comment").(string)),
		DefaultTTL:                               aws.Int64(int64(d.Get("default_ttl").(int))),
		MaxTTL:                                   aws.Int64(int64(d.Get("max_ttl").(int))),
		MinTTL:                                   aws.Int64(int64(d.Get("min_ttl").(int))),
		Name:                                     aws.String(d.Get("name").(string)),
		ParametersInCacheKeyAndForwardedToOrigin: parametersConfig,
	}

	return cachePolicy
}

func flattenCloudFrontCachePolicyCookiesConfig(cookiesConfig *cloudfront.CachePolicyCookiesConfig) []map[string]interface{} {
	cookiesConfigFlat := map[string]interface{}{}

	cookies := []map[string]interface{}{}
	if cookiesConfig.Cookies != nil {
		cookies = []map[string]interface{}{
			{
				"items": cookiesConfig.Cookies.Items,
			},
		}
	}

	cookiesConfigFlat["cookie_behavior"] = aws.StringValue(cookiesConfig.CookieBehavior)
	cookiesConfigFlat["cookies"] = cookies

	return []map[string]interface{}{
		cookiesConfigFlat,
	}
}

func flattenCloudFrontCachePolicyHeadersConfig(headersConfig *cloudfront.CachePolicyHeadersConfig) []map[string]interface{} {
	headersConfigFlat := map[string]interface{}{}

	headers := []map[string]interface{}{}
	if headersConfig.Headers != nil {
		headers = []map[string]interface{}{
			{
				"items": headersConfig.Headers.Items,
			},
		}
	}

	headersConfigFlat["header_behavior"] = aws.StringValue(headersConfig.HeaderBehavior)
	headersConfigFlat["headers"] = headers

	return []map[string]interface{}{
		headersConfigFlat,
	}
}

func flattenCloudFrontCachePolicyQueryStringsConfig(queryStringsConfig *cloudfront.CachePolicyQueryStringsConfig) []map[string]interface{} {
	queryStringsConfigFlat := map[string]interface{}{}

	queryStrings := []map[string]interface{}{}
	if queryStringsConfig.QueryStrings != nil {
		queryStrings = []map[string]interface{}{
			{
				"items": queryStringsConfig.QueryStrings.Items,
			},
		}
	}

	queryStringsConfigFlat["query_string_behavior"] = aws.StringValue(queryStringsConfig.QueryStringBehavior)
	queryStringsConfigFlat["query_strings"] = queryStrings

	return []map[string]interface{}{
		queryStringsConfigFlat,
	}
}

func setParametersConfig(parametersConfig *cloudfront.ParametersInCacheKeyAndForwardedToOrigin) []map[string]interface{} {
	parametersConfigFlat := map[string]interface{}{
		"enable_accept_encoding_brotli": aws.BoolValue(parametersConfig.EnableAcceptEncodingBrotli),
		"enable_accept_encoding_gzip":   aws.BoolValue(parametersConfig.EnableAcceptEncodingGzip),
		"cookies_config":                flattenCloudFrontCachePolicyCookiesConfig(parametersConfig.CookiesConfig),
		"headers_config":                flattenCloudFrontCachePolicyHeadersConfig(parametersConfig.HeadersConfig),
		"query_strings_config":          flattenCloudFrontCachePolicyQueryStringsConfig(parametersConfig.QueryStringsConfig),
	}

	return []map[string]interface{}{
		parametersConfigFlat,
	}
}

func setCloudFrontCachePolicy(d *schema.ResourceData, cachePolicy *cloudfront.CachePolicyConfig) {
	d.Set("comment", aws.StringValue(cachePolicy.Comment))
	d.Set("default_ttl", aws.Int64Value(cachePolicy.DefaultTTL))
	d.Set("max_ttl", aws.Int64Value(cachePolicy.MaxTTL))
	d.Set("min_ttl", aws.Int64Value(cachePolicy.MinTTL))
	d.Set("name", aws.StringValue(cachePolicy.Name))
	d.Set("parameters_in_cache_key_and_forwarded_to_origin", setParametersConfig(cachePolicy.ParametersInCacheKeyAndForwardedToOrigin))
}
