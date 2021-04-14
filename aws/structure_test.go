package aws

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDiffStringMaps(t *testing.T) {
	cases := []struct {
		Old, New                  map[string]interface{}
		Create, Remove, Unchanged map[string]interface{}
	}{
		// Add
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
			Create: map[string]interface{}{
				"bar": "baz",
			},
			Remove: map[string]interface{}{},
			Unchanged: map[string]interface{}{
				"foo": "bar",
			},
		},

		// Modify
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"foo": "baz",
			},
			Create: map[string]interface{}{
				"foo": "baz",
			},
			Remove: map[string]interface{}{
				"foo": "bar",
			},
			Unchanged: map[string]interface{}{},
		},

		// Overlap
		{
			Old: map[string]interface{}{
				"foo":   "bar",
				"hello": "world",
			},
			New: map[string]interface{}{
				"foo":   "baz",
				"hello": "world",
			},
			Create: map[string]interface{}{
				"foo": "baz",
			},
			Remove: map[string]interface{}{
				"foo": "bar",
			},
			Unchanged: map[string]interface{}{
				"hello": "world",
			},
		},

		// Remove
		{
			Old: map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
			New: map[string]interface{}{
				"foo": "bar",
			},
			Create: map[string]interface{}{},
			Remove: map[string]interface{}{
				"bar": "baz",
			},
			Unchanged: map[string]interface{}{
				"foo": "bar",
			},
		},
	}

	for i, tc := range cases {
		c, r, u := diffStringMaps(tc.Old, tc.New)
		cm := pointersMapToStringList(c)
		rm := pointersMapToStringList(r)
		um := pointersMapToStringList(u)
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		if !reflect.DeepEqual(rm, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, rm)
		}
		if !reflect.DeepEqual(um, tc.Unchanged) {
			t.Fatalf("%d: bad unchanged: %#v", i, rm)
		}
	}
}

func TestExpandIPPerms(t *testing.T) {
	hash := schema.HashString

	expanded := []interface{}{
		map[string]interface{}{
			"protocol":    "icmp",
			"from_port":   1,
			"to_port":     -1,
			"cidr_blocks": []interface{}{"0.0.0.0/0"},
			"security_groups": schema.NewSet(hash, []interface{}{
				"sg-11111",
				"foo/sg-22222",
			}),
			"description": "desc",
		},
		map[string]interface{}{
			"protocol":  "icmp",
			"from_port": 1,
			"to_port":   -1,
			"self":      true,
		},
	}
	group := &ec2.SecurityGroup{
		GroupId: aws.String("foo"),
		VpcId:   aws.String("bar"),
	}
	perms, err := expandIPPerms(group, expanded)
	if err != nil {
		t.Fatalf("error expanding perms: %v", err)
	}

	expected := []ec2.IpPermission{
		{
			IpProtocol: aws.String("icmp"),
			FromPort:   aws.Int64(int64(1)),
			ToPort:     aws.Int64(int64(-1)),
			IpRanges: []*ec2.IpRange{
				{
					CidrIp:      aws.String("0.0.0.0/0"),
					Description: aws.String("desc"),
				},
			},
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				{
					UserId:      aws.String("foo"),
					GroupId:     aws.String("sg-22222"),
					Description: aws.String("desc"),
				},
				{
					GroupId:     aws.String("sg-11111"),
					Description: aws.String("desc"),
				},
			},
		},
		{
			IpProtocol: aws.String("icmp"),
			FromPort:   aws.Int64(int64(1)),
			ToPort:     aws.Int64(int64(-1)),
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				{
					GroupId: aws.String("foo"),
				},
			},
		},
	}

	exp := expected[0]
	perm := perms[0]

	if aws.Int64Value(exp.FromPort) != aws.Int64Value(perm.FromPort) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.Int64Value(perm.FromPort),
			aws.Int64Value(exp.FromPort))
	}

	if aws.StringValue(exp.IpRanges[0].CidrIp) != aws.StringValue(perm.IpRanges[0].CidrIp) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.IpRanges[0].CidrIp),
			aws.StringValue(exp.IpRanges[0].CidrIp))
	}

	if aws.StringValue(exp.UserIdGroupPairs[0].UserId) != aws.StringValue(perm.UserIdGroupPairs[0].UserId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[0].UserId),
			aws.StringValue(exp.UserIdGroupPairs[0].UserId))
	}

	if aws.StringValue(exp.UserIdGroupPairs[0].GroupId) != aws.StringValue(perm.UserIdGroupPairs[0].GroupId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[0].GroupId),
			aws.StringValue(exp.UserIdGroupPairs[0].GroupId))
	}

	if aws.StringValue(exp.UserIdGroupPairs[1].GroupId) != aws.StringValue(perm.UserIdGroupPairs[1].GroupId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[1].GroupId),
			aws.StringValue(exp.UserIdGroupPairs[1].GroupId))
	}

	exp = expected[1]
	perm = perms[1]

	if aws.StringValue(exp.UserIdGroupPairs[0].GroupId) != aws.StringValue(perm.UserIdGroupPairs[0].GroupId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[0].GroupId),
			aws.StringValue(exp.UserIdGroupPairs[0].GroupId))
	}
}

