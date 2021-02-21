// CloudFront DistributionConfig structure helpers.
//
// These functions assist in pulling in data from Terraform resource
// configuration for the aws_cloudfront_distribution resource, as there are
// several sub-fields that require their own data type, and do not necessarily
// 1-1 translate to resource configuration.

package aws

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
)

// cloudFrontRoute53ZoneID defines the route 53 zone ID for CloudFront. This
// is used to set the zone_id attribute.
const cloudFrontRoute53ZoneID = "Z2FDTNDATAQYW2"

// Assemble the *cloudfront.DistributionConfig variable. Calls out to various
// expander functions to convert attributes and sub-attributes to the various
// complex structures which are necessary to properly build the
// DistributionConfig structure.
//
// Used by the aws_cloudfront_distribution Create and Update functions.
func expandDistributionConfig(d *schema.ResourceData) *cloudfront.DistributionConfig {
	distributionConfig := &cloudfront.DistributionConfig{
		CacheBehaviors:       expandCacheBehaviors(d.Get("ordered_cache_behavior").([]interface{})),
		CallerReference:      aws.String(resource.UniqueId()),
		Comment:              aws.String(d.Get("comment").(string)),
		CustomErrorResponses: expandCustomErrorResponses(d.Get("custom_error_response").(*schema.Set)),
		DefaultCacheBehavior: expandCloudFrontDefaultCacheBehavior(d.Get("default_cache_behavior").([]interface{})[0].(map[string]interface{})),
		DefaultRootObject:    aws.String(d.Get("default_root_object").(string)),
		Enabled:              aws.Bool(d.Get("enabled").(bool)),
		IsIPV6Enabled:        aws.Bool(d.Get("is_ipv6_enabled").(bool)),
		HttpVersion:          aws.String(d.Get("http_version").(string)),
		Origins:              expandOrigins(d.Get("origin").(*schema.Set)),
		PriceClass:           aws.String(d.Get("price_class").(string)),
		WebACLId:             aws.String(d.Get("web_acl_id").(string)),
	}

	// This sets CallerReference if it's still pending computation (ie: new resource)
	if v, ok := d.GetOk("caller_reference"); ok {
		distributionConfig.CallerReference = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging_config"); ok {
		distributionConfig.Logging = expandLoggingConfig(v.([]interface{})[0].(map[string]interface{}))
	} else {
		distributionConfig.Logging = expandLoggingConfig(nil)
	}
	if v, ok := d.GetOk("aliases"); ok {
		distributionConfig.Aliases = expandAliases(v.(*schema.Set))
	} else {
		distributionConfig.Aliases = expandAliases(schema.NewSet(aliasesHash, []interface{}{}))
	}
	if v, ok := d.GetOk("restrictions"); ok {
		distributionConfig.Restrictions = expandRestrictions(v.([]interface{})[0].(map[string]interface{}))
	}
	if v, ok := d.GetOk("viewer_certificate"); ok {
		distributionConfig.ViewerCertificate = expandViewerCertificate(v.([]interface{})[0].(map[string]interface{}))
	}
	if v, ok := d.GetOk("origin_group"); ok {
		distributionConfig.OriginGroups = expandOriginGroups(v.(*schema.Set))
	}
	return distributionConfig
}

// Unpack the *cloudfront.DistributionConfig variable and set resource data.
// Calls out to flatten functions to convert the DistributionConfig
// sub-structures to their respective attributes in the
// aws_cloudfront_distribution resource.
//
// Used by the aws_cloudfront_distribution Read function.
func flattenDistributionConfig(d *schema.ResourceData, distributionConfig *cloudfront.DistributionConfig) error {
	var err error

	d.Set("enabled", distributionConfig.Enabled)
	d.Set("is_ipv6_enabled", distributionConfig.IsIPV6Enabled)
	d.Set("price_class", distributionConfig.PriceClass)
	d.Set("hosted_zone_id", cloudFrontRoute53ZoneID)

	err = d.Set("default_cache_behavior", flattenDefaultCacheBehavior(distributionConfig.DefaultCacheBehavior))
	if err != nil {
		return err
	}
	err = d.Set("viewer_certificate", flattenViewerCertificate(distributionConfig.ViewerCertificate))
	if err != nil {
		return err
	}

	if distributionConfig.CallerReference != nil {
		d.Set("caller_reference", distributionConfig.CallerReference)
	}
	if distributionConfig.Comment != nil {
		if *distributionConfig.Comment != "" {
			d.Set("comment", distributionConfig.Comment)
		}
	}
	if distributionConfig.DefaultRootObject != nil {
		d.Set("default_root_object", distributionConfig.DefaultRootObject)
	}
	if distributionConfig.HttpVersion != nil {
		d.Set("http_version", distributionConfig.HttpVersion)
	}
	if distributionConfig.WebACLId != nil {
		d.Set("web_acl_id", distributionConfig.WebACLId)
	}

	if distributionConfig.CustomErrorResponses != nil {
		err = d.Set("custom_error_response", flattenCustomErrorResponses(distributionConfig.CustomErrorResponses))
		if err != nil {
			return err
		}
	}
	if distributionConfig.CacheBehaviors != nil {
		if err := d.Set("ordered_cache_behavior", flattenCacheBehaviors(distributionConfig.CacheBehaviors)); err != nil {
			return err
		}
	}

	if distributionConfig.Logging != nil && *distributionConfig.Logging.Enabled {
		err = d.Set("logging_config", flattenLoggingConfig(distributionConfig.Logging))
	} else {
		err = d.Set("logging_config", []interface{}{})
	}
	if err != nil {
		return err
	}

	if distributionConfig.Aliases != nil {
		err = d.Set("aliases", flattenAliases(distributionConfig.Aliases))
		if err != nil {
			return err
		}
	}
	if distributionConfig.Restrictions != nil {
		err = d.Set("restrictions", flattenRestrictions(distributionConfig.Restrictions))
		if err != nil {
			return err
		}
	}
	if *distributionConfig.Origins.Quantity > 0 {
		err = d.Set("origin", flattenOrigins(distributionConfig.Origins))
		if err != nil {
			return err
		}
	}
	if *distributionConfig.OriginGroups.Quantity > 0 {
		err = d.Set("origin_group", flattenOriginGroups(distributionConfig.OriginGroups))
		if err != nil {
			return err
		}
	}

	return nil
}

