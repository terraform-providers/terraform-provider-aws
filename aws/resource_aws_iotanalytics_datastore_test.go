package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSIoTAnalyticsDatastore_basic(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsDatastoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsDatastore_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsDatastoreExists_basic("aws_iotanalytics_datastore.datastore"),
					resource.TestCheckResourceAttr("aws_iotanalytics_datastore.datastore", "name", fmt.Sprintf("test_datastore_%s", rString)),
					resource.TestCheckResourceAttr("aws_iotanalytics_datastore.datastore", "tags.tagKey", "tagValue"),
				),
			},
		},
	})
}

func TestAccAWSIoTAnalyticsDatastore_CustomerManagedS3(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsDatastoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsDatastore_CustomerManagedS3(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsDatastoreExists_basic("aws_iotanalytics_datastore.datastore"),
					resource.TestCheckResourceAttr("aws_iotanalytics_datastore.datastore", "name", fmt.Sprintf("test_datastore_%s", rString)),
					testAccCheckAWSIoTAnalyticsDatastore_CustomerManagedS3,
				),
			},
		},
	})
}

func TestAccAWSIoTAnalyticsDatastore_RetentionPeriodNumberOfDays(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsDatastoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsDatastore_RetentionPeriodNumberOfDays(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsDatastoreExists_basic("aws_iotanalytics_datastore.datastore"),
					resource.TestCheckResourceAttr("aws_iotanalytics_datastore.datastore", "name", fmt.Sprintf("test_datastore_%s", rString)),
					testAccCheckAWSIoTAnalyticsDatastore_RetentionPeriodNumberOfDays,
				),
			},
		},
	})
}

func TestAccAWSIoTAnalyticsDatastore_RetentionPeriodUnlimited(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsDatastoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsDatastore_RetentionPeriodUnlimited(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsDatastoreExists_basic("aws_iotanalytics_datastore.datastore"),
					resource.TestCheckResourceAttr("aws_iotanalytics_datastore.datastore", "name", fmt.Sprintf("test_datastore_%s", rString)),
					testAccCheckAWSIoTAnalyticsDatastore_RetentionPeriodUnlimited,
				),
			},
		},
	})
}

func testAccCheckAWSIoTAnalyticsDatastoreDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotanalytics_datastore" {
			continue
		}

		params := &iotanalytics.DescribeDatastoreInput{
			DatastoreName: aws.String(rs.Primary.ID),
		}
		_, err := conn.DescribeDatastore(params)

		if err != nil {
			if isAWSErr(err, iotanalytics.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
		return fmt.Errorf("Expected IoTAnalytics Datastore to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSIoTAnalyticsDatastoreExists_basic(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccCheckAWSIoTAnalyticsDatastore_CustomerManagedS3(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotanalytics_datastore" {
			continue
		}

		params := &iotanalytics.DescribeDatastoreInput{
			DatastoreName: aws.String(rs.Primary.ID),
		}
		out, err := conn.DescribeDatastore(params)

		if err != nil {
			return err
		}

		customerS3 := out.Datastore.Storage.CustomerManagedS3
		if customerS3 == nil {
			return fmt.Errorf("Datastore CustomerManagedS3 is equal nil")
		}

		expectedBucket := "fakebucket"
		expectedKeyPrefix := "fakeprefix/"
		if *customerS3.Bucket != expectedBucket {
			return fmt.Errorf("CustomerManagedDatastoreS3 Bucket %s is not equal expected bucket %s", *customerS3.Bucket, expectedBucket)
		}

		if *customerS3.KeyPrefix != expectedKeyPrefix {
			return fmt.Errorf("CustomerManagedDatastoreS3 KeyPrefix %s is not equal expected key prefix %s", *customerS3.KeyPrefix, expectedKeyPrefix)
		}
	}

	return nil
}

func testAccCheckAWSIoTAnalyticsDatastore_RetentionPeriodNumberOfDays(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotanalytics_datastore" {
			continue
		}

		params := &iotanalytics.DescribeDatastoreInput{
			DatastoreName: aws.String(rs.Primary.ID),
		}
		out, err := conn.DescribeDatastore(params)

		if err != nil {
			return err
		}

		retentionPeriod := out.Datastore.RetentionPeriod
		if retentionPeriod == nil {
			return fmt.Errorf("Datastore RetentionPeriod is equal nil")
		}

		expectedNumberOfDays := int64(2)
		expectedUnlimited := false
		if *retentionPeriod.NumberOfDays != expectedNumberOfDays {
			return fmt.Errorf("RetentionPeriod NumberOfDays %d is not equal expected number of days %d", *retentionPeriod.NumberOfDays, expectedNumberOfDays)
		}

		if *retentionPeriod.Unlimited != expectedUnlimited {
			return fmt.Errorf("RetentionPeriod Unlimited %t is not equal expected unlimited %t", *retentionPeriod.Unlimited, expectedUnlimited)
		}
	}

	return nil
}

