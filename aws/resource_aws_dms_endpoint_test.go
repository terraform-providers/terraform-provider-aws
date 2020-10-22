package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsDmsEndpoint_basic(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := acctest.RandString(8) + "-basic"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dmsEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointBasicConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointBasicConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", "tf-test-dms-db-updated"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", "extra"),
					resource.TestCheckResourceAttr(resourceName, "password", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, "port", "3303"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, "username", "tftestupdate"),
				),
			},
		},
	})
}

func TestAccAwsDmsEndpoint_S3(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := acctest.RandString(8) + "-s3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dmsEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointS3Config(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.external_table_definition", ""),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.csv_row_delimiter", "\\n"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.csv_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.bucket_folder", ""),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.bucket_name", "bucket_name"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.compression_type", "NONE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointS3ConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", "key=value;"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.external_table_definition", "new-external_table_definition"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.csv_row_delimiter", "\\r"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.csv_delimiter", "."),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.bucket_folder", "new-bucket_folder"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.bucket_name", "new-bucket_name"),
					resource.TestCheckResourceAttr(resourceName, "s3_settings.0.compression_type", "GZIP"),
				),
			},
		},
	})
}

func TestAccAwsDmsEndpoint_DynamoDb(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := acctest.RandString(8) + "-dynamodb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dmsEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointDynamoDbConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointDynamoDbConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
				),
			},
		},
	})
}

func TestAccAwsDmsEndpoint_Elasticsearch(t *testing.T) {
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dmsEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointElasticsearchConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
					testAccCheckResourceAttrRegionalHostname(resourceName, "elasticsearch_settings.0.endpoint_uri", "es", "search-estest"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.full_load_error_percentage", "10"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.error_retry_duration", "300"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccAwsDmsEndpoint_Elasticsearch_ErrorRetryDuration(t *testing.T) {
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dmsEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointElasticsearchConfigErrorRetryDuration(rName, 60),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.error_retry_duration", "60"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			// Resource needs additional creation retry handling for the following:
			// Error creating DMS endpoint: ResourceAlreadyExistsFault: ReplicationEndpoint "xxx" already in use
			// {
			// 	Config: dmsEndpointElasticsearchConfigErrorRetryDuration(rName, 120),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		checkDmsEndpointExists(resourceName),
			// 		resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
			// 		resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.error_retry_duration", "120"),
			// 	),
			// },
		},
	})
}

func TestAccAwsDmsEndpoint_Elasticsearch_FullLoadErrorPercentage(t *testing.T) {
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dmsEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointElasticsearchConfigFullLoadErrorPercentage(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.full_load_error_percentage", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			// Resource needs additional creation retry handling for the following:
			// Error creating DMS endpoint: ResourceAlreadyExistsFault: ReplicationEndpoint "xxx" already in use
			// {
			// 	Config: dmsEndpointElasticsearchConfigFullLoadErrorPercentage(rName, 2),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		checkDmsEndpointExists(resourceName),
			// 		resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.#", "1"),
			// 		resource.TestCheckResourceAttr(resourceName, "elasticsearch_settings.0.full_load_error_percentage", "2"),
			// 	),
			// },
		},
	})
}

func TestAccAwsDmsEndpoint_Kafka_Broker(t *testing.T) {
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	brokerPrefix := "ec2-12-345-678-901"
	brokerService := "compute-1"
	brokerPort1 := 2345
	brokerPort2 := 3456

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dmsEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointKafkaConfigBroker(rName, brokerPrefix, brokerService, brokerPort1),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.#", "1"),
					testAccCheckResourceAttrHostnameWithPort(resourceName, "kafka_settings.0.broker", brokerService, brokerPrefix, brokerPort1),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.topic", "kafka-default-topic"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointKafkaConfigBroker(rName, brokerPrefix, brokerService, brokerPort2),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.#", "1"),
					testAccCheckResourceAttrHostnameWithPort(resourceName, "kafka_settings.0.broker", brokerService, brokerPrefix, brokerPort2),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.topic", "kafka-default-topic"),
				),
			},
		},
	})
}

func TestAccAwsDmsEndpoint_Kafka_Topic(t *testing.T) {
	resourceName := "aws_dms_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dmsEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointKafkaConfigTopic(rName, "topic1"),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.topic", "topic1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointKafkaConfigTopic(rName, "topic2"),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kafka_settings.0.topic", "topic2"),
				),
			},
		},
	})
}