func flattenDefaultCacheBehavior(dcb *cloudfront.DefaultCacheBehavior) []interface{} {
	return []interface{}{flattenCloudFrontDefaultCacheBehavior(dcb)}
}

func expandCacheBehaviors(lst []interface{}) *cloudfront.CacheBehaviors {
	var qty int64
	var items []*cloudfront.CacheBehavior
	for _, v := range lst {
		items = append(items, expandCacheBehavior(v.(map[string]interface{})))
		qty++
	}
	return &cloudfront.CacheBehaviors{
		Quantity: aws.Int64(qty),
		Items:    items,
	}
}

func flattenCacheBehaviors(cbs *cloudfront.CacheBehaviors) []interface{} {
	lst := []interface{}{}
	for _, v := range cbs.Items {
		lst = append(lst, flattenCacheBehavior(v))
	}
	return lst
}

func expandCloudFrontDefaultCacheBehavior(m map[string]interface{}) *cloudfront.DefaultCacheBehavior {
	dcb := &cloudfront.DefaultCacheBehavior{
		CachePolicyId:          aws.String(m["cache_policy_id"].(string)),
		Compress:               aws.Bool(m["compress"].(bool)),
		FieldLevelEncryptionId: aws.String(m["field_level_encryption_id"].(string)),
		OriginRequestPolicyId:  aws.String(m["origin_request_policy_id"].(string)),
		TargetOriginId:         aws.String(m["target_origin_id"].(string)),
		ViewerProtocolPolicy:   aws.String(m["viewer_protocol_policy"].(string)),
	}

	if forwardedValuesFlat, ok := m["forwarded_values"].([]interface{}); ok && len(forwardedValuesFlat) == 1 {
		dcb.ForwardedValues = expandForwardedValues(m["forwarded_values"].([]interface{})[0].(map[string]interface{}))
	}

	if m["cache_policy_id"].(string) == "" {
		dcb.MinTTL = aws.Int64(int64(m["min_ttl"].(int)))
		dcb.MaxTTL = aws.Int64(int64(m["max_ttl"].(int)))
		dcb.DefaultTTL = aws.Int64(int64(m["default_ttl"].(int)))
	}

	if v, ok := m["trusted_signers"]; ok {
		dcb.TrustedSigners = expandTrustedSigners(v.([]interface{}))
	} else {
		dcb.TrustedSigners = expandTrustedSigners([]interface{}{})
	}

	if v, ok := m["lambda_function_association"]; ok {
		dcb.LambdaFunctionAssociations = expandLambdaFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := m["smooth_streaming"]; ok {
		dcb.SmoothStreaming = aws.Bool(v.(bool))
	}
	if v, ok := m["allowed_methods"]; ok {
		dcb.AllowedMethods = expandAllowedMethods(v.(*schema.Set))
	}
	if v, ok := m["cached_methods"]; ok {
		dcb.AllowedMethods.CachedMethods = expandCachedMethods(v.(*schema.Set))
	}
	if v, ok := m["realtime_log_config_arn"]; ok && v.(string) != "" {
		dcb.RealtimeLogConfigArn = aws.String(v.(string))
	}

	return dcb
}

func expandCacheBehavior(m map[string]interface{}) *cloudfront.CacheBehavior {
	var forwardedValues *cloudfront.ForwardedValues
	if forwardedValuesFlat, ok := m["forwarded_values"].([]interface{}); ok && len(forwardedValuesFlat) == 1 {
		forwardedValues = expandForwardedValues(m["forwarded_values"].([]interface{})[0].(map[string]interface{}))
	}

	cb := &cloudfront.CacheBehavior{
		CachePolicyId:          aws.String(m["cache_policy_id"].(string)),
		Compress:               aws.Bool(m["compress"].(bool)),
		FieldLevelEncryptionId: aws.String(m["field_level_encryption_id"].(string)),
		ForwardedValues:        forwardedValues,
		OriginRequestPolicyId:  aws.String(m["origin_request_policy_id"].(string)),
		TargetOriginId:         aws.String(m["target_origin_id"].(string)),
		ViewerProtocolPolicy:   aws.String(m["viewer_protocol_policy"].(string)),
	}

	if m["cache_policy_id"].(string) == "" {
		cb.MinTTL = aws.Int64(int64(m["min_ttl"].(int)))
		cb.MaxTTL = aws.Int64(int64(m["max_ttl"].(int)))
		cb.DefaultTTL = aws.Int64(int64(m["default_ttl"].(int)))
	}

	if v, ok := m["trusted_signers"]; ok {
		cb.TrustedSigners = expandTrustedSigners(v.([]interface{}))
	} else {
		cb.TrustedSigners = expandTrustedSigners([]interface{}{})
	}

	if v, ok := m["lambda_function_association"]; ok {
		cb.LambdaFunctionAssociations = expandLambdaFunctionAssociations(v.(*schema.Set).List())
	}

	if v, ok := m["smooth_streaming"]; ok {
		cb.SmoothStreaming = aws.Bool(v.(bool))
	}
	if v, ok := m["allowed_methods"]; ok {
		cb.AllowedMethods = expandAllowedMethods(v.(*schema.Set))
	}
	if v, ok := m["cached_methods"]; ok {
		cb.AllowedMethods.CachedMethods = expandCachedMethods(v.(*schema.Set))
	}
	if v, ok := m["path_pattern"]; ok {
		cb.PathPattern = aws.String(v.(string))
	}
	if v, ok := m["realtime_log_config_arn"]; ok && v.(string) != "" {
		cb.RealtimeLogConfigArn = aws.String(v.(string))
	}

	return cb
}

func flattenCloudFrontDefaultCacheBehavior(dcb *cloudfront.DefaultCacheBehavior) map[string]interface{} {
	m := map[string]interface{}{
		"cache_policy_id":           aws.StringValue(dcb.CachePolicyId),
		"compress":                  aws.BoolValue(dcb.Compress),
		"field_level_encryption_id": aws.StringValue(dcb.FieldLevelEncryptionId),
		"viewer_protocol_policy":    aws.StringValue(dcb.ViewerProtocolPolicy),
		"target_origin_id":          aws.StringValue(dcb.TargetOriginId),
		"min_ttl":                   aws.Int64Value(dcb.MinTTL),
		"origin_request_policy_id":  aws.StringValue(dcb.OriginRequestPolicyId),
		"realtime_log_config_arn":   aws.StringValue(dcb.RealtimeLogConfigArn),
	}

	if dcb.ForwardedValues != nil {
		m["forwarded_values"] = []interface{}{flattenForwardedValues(dcb.ForwardedValues)}
	}
	if len(dcb.TrustedSigners.Items) > 0 {
		m["trusted_signers"] = flattenTrustedSigners(dcb.TrustedSigners)
	}
	if len(dcb.LambdaFunctionAssociations.Items) > 0 {
		m["lambda_function_association"] = flattenLambdaFunctionAssociations(dcb.LambdaFunctionAssociations)
	}
	if dcb.MaxTTL != nil {
		m["max_ttl"] = aws.Int64Value(dcb.MaxTTL)
	}
	if dcb.SmoothStreaming != nil {
		m["smooth_streaming"] = aws.BoolValue(dcb.SmoothStreaming)
	}
	if dcb.DefaultTTL != nil {
		m["default_ttl"] = int(aws.Int64Value(dcb.DefaultTTL))
	}
	if dcb.AllowedMethods != nil {
		m["allowed_methods"] = flattenAllowedMethods(dcb.AllowedMethods)
	}
	if dcb.AllowedMethods.CachedMethods != nil {
		m["cached_methods"] = flattenCachedMethods(dcb.AllowedMethods.CachedMethods)
	}

	return m
}

func flattenCacheBehavior(cb *cloudfront.CacheBehavior) map[string]interface{} {
	m := make(map[string]interface{})

	m["cache_policy_id"] = aws.StringValue(cb.CachePolicyId)
	m["compress"] = aws.BoolValue(cb.Compress)
	m["field_level_encryption_id"] = aws.StringValue(cb.FieldLevelEncryptionId)
	m["viewer_protocol_policy"] = aws.StringValue(cb.ViewerProtocolPolicy)
	m["target_origin_id"] = aws.StringValue(cb.TargetOriginId)
	m["min_ttl"] = int(aws.Int64Value(cb.MinTTL))
	m["origin_request_policy_id"] = aws.StringValue(cb.OriginRequestPolicyId)
	m["realtime_log_config_arn"] = aws.StringValue(cb.RealtimeLogConfigArn)

	if cb.ForwardedValues != nil {
		m["forwarded_values"] = []interface{}{flattenForwardedValues(cb.ForwardedValues)}
	}
	if len(cb.TrustedSigners.Items) > 0 {
		m["trusted_signers"] = flattenTrustedSigners(cb.TrustedSigners)
	}
	if len(cb.LambdaFunctionAssociations.Items) > 0 {
		m["lambda_function_association"] = flattenLambdaFunctionAssociations(cb.LambdaFunctionAssociations)
	}
	if cb.MaxTTL != nil {
		m["max_ttl"] = int(*cb.MaxTTL)
	}
	if cb.SmoothStreaming != nil {
		m["smooth_streaming"] = *cb.SmoothStreaming
	}
	if cb.DefaultTTL != nil {
		m["default_ttl"] = int(*cb.DefaultTTL)
	}
	if cb.AllowedMethods != nil {
		m["allowed_methods"] = flattenAllowedMethods(cb.AllowedMethods)
	}
	if cb.AllowedMethods.CachedMethods != nil {
		m["cached_methods"] = flattenCachedMethods(cb.AllowedMethods.CachedMethods)
	}
	if cb.PathPattern != nil {
		m["path_pattern"] = *cb.PathPattern
	}
	return m
}

func expandTrustedSigners(s []interface{}) *cloudfront.TrustedSigners {
	var ts cloudfront.TrustedSigners
	if len(s) > 0 {
		ts.Quantity = aws.Int64(int64(len(s)))
		ts.Items = expandStringList(s)
		ts.Enabled = aws.Bool(true)
	} else {
		ts.Quantity = aws.Int64(0)
		ts.Enabled = aws.Bool(false)
	}
	return &ts
}

func flattenTrustedSigners(ts *cloudfront.TrustedSigners) []interface{} {
	if ts.Items != nil {
		return flattenStringList(ts.Items)
	}
	return []interface{}{}
}

func lambdaFunctionAssociationHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["event_type"].(string)))
	buf.WriteString(m["lambda_arn"].(string))
	buf.WriteString(fmt.Sprintf("%t", m["include_body"].(bool)))
	return hashcode.String(buf.String())
}