func testAccCheckAWSIoTAnalyticsDatastore_RetentionPeriodUnlimited(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotanalytics_datastore" {
			continue
		}

		params := &iotanalytics.DescribeDatastoreInput{
			DatastoreName: aws.String(rs.Primary.ID),
		}
		out, err := conn.DescribeDatastore(params)

		if err != nil {
			return err
		}

		retentionPeriod := out.Datastore.RetentionPeriod
		if retentionPeriod == nil {
			return fmt.Errorf("Datastore RetentionPeriod is equal nil")
		}

		expectedUnlimited := true
		if retentionPeriod.NumberOfDays != nil {
			return fmt.Errorf("RetentionPeriod NumberOfDays %d is not nil", *retentionPeriod.NumberOfDays)
		}

		if *retentionPeriod.Unlimited != expectedUnlimited {
			return fmt.Errorf("RetentionPeriod Unlimited %t is not equal expected unlimited %t", *retentionPeriod.Unlimited, expectedUnlimited)
		}
	}
	return nil
}

func testAccAWSIoTAnalyticsDatastore_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_iotanalytics_datastore" "datastore" {
  name = "test_datastore_%[1]s"

  tags = {
	"tagKey" = "tagValue",
  }

  storage {
	  service_managed_s3 {}
  }

  retention_period {
	  unlimited = true
  }
}
`, rString)
}

func testAccAWSIoTAnalyticsDatastore_CustomerManagedS3(rString string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iotanalytics_role" {
    name = "test_role_%[1]s"
    assume_role_policy = <<EOF
{
    "Version":"2012-10-17",
    "Statement":[{
        "Effect": "Allow",
        "Principal": {
            "Service": "iotanalytics.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
    }]
}
EOF
}

resource "aws_iam_policy" "policy" {
    name = "test_policy_%[1]s"
    path = "/"
    description = "My test policy"
    policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Action": "*",
        "Resource": "*"
    }]
}
EOF
}

resource "aws_iam_policy_attachment" "attach_policy" {
    name = "test_policy_attachment_%[1]s"
    roles = ["${aws_iam_role.iotanalytics_role.name}"]
    policy_arn = "${aws_iam_policy.policy.arn}"
}
	
		
resource "aws_iotanalytics_datastore" "datastore" {
  name = "test_datastore_%[1]s"

  storage {
	customer_managed_s3 {
		bucket = "fakebucket"
		key_prefix = "fakeprefix/"
		role_arn = "${aws_iam_role.iotanalytics_role.arn}"
	}
  }
}
`, rString)
}

func testAccAWSIoTAnalyticsDatastore_RetentionPeriodNumberOfDays(rString string) string {
	return fmt.Sprintf(`
resource "aws_iotanalytics_datastore" "datastore" {
  name = "test_datastore_%[1]s"

  storage {
	  service_managed_s3 {}
  }
  retention_period {
	number_of_days = 2
  }
}
`, rString)
}

func testAccAWSIoTAnalyticsDatastore_RetentionPeriodUnlimited(rString string) string {
	return fmt.Sprintf(`
resource "aws_iotanalytics_datastore" "datastore" {
  name = "test_datastore_%[1]s"

  storage {
	  service_managed_s3 {}
  }
  retention_period {
	unlimited = true
  }
}
`, rString)
}