func TestAccAwsDmsEndpoint_Kinesis(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := acctest.RandString(8) + "-kinesis"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dmsEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointKinesisConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.message_format", "json"),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_settings.0.stream_arn", "aws_kinesis_stream.stream1", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointKinesisConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kinesis_settings.0.message_format", "json"),
					resource.TestCheckResourceAttrPair(resourceName, "kinesis_settings.0.stream_arn", "aws_kinesis_stream.stream2", "arn"),
				),
			},
		},
	})
}

func TestAccAwsDmsEndpoint_MongoDb(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := acctest.RandString(8) + "-mongodb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dmsEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointMongoDbConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointMongoDbConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftest-new-server_name"),
					resource.TestCheckResourceAttr(resourceName, "port", "27018"),
					resource.TestCheckResourceAttr(resourceName, "username", "tftest-new-username"),
					resource.TestCheckResourceAttr(resourceName, "password", "tftest-new-password"),
					resource.TestCheckResourceAttr(resourceName, "database_name", "tftest-new-database_name"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "require"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", "key=value;"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.0.auth_mechanism", "scram-sha-1"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.0.nesting_level", "one"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.0.extract_doc_id", "true"),
					resource.TestCheckResourceAttr(resourceName, "mongodb_settings.0.docs_to_investigate", "1001"),
				),
			},
		},
	})
}

func TestAccAwsDmsEndpoint_DocDB(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := acctest.RandString(8) + "-docdb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dmsEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointDocDBConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointDocDBConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", "tf-test-dms-db-updated"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", "extra"),
					resource.TestCheckResourceAttr(resourceName, "password", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, "port", "27019"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, "username", "tftestupdate"),
				),
			},
		},
	})
}

func TestAccAwsDmsEndpoint_Db2(t *testing.T) {
	resourceName := "aws_dms_endpoint.dms_endpoint"
	randId := acctest.RandString(8) + "-db2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: dmsEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: dmsEndpointDb2Config(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: dmsEndpointDb2ConfigUpdate(randId),
				Check: resource.ComposeTestCheckFunc(
					checkDmsEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", "tf-test-dms-db-updated"),
					resource.TestCheckResourceAttr(resourceName, "extra_connection_attributes", "extra"),
					resource.TestCheckResourceAttr(resourceName, "password", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, "port", "27019"),
					resource.TestCheckResourceAttr(resourceName, "ssl_mode", "none"),
					resource.TestCheckResourceAttr(resourceName, "server_name", "tftestupdate"),
					resource.TestCheckResourceAttr(resourceName, "username", "tftestupdate"),
				),
			},
		},
	})
}

func dmsEndpointDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dms_endpoint" {
			continue
		}

		err := checkDmsEndpointExists(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Found an endpoint that was not destroyed: %s", rs.Primary.ID)
		}
	}

	return nil
}

func checkDmsEndpointExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).dmsconn
		resp, err := conn.DescribeEndpoints(&dms.DescribeEndpointsInput{
			Filters: []*dms.Filter{
				{
					Name:   aws.String("endpoint-id"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		})

		if err != nil {
			return fmt.Errorf("DMS endpoint error: %v", err)
		}

		if resp.Endpoints == nil {
			return fmt.Errorf("DMS endpoint not found")
		}

		return nil
	}
}

func dmsEndpointBasicConfig(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  database_name               = "tf-test-dms-db"
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "aurora"
  extra_connection_attributes = ""
  password                    = "tftest"
  port                        = 3306
  server_name                 = "tftest"
  ssl_mode                    = "none"

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }

  username = "tftest"
}
`, randId)
}

func dmsEndpointBasicConfigUpdate(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  database_name               = "tf-test-dms-db-updated"
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "aurora"
  extra_connection_attributes = "extra"
  password                    = "tftestupdate"
  port                        = 3303
  server_name                 = "tftestupdate"
  ssl_mode                    = "none"

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "updated"
    Add    = "added"
  }

  username = "tftestupdate"
}
`, randId)
}