func expandLambdaFunctionAssociations(v interface{}) *cloudfront.LambdaFunctionAssociations {
	if v == nil {
		return &cloudfront.LambdaFunctionAssociations{
			Quantity: aws.Int64(0),
		}
	}

	s := v.([]interface{})
	var lfa cloudfront.LambdaFunctionAssociations
	lfa.Quantity = aws.Int64(int64(len(s)))
	lfa.Items = make([]*cloudfront.LambdaFunctionAssociation, len(s))
	for i, lf := range s {
		lfa.Items[i] = expandLambdaFunctionAssociation(lf.(map[string]interface{}))
	}
	return &lfa
}

func expandLambdaFunctionAssociation(lf map[string]interface{}) *cloudfront.LambdaFunctionAssociation {
	var lfa cloudfront.LambdaFunctionAssociation
	if v, ok := lf["event_type"]; ok {
		lfa.EventType = aws.String(v.(string))
	}
	if v, ok := lf["lambda_arn"]; ok {
		lfa.LambdaFunctionARN = aws.String(v.(string))
	}
	if v, ok := lf["include_body"]; ok {
		lfa.IncludeBody = aws.Bool(v.(bool))
	}
	return &lfa
}

func flattenLambdaFunctionAssociations(lfa *cloudfront.LambdaFunctionAssociations) *schema.Set {
	s := schema.NewSet(lambdaFunctionAssociationHash, []interface{}{})
	for _, v := range lfa.Items {
		s.Add(flattenLambdaFunctionAssociation(v))
	}
	return s
}

