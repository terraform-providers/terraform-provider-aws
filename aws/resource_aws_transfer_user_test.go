package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/transfer/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func TestAccAWSTransferUser_basic(t *testing.T) {
	var conf transfer.DescribedUser
	resourceName := "aws_transfer_user.test"
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferUserExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`user/.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "server_id", "aws_transfer_server.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSTransferUser_posix(t *testing.T) {
	var conf transfer.DescribedUser
	resourceName := "aws_transfer_user.test"
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferUserConfigPosix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferUserExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.0.gid", "1000"),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.0.uid", "1000"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSTransferUserConfigPosixUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferUserExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.0.gid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.0.uid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_profile.0.secondary_gids.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSTransferUser_modifyWithOptions(t *testing.T) {
	var conf transfer.DescribedUser
	resourceName := "aws_transfer_user.test"
	rName := acctest.RandString(10)
	rName2 := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferUserConfig_options(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferUserExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "home_directory", "/home/tftestuser"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.NAME", "tftestuser"),
					resource.TestCheckResourceAttr(resourceName, "tags.ENV", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.ADMIN", "test"),
				),
			},
			{
				Config: testAccAWSTransferUserConfig_modify(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferUserExists(resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "home_directory", "/test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.NAME", "tf-test-user"),
					resource.TestCheckResourceAttr(resourceName, "tags.TEST", "test2"),
				),
			},
			{
				Config: testAccAWSTransferUserConfig_forceNew(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferUserExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "user_name", "tftestuser2"),
					resource.TestCheckResourceAttrPair(resourceName, "role", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "home_directory", "/home/tftestuser2"),
				),
			},
		},
	})
}

func TestAccAWSTransferUser_disappears(t *testing.T) {
	var serverConf transfer.DescribedServer
	var userConf transfer.DescribedUser
	rName := acctest.RandString(10)
	resourceName := "aws_transfer_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferUserConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists("aws_transfer_server.test", &serverConf),
					testAccCheckAWSTransferUserExists("aws_transfer_user.test", &userConf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsTransferUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSTransferUser_UserName_Validation(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferUserDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSTransferUserName_validation("!@#$%^"),
				ExpectError: regexp.MustCompile(`Invalid "user_name": `),
			},
			{
				Config:      testAccAWSTransferUserName_validation(acctest.RandString(2)),
				ExpectError: regexp.MustCompile(`Invalid "user_name": `),
			},
			{
				Config:             testAccAWSTransferUserName_validation(acctest.RandString(33)),
				ExpectNonEmptyPlan: true,
				PlanOnly:           true,
			},
			{
				Config:      testAccAWSTransferUserName_validation(acctest.RandString(101)),
				ExpectError: regexp.MustCompile(`Invalid "user_name": `),
			},
			{
				Config:      testAccAWSTransferUserName_validation("-abcdef"),
				ExpectError: regexp.MustCompile(`Invalid "user_name": `),
			},
			{
				Config:             testAccAWSTransferUserName_validation("valid_username"),
				ExpectNonEmptyPlan: true,
				PlanOnly:           true,
			},
		},
	})
}

func TestAccAWSTransferUser_homeDirectoryMappings(t *testing.T) {
	var conf transfer.DescribedUser
	rName := acctest.RandString(10)
	resourceName := "aws_transfer_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferUserConfig_homeDirectoryMappings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferUserExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "home_directory_mappings.#", "1"),
				),
			},
			{
				Config: testAccAWSTransferUserConfig_homeDirectoryMappingsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferUserExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "home_directory_mappings.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSTransferUserExists(n string, res *transfer.DescribedUser) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer User ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).transferconn
		userName := rs.Primary.Attributes["user_name"]
		serverID := rs.Primary.Attributes["server_id"]

		describe, err := finder.UserByID(conn, serverID, userName)
		if err != nil {
			return err
		}

		*res = *describe.User

		return nil
	}
}

func testAccCheckAWSTransferUserDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).transferconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transfer_user" {
			continue
		}

		userName := rs.Primary.Attributes["user_name"]
		serverID := rs.Primary.Attributes["server_id"]

		_, err := finder.UserByID(conn, serverID, userName)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

const testAccAWSTransferUserConfig_base = `
resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"

  tags = {
    NAME = "tf-acc-test-transfer-server"
  }
}

data "aws_partition" "current" {}
`

func testAccAWSTransferUserConfig_basic(rName string) string {
	return composeConfig(testAccAWSTransferUserConfig_base, fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "tf-test-transfer-user-iam-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-test-transfer-user-iam-policy-%[1]s"
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoS3",
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_transfer_user" "test" {
  server_id = aws_transfer_server.test.id
  user_name = "tftestuser"
  role      = aws_iam_role.test.arn
}
`, rName))
}

func testAccAWSTransferUserName_validation(rName string) string {
	return composeConfig(testAccAWSTransferUserConfig_base, fmt.Sprintf(`
resource "aws_transfer_user" "test" {
  server_id = aws_transfer_server.test.id
  user_name = "%s"
  role      = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = "tf-test-transfer-user-iam-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}
`, rName))
}