func TestExpandIPPerms_NegOneProtocol(t *testing.T) {
	hash := schema.HashString

	expanded := []interface{}{
		map[string]interface{}{
			"protocol":    "-1",
			"from_port":   0,
			"to_port":     0,
			"cidr_blocks": []interface{}{"0.0.0.0/0"},
			"security_groups": schema.NewSet(hash, []interface{}{
				"sg-11111",
				"foo/sg-22222",
			}),
		},
	}
	group := &ec2.SecurityGroup{
		GroupId: aws.String("foo"),
		VpcId:   aws.String("bar"),
	}

	perms, err := expandIPPerms(group, expanded)
	if err != nil {
		t.Fatalf("error expanding perms: %v", err)
	}

	expected := []ec2.IpPermission{
		{
			IpProtocol: aws.String("-1"),
			FromPort:   aws.Int64(int64(0)),
			ToPort:     aws.Int64(int64(0)),
			IpRanges:   []*ec2.IpRange{{CidrIp: aws.String("0.0.0.0/0")}},
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				{
					UserId:  aws.String("foo"),
					GroupId: aws.String("sg-22222"),
				},
				{
					GroupId: aws.String("sg-11111"),
				},
			},
		},
	}

	exp := expected[0]
	perm := perms[0]

	if aws.Int64Value(exp.FromPort) != aws.Int64Value(perm.FromPort) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.Int64Value(perm.FromPort),
			aws.Int64Value(exp.FromPort))
	}

	if aws.StringValue(exp.IpRanges[0].CidrIp) != aws.StringValue(perm.IpRanges[0].CidrIp) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.IpRanges[0].CidrIp),
			aws.StringValue(exp.IpRanges[0].CidrIp))
	}

	if aws.StringValue(exp.UserIdGroupPairs[0].UserId) != aws.StringValue(perm.UserIdGroupPairs[0].UserId) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[0].UserId),
			aws.StringValue(exp.UserIdGroupPairs[0].UserId))
	}

	// Now test the error case. This *should* error when either from_port
	// or to_port is not zero, but protocol is "-1".
	errorCase := []interface{}{
		map[string]interface{}{
			"protocol":    "-1",
			"from_port":   0,
			"to_port":     65535,
			"cidr_blocks": []interface{}{"0.0.0.0/0"},
			"security_groups": schema.NewSet(hash, []interface{}{
				"sg-11111",
				"foo/sg-22222",
			}),
		},
	}
	securityGroups := &ec2.SecurityGroup{
		GroupId: aws.String("foo"),
		VpcId:   aws.String("bar"),
	}

	_, expandErr := expandIPPerms(securityGroups, errorCase)
	if expandErr == nil {
		t.Fatal("expandIPPerms should have errored!")
	}
}

func TestExpandIPPerms_nonVPC(t *testing.T) {
	hash := schema.HashString

	expanded := []interface{}{
		map[string]interface{}{
			"protocol":    "icmp",
			"from_port":   1,
			"to_port":     -1,
			"cidr_blocks": []interface{}{"0.0.0.0/0"},
			"security_groups": schema.NewSet(hash, []interface{}{
				"sg-11111",
				"foo/sg-22222",
			}),
		},
		map[string]interface{}{
			"protocol":  "icmp",
			"from_port": 1,
			"to_port":   -1,
			"self":      true,
		},
	}
	group := &ec2.SecurityGroup{
		GroupName: aws.String("foo"),
	}
	perms, err := expandIPPerms(group, expanded)
	if err != nil {
		t.Fatalf("error expanding perms: %v", err)
	}

	expected := []ec2.IpPermission{
		{
			IpProtocol: aws.String("icmp"),
			FromPort:   aws.Int64(int64(1)),
			ToPort:     aws.Int64(int64(-1)),
			IpRanges:   []*ec2.IpRange{{CidrIp: aws.String("0.0.0.0/0")}},
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				{
					GroupName: aws.String("sg-22222"),
				},
				{
					GroupName: aws.String("sg-11111"),
				},
			},
		},
		{
			IpProtocol: aws.String("icmp"),
			FromPort:   aws.Int64(int64(1)),
			ToPort:     aws.Int64(int64(-1)),
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				{
					GroupName: aws.String("foo"),
				},
			},
		},
	}

	exp := expected[0]
	perm := perms[0]

	if aws.Int64Value(exp.FromPort) != aws.Int64Value(perm.FromPort) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.Int64Value(perm.FromPort),
			aws.Int64Value(exp.FromPort))
	}

	if aws.StringValue(exp.IpRanges[0].CidrIp) != aws.StringValue(perm.IpRanges[0].CidrIp) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.IpRanges[0].CidrIp),
			aws.StringValue(exp.IpRanges[0].CidrIp))
	}

	if aws.StringValue(exp.UserIdGroupPairs[0].GroupName) != aws.StringValue(perm.UserIdGroupPairs[0].GroupName) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[0].GroupName),
			aws.StringValue(exp.UserIdGroupPairs[0].GroupName))
	}

	if aws.StringValue(exp.UserIdGroupPairs[1].GroupName) != aws.StringValue(perm.UserIdGroupPairs[1].GroupName) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[1].GroupName),
			aws.StringValue(exp.UserIdGroupPairs[1].GroupName))
	}

	exp = expected[1]
	perm = perms[1]

	if aws.StringValue(exp.UserIdGroupPairs[0].GroupName) != aws.StringValue(perm.UserIdGroupPairs[0].GroupName) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			aws.StringValue(perm.UserIdGroupPairs[0].GroupName),
			aws.StringValue(exp.UserIdGroupPairs[0].GroupName))
	}
}

