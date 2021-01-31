package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDataSourceIAMPolicyDocument_basic(t *testing.T) {
	// This really ought to be able to be a unit test rather than an
	// acceptance test, but just instantiating the AWS provider requires
	// some AWS API calls, and so this needs valid AWS credentials to work.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMPolicyDocumentConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test", "json",
						testAccAWSIAMPolicyDocumentExpectedJSON(),
					),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMPolicyDocument_source(t *testing.T) {
	// This really ought to be able to be a unit test rather than an
	// acceptance test, but just instantiating the AWS provider requires
	// some AWS API calls, and so this needs valid AWS credentials to work.
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMPolicyDocumentSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test_source", "json",
						testAccAWSIAMPolicyDocumentSourceExpectedJSON(),
					),
				),
			},
			{
				Config: testAccAWSIAMPolicyDocumentSourceBlankConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test_source_blank", "json",
						testAccAWSIAMPolicyDocumentSourceBlankExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMPolicyDocument_sourceConflicting(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMPolicyDocumentSourceConflictingConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test_source_conflicting", "json",
						testAccAWSIAMPolicyDocumentSourceConflictingExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMPolicyDocument_override(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMPolicyDocumentOverrideConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test_override", "json",
						testAccAWSIAMPolicyDocumentOverrideExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMPolicyDocument_noStatementMerge(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMPolicyDocumentNoStatementMergeConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.yak_politik", "json",
						testAccAWSIAMPolicyDocumentNoStatementMergeExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMPolicyDocument_noStatementOverride(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMPolicyDocumentNoStatementOverrideConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.yak_politik", "json",
						testAccAWSIAMPolicyDocumentNoStatementOverrideExpectedJSON,
					),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMPolicyDocument_duplicateSid(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSIAMPolicyDocumentDuplicateSidConfig,
				ExpectError: regexp.MustCompile(`Found duplicate sid`),
			},
			{
				Config: testAccAWSIAMPolicyDocumentDuplicateBlankSidConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test", "json",
						testAccAWSIAMPolicyDocumentDuplicateBlankSidExpectedJSON,
					),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/10777
func TestAccAWSDataSourceIAMPolicyDocument_Statement_Principal_Identifiers_StringAndSlice(t *testing.T) {
	dataSourceName := "data.aws_iam_policy_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMPolicyDocumentConfigStatementPrincipalIdentifiersStringAndSlice,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json", testAccAWSIAMPolicyDocumentExpectedJSONStatementPrincipalIdentifiersStringAndSlice()),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/10777
func TestAccAWSDataSourceIAMPolicyDocument_Statement_Principal_Identifiers_MultiplePrincipals(t *testing.T) {
	dataSourceName := "data.aws_iam_policy_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionPreCheck(endpoints.AwsPartitionID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMPolicyDocumentConfigStatementPrincipalIdentifiersMultiplePrincipals,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json", testAccAWSIAMPolicyDocumentExpectedJSONStatementPrincipalIdentifiersMultiplePrincipalsAWS()),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMPolicyDocument_Statement_Principal_Identifiers_MultiplePrincipalsGov(t *testing.T) {
	dataSourceName := "data.aws_iam_policy_document.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionPreCheck(endpoints.AwsUsGovPartitionID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMPolicyDocumentConfigStatementPrincipalIdentifiersMultiplePrincipals,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "json", testAccAWSIAMPolicyDocumentExpectedJSONStatementPrincipalIdentifiersMultiplePrincipalsGov()),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMPolicyDocument_Version_20081017(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSIAMPolicyDocumentDataSourceConfigVersion20081017ConversionCondition,
				ExpectError: regexp.MustCompile(`found \&\{ sequence in \(.+\), which is not supported in document version 2008-10-17`),
			},
			{
				Config:      testAccAWSIAMPolicyDocumentDataSourceConfigVersion20081017ConversionNotPrincipals,
				ExpectError: regexp.MustCompile(`found \&\{ sequence in \(.+\), which is not supported in document version 2008-10-17`),
			},
			{
				Config:      testAccAWSIAMPolicyDocumentDataSourceConfigVersion20081017ConversionNotResources,
				ExpectError: regexp.MustCompile(`found \&\{ sequence in \(.+\), which is not supported in document version 2008-10-17`),
			},
			{
				Config:      testAccAWSIAMPolicyDocumentDataSourceConfigVersion20081017ConversionPrincipal,
				ExpectError: regexp.MustCompile(`found \&\{ sequence in \(.+\), which is not supported in document version 2008-10-17`),
			},
			{
				Config:      testAccAWSIAMPolicyDocumentDataSourceConfigVersion20081017ConversionResources,
				ExpectError: regexp.MustCompile(`found \&\{ sequence in \(.+\), which is not supported in document version 2008-10-17`),
			},
			{
				Config: testAccAWSIAMPolicyDocumentDataSourceConfigVersion20081017,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy_document.test", "json", testAccAWSIAMPolicyDocumentDataSourceConfigVersion20081017ExpectedJSON),
				),
			},
		},
	})
}