func testAccAWSTransferUserConfig_options(rName string) string {
	return composeConfig(testAccAWSTransferUserConfig_base, fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "tf-test-transfer-user-iam-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-test-transfer-user-iam-policy-%[1]s"
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoS3",
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

data "aws_iam_policy_document" "test" {
  statement {
    sid = "ListHomeDir"

    actions = [
      "s3:ListBucket",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::&{transfer:HomeBucket}",
    ]
  }

  statement {
    sid = "AWSTransferRequirements"

    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]

    resources = [
      "*",
    ]
  }

  statement {
    sid = "HomeDirObjectAccess"

    actions = [
      "s3:PutObject",
      "s3:GetObject",
      "s3:DeleteObjectVersion",
      "s3:DeleteObject",
      "s3:GetObjectVersion",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::&{transfer:HomeDirectory}*",
    ]
  }
}

resource "aws_transfer_user" "test" {
  server_id      = aws_transfer_server.test.id
  user_name      = "tftestuser"
  role           = aws_iam_role.test.arn
  policy         = data.aws_iam_policy_document.test.json
  home_directory = "/home/tftestuser"

  tags = {
    NAME  = "tftestuser"
    ENV   = "test"
    ADMIN = "test"
  }
}
`, rName))
}

func testAccAWSTransferUserConfig_modify(rName string) string {
	return composeConfig(testAccAWSTransferUserConfig_base, fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "tf-test-transfer-user-iam-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-test-transfer-user-iam-policy-%[1]s"
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoS3",
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

data "aws_iam_policy_document" "test" {
  statement {
    sid = "ListHomeDir"

    actions = [
      "s3:ListBucket",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::&{transfer:HomeBucket}",
    ]
  }

  statement {
    sid = "AWSTransferRequirements"

    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]

    resources = [
      "*",
    ]
  }

  statement {
    sid = "HomeDirObjectAccess"

    actions = [
      "s3:PutObject",
      "s3:GetObject",
      "s3:GetObjectVersion",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::&{transfer:HomeDirectory}*",
    ]
  }
}

resource "aws_transfer_user" "test" {
  server_id      = aws_transfer_server.test.id
  user_name      = "tftestuser"
  role           = aws_iam_role.test.arn
  policy         = data.aws_iam_policy_document.test.json
  home_directory = "/test"

  tags = {
    NAME = "tf-test-user"
    TEST = "test2"
  }
}
`, rName))
}

func testAccAWSTransferUserConfig_forceNew(rName string) string {
	return composeConfig(testAccAWSTransferUserConfig_base, fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "tf-test-transfer-user-iam-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-test-transfer-user-iam-policy-%[1]s"
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoS3",
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

data "aws_iam_policy_document" "test" {
  statement {
    sid = "ListHomeDir"

    actions = [
      "s3:ListBucket",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::&{transfer:HomeBucket}",
    ]
  }

  statement {
    sid = "AWSTransferRequirements"

    actions = [
      "s3:ListAllMyBuckets",
      "s3:GetBucketLocation",
    ]

    resources = [
      "*",
    ]
  }

  statement {
    sid = "HomeDirObjectAccess"

    actions = [
      "s3:PutObject",
      "s3:GetObject",
      "s3:DeleteObjectVersion",
      "s3:DeleteObject",
      "s3:GetObjectVersion",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::&{transfer:HomeDirectory}*",
    ]
  }
}

resource "aws_transfer_user" "test" {
  server_id      = aws_transfer_server.test.id
  user_name      = "tftestuser2"
  role           = aws_iam_role.test.arn
  policy         = data.aws_iam_policy_document.test.json
  home_directory = "/home/tftestuser2"

  tags = {
    NAME = "tf-test-user"
    TEST = "test2"
  }
}
`, rName))
}

func testAccAWSTransferUserConfig_homeDirectoryMappings(rName string) string {
	return composeConfig(
		testAccAWSTransferUserConfig_base,
		fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "tf-test-transfer-user-iam-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-test-transfer-user-iam-policy-%[1]s"
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoS3",
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_transfer_user" "test" {
  home_directory_type = "LOGICAL"
  role                = aws_iam_role.test.arn
  server_id           = aws_transfer_server.test.id
  user_name           = "tftestuser"

  home_directory_mappings {
    entry  = "/your-personal-report.pdf"
    target = "/bucket3/customized-reports/tftestuser.pdf"
  }
}
`, rName))
}

func testAccAWSTransferUserConfig_homeDirectoryMappingsUpdate(rName string) string {
	return composeConfig(
		testAccAWSTransferUserConfig_base,
		fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "tf-test-transfer-user-iam-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-test-transfer-user-iam-policy-%[1]s"
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoS3",
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_transfer_user" "test" {
  home_directory_type = "LOGICAL"
  role                = aws_iam_role.test.arn
  server_id           = aws_transfer_server.test.id
  user_name           = "tftestuser"

  home_directory_mappings {
    entry  = "/your-personal-report.pdf"
    target = "/bucket3/customized-reports/tftestuser.pdf"
  }

  home_directory_mappings {
    entry  = "/your-personal-report2.pdf"
    target = "/bucket3/customized-reports2/tftestuser.pdf"
  }
}
`, rName))
}

func testAccAWSTransferUserConfigPosix(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  domain = "EFS"
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = "tf-test-transfer-user-iam-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-test-transfer-user-iam-policy-%[1]s"
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoS3",
      "Effect": "Allow",
      "Action": [
        "efs:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_transfer_user" "test" {
  server_id = aws_transfer_server.test.id
  user_name = "tftestuser"
  role      = aws_iam_role.test.arn

  posix_profile {
    gid = 1000
    uid = 1000
  }
}
`, rName)
}

func testAccAWSTransferUserConfigPosixUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  domain = "EFS"
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = "tf-test-transfer-user-iam-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "tf-test-transfer-user-iam-policy-%[1]s"
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoS3",
      "Effect": "Allow",
      "Action": [
        "efs:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_transfer_user" "test" {
  server_id = aws_transfer_server.test.id
  user_name = "tftestuser"
  role      = aws_iam_role.test.arn

  posix_profile {
    gid            = 1001
    uid            = 1001
    secondary_gids = [1000, 1002]
  }
}
`, rName)
}