func TestExpandListeners(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"instance_port":     8000,
			"lb_port":           80,
			"instance_protocol": "http",
			"lb_protocol":       "http",
		},
		map[string]interface{}{
			"instance_port":      8000,
			"lb_port":            80,
			"instance_protocol":  "https",
			"lb_protocol":        "https",
			"ssl_certificate_id": "something",
		},
	}
	listeners, err := expandListeners(expanded)
	if err != nil {
		t.Fatalf("bad: %#v", err)
	}

	expected := &elb.Listener{
		InstancePort:     aws.Int64(int64(8000)),
		LoadBalancerPort: aws.Int64(int64(80)),
		InstanceProtocol: aws.String("http"),
		Protocol:         aws.String("http"),
	}

	if !reflect.DeepEqual(listeners[0], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			listeners[0],
			expected)
	}
}

// this test should produce an error from expandlisteners on an invalid
// combination
func TestExpandListeners_invalid(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"instance_port":      8000,
			"lb_port":            80,
			"instance_protocol":  "http",
			"lb_protocol":        "http",
			"ssl_certificate_id": "something",
		},
	}
	_, err := expandListeners(expanded)
	if err != nil {
		// Check the error we got
		if !strings.Contains(err.Error(), "ssl_certificate_id may be set only when protocol") {
			t.Fatalf("Got error in TestExpandListeners_invalid, but not what we expected: %s", err)
		}
	}

	if err == nil {
		t.Fatalf("Expected TestExpandListeners_invalid to fail, but passed")
	}
}

func TestFlattenHealthCheck(t *testing.T) {
	cases := []struct {
		Input  *elb.HealthCheck
		Output []map[string]interface{}
	}{
		{
			Input: &elb.HealthCheck{
				UnhealthyThreshold: aws.Int64(int64(10)),
				HealthyThreshold:   aws.Int64(int64(10)),
				Target:             aws.String("HTTP:80/"),
				Timeout:            aws.Int64(int64(30)),
				Interval:           aws.Int64(int64(30)),
			},
			Output: []map[string]interface{}{
				{
					"unhealthy_threshold": int64(10),
					"healthy_threshold":   int64(10),
					"target":              "HTTP:80/",
					"timeout":             int64(30),
					"interval":            int64(30),
				},
			},
		},
	}

	for _, tc := range cases {
		output := flattenHealthCheck(tc.Input)
		if !reflect.DeepEqual(output, tc.Output) {
			t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v", output, tc.Output)
		}
	}
}

func TestFlattenOrganizationsOrganizationalUnits(t *testing.T) {
	input := []*organizations.OrganizationalUnit{
		{
			Arn:  aws.String("arn:aws:organizations::123456789012:ou/o-abcde12345/ou-ab12-abcd1234"), //lintignore:AWSAT005
			Id:   aws.String("ou-ab12-abcd1234"),
			Name: aws.String("Engineering"),
		},
	}

	expected_output := []map[string]interface{}{
		{
			"arn":  "arn:aws:organizations::123456789012:ou/o-abcde12345/ou-ab12-abcd1234", //lintignore:AWSAT005
			"id":   "ou-ab12-abcd1234",
			"name": "Engineering",
		},
	}

	output := flattenOrganizationsOrganizationalUnits(input)
	if !reflect.DeepEqual(expected_output, output) {
		t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v", output, expected_output)
	}
}

func TestExpandStringList(t *testing.T) {
	expanded := []interface{}{"us-east-1a", "us-east-1b"} //lintignore:AWSAT003
	stringList := expandStringList(expanded)
	expected := []*string{
		aws.String("us-east-1a"), //lintignore:AWSAT003
		aws.String("us-east-1b"), //lintignore:AWSAT003
	}

	if !reflect.DeepEqual(stringList, expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			stringList,
			expected)
	}
}

func TestExpandStringListEmptyItems(t *testing.T) {
	expanded := []interface{}{"foo", "bar", "", "baz"}
	stringList := expandStringList(expanded)
	expected := []*string{
		aws.String("foo"),
		aws.String("bar"),
		aws.String("baz"),
	}

	if !reflect.DeepEqual(stringList, expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			stringList,
			expected)
	}
}

func TestExpandParameters(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"name":         "character_set_client",
			"value":        "utf8",
			"apply_method": "immediate",
		},
	}
	parameters := expandParameters(expanded)

	expected := &rds.Parameter{
		ParameterName:  aws.String("character_set_client"),
		ParameterValue: aws.String("utf8"),
		ApplyMethod:    aws.String("immediate"),
	}

	if !reflect.DeepEqual(parameters[0], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			parameters[0],
			expected)
	}
}

func TestExpandRedshiftParameters(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"name":  "character_set_client",
			"value": "utf8",
		},
	}
	parameters := expandRedshiftParameters(expanded)

	expected := &redshift.Parameter{
		ParameterName:  aws.String("character_set_client"),
		ParameterValue: aws.String("utf8"),
	}

	if !reflect.DeepEqual(parameters[0], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			parameters[0],
			expected)
	}
}