var testAccAWSIAMPolicyDocumentConfig = `
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  policy_id = "policy_id"

  statement {
    sid = "1"
    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::*",
    ]
  }

  statement {
    actions = [
      "s3:ListBucket",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::foo",
    ]

    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values = [
        "home/",
        "home/&{aws:username}/",
      ]
    }

    not_principals {
      type        = "AWS"
      identifiers = ["arn:blahblah:example"]
    }
  }

  statement {
    actions = [
      "s3:*",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}",
      "arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:blahblah:example"]
    }
  }

  statement {
    effect        = "Deny"
    not_actions   = ["s3:*"]
    not_resources = ["arn:${data.aws_partition.current.partition}:s3:::*"]
  }

  # Normalization of wildcard principals

  statement {
    effect  = "Allow"
    actions = ["kinesis:*"]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }

  statement {
    effect  = "Allow"
    actions = ["firehose:*"]

    principals {
      type        = "*"
      identifiers = ["*"]
    }
  }
}
`

func testAccAWSIAMPolicyDocumentExpectedJSON() string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Id": "policy_id",
  "Statement": [
    {
      "Sid": "1",
      "Effect": "Allow",
      "Action": [
        "s3:ListAllMyBuckets",
        "s3:GetBucketLocation"
      ],
      "Resource": "arn:%[1]s:s3:::*"
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "s3:ListBucket",
      "Resource": "arn:%[1]s:s3:::foo",
      "NotPrincipal": {
        "AWS": "arn:blahblah:example"
      },
      "Condition": {
        "StringLike": {
          "s3:prefix": [
            "home/${aws:username}/",
            "home/"
          ]
        }
      }
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "arn:%[1]s:s3:::foo/home/${aws:username}/*",
        "arn:%[1]s:s3:::foo/home/${aws:username}"
      ],
      "Principal": {
        "AWS": "arn:blahblah:example"
      }
    },
    {
      "Sid": "",
      "Effect": "Deny",
      "NotAction": "s3:*",
      "NotResource": "arn:%[1]s:s3:::*"
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "kinesis:*",
      "Principal": {
        "AWS": "*"
      }
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "firehose:*",
      "Principal": "*"
    }
  ]
}`, testAccGetPartition())
}

var testAccAWSIAMPolicyDocumentSourceConfig = `
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  policy_id = "policy_id"

  statement {
    sid = "1"
    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::*",
    ]
  }

  statement {
    actions = [
      "s3:ListBucket",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::foo",
    ]

    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values = [
        "home/",
        "home/&{aws:username}/",
      ]
    }

    not_principals {
      type        = "AWS"
      identifiers = ["arn:blahblah:example"]
    }
  }

  statement {
    actions = [
      "s3:*",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}",
      "arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}/*",
    ]

    principals {
      type = "AWS"
      identifiers = [
        "arn:blahblah:example",
        "arn:blahblahblah:example",
      ]
    }
  }

  statement {
    effect        = "Deny"
    not_actions   = ["s3:*"]
    not_resources = ["arn:${data.aws_partition.current.partition}:s3:::*"]
  }

  # Normalization of wildcard principals

  statement {
    effect  = "Allow"
    actions = ["kinesis:*"]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }

  statement {
    effect  = "Allow"
    actions = ["firehose:*"]

    principals {
      type        = "*"
      identifiers = ["*"]
    }
  }
}

