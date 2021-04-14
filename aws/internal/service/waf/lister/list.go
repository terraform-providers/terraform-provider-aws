//go:generate go run ../../../generators/listpages/main.go -function=ListByteMatchSets,ListGeoMatchSets,ListIPSets,ListRateBasedRules,ListRegexMatchSets,ListRegexPatternSets,ListRuleGroups,ListRules,ListSizeConstraintSets,ListSqlInjectionMatchSets,ListWebACLs,ListXssMatchSets -paginator=NextMarker github.com/aws/aws-sdk-go/service/waf

package lister