func dmsEndpointDynamoDbConfig(randId string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id         = "tf-test-dms-endpoint-%[1]s"
  endpoint_type       = "target"
  engine_name         = "dynamodb"
  service_access_role = aws_iam_role.iam_role.arn
  ssl_mode            = "none"

  tags = {
    Name   = "tf-test-dynamodb-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }

  depends_on = [aws_iam_role_policy.dms_dynamodb_access]
}

resource "aws_iam_role" "iam_role" {
  name = "tf-test-iam-dynamodb-role-%[1]s"

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "dms.${data.aws_partition.current.dns_suffix}"
			},
			"Effect": "Allow"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "dms_dynamodb_access" {
  name = "tf-test-iam-dynamodb-role-policy-%[1]s"
  role = aws_iam_role.iam_role.name

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"dynamodb:PutItem",
				"dynamodb:CreateTable",
				"dynamodb:DescribeTable",
				"dynamodb:DeleteTable",
				"dynamodb:DeleteItem",
				"dynamodb:ListTables"
			],
			"Resource": "*"
		}
	]
}
EOF
}
`, randId)
}

func dmsEndpointDynamoDbConfigUpdate(randId string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id         = "tf-test-dms-endpoint-%[1]s"
  endpoint_type       = "target"
  engine_name         = "dynamodb"
  service_access_role = aws_iam_role.iam_role.arn
  ssl_mode            = "none"

  tags = {
    Name   = "tf-test-dynamodb-endpoint-%[1]s"
    Update = "updated"
    Add    = "added"
  }
}

resource "aws_iam_role" "iam_role" {
  name = "tf-test-iam-dynamodb-role-%[1]s"

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "dms.${data.aws_partition.current.dns_suffix}"
			},
			"Effect": "Allow"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "dms_dynamodb_access" {
  name = "tf-test-iam-dynamodb-role-policy-%[1]s"
  role = aws_iam_role.iam_role.name

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"dynamodb:PutItem",
				"dynamodb:CreateTable",
				"dynamodb:DescribeTable",
				"dynamodb:DeleteTable",
				"dynamodb:DeleteItem",
				"dynamodb:ListTables"
			],
			"Resource": "*"
		}
	]
}
EOF
}
`, randId)
}

func dmsEndpointS3Config(randId string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "target"
  engine_name                 = "s3"
  ssl_mode                    = "none"
  extra_connection_attributes = ""

  tags = {
    Name   = "tf-test-s3-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }

  s3_settings {
    service_access_role_arn = aws_iam_role.iam_role.arn
    bucket_name             = "bucket_name"
  }

  depends_on = [aws_iam_role_policy.dms_s3_access]
}

resource "aws_iam_role" "iam_role" {
  name = "tf-test-iam-s3-role-%[1]s"

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "dms.${data.aws_partition.current.dns_suffix}"
			},
			"Effect": "Allow"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "dms_s3_access" {
  name = "tf-test-iam-s3-role-policy-%[1]s"
  role = aws_iam_role.iam_role.name

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"s3:CreateBucket",
				"s3:ListBucket",
				"s3:DeleteBucket",
				"s3:GetBucketLocation",
				"s3:GetObject",
				"s3:PutObject",
				"s3:DeleteObject",
				"s3:GetObjectVersion",
				"s3:GetBucketPolicy",
				"s3:PutBucketPolicy",
				"s3:DeleteBucketPolicy"
			],
			"Resource": "*"
		}
	]
}
EOF
}
`, randId)
}

func dmsEndpointS3ConfigUpdate(randId string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "target"
  engine_name                 = "s3"
  ssl_mode                    = "none"
  extra_connection_attributes = "key=value;"

  tags = {
    Name   = "tf-test-s3-endpoint-%[1]s"
    Update = "updated"
    Add    = "added"
  }

  s3_settings {
    service_access_role_arn   = aws_iam_role.iam_role.arn
    external_table_definition = "new-external_table_definition"
    csv_row_delimiter         = "\\r"
    csv_delimiter             = "."
    bucket_folder             = "new-bucket_folder"
    bucket_name               = "new-bucket_name"
    compression_type          = "GZIP"
  }
}

resource "aws_iam_role" "iam_role" {
  name = "tf-test-iam-s3-role-%[1]s"

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "dms.${data.aws_partition.current.dns_suffix}"
			},
			"Effect": "Allow"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "dms_s3_access" {
  name = "tf-test-iam-s3-role-policy-%[1]s"
  role = aws_iam_role.iam_role.name

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"s3:CreateBucket",
				"s3:ListBucket",
				"s3:DeleteBucket",
				"s3:GetBucketLocation",
				"s3:GetObject",
				"s3:PutObject",
				"s3:DeleteObject",
				"s3:GetObjectVersion",
				"s3:GetBucketPolicy",
				"s3:PutBucketPolicy",
				"s3:DeleteBucketPolicy"
			],
			"Resource": "*"
		}
	]
}
EOF
}
`, randId)
}

func dmsEndpointElasticsearchConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "dms.${data.aws_partition.current.dns_suffix}"
			},
			"Effect": "Allow"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				 "es:ESHttpDelete",
				 "es:ESHttpGet",
				 "es:ESHttpHead",
				 "es:ESHttpPost",
				 "es:ESHttpPut"
			],
			"Resource": "*"
		}
	]
}
EOF
}
`, rName)
}

func dmsEndpointElasticsearchConfig(rName string) string {
	return composeConfig(
		dmsEndpointElasticsearchConfigBase(rName),
		fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "elasticsearch"

  elasticsearch_settings {
    endpoint_uri            = "search-estest.es.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
    service_access_role_arn = aws_iam_role.test.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func dmsEndpointElasticsearchConfigErrorRetryDuration(rName string, errorRetryDuration int) string {
	return composeConfig(
		dmsEndpointElasticsearchConfigBase(rName),
		fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "elasticsearch"

  elasticsearch_settings {
    endpoint_uri            = "search-estest.${data.aws_region.current.name}.es.${data.aws_partition.current.dns_suffix}"
    error_retry_duration    = %[2]d
    service_access_role_arn = aws_iam_role.test.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, errorRetryDuration))
}

func dmsEndpointElasticsearchConfigFullLoadErrorPercentage(rName string, fullLoadErrorPercentage int) string {
	return composeConfig(
		dmsEndpointElasticsearchConfigBase(rName),
		fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "elasticsearch"

  elasticsearch_settings {
    endpoint_uri               = "search-estest.${data.aws_region.current.name}.es.${data.aws_partition.current.dns_suffix}"
    full_load_error_percentage = %[2]d
    service_access_role_arn    = aws_iam_role.test.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, fullLoadErrorPercentage))
}

func dmsEndpointKafkaConfigBroker(rName, brokerPrefix, brokerServiceName string, brokerPort int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "kafka"

  kafka_settings {
    # example kafka broker: "ec2-12-345-678-901.compute-1.amazonaws.com:2345"
    broker = "%[2]s.%[3]s.${data.aws_partition.current.dns_suffix}:%[4]d"
  }
}
`, rName, brokerPrefix, brokerServiceName, brokerPort)
}

func dmsEndpointKafkaConfigTopic(rName string, topic string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  engine_name   = "kafka"

  kafka_settings {
    broker = "ec2-12-345-678-901.compute-1.${data.aws_partition.current.dns_suffix}:2345"
    topic  = %[2]q
  }
}
`, rName, topic)
}

func dmsEndpointKinesisConfig(randId string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id   = "tf-test-dms-endpoint-%[1]s"
  endpoint_type = "target"
  engine_name   = "kinesis"

  kinesis_settings {
    service_access_role_arn = aws_iam_role.iam_role.arn
    stream_arn              = aws_kinesis_stream.stream1.arn
  }

  depends_on = [aws_iam_role_policy.dms_kinesis_access]
}

resource "aws_kinesis_stream" "stream1" {
  name        = "tf-test-dms-kinesis-1-%[1]s"
  shard_count = 1
}

resource "aws_kinesis_stream" "stream2" {
  name        = "tf-test-dms-kinesis-2-%[1]s"
  shard_count = 1
}

resource "aws_iam_role" "iam_role" {
  name_prefix = "tf-test-iam-kinesis-role"

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "dms.${data.aws_partition.current.dns_suffix}"
			},
			"Effect": "Allow"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "dms_kinesis_access" {
  name_prefix = "tf-test-iam-kinesis-role-policy"
  role        = aws_iam_role.iam_role.name
  policy      = data.aws_iam_policy_document.dms_kinesis_access.json
}

data "aws_iam_policy_document" "dms_kinesis_access" {
  statement {
    actions = [
      "kinesis:DescribeStream",
      "kinesis:PutRecord",
      "kinesis:PutRecords",
    ]
    resources = [
      aws_kinesis_stream.stream1.arn,
      aws_kinesis_stream.stream2.arn,
    ]
  }
}
`, randId)
}

func dmsEndpointKinesisConfigUpdate(randId string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id   = "tf-test-dms-endpoint-%[1]s"
  endpoint_type = "target"
  engine_name   = "kinesis"

  kinesis_settings {
    service_access_role_arn = aws_iam_role.iam_role.arn
    stream_arn              = aws_kinesis_stream.stream2.arn
  }

  depends_on = [aws_iam_role_policy.dms_kinesis_access]
}