func TestExpandStepAdjustments(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"metric_interval_lower_bound": "1.0",
			"metric_interval_upper_bound": "2.0",
			"scaling_adjustment":          1,
		},
	}
	parameters, err := expandStepAdjustments(expanded)
	if err != nil {
		t.Fatalf("bad: %#v", err)
	}

	expected := &autoscaling.StepAdjustment{
		MetricIntervalLowerBound: aws.Float64(1.0),
		MetricIntervalUpperBound: aws.Float64(2.0),
		ScalingAdjustment:        aws.Int64(int64(1)),
	}

	if !reflect.DeepEqual(parameters[0], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			parameters[0],
			expected)
	}
}

func TestFlattenParameters(t *testing.T) {
	cases := []struct {
		Input  []*rds.Parameter
		Output []map[string]interface{}
	}{
		{
			Input: []*rds.Parameter{
				{
					ParameterName:  aws.String("character_set_client"),
					ParameterValue: aws.String("utf8"),
				},
			},
			Output: []map[string]interface{}{
				{
					"name":  "character_set_client",
					"value": "utf8",
				},
			},
		},
	}

	for _, tc := range cases {
		output := flattenParameters(tc.Input)
		if !reflect.DeepEqual(output, tc.Output) {
			t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v", output, tc.Output)
		}
	}
}

func TestFlattenRedshiftParameters(t *testing.T) {
	cases := []struct {
		Input  []*redshift.Parameter
		Output []map[string]interface{}
	}{
		{
			Input: []*redshift.Parameter{
				{
					ParameterName:  aws.String("character_set_client"),
					ParameterValue: aws.String("utf8"),
				},
			},
			Output: []map[string]interface{}{
				{
					"name":  "character_set_client",
					"value": "utf8",
				},
			},
		},
	}

	for _, tc := range cases {
		output := flattenRedshiftParameters(tc.Input)
		if !reflect.DeepEqual(output, tc.Output) {
			t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v", output, tc.Output)
		}
	}
}

func TestExpandInstanceString(t *testing.T) {

	expected := []*elb.Instance{
		{InstanceId: aws.String("test-one")},
		{InstanceId: aws.String("test-two")},
	}

	ids := []interface{}{
		"test-one",
		"test-two",
	}

	expanded := expandInstanceString(ids)

	if !reflect.DeepEqual(expanded, expected) {
		t.Fatalf("Expand Instance String output did not match.\nGot:\n%#v\n\nexpected:\n%#v", expanded, expected)
	}
}

func TestFlattenNetworkInterfacesPrivateIPAddresses(t *testing.T) {
	expanded := []*ec2.NetworkInterfacePrivateIpAddress{
		{PrivateIpAddress: aws.String("192.168.0.1")},
		{PrivateIpAddress: aws.String("192.168.0.2")},
	}

	result := flattenNetworkInterfacesPrivateIPAddresses(expanded)

	if result == nil {
		t.Fatal("result was nil")
	}

	if len(result) != 2 {
		t.Fatalf("expected result had %d elements, but got %d", 2, len(result))
	}

	if result[0] != "192.168.0.1" {
		t.Fatalf("expected ip to be 192.168.0.1, but was %s", result[0])
	}

	if result[1] != "192.168.0.2" {
		t.Fatalf("expected ip to be 192.168.0.2, but was %s", result[1])
	}
}

func TestFlattenGroupIdentifiers(t *testing.T) {
	expanded := []*ec2.GroupIdentifier{
		{GroupId: aws.String("sg-001")},
		{GroupId: aws.String("sg-002")},
	}

	result := flattenGroupIdentifiers(expanded)

	if len(result) != 2 {
		t.Fatalf("expected result had %d elements, but got %d", 2, len(result))
	}

	if result[0] != "sg-001" {
		t.Fatalf("expected id to be sg-001, but was %s", result[0])
	}

	if result[1] != "sg-002" {
		t.Fatalf("expected id to be sg-002, but was %s", result[1])
	}
}

func TestExpandPrivateIPAddresses(t *testing.T) {

	ip1 := "192.168.0.1"
	ip2 := "192.168.0.2"
	flattened := []interface{}{
		ip1,
		ip2,
	}

	result := expandPrivateIPAddresses(flattened)

	if len(result) != 2 {
		t.Fatalf("expected result had %d elements, but got %d", 2, len(result))
	}

	if aws.StringValue(result[0].PrivateIpAddress) != "192.168.0.1" || !aws.BoolValue(result[0].Primary) {
		t.Fatalf("expected ip to be 192.168.0.1 and Primary, but got %v, %t", aws.StringValue(result[0].PrivateIpAddress), aws.BoolValue(result[0].Primary))
	}

	if aws.StringValue(result[1].PrivateIpAddress) != "192.168.0.2" || aws.BoolValue(result[1].Primary) {
		t.Fatalf("expected ip to be 192.168.0.2 and not Primary, but got %v, %t", aws.StringValue(result[1].PrivateIpAddress), aws.BoolValue(result[1].Primary))
	}
}

func TestFlattenAttachment(t *testing.T) {
	expanded := &ec2.NetworkInterfaceAttachment{
		InstanceId:   aws.String("i-00001"),
		DeviceIndex:  aws.Int64(int64(1)),
		AttachmentId: aws.String("at-002"),
	}

	result := flattenAttachment(expanded)

	if result == nil {
		t.Fatal("expected result to have value, but got nil")
	}

	if result["instance"] != "i-00001" {
		t.Fatalf("expected instance to be i-00001, but got %s", result["instance"])
	}

	if result["device_index"] != int64(1) {
		t.Fatalf("expected device_index to be 1, but got %d", result["device_index"])
	}

	if result["attachment_id"] != "at-002" {
		t.Fatalf("expected attachment_id to be at-002, but got %s", result["attachment_id"])
	}
}