func flattenLambdaFunctionAssociation(lfa *cloudfront.LambdaFunctionAssociation) map[string]interface{} {
	m := map[string]interface{}{}
	if lfa != nil {
		m["event_type"] = aws.StringValue(lfa.EventType)
		m["lambda_arn"] = aws.StringValue(lfa.LambdaFunctionARN)
		m["include_body"] = aws.BoolValue(lfa.IncludeBody)
	}
	return m
}

func expandForwardedValues(m map[string]interface{}) *cloudfront.ForwardedValues {
	if len(m) < 1 {
		return nil
	}

	fv := &cloudfront.ForwardedValues{
		QueryString: aws.Bool(m["query_string"].(bool)),
	}
	if v, ok := m["cookies"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		fv.Cookies = expandCookiePreference(v.([]interface{})[0].(map[string]interface{}))
	}
	if v, ok := m["headers"]; ok {
		fv.Headers = expandHeaders(v.(*schema.Set).List())
	}
	if v, ok := m["query_string_cache_keys"]; ok {
		fv.QueryStringCacheKeys = expandQueryStringCacheKeys(v.([]interface{}))
	}
	return fv
}

func flattenForwardedValues(fv *cloudfront.ForwardedValues) map[string]interface{} {
	m := make(map[string]interface{})
	m["query_string"] = aws.BoolValue(fv.QueryString)
	if fv.Cookies != nil {
		m["cookies"] = []interface{}{flattenCookiePreference(fv.Cookies)}
	}
	if fv.Headers != nil {
		m["headers"] = schema.NewSet(schema.HashString, flattenHeaders(fv.Headers))
	}
	if fv.QueryStringCacheKeys != nil {
		m["query_string_cache_keys"] = flattenQueryStringCacheKeys(fv.QueryStringCacheKeys)
	}
	return m
}

func expandHeaders(d []interface{}) *cloudfront.Headers {
	return &cloudfront.Headers{
		Quantity: aws.Int64(int64(len(d))),
		Items:    expandStringList(d),
	}
}

func flattenHeaders(h *cloudfront.Headers) []interface{} {
	if h.Items != nil {
		return flattenStringList(h.Items)
	}
	return []interface{}{}
}

func expandQueryStringCacheKeys(d []interface{}) *cloudfront.QueryStringCacheKeys {
	return &cloudfront.QueryStringCacheKeys{
		Quantity: aws.Int64(int64(len(d))),
		Items:    expandStringList(d),
	}
}

func flattenQueryStringCacheKeys(k *cloudfront.QueryStringCacheKeys) []interface{} {
	if k.Items != nil {
		return flattenStringList(k.Items)
	}
	return []interface{}{}
}

func expandCookiePreference(m map[string]interface{}) *cloudfront.CookiePreference {
	cp := &cloudfront.CookiePreference{
		Forward: aws.String(m["forward"].(string)),
	}
	if v, ok := m["whitelisted_names"]; ok {
		cp.WhitelistedNames = expandCookieNames(v.(*schema.Set).List())
	}
	return cp
}

func flattenCookiePreference(cp *cloudfront.CookiePreference) map[string]interface{} {
	m := make(map[string]interface{})
	m["forward"] = aws.StringValue(cp.Forward)
	if cp.WhitelistedNames != nil {
		m["whitelisted_names"] = schema.NewSet(schema.HashString, flattenCookieNames(cp.WhitelistedNames))
	}
	return m
}