resource "aws_kinesis_stream" "stream1" {
  name        = "tf-test-dms-kinesis-1-%[1]s"
  shard_count = 1
}

resource "aws_kinesis_stream" "stream2" {
  name        = "tf-test-dms-kinesis-2-%[1]s"
  shard_count = 1
}

resource "aws_iam_role" "iam_role" {
  name_prefix = "tf-test-iam-kinesis-role"

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "dms.${data.aws_partition.current.dns_suffix}"
			},
			"Effect": "Allow"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "dms_kinesis_access" {
  name_prefix = "tf-test-iam-kinesis-role-policy"
  role        = aws_iam_role.iam_role.name
  policy      = data.aws_iam_policy_document.dms_kinesis_access.json
}

data "aws_iam_policy_document" "dms_kinesis_access" {
  statement {
    actions = [
      "kinesis:DescribeStream",
      "kinesis:PutRecord",
      "kinesis:PutRecords",
    ]
    resources = [
      aws_kinesis_stream.stream1.arn,
      aws_kinesis_stream.stream2.arn,
    ]
  }
}
`, randId)
}

func dmsEndpointMongoDbConfig(randId string) string {
	return fmt.Sprintf(`
data "aws_kms_alias" "dms" {
  name = "alias/aws/dms"
}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "mongodb"
  server_name                 = "tftest"
  port                        = 27017
  username                    = "tftest"
  password                    = "tftest"
  database_name               = "tftest"
  ssl_mode                    = "none"
  extra_connection_attributes = ""
  kms_key_arn                 = data.aws_kms_alias.dms.target_key_arn

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }

  mongodb_settings {
    auth_type           = "password"
    auth_mechanism      = "default"
    nesting_level       = "none"
    extract_doc_id      = "false"
    docs_to_investigate = "1000"
    auth_source         = "admin"
  }
}
`, randId)
}

func dmsEndpointMongoDbConfigUpdate(randId string) string {
	return fmt.Sprintf(`
data "aws_kms_alias" "dms" {
  name = "alias/aws/dms"
}

resource "aws_dms_endpoint" "dms_endpoint" {
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "mongodb"
  server_name                 = "tftest-new-server_name"
  port                        = 27018
  username                    = "tftest-new-username"
  password                    = "tftest-new-password"
  database_name               = "tftest-new-database_name"
  ssl_mode                    = "require"
  extra_connection_attributes = "key=value;"
  kms_key_arn                 = data.aws_kms_alias.dms.target_key_arn

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "updated"
    Add    = "added"
  }

  mongodb_settings {
    auth_mechanism      = "scram-sha-1"
    nesting_level       = "one"
    extract_doc_id      = "true"
    docs_to_investigate = "1001"
  }
}
`, randId)
}

func dmsEndpointDocDBConfig(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  database_name               = "tf-test-dms-db"
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "target"
  engine_name                 = "docdb"
  extra_connection_attributes = ""
  password                    = "tftest"
  port                        = 27017
  server_name                 = "tftest"
  ssl_mode                    = "none"

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }

  username = "tftest"
}
`, randId)
}

func dmsEndpointDocDBConfigUpdate(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  database_name               = "tf-test-dms-db-updated"
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "target"
  engine_name                 = "docdb"
  extra_connection_attributes = "extra"
  password                    = "tftestupdate"
  port                        = 27019
  server_name                 = "tftestupdate"
  ssl_mode                    = "none"

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "updated"
    Add    = "added"
  }

  username = "tftestupdate"
}
`, randId)
}

func dmsEndpointDb2Config(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  database_name               = "tf-test-dms-db"
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "db2"
  extra_connection_attributes = ""
  password                    = "tftest"
  port                        = 27017
  server_name                 = "tftest"
  ssl_mode                    = "none"

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "to-update"
    Remove = "to-remove"
  }

  username = "tftest"
}
`, randId)
}

func dmsEndpointDb2ConfigUpdate(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "dms_endpoint" {
  database_name               = "tf-test-dms-db-updated"
  endpoint_id                 = "tf-test-dms-endpoint-%[1]s"
  endpoint_type               = "source"
  engine_name                 = "db2"
  extra_connection_attributes = "extra"
  password                    = "tftestupdate"
  port                        = 27019
  server_name                 = "tftestupdate"
  ssl_mode                    = "none"

  tags = {
    Name   = "tf-test-dms-endpoint-%[1]s"
    Update = "updated"
    Add    = "added"
  }

  username = "tftestupdate"
}
`, randId)
}