func TestFlattenAttachmentWhenNoInstanceId(t *testing.T) {
	expanded := &ec2.NetworkInterfaceAttachment{
		DeviceIndex:  aws.Int64(int64(1)),
		AttachmentId: aws.String("at-002"),
	}

	result := flattenAttachment(expanded)

	if result == nil {
		t.Fatal("expected result to have value, but got nil")
	}

	if result["instance"] != nil {
		t.Fatalf("expected instance to be nil, but got %s", result["instance"])
	}
}

func TestFlattenStepAdjustments(t *testing.T) {
	expanded := []*autoscaling.StepAdjustment{
		{
			MetricIntervalLowerBound: aws.Float64(1.0),
			MetricIntervalUpperBound: aws.Float64(2.5),
			ScalingAdjustment:        aws.Int64(int64(1)),
		},
	}

	result := flattenStepAdjustments(expanded)[0]
	if result == nil {
		t.Fatal("expected result to have value, but got nil")
	}
	if result["metric_interval_lower_bound"] != "1" {
		t.Fatalf("expected metric_interval_lower_bound to be 1, but got %s", result["metric_interval_lower_bound"])
	}
	if result["metric_interval_upper_bound"] != "2.5" {
		t.Fatalf("expected metric_interval_upper_bound to be 2.5, but got %s", result["metric_interval_upper_bound"])
	}
	if result["scaling_adjustment"] != int64(1) {
		t.Fatalf("expected scaling_adjustment to be 1, but got %d", result["scaling_adjustment"])
	}
}

func TestFlattenResourceRecords(t *testing.T) {
	original := []string{
		`127.0.0.1`,
		`"abc def"`,
		`"abc" "def"`,
		`"abc" ""`,
	}

	dequoted := []string{
		`127.0.0.1`,
		`abc def`,
		`abc" "def`,
		`abc" "`,
	}

	var wrapped []*route53.ResourceRecord = nil
	for _, original := range original {
		wrapped = append(wrapped, &route53.ResourceRecord{Value: aws.String(original)})
	}

	sub := func(recordType string, expected []string) {
		t.Run(recordType, func(t *testing.T) {
			checkFlattenResourceRecords(t, recordType, wrapped, expected)
		})
	}

	// These record types should be dequoted.
	sub("TXT", dequoted)
	sub("SPF", dequoted)

	// These record types should not be touched.
	sub("CNAME", original)
	sub("MX", original)
}

func checkFlattenResourceRecords(
	t *testing.T,
	recordType string,
	expanded []*route53.ResourceRecord,
	expected []string) {

	result := flattenResourceRecords(expanded, recordType)

	if result == nil {
		t.Fatal("expected result to have value, but got nil")
	}

	if len(result) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, result)
	}

	for i, e := range expected {
		if result[i] != e {
			t.Fatalf("expected %v, got %v", expected, result)
		}
	}
}

func TestFlattenAsgEnabledMetrics(t *testing.T) {
	expanded := []*autoscaling.EnabledMetric{
		{Granularity: aws.String("1Minute"), Metric: aws.String("GroupTotalInstances")},
		{Granularity: aws.String("1Minute"), Metric: aws.String("GroupMaxSize")},
	}

	result := flattenAsgEnabledMetrics(expanded)

	if len(result) != 2 {
		t.Fatalf("expected result had %d elements, but got %d", 2, len(result))
	}

	if result[0] != "GroupTotalInstances" {
		t.Fatalf("expected id to be GroupTotalInstances, but was %s", result[0])
	}

	if result[1] != "GroupMaxSize" {
		t.Fatalf("expected id to be GroupMaxSize, but was %s", result[1])
	}
}

func TestFlattenKinesisShardLevelMetrics(t *testing.T) {
	expanded := []*kinesis.EnhancedMetrics{
		{
			ShardLevelMetrics: []*string{
				aws.String("IncomingBytes"),
				aws.String("IncomingRecords"),
			},
		},
	}
	result := flattenKinesisShardLevelMetrics(expanded)
	if len(result) != 2 {
		t.Fatalf("expected result had %d elements, but got %d", 2, len(result))
	}
	if result[0] != "IncomingBytes" {
		t.Fatalf("expected element 0 to be IncomingBytes, but was %s", result[0])
	}
	if result[1] != "IncomingRecords" {
		t.Fatalf("expected element 0 to be IncomingRecords, but was %s", result[1])
	}
}