func expandCookieNames(d []interface{}) *cloudfront.CookieNames {
	return &cloudfront.CookieNames{
		Quantity: aws.Int64(int64(len(d))),
		Items:    expandStringList(d),
	}
}

func flattenCookieNames(cn *cloudfront.CookieNames) []interface{} {
	if cn.Items != nil {
		return flattenStringList(cn.Items)
	}
	return []interface{}{}
}

func expandAllowedMethods(s *schema.Set) *cloudfront.AllowedMethods {
	return &cloudfront.AllowedMethods{
		Quantity: aws.Int64(int64(s.Len())),
		Items:    expandStringSet(s),
	}
}

func flattenAllowedMethods(am *cloudfront.AllowedMethods) *schema.Set {
	if am.Items != nil {
		return flattenStringSet(am.Items)
	}
	return nil
}

func expandCachedMethods(s *schema.Set) *cloudfront.CachedMethods {
	return &cloudfront.CachedMethods{
		Quantity: aws.Int64(int64(s.Len())),
		Items:    expandStringSet(s),
	}
}

func flattenCachedMethods(cm *cloudfront.CachedMethods) *schema.Set {
	if cm.Items != nil {
		return flattenStringSet(cm.Items)
	}
	return nil
}

func expandOrigins(s *schema.Set) *cloudfront.Origins {
	qty := 0
	items := []*cloudfront.Origin{}
	for _, v := range s.List() {
		items = append(items, expandOrigin(v.(map[string]interface{})))
		qty++
	}
	return &cloudfront.Origins{
		Quantity: aws.Int64(int64(qty)),
		Items:    items,
	}
}

func flattenOrigins(ors *cloudfront.Origins) *schema.Set {
	s := []interface{}{}
	for _, v := range ors.Items {
		s = append(s, flattenOrigin(v))
	}
	return schema.NewSet(originHash, s)
}

func expandOrigin(m map[string]interface{}) *cloudfront.Origin {
	origin := &cloudfront.Origin{
		Id:         aws.String(m["origin_id"].(string)),
		DomainName: aws.String(m["domain_name"].(string)),
	}
	if v, ok := m["custom_header"]; ok {
		origin.CustomHeaders = expandCustomHeaders(v.(*schema.Set))
	}
	if v, ok := m["custom_origin_config"]; ok {
		if s := v.([]interface{}); len(s) > 0 {
			origin.CustomOriginConfig = expandCustomOriginConfig(s[0].(map[string]interface{}))
		}
	}
	if v, ok := m["origin_path"]; ok {
		origin.OriginPath = aws.String(v.(string))
	}
	if v, ok := m["s3_origin_config"]; ok {
		if s := v.([]interface{}); len(s) > 0 {
			origin.S3OriginConfig = expandS3OriginConfig(s[0].(map[string]interface{}))
		}
	}

	// if both custom and s3 origin are missing, add an empty s3 origin
	// One or the other must be specified, but the S3 origin can be "empty"
	if origin.S3OriginConfig == nil && origin.CustomOriginConfig == nil {
		origin.S3OriginConfig = &cloudfront.S3OriginConfig{
			OriginAccessIdentity: aws.String(""),
		}
	}

	return origin
}

func flattenOrigin(or *cloudfront.Origin) map[string]interface{} {
	m := make(map[string]interface{})
	m["origin_id"] = aws.StringValue(or.Id)
	m["domain_name"] = aws.StringValue(or.DomainName)
	if or.CustomHeaders != nil {
		m["custom_header"] = flattenCustomHeaders(or.CustomHeaders)
	}
	if or.CustomOriginConfig != nil {
		m["custom_origin_config"] = []interface{}{flattenCustomOriginConfig(or.CustomOriginConfig)}
	}
	if or.OriginPath != nil {
		m["origin_path"] = aws.StringValue(or.OriginPath)
	}
	if or.S3OriginConfig != nil && aws.StringValue(or.S3OriginConfig.OriginAccessIdentity) != "" {
		m["s3_origin_config"] = []interface{}{flattenS3OriginConfig(or.S3OriginConfig)}
	}
	return m
}

func expandOriginGroups(s *schema.Set) *cloudfront.OriginGroups {
	qty := 0
	items := []*cloudfront.OriginGroup{}
	for _, v := range s.List() {
		items = append(items, expandOriginGroup(v.(map[string]interface{})))
		qty++
	}
	return &cloudfront.OriginGroups{
		Quantity: aws.Int64(int64(qty)),
		Items:    items,
	}
}

func flattenOriginGroups(ogs *cloudfront.OriginGroups) *schema.Set {
	s := []interface{}{}
	for _, v := range ogs.Items {
		s = append(s, flattenOriginGroup(v))
	}
	return schema.NewSet(originGroupHash, s)
}

func expandOriginGroup(m map[string]interface{}) *cloudfront.OriginGroup {
	failoverCriteria := m["failover_criteria"].([]interface{})[0].(map[string]interface{})
	members := m["member"].([]interface{})
	originGroup := &cloudfront.OriginGroup{
		Id:               aws.String(m["origin_id"].(string)),
		FailoverCriteria: expandOriginGroupFailoverCriteria(failoverCriteria),
		Members:          expandMembers(members),
	}
	return originGroup
}

func flattenOriginGroup(og *cloudfront.OriginGroup) map[string]interface{} {
	m := make(map[string]interface{})
	m["origin_id"] = aws.StringValue(og.Id)
	if og.FailoverCriteria != nil {
		m["failover_criteria"] = flattenOriginGroupFailoverCriteria(og.FailoverCriteria)
	}
	if og.Members != nil {
		m["member"] = flattenOriginGroupMembers(og.Members)
	}
	return m
}

