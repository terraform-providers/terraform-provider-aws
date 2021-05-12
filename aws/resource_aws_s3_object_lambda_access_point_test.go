package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	awspolicy "github.com/jen20/awspolicyequivalence"
)

func init() {
	resource.AddTestSweepers("aws_s3_object_lambda_access_point", &resource.Sweeper{
		Name: "aws_s3_object_lambda_access_point",
		F:    testSweepS3ObjectLambdaAccessPoints,
	})
}

func testSweepS3ObjectLambdaAccessPoints(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	accountId := client.(*AWSClient).accountid
	conn := client.(*AWSClient).s3controlconn

	input := &s3control.ListAccessPointsForObjectLambdaInput{
		AccountId: aws.String(accountId),
	}
	var sweeperErrs *multierror.Error

	conn.ListAccessPointsForObjectLambdaPages(input, func(page *s3control.ListAccessPointsForObjectLambdaOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, ObjectLambdaAccessPoint := range page.ObjectLambdaAccessPointList {
			input := &s3control.DeleteAccessPointForObjectLambdaInput{
				AccountId: aws.String(accountId),
				Name:      ObjectLambdaAccessPoint.Name,
			}
			name := aws.StringValue(ObjectLambdaAccessPoint.Name)

			log.Printf("[INFO] Deleting S3 Object Lambda Access Point: %s", name)
			_, err := conn.DeleteAccessPointForObjectLambda(input)

			if isAWSErr(err, "NoSuchAccessPoint", "") {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting S3 Object Lambda Access Point (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping S3 Object Lambda Access Point sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing S3 Object Lambda Access Points: %w", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSS3ObjectLambdaAccessPoint_basic(t *testing.T) {
	var v s3control.GetAccessPointForObjectLambdaOutput
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	accessPointName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_object_lambda_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3control.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3ObjectLambdaAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3ObjectLambdaAccessPointConfig_basic(bucketName, accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ObjectLambdaAccessPointExists(resourceName, &v),
					testAccCheckResourceAttrAccountID(resourceName, "account_id"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "s3", fmt.Sprintf("accesspoint/%s", accessPointName)),
					resource.TestCheckResourceAttr(resourceName, "bucket", bucketName),
					testAccMatchResourceAttrRegionalHostname(resourceName, "domain_name", "s3-accesspoint", regexp.MustCompile(fmt.Sprintf("^%s-\\d{12}", accessPointName))),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", accessPointName),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Internet"),
					resource.TestCheckResourceAttr(resourceName, "policy", ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", "true"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
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

func TestAccAWSS3ObjectLambdaAccessPoint_disappears(t *testing.T) {
	var v s3control.GetAccessPointForObjectLambdaOutput
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	accessPointName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_object_lambda_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3control.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3ObjectLambdaAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3ObjectLambdaAccessPointConfig_basic(bucketName, accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ObjectLambdaAccessPointExists(resourceName, &v),
					testAccCheckAWSS3ObjectLambdaAccessPointDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSS3ObjectLambdaAccessPoint_disappears_Bucket(t *testing.T) {
	var v s3control.GetAccessPointForObjectLambdaOutput
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	accessPointName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_object_lambda_access_point.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3control.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3ObjectLambdaAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3ObjectLambdaAccessPointConfig_basic(bucketName, accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ObjectLambdaAccessPointExists(resourceName, &v),
					testAccCheckAWSS3DestroyBucket(bucketResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSS3ObjectLambdaAccessPoint_Bucket_Arn(t *testing.T) {
	var v s3control.GetAccessPointForObjectLambdaOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_object_lambda_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		ErrorCheck:   testAccErrorCheck(t, s3control.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3ObjectLambdaAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3ObjectLambdaAccessPointConfig_Bucket_Arn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ObjectLambdaAccessPointExists(resourceName, &v),
					testAccCheckResourceAttrAccountID(resourceName, "account_id"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "s3-outposts", fmt.Sprintf("outpost/[^/]+/accesspoint/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", "aws_s3control_bucket.test", "arn"),
					testAccMatchResourceAttrRegionalHostname(resourceName, "domain_name", "s3-accesspoint", regexp.MustCompile(fmt.Sprintf("^%s-\\d{12}", rName))),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Vpc"),
					resource.TestCheckResourceAttr(resourceName, "policy", ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", "true"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_configuration.0.vpc_id", "aws_vpc.test", "id"),
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

func TestAccAWSS3ObjectLambdaAccessPoint_Policy(t *testing.T) {
	var v s3control.GetAccessPointForObjectLambdaOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_object_lambda_access_point.test"

	expectedPolicyText1 := func() string {
		return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "s3:GetObjectTagging",
      "Resource": [
        "arn:%s:s3:%s:%s:accesspoint/%s/object/*"
      ]
    }
  ]
}`, testAccGetPartition(), testAccGetRegion(), testAccGetAccountID(), rName)
	}
	expectedPolicyText2 := func() string {
		return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": [
        "s3:GetObjectLegalHold",
        "s3:GetObjectRetention"
      ],
      "Resource": [
        "arn:%s:s3:%s:%s:accesspoint/%s/object/*"
      ]
    }
  ]
}`, testAccGetPartition(), testAccGetRegion(), testAccGetAccountID(), rName)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3control.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3ObjectLambdaAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3ObjectLambdaAccessPointConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ObjectLambdaAccessPointExists(resourceName, &v),
					testAccCheckAWSS3ObjectLambdaAccessPointHasPolicy(resourceName, expectedPolicyText1),
					testAccCheckResourceAttrAccountID(resourceName, "account_id"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "s3", fmt.Sprintf("accesspoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Internet"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", "false"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", "false"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSS3ObjectLambdaAccessPointConfig_policyUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ObjectLambdaAccessPointExists(resourceName, &v),
					testAccCheckAWSS3ObjectLambdaAccessPointHasPolicy(resourceName, expectedPolicyText2),
				),
			},
			{
				Config: testAccAWSS3ObjectLambdaAccessPointConfig_noPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ObjectLambdaAccessPointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy", ""),
				),
			},
		},
	})
}

func TestAccAWSS3ObjectLambdaAccessPoint_PublicAccessBlockConfiguration(t *testing.T) {
	var v s3control.GetAccessPointForObjectLambdaOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_object_lambda_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3control.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3ObjectLambdaAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3ObjectLambdaAccessPointConfig_publicAccessBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ObjectLambdaAccessPointExists(resourceName, &v),
					testAccCheckResourceAttrAccountID(resourceName, "account_id"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "s3", fmt.Sprintf("accesspoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "Internet"),
					resource.TestCheckResourceAttr(resourceName, "policy", ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", "false"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", "false"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", "false"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", "false"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "0"),
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

func TestAccAWSS3ObjectLambdaAccessPoint_VpcConfiguration(t *testing.T) {
	var v s3control.GetAccessPointForObjectLambdaOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_object_lambda_access_point.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, s3control.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3ObjectLambdaAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3ObjectLambdaAccessPointConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ObjectLambdaAccessPointExists(resourceName, &v),
					testAccCheckResourceAttrAccountID(resourceName, "account_id"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "s3", fmt.Sprintf("accesspoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "has_public_access_policy", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "network_origin", "VPC"),
					resource.TestCheckResourceAttr(resourceName, "policy", ""),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_acls", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.block_public_policy", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.ignore_public_acls", "true"),
					resource.TestCheckResourceAttr(resourceName, "public_access_block_configuration.0.restrict_public_buckets", "true"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_configuration.0.vpc_id", vpcResourceName, "id"),
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

func testAccCheckAWSS3ObjectLambdaAccessPointDisappears(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Access Point ID is set")
		}

		accountId, name, err := s3AccessPointParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).s3controlconn

		_, err = conn.DeleteAccessPoint(&s3control.DeleteAccessPointInput{
			AccountId: aws.String(accountId),
			Name:      aws.String(name),
		})
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckAWSS3ObjectLambdaAccessPointDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).s3controlconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_object_lambda_access_point" {
			continue
		}

		accountId, name, err := s3AccessPointParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.GetAccessPointForObjectLambda(&s3control.GetAccessPointForObjectLambdaInput{
			AccountId: aws.String(accountId),
			Name:      aws.String(name),
		})
		if err == nil {
			return fmt.Errorf("S3 Access Point still exists")
		}
	}
	return nil
}

func testAccCheckAWSS3ObjectLambdaAccessPointExists(n string, output *s3control.GetAccessPointForObjectLambdaOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Access Point ID is set")
		}

		accountId, name, err := s3AccessPointParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).s3controlconn

		resp, err := conn.GetAccessPointForObjectLambda(&s3control.GetAccessPointForObjectLambdaInput{
			AccountId: aws.String(accountId),
			Name:      aws.String(name),
		})
		if err != nil {
			return err
		}

		*output = *resp

		return nil
	}
}

func testAccCheckAWSS3ObjectLambdaAccessPointHasPolicy(n string, fn func() string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Access Point ID is set")
		}

		accountId, name, err := s3AccessPointParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).s3controlconn

		resp, err := conn.GetAccessPointPolicyForObjectLambda(&s3control.GetAccessPointForObjectLambdaPolicyInput{
			AccountId: aws.String(accountId),
			Name:      aws.String(name),
		})
		if err != nil {
			return err
		}

		actualPolicyText := *resp.Policy
		expectedPolicyText := fn()

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccAWSS3ObjectLambdaAccessPointConfig_basic(bucketName, accessPointName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object_lambda_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[2]q
}
`, bucketName, accessPointName)
}

func testAccAWSS3ObjectLambdaAccessPointConfig_Bucket_Arn(rName string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_s3control_bucket" "test" {
  bucket     = %[1]q
  outpost_id = data.aws_outposts_outpost.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_s3_object_lambda_access_point" "test" {
  bucket = aws_s3control_bucket.test.arn
  name   = %[1]q

  vpc_configuration {
    vpc_id = aws_vpc.test.id
  }
}
`, rName)
}

func testAccAWSS3ObjectLambdaAccessPointConfig_policy(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object_lambda_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q
  policy = data.aws_iam_policy_document.test.json

  public_access_block_configuration {
    block_public_acls       = true
    block_public_policy     = false
    ignore_public_acls      = true
    restrict_public_buckets = false
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    actions = [
      "s3:GetObjectTagging",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:accesspoint/%[1]s/object/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }
}
`, rName)
}

func testAccAWSS3ObjectLambdaAccessPointConfig_policyUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object_lambda_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q
  policy = data.aws_iam_policy_document.test.json

  public_access_block_configuration {
    block_public_acls       = true
    block_public_policy     = false
    ignore_public_acls      = true
    restrict_public_buckets = false
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    actions = [
      "s3:GetObjectLegalHold",
      "s3:GetObjectRetention"
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:accesspoint/%[1]s/object/*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
  }
}
`, rName)
}

func testAccAWSS3ObjectLambdaAccessPointConfig_noPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object_lambda_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  public_access_block_configuration {
    block_public_acls       = true
    block_public_policy     = false
    ignore_public_acls      = true
    restrict_public_buckets = false
  }
}
`, rName)
}

func testAccAWSS3ObjectLambdaAccessPointConfig_publicAccessBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object_lambda_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  public_access_block_configuration {
    block_public_acls       = false
    block_public_policy     = false
    ignore_public_acls      = false
    restrict_public_buckets = false
  }
}
`, rName)
}

func testAccAWSS3ObjectLambdaAccessPointConfig_vpc(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object_lambda_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  vpc_configuration {
    vpc_id = aws_vpc.test.id
  }
}
`, rName)
}