func TestFlattenSecurityGroups(t *testing.T) {
	cases := []struct {
		ownerId  *string
		pairs    []*ec2.UserIdGroupPair
		expected []*GroupIdentifier
	}{
		// simple, no user id included (we ignore it mostly)
		{
			ownerId: aws.String("user1234"),
			pairs: []*ec2.UserIdGroupPair{
				{
					GroupId: aws.String("sg-12345"),
				},
			},
			expected: []*GroupIdentifier{
				{
					GroupId: aws.String("sg-12345"),
				},
			},
		},
		// include the owner id, but keep it consitent with the same account. Tests
		// EC2 classic situation
		{
			ownerId: aws.String("user1234"),
			pairs: []*ec2.UserIdGroupPair{
				{
					GroupId: aws.String("sg-12345"),
					UserId:  aws.String("user1234"),
				},
			},
			expected: []*GroupIdentifier{
				{
					GroupId: aws.String("sg-12345"),
				},
			},
		},

		// include the owner id, but from a different account. This is reflects
		// EC2 Classic when referring to groups by name
		{
			ownerId: aws.String("user1234"),
			pairs: []*ec2.UserIdGroupPair{
				{
					GroupId:   aws.String("sg-12345"),
					GroupName: aws.String("somegroup"), // GroupName is only included in Classic
					UserId:    aws.String("user4321"),
				},
			},
			expected: []*GroupIdentifier{
				{
					GroupId:   aws.String("sg-12345"),
					GroupName: aws.String("user4321/somegroup"),
				},
			},
		},

		// include the owner id, but from a different account. This reflects in
		// EC2 VPC when referring to groups by id
		{
			ownerId: aws.String("user1234"),
			pairs: []*ec2.UserIdGroupPair{
				{
					GroupId: aws.String("sg-12345"),
					UserId:  aws.String("user4321"),
				},
			},
			expected: []*GroupIdentifier{
				{
					GroupId: aws.String("user4321/sg-12345"),
				},
			},
		},

		// include description
		{
			ownerId: aws.String("user1234"),
			pairs: []*ec2.UserIdGroupPair{
				{
					GroupId:     aws.String("sg-12345"),
					Description: aws.String("desc"),
				},
			},
			expected: []*GroupIdentifier{
				{
					GroupId:     aws.String("sg-12345"),
					Description: aws.String("desc"),
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenSecurityGroups(c.pairs, c.ownerId)
		if !reflect.DeepEqual(out, c.expected) {
			t.Fatalf("Error matching output and expected: %#v vs %#v", out, c.expected)
		}
	}
}

func TestFlattenApiGatewayThrottleSettings(t *testing.T) {
	expectedBurstLimit := int64(140)
	expectedRateLimit := 120.0

	ts := &apigateway.ThrottleSettings{
		BurstLimit: aws.Int64(expectedBurstLimit),
		RateLimit:  aws.Float64(expectedRateLimit),
	}
	result := flattenApiGatewayThrottleSettings(ts)

	if len(result) != 1 {
		t.Fatalf("Expected map to have exactly 1 element, got %d", len(result))
	}

	burstLimit, ok := result[0]["burst_limit"]
	if !ok {
		t.Fatal("Expected 'burst_limit' key in the map")
	}
	burstLimitInt, ok := burstLimit.(int64)
	if !ok {
		t.Fatal("Expected 'burst_limit' to be int")
	}
	if burstLimitInt != expectedBurstLimit {
		t.Fatalf("Expected 'burst_limit' to equal %d, got %d", expectedBurstLimit, burstLimitInt)
	}

	rateLimit, ok := result[0]["rate_limit"]
	if !ok {
		t.Fatal("Expected 'rate_limit' key in the map")
	}
	rateLimitFloat, ok := rateLimit.(float64)
	if !ok {
		t.Fatal("Expected 'rate_limit' to be float64")
	}
	if rateLimitFloat != expectedRateLimit {
		t.Fatalf("Expected 'rate_limit' to equal %f, got %f", expectedRateLimit, rateLimitFloat)
	}
}

func TestExpandPolicyAttributes(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"name":  "Protocol-TLSv1",
			"value": "false",
		},
		map[string]interface{}{
			"name":  "Protocol-TLSv1.1",
			"value": "false",
		},
		map[string]interface{}{
			"name":  "Protocol-TLSv1.2",
			"value": "true",
		},
	}
	attributes := expandPolicyAttributes(expanded)

	if len(attributes) != 3 {
		t.Fatalf("expected number of attributes to be 3, but got %d", len(attributes))
	}

	expected := &elb.PolicyAttribute{
		AttributeName:  aws.String("Protocol-TLSv1.2"),
		AttributeValue: aws.String("true"),
	}

	if !reflect.DeepEqual(attributes[2], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			attributes[2],
			expected)
	}
}

func TestExpandPolicyAttributes_invalid(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"name":  "Protocol-TLSv1.2",
			"value": "true",
		},
	}
	attributes := expandPolicyAttributes(expanded)

	expected := &elb.PolicyAttribute{
		AttributeName:  aws.String("Protocol-TLSv1.2"),
		AttributeValue: aws.String("false"),
	}

	if reflect.DeepEqual(attributes[0], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			attributes[0],
			expected)
	}
}

func TestExpandPolicyAttributes_empty(t *testing.T) {
	var expanded []interface{}

	attributes := expandPolicyAttributes(expanded)

	if len(attributes) != 0 {
		t.Fatalf("expected number of attributes to be 0, but got %d", len(attributes))
	}
}

func TestFlattenPolicyAttributes(t *testing.T) {
	cases := []struct {
		Input  []*elb.PolicyAttributeDescription
		Output []interface{}
	}{
		{
			Input: []*elb.PolicyAttributeDescription{
				{
					AttributeName:  aws.String("Protocol-TLSv1.2"),
					AttributeValue: aws.String("true"),
				},
			},
			Output: []interface{}{
				map[string]string{
					"name":  "Protocol-TLSv1.2",
					"value": "true",
				},
			},
		},
	}

	for _, tc := range cases {
		output := flattenPolicyAttributes(tc.Input)
		if !reflect.DeepEqual(output, tc.Output) {
			t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v", output, tc.Output)
		}
	}
}