func expandOriginGroupFailoverCriteria(m map[string]interface{}) *cloudfront.OriginGroupFailoverCriteria {
	failoverCriteria := &cloudfront.OriginGroupFailoverCriteria{}
	if v, ok := m["status_codes"]; ok {
		codes := []*int64{}
		for _, code := range v.(*schema.Set).List() {
			codes = append(codes, aws.Int64(int64(code.(int))))
		}
		failoverCriteria.StatusCodes = &cloudfront.StatusCodes{
			Items:    codes,
			Quantity: aws.Int64(int64(len(codes))),
		}
	}
	return failoverCriteria
}

func flattenOriginGroupFailoverCriteria(ogfc *cloudfront.OriginGroupFailoverCriteria) []interface{} {
	m := make(map[string]interface{})
	if ogfc.StatusCodes.Items != nil {
		l := []interface{}{}
		for _, i := range ogfc.StatusCodes.Items {
			l = append(l, int(aws.Int64Value(i)))
		}
		m["status_codes"] = schema.NewSet(schema.HashInt, l)
	}
	return []interface{}{m}
}

func expandMembers(l []interface{}) *cloudfront.OriginGroupMembers {
	qty := 0
	items := []*cloudfront.OriginGroupMember{}
	for _, m := range l {
		ogm := &cloudfront.OriginGroupMember{
			OriginId: aws.String(m.(map[string]interface{})["origin_id"].(string)),
		}
		items = append(items, ogm)
		qty++
	}
	return &cloudfront.OriginGroupMembers{
		Quantity: aws.Int64(int64(qty)),
		Items:    items,
	}
}

func flattenOriginGroupMembers(ogm *cloudfront.OriginGroupMembers) []interface{} {
	s := []interface{}{}
	for _, i := range ogm.Items {
		m := map[string]interface{}{
			"origin_id": aws.StringValue(i.OriginId),
		}
		s = append(s, m)
	}
	return s
}

// Assemble the hash for the aws_cloudfront_distribution origin
// TypeSet attribute.
func originHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["origin_id"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["domain_name"].(string)))
	if v, ok := m["custom_header"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", customHeadersHash(v.(*schema.Set))))
	}
	if v, ok := m["custom_origin_config"]; ok {
		if s := v.([]interface{}); len(s) > 0 && s[0] != nil {
			buf.WriteString(fmt.Sprintf("%d-", customOriginConfigHash((s[0].(map[string]interface{})))))
		}
	}
	if v, ok := m["origin_path"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["s3_origin_config"]; ok {
		if s := v.([]interface{}); len(s) > 0 && s[0] != nil {
			buf.WriteString(fmt.Sprintf("%d-", s3OriginConfigHash((s[0].(map[string]interface{})))))
		}
	}
	return hashcode.String(buf.String())
}

// Assemble the hash for the aws_cloudfront_distribution origin group
// TypeSet attribute.
func originGroupHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["origin_id"].(string)))
	if v, ok := m["failover_criteria"]; ok {
		if l := v.([]interface{}); len(l) > 0 {
			buf.WriteString(fmt.Sprintf("%d-", failoverCriteriaHash(l[0])))
		}
	}
	if v, ok := m["member"]; ok {
		if members := v.([]interface{}); len(members) > 0 {
			for _, member := range members {
				buf.WriteString(fmt.Sprintf("%d-", memberHash(member)))
			}
		}
	}
	return hashcode.String(buf.String())
}

func memberHash(v interface{}) int {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", v.(map[string]interface{})["origin_id"]))
	return hashcode.String(buf.String())
}

func failoverCriteriaHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["status_codes"]; ok {
		for _, w := range v.(*schema.Set).List() {
			buf.WriteString(fmt.Sprintf("%d-", w))
		}
	}
	return hashcode.String(buf.String())
}

func expandCustomHeaders(s *schema.Set) *cloudfront.CustomHeaders {
	qty := 0
	items := []*cloudfront.OriginCustomHeader{}
	for _, v := range s.List() {
		items = append(items, expandOriginCustomHeader(v.(map[string]interface{})))
		qty++
	}
	return &cloudfront.CustomHeaders{
		Quantity: aws.Int64(int64(qty)),
		Items:    items,
	}
}

func flattenCustomHeaders(chs *cloudfront.CustomHeaders) *schema.Set {
	s := []interface{}{}
	for _, v := range chs.Items {
		s = append(s, flattenOriginCustomHeader(v))
	}
	return schema.NewSet(originCustomHeaderHash, s)
}

func expandOriginCustomHeader(m map[string]interface{}) *cloudfront.OriginCustomHeader {
	return &cloudfront.OriginCustomHeader{
		HeaderName:  aws.String(m["name"].(string)),
		HeaderValue: aws.String(m["value"].(string)),
	}
}

func flattenOriginCustomHeader(och *cloudfront.OriginCustomHeader) map[string]interface{} {
	return map[string]interface{}{
		"name":  aws.StringValue(och.HeaderName),
		"value": aws.StringValue(och.HeaderValue),
	}
}

// Helper function used by originHash to get a composite hash for all
// aws_cloudfront_distribution custom_header attributes.
func customHeadersHash(s *schema.Set) int {
	var buf bytes.Buffer
	for _, v := range s.List() {
		buf.WriteString(fmt.Sprintf("%d-", originCustomHeaderHash(v)))
	}
	return hashcode.String(buf.String())
}

