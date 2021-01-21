package aws

import (
	"context"
	"reflect"
	"testing"
)

func testResourceAwsKinesisStreamStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"arn":                 "arn:aws:test:us-east-1:123456789012:test", //lintignore:AWSAT003,AWSAT005
		"encryption_type":     "NONE",
		"kms_key_id":          "",
		"name":                "test",
		"retention_period":    24,
		"shard_count":         1,
		"shard_level_metrics": []interface{}{},
		"tags":                map[string]interface{}{"key1": "value1"},
	}
}

func testResourceAwsKinesisStreamStateDataV1() map[string]interface{} {
	v0 := testResourceAwsKinesisStreamStateDataV0()
	return map[string]interface{}{
		"arn":                       v0["arn"],
		"encryption_type":           v0["encryption_type"],
		"enforce_consumer_deletion": false,
		"kms_key_id":                v0["kms_key_id"],
		"name":                      v0["name"],
		"retention_period":          v0["retention_period"],
		"shard_count":               v0["shard_count"],
		"shard_level_metrics":       v0["shard_level_metrics"],
		"tags":                      v0["tags"],
	}
}

func TestResourceAwsKinesisStreamStateUpgradeV0(t *testing.T) {
	expected := testResourceAwsKinesisStreamStateDataV1()
	actual, err := resourceAwsKinesisStreamStateUpgradeV0(context.Background(), testResourceAwsKinesisStreamStateDataV0(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