func TestCheckYamlString(t *testing.T) {
	var err error
	var actual string

	validYaml := `---
abc:
  def: 123
  xyz:
    -
      a: "ホリネズミ"
      b: "1"
`

	actual, err = checkYamlString(validYaml)
	if err != nil {
		t.Fatalf("Expected not to throw an error while parsing YAML, but got: %s", err)
	}

	// We expect the same YAML string back
	if actual != validYaml {
		t.Fatalf("Got:\n\n%s\n\nExpected:\n\n%s\n", actual, validYaml)
	}

	invalidYaml := `abc: [`

	actual, err = checkYamlString(invalidYaml)
	if err == nil {
		t.Fatalf("Expected to throw an error while parsing YAML, but got: %s", err)
	}

	// We expect the invalid YAML to be shown back to us again.
	if actual != invalidYaml {
		t.Fatalf("Got:\n\n%s\n\nExpected:\n\n%s\n", actual, invalidYaml)
	}
}

func TestNormalizeJsonOrYamlString(t *testing.T) {
	var err error
	var actual string

	validNormalizedJson := `{"abc":"1"}`
	actual, err = normalizeJsonOrYamlString(validNormalizedJson)
	if err != nil {
		t.Fatalf("Expected not to throw an error while parsing template, but got: %s", err)
	}
	if actual != validNormalizedJson {
		t.Fatalf("Got:\n\n%s\n\nExpected:\n\n%s\n", actual, validNormalizedJson)
	}

	validNormalizedYaml := `abc: 1
`
	actual, err = normalizeJsonOrYamlString(validNormalizedYaml)
	if err != nil {
		t.Fatalf("Expected not to throw an error while parsing template, but got: %s", err)
	}
	if actual != validNormalizedYaml {
		t.Fatalf("Got:\n\n%s\n\nExpected:\n\n%s\n", actual, validNormalizedYaml)
	}
}

func TestCognitoUserPoolSchemaAttributeMatchesStandardAttribute(t *testing.T) {
	cases := []struct {
		Input    *cognitoidentityprovider.SchemaAttributeType
		Expected bool
	}{
		{
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("10"),
					MinLength: aws.String("10"),
				},
			},
			Expected: true,
		},
		{
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(true),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("10"),
					MinLength: aws.String("10"),
				},
			},
			Expected: false,
		},
		{
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(false),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("10"),
					MinLength: aws.String("10"),
				},
			},
			Expected: false,
		},
		{
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("non-existent"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("10"),
					MinLength: aws.String("10"),
				},
			},
			Expected: false,
		},
		{
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(true),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("10"),
					MinLength: aws.String("10"),
				},
			},
			Expected: false,
		},
		{
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("999"),
					MinLength: aws.String("10"),
				},
			},
			Expected: false,
		},
		{
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("10"),
					MinLength: aws.String("999"),
				},
			},
			Expected: false,
		},
		{
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeBoolean),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("email_verified"),
				Required:               aws.Bool(false),
			},
			Expected: true,
		},
		{
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeNumber),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("updated_at"),
				NumberAttributeConstraints: &cognitoidentityprovider.NumberAttributeConstraintsType{
					MinValue: aws.String("0"),
				},
				Required: aws.Bool(false),
			},
			Expected: true,
		},
	}

	for _, tc := range cases {
		output := cognitoUserPoolSchemaAttributeMatchesStandardAttribute(tc.Input)
		if output != tc.Expected {
			t.Fatalf("Expected %t match with standard attribute on input: \n\n%#v\n\n", tc.Expected, tc.Input)
		}
	}
}

func TestCanonicalXML(t *testing.T) {
	cases := []struct {
		Name        string
		Config      string
		Expected    string
		ExpectError bool
	}{
		{
			Name:     "Config sample from MSDN",
			Config:   testExampleXML_from_msdn,
			Expected: testExampleXML_from_msdn,
		},
		{
			Name:     "Config sample from MSDN, modified",
			Config:   testExampleXML_from_msdn,
			Expected: testExampleXML_from_msdn_modified,
		},
		{
			Name:        "Config sample from MSDN, flaw",
			Config:      testExampleXML_from_msdn,
			Expected:    testExampleXML_from_msdn_flawed,
			ExpectError: true,
		},
		{
			Name: "A note",
			Config: `
<?xml version="1.0"?>
<note>
<to>You</to>
<from>Me</from>
<heading>Reminder</heading>
<body>You're awesome</body>
<rant/>
<rant/>
</note>
`,
			Expected: `
<?xml version="1.0"?>
<note>
	<to>You</to>
	<from>Me</from>
	<heading>
    Reminder
    </heading>
	<body>You're awesome</body>
	<rant/>
	<rant>
</rant>
</note>`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			config, err := canonicalXML(tc.Config)
			if err != nil {
				t.Fatalf("Error getting canonical xml for given config: %s", err)
			}
			expected, err := canonicalXML(tc.Expected)
			if err != nil {
				t.Fatalf("Error getting canonical xml for expected config: %s", err)
			}

			if config != expected {
				if !tc.ExpectError {
					t.Fatalf("Error matching canonical xmls:\n\tconfig: %s\n\n\texpected: %s\n", config, expected)
				}
			}
		})
	}
}