data "aws_iam_policy_document" "test_source" {
  source_json = data.aws_iam_policy_document.test.json

  statement {
    sid       = "SourceJSONTest1"
    actions   = ["*"]
    resources = ["*"]
  }
}
`

func testAccAWSIAMPolicyDocumentSourceExpectedJSON() string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Id": "policy_id",
  "Statement": [
    {
      "Sid": "1",
      "Effect": "Allow",
      "Action": [
        "s3:ListAllMyBuckets",
        "s3:GetBucketLocation"
      ],
      "Resource": "arn:%[1]s:s3:::*"
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "s3:ListBucket",
      "Resource": "arn:%[1]s:s3:::foo",
      "NotPrincipal": {
        "AWS": "arn:blahblah:example"
      },
      "Condition": {
        "StringLike": {
          "s3:prefix": [
            "home/${aws:username}/",
            "home/"
          ]
        }
      }
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "arn:%[1]s:s3:::foo/home/${aws:username}/*",
        "arn:%[1]s:s3:::foo/home/${aws:username}"
      ],
      "Principal": {
        "AWS": [
          "arn:blahblahblah:example",
          "arn:blahblah:example"
        ]
      }
    },
    {
      "Sid": "",
      "Effect": "Deny",
      "NotAction": "s3:*",
      "NotResource": "arn:%[1]s:s3:::*"
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "kinesis:*",
      "Principal": {
        "AWS": "*"
      }
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "firehose:*",
      "Principal": "*"
    },
    {
      "Sid": "SourceJSONTest1",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*"
    }
  ]
}`, testAccGetPartition())
}

var testAccAWSIAMPolicyDocumentSourceBlankConfig = `
data "aws_iam_policy_document" "test_source_blank" {
  source_json = ""

  statement {
    sid       = "SourceJSONTest2"
    actions   = ["*"]
    resources = ["*"]
  }
}
`

