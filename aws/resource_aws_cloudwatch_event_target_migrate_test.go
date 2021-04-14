package aws

import (
	"context"
	"reflect"
	"testing"
)

func testResourceAwsCloudWatchEventTargetStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"arn":       "arn:aws:test:us-east-1:123456789012:test", //lintignore:AWSAT003,AWSAT005
		"rule":      "testrule",
		"target_id": "testtargetid",
	}
}

func testResourceAwsCloudWatchEventTargetStateDataV0EventBusName() map[string]interface{} {
	return map[string]interface{}{
		"arn":            "arn:aws:test:us-east-1:123456789012:test", //lintignore:AWSAT003,AWSAT005
		"event_bus_name": "testbus",
		"rule":           "testrule",
		"target_id":      "testtargetid",
	}
}

func testResourceAwsCloudWatchEventTargetStateDataV1() map[string]interface{} {
	v0 := testResourceAwsCloudWatchEventTargetStateDataV0()
	return map[string]interface{}{
		"arn":            v0["arn"],
		"event_bus_name": "default",
		"rule":           v0["rule"],
		"target_id":      v0["target_id"],
	}
}

func testResourceAwsCloudWatchEventTargetStateDataV1EventBusName() map[string]interface{} {
	v0 := testResourceAwsCloudWatchEventTargetStateDataV0EventBusName()
	return map[string]interface{}{
		"arn":            v0["arn"],
		"event_bus_name": v0["event_bus_name"],
		"rule":           v0["rule"],
		"target_id":      v0["target_id"],
	}
}

func TestResourceAwsCloudWatchEventTargetStateUpgradeV0(t *testing.T) {
	expected := testResourceAwsCloudWatchEventTargetStateDataV1()
	actual, err := resourceAwsCloudWatchEventTargetStateUpgradeV0(context.Background(), testResourceAwsCloudWatchEventTargetStateDataV0(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}

func TestResourceAwsCloudWatchEventTargetStateUpgradeV0EventBusName(t *testing.T) {
	expected := testResourceAwsCloudWatchEventTargetStateDataV1EventBusName()
	actual, err := resourceAwsCloudWatchEventTargetStateUpgradeV0(context.Background(), testResourceAwsCloudWatchEventTargetStateDataV0EventBusName(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
