package aws

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func configDescribeOrganizationConfigRule(conn *configservice.ConfigService, name string) (*configservice.OrganizationConfigRule, error) {
	input := &configservice.DescribeOrganizationConfigRulesInput{
		OrganizationConfigRuleNames: []*string{aws.String(name)},
	}

	for {
		output, err := conn.DescribeOrganizationConfigRules(input)

		if err != nil {
			return nil, err
		}

		for _, rule := range output.OrganizationConfigRules {
			if aws.StringValue(rule.OrganizationConfigRuleName) == name {
				return rule, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func configDescribeOrganizationConfigRuleStatus(conn *configservice.ConfigService, name string) (*configservice.OrganizationConfigRuleStatus, error) {
	input := &configservice.DescribeOrganizationConfigRuleStatusesInput{
		OrganizationConfigRuleNames: []*string{aws.String(name)},
	}

	for {
		output, err := conn.DescribeOrganizationConfigRuleStatuses(input)

		if err != nil {
			return nil, err
		}

		for _, status := range output.OrganizationConfigRuleStatuses {
			if aws.StringValue(status.OrganizationConfigRuleName) == name {
				return status, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func configGetOrganizationConfigRuleDetailedStatus(conn *configservice.ConfigService, ruleName, ruleStatus string) ([]*configservice.MemberAccountStatus, error) {
	input := &configservice.GetOrganizationConfigRuleDetailedStatusInput{
		Filters: &configservice.StatusDetailFilters{
			MemberAccountRuleStatus: aws.String(ruleStatus),
		},
		OrganizationConfigRuleName: aws.String(ruleName),
	}
	var statuses []*configservice.MemberAccountStatus

	for {
		output, err := conn.GetOrganizationConfigRuleDetailedStatus(input)

		if err != nil {
			return nil, err
		}

		statuses = append(statuses, output.OrganizationConfigRuleDetailedStatus...)

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return statuses, nil
}

func configRefreshOrganizationConfigRuleStatus(conn *configservice.ConfigService, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		status, err := configDescribeOrganizationConfigRuleStatus(conn, name)

		if err != nil {
			return nil, "", err
		}

		if status == nil {
			return nil, "", fmt.Errorf("status not found")
		}

		if status.ErrorCode != nil {
			return status, aws.StringValue(status.OrganizationRuleStatus), fmt.Errorf("%s: %s", aws.StringValue(status.ErrorCode), aws.StringValue(status.ErrorMessage))
		}

		switch aws.StringValue(status.OrganizationRuleStatus) {
		case configservice.OrganizationRuleStatusCreateFailed, configservice.OrganizationRuleStatusDeleteFailed, configservice.OrganizationRuleStatusUpdateFailed:
			// Display detailed errors for failed member accounts
			memberAccountStatuses, err := configGetOrganizationConfigRuleDetailedStatus(conn, name, aws.StringValue(status.OrganizationRuleStatus))

			if err != nil {
				return status, aws.StringValue(status.OrganizationRuleStatus), fmt.Errorf("unable to get Organization Config Rule detailed status for showing member account errors: %s", err)
			}

			var errBuilder strings.Builder

			for _, mas := range memberAccountStatuses {
				errBuilder.WriteString(fmt.Sprintf("Account ID (%s): %s: %s\n", aws.StringValue(mas.AccountId), aws.StringValue(mas.ErrorCode), aws.StringValue(mas.ErrorMessage)))
			}

			return status, aws.StringValue(status.OrganizationRuleStatus), fmt.Errorf("Failed in %d account(s):\n\n%s", len(memberAccountStatuses), errBuilder.String())
		}

		return status, aws.StringValue(status.OrganizationRuleStatus), nil
	}
}

func configWaitForOrganizationRuleStatusCreateSuccessful(conn *configservice.ConfigService, name string, timeout time.Duration) error {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{configservice.OrganizationRuleStatusCreateInProgress},
		Target:  []string{configservice.OrganizationRuleStatusCreateSuccessful},
		Refresh: configRefreshOrganizationConfigRuleStatus(conn, name),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	_, err := stateChangeConf.WaitForState()

	return err
}

func configWaitForOrganizationRuleStatusDeleteSuccessful(conn *configservice.ConfigService, name string, timeout time.Duration) error {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{configservice.OrganizationRuleStatusDeleteInProgress},
		Target:  []string{configservice.OrganizationRuleStatusDeleteSuccessful},
		Refresh: configRefreshOrganizationConfigRuleStatus(conn, name),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	_, err := stateChangeConf.WaitForState()

	if isAWSErr(err, configservice.ErrCodeNoSuchOrganizationConfigRuleException, "") {
		return nil
	}

	return err
}

func configWaitForOrganizationRuleStatusUpdateSuccessful(conn *configservice.ConfigService, name string, timeout time.Duration) error {
	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{configservice.OrganizationRuleStatusUpdateInProgress},
		Target:  []string{configservice.OrganizationRuleStatusUpdateSuccessful},
		Refresh: configRefreshOrganizationConfigRuleStatus(conn, name),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	_, err := stateChangeConf.WaitForState()

	return err
}