var testAccAWSIAMPolicyDocumentSourceBlankExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "SourceJSONTest2",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*"
    }
  ]
}`

var testAccAWSIAMPolicyDocumentSourceConflictingConfig = `
data "aws_iam_policy_document" "test_source" {
  statement {
    sid       = "SourceJSONTestConflicting"
    actions   = ["iam:*"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "test_source_conflicting" {
  source_json = data.aws_iam_policy_document.test_source.json

  statement {
    sid       = "SourceJSONTestConflicting"
    actions   = ["*"]
    resources = ["*"]
  }
}
`

var testAccAWSIAMPolicyDocumentSourceConflictingExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "SourceJSONTestConflicting",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*"
    }
  ]
}`

var testAccAWSIAMPolicyDocumentOverrideConfig = `
data "aws_partition" "current" {}

data "aws_iam_policy_document" "override" {
  statement {
    sid = "SidToOverwrite"

    actions   = ["s3:*"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "test_override" {
  override_json = data.aws_iam_policy_document.override.json

  statement {
    actions   = ["ec2:*"]
    resources = ["*"]
  }

  statement {
    sid = "SidToOverwrite"

    actions = ["s3:*"]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::somebucket",
      "arn:${data.aws_partition.current.partition}:s3:::somebucket/*",
    ]
  }
}
`

var testAccAWSIAMPolicyDocumentOverrideExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "ec2:*",
      "Resource": "*"
    },
    {
      "Sid": "SidToOverwrite",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": "*"
    }
  ]
}`

var testAccAWSIAMPolicyDocumentNoStatementMergeConfig = `
data "aws_iam_policy_document" "source" {
  statement {
    sid       = ""
    actions   = ["ec2:DescribeAccountAttributes"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "override" {
  statement {
    sid       = "OverridePlaceholder"
    actions   = ["s3:GetObject"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "yak_politik" {
  source_json   = data.aws_iam_policy_document.source.json
  override_json = data.aws_iam_policy_document.override.json
}
`

var testAccAWSIAMPolicyDocumentNoStatementMergeExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "ec2:DescribeAccountAttributes",
      "Resource": "*"
    },
    {
      "Sid": "OverridePlaceholder",
      "Effect": "Allow",
      "Action": "s3:GetObject",
      "Resource": "*"
    }
  ]
}`

var testAccAWSIAMPolicyDocumentNoStatementOverrideConfig = `
data "aws_iam_policy_document" "source" {
  statement {
    sid       = "OverridePlaceholder"
    actions   = ["ec2:DescribeAccountAttributes"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "override" {
  statement {
    sid       = "OverridePlaceholder"
    actions   = ["s3:GetObject"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "yak_politik" {
  source_json   = data.aws_iam_policy_document.source.json
  override_json = data.aws_iam_policy_document.override.json
}
`

var testAccAWSIAMPolicyDocumentNoStatementOverrideExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "OverridePlaceholder",
      "Effect": "Allow",
      "Action": "s3:GetObject",
      "Resource": "*"
    }
  ]
}`

var testAccAWSIAMPolicyDocumentDuplicateSidConfig = `
data "aws_iam_policy_document" "test" {
  statement {
    sid       = "1"
    effect    = "Allow"
    actions   = ["ec2:DescribeAccountAttributes"]
    resources = ["*"]
  }

  statement {
    sid       = "1"
    effect    = "Allow"
    actions   = ["s3:GetObject"]
    resources = ["*"]
  }
}
`

var testAccAWSIAMPolicyDocumentDuplicateBlankSidConfig = `
data "aws_iam_policy_document" "test" {
  statement {
    sid       = ""
    effect    = "Allow"
    actions   = ["ec2:DescribeAccountAttributes"]
    resources = ["*"]
  }

  statement {
    sid       = ""
    effect    = "Allow"
    actions   = ["s3:GetObject"]
    resources = ["*"]
  }
}
`

var testAccAWSIAMPolicyDocumentDuplicateBlankSidExpectedJSON = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "ec2:DescribeAccountAttributes",
      "Resource": "*"
    },
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "s3:GetObject",
      "Resource": "*"
    }
  ]
}`

const testAccAWSIAMPolicyDocumentDataSourceConfigVersion20081017 = `
data "aws_iam_policy_document" "test" {
  version = "2008-10-17"

  statement {
    actions   = ["ec2:*"]
    resources = ["*"]
  }
}
`

const testAccAWSIAMPolicyDocumentDataSourceConfigVersion20081017ExpectedJSON = `{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "ec2:*",
      "Resource": "*"
    }
  ]
}`

const testAccAWSIAMPolicyDocumentDataSourceConfigVersion20081017ConversionCondition = `
data "aws_iam_policy_document" "test" {
  version = "2008-10-17"

  statement {
    actions   = ["*"]
    resources = ["*"]

    condition {
      test = "StringLike"
      values = [
        "home/",
        "home/&{aws:username}/",
      ]
      variable = "s3:prefix"
    }
  }
}
`

const testAccAWSIAMPolicyDocumentDataSourceConfigVersion20081017ConversionNotPrincipals = `
data "aws_iam_policy_document" "test" {
  version = "2008-10-17"

  statement {
    actions   = ["*"]
    resources = ["*"]

    not_principals {
      identifiers = ["&{aws:username}"]
      type        = "AWS"
    }
  }
}
`

const testAccAWSIAMPolicyDocumentDataSourceConfigVersion20081017ConversionNotResources = `
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  version = "2008-10-17"

  statement {
    actions       = ["*"]
    not_resources = ["arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}", ]
  }
}
`

const testAccAWSIAMPolicyDocumentDataSourceConfigVersion20081017ConversionPrincipal = `
data "aws_iam_policy_document" "test" {
  version = "2008-10-17"

  statement {
    actions   = ["*"]
    resources = ["*"]

    principals {
      identifiers = ["&{aws:username}"]
      type        = "AWS"
    }
  }
}
`

const testAccAWSIAMPolicyDocumentDataSourceConfigVersion20081017ConversionResources = `
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  version = "2008-10-17"

  statement {
    actions   = ["*"]
    resources = ["arn:${data.aws_partition.current.partition}:s3:::foo/home/&{aws:username}", ]
  }
}
`

var testAccAWSIAMPolicyDocumentConfigStatementPrincipalIdentifiersStringAndSlice = `
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["*"]
    resources = ["*"]
    sid       = "StatementPrincipalIdentifiersStringAndSlice"

    principals {
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::111111111111:root"]
      type        = "AWS"
    }

    principals {
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::222222222222:root", "arn:${data.aws_partition.current.partition}:iam::333333333333:root"]
      type        = "AWS"
    }
  }
}
`

func testAccAWSIAMPolicyDocumentExpectedJSONStatementPrincipalIdentifiersStringAndSlice() string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "StatementPrincipalIdentifiersStringAndSlice",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*",
      "Principal": {
        "AWS": [
          "arn:%[1]s:iam::111111111111:root",
          "arn:%[1]s:iam::333333333333:root",
          "arn:%[1]s:iam::222222222222:root"
        ]
      }
    }
  ]
}`, testAccGetPartition())
}

var testAccAWSIAMPolicyDocumentConfigStatementPrincipalIdentifiersMultiplePrincipals = `
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["*"]
    resources = ["*"]
    sid       = "StatementPrincipalIdentifiersStringAndSlice"

    principals {
      identifiers = [
        "arn:${data.aws_partition.current.partition}:iam::111111111111:root",
        "arn:${data.aws_partition.current.partition}:iam::222222222222:root",
      ]
      type = "AWS"
    }

    principals {
      identifiers = [
        "arn:${data.aws_partition.current.partition}:iam::333333333333:root",
      ]
      type = "AWS"
    }

    principals {
      identifiers = [
        "arn:${data.aws_partition.current.partition}:iam::444444444444:root",
      ]
      type = "AWS"
    }
  }
}
`

func testAccAWSIAMPolicyDocumentExpectedJSONStatementPrincipalIdentifiersMultiplePrincipalsAWS() string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "StatementPrincipalIdentifiersStringAndSlice",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*",
      "Principal": {
        "AWS": [
          "arn:%[1]s:iam::333333333333:root",
          "arn:%[1]s:iam::444444444444:root",
          "arn:%[1]s:iam::222222222222:root",
          "arn:%[1]s:iam::111111111111:root"
        ]
      }
    }
  ]
}`, testAccGetPartition())
}

func testAccAWSIAMPolicyDocumentExpectedJSONStatementPrincipalIdentifiersMultiplePrincipalsGov() string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "StatementPrincipalIdentifiersStringAndSlice",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*",
      "Principal": {
        "AWS": [
          "arn:%[1]s:iam::333333333333:root",
          "arn:%[1]s:iam::222222222222:root",
          "arn:%[1]s:iam::111111111111:root",
          "arn:%[1]s:iam::444444444444:root"
        ]
      }
    }
  ]
}`, testAccGetPartition())
}