// Assemble the hash for the aws_cloudfront_distribution custom_header
// TypeSet attribute.
func originCustomHeaderHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["value"].(string)))
	return hashcode.String(buf.String())
}

func expandCustomOriginConfig(m map[string]interface{}) *cloudfront.CustomOriginConfig {

	customOrigin := &cloudfront.CustomOriginConfig{
		OriginProtocolPolicy:   aws.String(m["origin_protocol_policy"].(string)),
		HTTPPort:               aws.Int64(int64(m["http_port"].(int))),
		HTTPSPort:              aws.Int64(int64(m["https_port"].(int))),
		OriginSslProtocols:     expandCustomOriginConfigSSL(m["origin_ssl_protocols"].(*schema.Set).List()),
		OriginReadTimeout:      aws.Int64(int64(m["origin_read_timeout"].(int))),
		OriginKeepaliveTimeout: aws.Int64(int64(m["origin_keepalive_timeout"].(int))),
	}

	return customOrigin
}

func flattenCustomOriginConfig(cor *cloudfront.CustomOriginConfig) map[string]interface{} {

	customOrigin := map[string]interface{}{
		"origin_protocol_policy":   aws.StringValue(cor.OriginProtocolPolicy),
		"http_port":                int(aws.Int64Value(cor.HTTPPort)),
		"https_port":               int(aws.Int64Value(cor.HTTPSPort)),
		"origin_ssl_protocols":     flattenCustomOriginConfigSSL(cor.OriginSslProtocols),
		"origin_read_timeout":      int(aws.Int64Value(cor.OriginReadTimeout)),
		"origin_keepalive_timeout": int(aws.Int64Value(cor.OriginKeepaliveTimeout)),
	}

	return customOrigin
}

// Assemble the hash for the aws_cloudfront_distribution custom_origin_config
// TypeSet attribute.
func customOriginConfigHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["origin_protocol_policy"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["http_port"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["https_port"].(int)))
	for _, v := range sortInterfaceSlice(m["origin_ssl_protocols"].(*schema.Set).List()) {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	buf.WriteString(fmt.Sprintf("%d-", m["origin_keepalive_timeout"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["origin_read_timeout"].(int)))

	return hashcode.String(buf.String())
}

func expandCustomOriginConfigSSL(s []interface{}) *cloudfront.OriginSslProtocols {
	items := expandStringList(s)
	return &cloudfront.OriginSslProtocols{
		Quantity: aws.Int64(int64(len(items))),
		Items:    items,
	}
}

func flattenCustomOriginConfigSSL(osp *cloudfront.OriginSslProtocols) *schema.Set {
	return flattenStringSet(osp.Items)
}

func expandS3OriginConfig(m map[string]interface{}) *cloudfront.S3OriginConfig {
	return &cloudfront.S3OriginConfig{
		OriginAccessIdentity: aws.String(m["origin_access_identity"].(string)),
	}
}

func flattenS3OriginConfig(s3o *cloudfront.S3OriginConfig) map[string]interface{} {
	return map[string]interface{}{
		"origin_access_identity": aws.StringValue(s3o.OriginAccessIdentity),
	}
}

// Assemble the hash for the aws_cloudfront_distribution s3_origin_config
// TypeSet attribute.
func s3OriginConfigHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["origin_access_identity"].(string)))
	return hashcode.String(buf.String())
}

func expandCustomErrorResponses(s *schema.Set) *cloudfront.CustomErrorResponses {
	qty := 0
	items := []*cloudfront.CustomErrorResponse{}
	for _, v := range s.List() {
		items = append(items, expandCustomErrorResponse(v.(map[string]interface{})))
		qty++
	}
	return &cloudfront.CustomErrorResponses{
		Quantity: aws.Int64(int64(qty)),
		Items:    items,
	}
}

func flattenCustomErrorResponses(ers *cloudfront.CustomErrorResponses) *schema.Set {
	s := []interface{}{}
	for _, v := range ers.Items {
		s = append(s, flattenCustomErrorResponse(v))
	}
	return schema.NewSet(customErrorResponseHash, s)
}

func expandCustomErrorResponse(m map[string]interface{}) *cloudfront.CustomErrorResponse {
	er := cloudfront.CustomErrorResponse{
		ErrorCode: aws.Int64(int64(m["error_code"].(int))),
	}
	if v, ok := m["error_caching_min_ttl"]; ok {
		er.ErrorCachingMinTTL = aws.Int64(int64(v.(int)))
	}
	if v, ok := m["response_code"]; ok && v.(int) != 0 {
		er.ResponseCode = aws.String(strconv.Itoa(v.(int)))
	} else {
		er.ResponseCode = aws.String("")
	}
	if v, ok := m["response_page_path"]; ok {
		er.ResponsePagePath = aws.String(v.(string))
	}

	return &er
}

func flattenCustomErrorResponse(er *cloudfront.CustomErrorResponse) map[string]interface{} {
	m := make(map[string]interface{})
	m["error_code"] = int(aws.Int64Value(er.ErrorCode))
	if er.ErrorCachingMinTTL != nil {
		m["error_caching_min_ttl"] = int(*er.ErrorCachingMinTTL)
	}
	if er.ResponseCode != nil {
		m["response_code"], _ = strconv.Atoi(*er.ResponseCode)
	}
	if er.ResponsePagePath != nil {
		m["response_page_path"] = *er.ResponsePagePath
	}
	return m
}

