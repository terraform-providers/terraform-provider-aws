package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestSuppressEquivalentJsonDiffsWhitespaceAndNoWhitespace(t *testing.T) {
	d := new(schema.ResourceData)

	noWhitespace := `{"test":"test"}`
	whitespace := `
{
  "test": "test"
}`

	if !suppressEquivalentJsonDiffs("", noWhitespace, whitespace, d) {
		t.Errorf("Expected suppressEquivalentJsonDiffs to return true for %s == %s", noWhitespace, whitespace)
	}

	noWhitespaceDiff := `{"test":"test"}`
	whitespaceDiff := `
{
  "test": "tested"
}`

	if suppressEquivalentJsonDiffs("", noWhitespaceDiff, whitespaceDiff, d) {
		t.Errorf("Expected suppressEquivalentJsonDiffs to return false for %s == %s", noWhitespaceDiff, whitespaceDiff)
	}
}

func TestSuppressEquivalentTypeStringBoolean(t *testing.T) {
	testCases := []struct {
		old        string
		new        string
		equivalent bool
	}{
		{
			old:        "false",
			new:        "0",
			equivalent: true,
		},
		{
			old:        "true",
			new:        "1",
			equivalent: true,
		},
		{
			old:        "",
			new:        "0",
			equivalent: false,
		},
		{
			old:        "",
			new:        "1",
			equivalent: false,
		},
	}

	for i, tc := range testCases {
		value := suppressEquivalentTypeStringBoolean("test_property", tc.old, tc.new, nil)

		if tc.equivalent && !value {
			t.Fatalf("expected test case %d to be equivalent", i)
		}

		if !tc.equivalent && value {
			t.Fatalf("expected test case %d to not be equivalent", i)
		}
	}
}

func TestSuppressEquivalentJsonOrYamlDiffs(t *testing.T) {
	testCases := []struct {
		description string
		equivalent  bool
		old         string
		new         string
	}{
		{
			description: `JSON no change`,
			equivalent:  true,
			old: `
{
   "Resources": {
      "TestVpc": {
         "Type": "AWS::EC2::VPC",
         "Properties": {
            "CidrBlock": "10.0.0.0/16"
         }
      }
   },
   "Outputs": {
      "TestVpcID": {
         "Value": { "Ref" : "TestVpc" }
      }
   }
}
`,
			new: `
{
   "Resources": {
      "TestVpc": {
         "Type": "AWS::EC2::VPC",
         "Properties": {
            "CidrBlock": "10.0.0.0/16"
         }
      }
   },
   "Outputs": {
      "TestVpcID": {
         "Value": { "Ref" : "TestVpc" }
      }
   }
}
`,
		},
		{
			description: `JSON whitespace`,
			equivalent:  true,
			old:         `{"Resources":{"TestVpc":{"Type":"AWS::EC2::VPC","Properties":{"CidrBlock":"10.0.0.0/16"}}},"Outputs":{"TestVpcID":{"Value":{"Ref":"TestVpc"}}}}`,
			new: `
{
   "Resources": {
      "TestVpc": {
         "Type": "AWS::EC2::VPC",
         "Properties": {
            "CidrBlock": "10.0.0.0/16"
         }
      }
   },
   "Outputs": {
      "TestVpcID": {
         "Value": { "Ref" : "TestVpc" }
      }
   }
}
`,
		},
		{
			description: `JSON change`,
			equivalent:  false,
			old: `
{
   "Resources": {
      "TestVpc": {
         "Type": "AWS::EC2::VPC",
         "Properties": {
            "CidrBlock": "10.0.0.0/16"
         }
      }
   },
   "Outputs": {
      "TestVpcID": {
         "Value": { "Ref" : "TestVpc" }
      }
   }
}
`,
			new: `
{
   "Resources": {
      "TestVpc": {
         "Type": "AWS::EC2::VPC",
         "Properties": {
            "CidrBlock": "172.16.0.0/16"
         }
      }
   },
   "Outputs": {
      "TestVpcID": {
         "Value": { "Ref" : "TestVpc" }
      }
   }
}
`,
		},
		{
			description: `YAML no change`,
			equivalent:  true,
			old: `
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
Outputs:
  TestVpcID:
    Value: !Ref TestVpc
`,
			new: `
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
Outputs:
  TestVpcID:
    Value: !Ref TestVpc
`,
		},
		{
			description: `YAML whitespace`,
			equivalent:  false,
			old: `
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16

Outputs:
  TestVpcID:
    Value: !Ref TestVpc

`,
			new: `
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
Outputs:
  TestVpcID:
    Value: !Ref TestVpc
`,
		},
		{
			description: `YAML change`,
			equivalent:  false,
			old: `
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 172.16.0.0/16
Outputs:
  TestVpcID:
    Value: !Ref TestVpc
`,
			new: `
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
Outputs:
  TestVpcID:
    Value: !Ref TestVpc
`,
		},
	}

	for _, tc := range testCases {
		value := suppressEquivalentJsonOrYamlDiffs("test_property", tc.old, tc.new, nil)

		if tc.equivalent && !value {
			t.Fatalf("expected test case (%s) to be equivalent", tc.description)
		}

		if !tc.equivalent && value {
			t.Fatalf("expected test case (%s) to not be equivalent", tc.description)
		}
	}
}
