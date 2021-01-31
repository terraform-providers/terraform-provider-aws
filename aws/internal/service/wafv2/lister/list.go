//go:generate go run ../../../generators/listpages/main.go -function=ListIPSets,ListRegexPatternSets,ListRuleGroups,ListWebACLs -paginator=NextMarker github.com/aws/aws-sdk-go/service/wafv2

package lister