// Assemble the hash for the aws_cloudfront_distribution custom_error_response
// TypeSet attribute.
func customErrorResponseHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m["error_code"].(int)))
	if v, ok := m["error_caching_min_ttl"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["response_code"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["response_page_path"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	return hashcode.String(buf.String())
}

func expandLoggingConfig(m map[string]interface{}) *cloudfront.LoggingConfig {
	var lc cloudfront.LoggingConfig
	if m != nil {
		lc.Prefix = aws.String(m["prefix"].(string))
		lc.Bucket = aws.String(m["bucket"].(string))
		lc.IncludeCookies = aws.Bool(m["include_cookies"].(bool))
		lc.Enabled = aws.Bool(true)
	} else {
		lc.Prefix = aws.String("")
		lc.Bucket = aws.String("")
		lc.IncludeCookies = aws.Bool(false)
		lc.Enabled = aws.Bool(false)
	}
	return &lc
}

func flattenLoggingConfig(lc *cloudfront.LoggingConfig) []interface{} {
	m := map[string]interface{}{
		"bucket":          aws.StringValue(lc.Bucket),
		"include_cookies": aws.BoolValue(lc.IncludeCookies),
		"prefix":          aws.StringValue(lc.Prefix),
	}

	return []interface{}{m}
}

func expandAliases(s *schema.Set) *cloudfront.Aliases {
	aliases := cloudfront.Aliases{
		Quantity: aws.Int64(int64(s.Len())),
	}
	if s.Len() > 0 {
		aliases.Items = expandStringSet(s)
	}
	return &aliases
}

func flattenAliases(aliases *cloudfront.Aliases) *schema.Set {
	if aliases.Items != nil {
		return flattenStringSet(aliases.Items)
	}
	return schema.NewSet(aliasesHash, []interface{}{})
}

// Assemble the hash for the aws_cloudfront_distribution aliases
// TypeSet attribute.
func aliasesHash(v interface{}) int {
	return hashcode.String(v.(string))
}

func expandRestrictions(m map[string]interface{}) *cloudfront.Restrictions {
	return &cloudfront.Restrictions{
		GeoRestriction: expandGeoRestriction(m["geo_restriction"].([]interface{})[0].(map[string]interface{})),
	}
}

func flattenRestrictions(r *cloudfront.Restrictions) []interface{} {
	m := map[string]interface{}{
		"geo_restriction": []interface{}{flattenGeoRestriction(r.GeoRestriction)},
	}

	return []interface{}{m}
}

func expandGeoRestriction(m map[string]interface{}) *cloudfront.GeoRestriction {
	gr := &cloudfront.GeoRestriction{
		Quantity:        aws.Int64(int64(0)),
		RestrictionType: aws.String(m["restriction_type"].(string)),
	}

	if v, ok := m["locations"]; ok {
		gr.Items = expandStringSet(v.(*schema.Set))
		gr.Quantity = aws.Int64(int64(v.(*schema.Set).Len()))
	}

	return gr
}

func flattenGeoRestriction(gr *cloudfront.GeoRestriction) map[string]interface{} {
	m := make(map[string]interface{})

	m["restriction_type"] = aws.StringValue(gr.RestrictionType)
	if gr.Items != nil {
		m["locations"] = flattenStringSet(gr.Items)
	}
	return m
}

func expandViewerCertificate(m map[string]interface{}) *cloudfront.ViewerCertificate {
	var vc cloudfront.ViewerCertificate
	if v, ok := m["iam_certificate_id"]; ok && v != "" {
		vc.IAMCertificateId = aws.String(v.(string))
		vc.SSLSupportMethod = aws.String(m["ssl_support_method"].(string))
	} else if v, ok := m["acm_certificate_arn"]; ok && v != "" {
		vc.ACMCertificateArn = aws.String(v.(string))
		vc.SSLSupportMethod = aws.String(m["ssl_support_method"].(string))
	} else {
		vc.CloudFrontDefaultCertificate = aws.Bool(m["cloudfront_default_certificate"].(bool))
	}
	if v, ok := m["minimum_protocol_version"]; ok && v != "" {
		vc.MinimumProtocolVersion = aws.String(v.(string))
	}
	return &vc
}

func flattenViewerCertificate(vc *cloudfront.ViewerCertificate) []interface{} {
	m := make(map[string]interface{})

	if vc.IAMCertificateId != nil {
		m["iam_certificate_id"] = *vc.IAMCertificateId
		m["ssl_support_method"] = *vc.SSLSupportMethod
	}
	if vc.ACMCertificateArn != nil {
		m["acm_certificate_arn"] = *vc.ACMCertificateArn
		m["ssl_support_method"] = *vc.SSLSupportMethod
	}
	if vc.CloudFrontDefaultCertificate != nil {
		m["cloudfront_default_certificate"] = *vc.CloudFrontDefaultCertificate
	}
	if vc.MinimumProtocolVersion != nil {
		m["minimum_protocol_version"] = *vc.MinimumProtocolVersion
	}
	return []interface{}{m}
}

func flattenCloudfrontActiveTrustedSigners(ats *cloudfront.ActiveTrustedSigners) []interface{} {
	if ats == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"enabled": aws.BoolValue(ats.Enabled),
		"items":   flattenCloudfrontSigners(ats.Items),
	}

	return []interface{}{m}
}

func flattenCloudfrontSigners(signers []*cloudfront.Signer) []interface{} {
	result := make([]interface{}, 0, len(signers))

	for _, signer := range signers {
		m := map[string]interface{}{
			"aws_account_number": aws.StringValue(signer.AwsAccountNumber),
			"key_pair_ids":       aws.StringValueSlice(signer.KeyPairIds.Items),
		}

		result = append(result, m)
	}

	return result
}