const testExampleXML_from_msdn = `
<?xml version="1.0"?>
<purchaseOrder xmlns="http://tempuri.org/po.xsd" orderDate="1999-10-20">
    <shipTo country="US">
        <name>Alice Smith</name>
        <street>123 Maple Street</street>
        <city>Mill Valley</city>
        <state>CA</state>
        <zip>90952</zip>
    </shipTo>
    <billTo country="US">
        <name>Robert Smith</name>
        <street>8 Oak Avenue</street>
        <city>Old Town</city>
        <state>PA</state>
        <zip>95819</zip>
    </billTo>
    <comment>Hurry, my lawn is going wild!</comment>
    <items>
        <item partNum="872-AA">
            <productName>Lawnmower</productName>
            <quantity>1</quantity>
            <USPrice>148.95</USPrice>
            <comment>Confirm this is electric</comment>
        </item>
        <item partNum="926-AA">
            <productName>Baby Monitor</productName>
            <quantity>1</quantity>
            <USPrice>39.98</USPrice>
            <shipDate>1999-05-21</shipDate>
        </item>
				<item/>
				<item/>
    </items>
</purchaseOrder>
`

const testExampleXML_from_msdn_modified = `
<?xml version="1.0"?>
<purchaseOrder xmlns="http://tempuri.org/po.xsd" orderDate="1999-10-20">
    <shipTo country="US">
        <name>Alice Smith</name>
        <street>123 Maple Street</street>
        <city>Mill Valley</city>
        <state>CA</state>
        <zip>90952</zip>
    </shipTo>
    <billTo country="US">
        <name>Robert Smith</name>
        <street>8 Oak Avenue</street>
        <city>Old Town</city>
        <state>PA</state>
        <zip>95819</zip>
    </billTo>
    <comment>Hurry, my lawn is going wild!</comment>
    <items>
        <item partNum="872-AA">
            <productName>Lawnmower</productName>
            <quantity>1</quantity>
            <USPrice>148.95</USPrice>
            <comment>Confirm this is electric</comment>
        </item>
        <item partNum="926-AA">
            <productName>Baby Monitor</productName>
            <quantity>1</quantity>
            <USPrice>39.98</USPrice>
            <shipDate>1999-05-21</shipDate>
        </item>
				  	 <item></item>
				<item>
</item>
    </items>
</purchaseOrder>
`

const testExampleXML_from_msdn_flawed = `
<?xml version="1.0"?>
<purchaseOrder xmlns="http://tempuri.org/po.xsd" orderDate="1999-10-20">
    <shipTo country="US">
        <name>Alice Smith</name>
        <street>123 Maple Street</street>
        <city>Mill Valley</city>
        <state>CA</state>
        <zip>90952</zip>
    </shipTo>
    <billTo country="US">
        <name>Robert Smith</name>
        <street>8 Oak Avenue</street>
        <city>Old Town</city>
        <state>PA</state>
        <zip>95819</zip>
    </billTo>
    <comment>Hurry, my lawn is going wild!</comment>
    <items>
        <item partNum="872-AA">
            <productName>Lawnmower</productName>
            <quantity>1</quantity>
            <USPrice>148.95</USPrice>
            <comment>Confirm this is electric</comment>
        </item>
        <item partNum="926-AA">
            <productName>Baby Monitor</productName>
            <quantity>1</quantity>
            <USPrice>39.98</USPrice>
            <shipDate>1999-05-21</shipDate>
        </item>
				<item>
				flaw
				</item>
    </items>
</purchaseOrder>
`

// TestExpandRdsClusterScalingConfiguration_serverless removed in v3.0.0
// as all engine_modes are treated equal when expanding scaling_configuration
// and an override of min_capacity is no longer needed
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11698

func TestExpandRdsClusterScalingConfiguration_basic(t *testing.T) {
	type testCase struct {
		EngineMode string
		Input      []interface{}
		ExpectNil  bool
	}
	cases := []testCase{}

	// RDS Cluster Scaling Configuration is only valid for serverless, but we're relying on AWS errors.
	// If Terraform adds whole-resource validation, we can do our own validation at plan time.
	for _, engineMode := range []string{"global", "multimaster", "parallelquery", "provisioned", "serverless"} {
		cases = append(cases, []testCase{
			{
				EngineMode: engineMode,
				Input: []interface{}{
					map[string]interface{}{
						"auto_pause":               false,
						"max_capacity":             32,
						"min_capacity":             4,
						"seconds_until_auto_pause": 600,
						"timeout_action":           "ForceApplyCapacityChange",
					},
				},
				ExpectNil: false,
			},
			{
				EngineMode: engineMode,
				Input:      []interface{}{},
				ExpectNil:  true,
			}, {
				EngineMode: engineMode,
				Input: []interface{}{
					nil,
				},
				ExpectNil: true,
			},
		}...)
	}

	for _, tc := range cases {
		output := expandRdsClusterScalingConfiguration(tc.Input)
		if tc.ExpectNil != (output == nil) {
			t.Errorf("EngineMode %q: Expected nil: %t, Got: %v", tc.EngineMode, tc.ExpectNil, output)
		}
	}
}
