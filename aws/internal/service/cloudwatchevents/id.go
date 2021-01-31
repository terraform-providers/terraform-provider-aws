package cloudwatchevents

import (
	"fmt"
	"strings"
)

const DefaultEventBusName = "default"

const PermissionIDSeparator = "/"

func PermissionCreateID(eventBusName, statementID string) string {
	if eventBusName == "" || eventBusName == DefaultEventBusName {
		return statementID
	}
	return eventBusName + PermissionIDSeparator + statementID
}

func PermissionParseID(id string) (string, string, error) {
	parts := strings.Split(id, PermissionIDSeparator)
	if len(parts) == 1 && parts[0] != "" {
		return DefaultEventBusName, parts[0], nil
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%q), expected <event-bus-name>"+PermissionIDSeparator+"<statement-id> or <statement-id>", id)
}

const ruleIDSeparator = "/"

func RuleCreateID(eventBusName, ruleName string) string {
	if eventBusName == "" || eventBusName == DefaultEventBusName {
		return ruleName
	}
	return eventBusName + ruleIDSeparator + ruleName
}

func RuleParseID(id string) (string, string, error) {
	parts := strings.Split(id, ruleIDSeparator)
	if len(parts) == 1 && parts[0] != "" {
		return DefaultEventBusName, parts[0], nil
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%q), expected <event-bus-name>"+ruleIDSeparator+"<rule-name> or <rule-name>", id)
}

// Terraform state IDs for Targets are not parseable, since the separator used ("-") is also a valid
// character in both the rule name and the target id.

const targetIDSeparator = "-"
const targetImportIDSeparator = "/"

func TargetCreateID(eventBusName, ruleName, targetID string) string {
	id := ruleName + targetIDSeparator + targetID
	if eventBusName != "" && eventBusName != DefaultEventBusName {
		id = eventBusName + targetIDSeparator + id
	}
	return id
}

func TargetParseImportID(id string) (string, string, string, error) {
	parts := strings.Split(id, targetImportIDSeparator)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return DefaultEventBusName, parts[0], parts[1], nil
	}
	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%q), expected <event-bus-name>"+targetImportIDSeparator+"<rule-name>"+targetImportIDSeparator+"<target-id> or <rule-name>"+targetImportIDSeparator+"<target-id>", id)
}
