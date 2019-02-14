package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSMqConfiguration_basic(t *testing.T) {
	configurationName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqConfigurationConfig(configurationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqConfigurationExists("aws_mq_configuration.test"),
					resource.TestCheckResourceAttrSet("aws_mq_configuration.test", "arn"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "description", "TfAccTest MQ Configuration"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "engine_version", "5.15.0"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "latest_revision", "2"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "name", configurationName),
				),
			},
			{
				Config: testAccMqConfigurationConfig_descriptionUpdated(configurationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqConfigurationExists("aws_mq_configuration.test"),
					resource.TestCheckResourceAttrSet("aws_mq_configuration.test", "arn"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "description", "TfAccTest MQ Configuration Updated"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "engine_version", "5.15.0"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "latest_revision", "3"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "name", configurationName),
				),
			},
		},
	})
}

func TestAccAWSMqConfiguration_withData(t *testing.T) {
	configurationName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqConfigurationWithDataConfig(configurationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqConfigurationExists("aws_mq_configuration.test"),
					resource.TestCheckResourceAttrSet("aws_mq_configuration.test", "arn"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "description", "TfAccTest MQ Configuration"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "engine_version", "5.15.0"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "latest_revision", "2"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "name", configurationName),
				),
			},
		},
	})
}

func TestAccAWSMqConfiguration_updateTags(t *testing.T) {
	configurationName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqConfigurationConfig_updateTags1(configurationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqConfigurationExists("aws_mq_configuration.test"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "tags.env", "test"),
				),
			},
			{
				Config: testAccMqConfigurationConfig_updateTags2(configurationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqConfigurationExists("aws_mq_configuration.test"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "tags.env", "test2"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "tags.role", "test-role"),
				),
			},
			{
				Config: testAccMqConfigurationConfig_updateTags3(configurationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqConfigurationExists("aws_mq_configuration.test"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_mq_configuration.test", "tags.role", "test-role"),
				),
			},
		},
	})
}

func testAccCheckAwsMqConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).mqconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_mq_configuration" {
			continue
		}

		input := &mq.DescribeConfigurationInput{
			ConfigurationId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeConfiguration(input)
		if err != nil {
			if isAWSErr(err, "NotFoundException", "") {
				return nil
			}
			return err
		}

		// TODO: Delete is not available in the API
		return nil
		//return fmt.Errorf("Expected MQ configuration to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsMqConfigurationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccMqConfigurationConfig(configurationName string) string {
	return fmt.Sprintf(`
resource "aws_mq_configuration" "test" {
  description = "TfAccTest MQ Configuration"
  name = "%s"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>
DATA
}`, configurationName)
}

func testAccMqConfigurationConfig_descriptionUpdated(configurationName string) string {
	return fmt.Sprintf(`
resource "aws_mq_configuration" "test" {
  description = "TfAccTest MQ Configuration Updated"
  name = "%s"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>
DATA
}`, configurationName)
}

func testAccMqConfigurationWithDataConfig(configurationName string) string {
	return fmt.Sprintf(`
resource "aws_mq_configuration" "test" {
  description = "TfAccTest MQ Configuration"
  name = "%s"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
  <plugins>
    <authorizationPlugin>
      <map>
        <authorizationMap>
          <authorizationEntries>
            <authorizationEntry admin="guests,users" queue="GUEST.&gt;" read="guests" write="guests,users"/>
            <authorizationEntry admin="guests,users" read="guests,users" topic="ActiveMQ.Advisory.&gt;" write="guests,users"/>
          </authorizationEntries>
          <tempDestinationAuthorizationEntry>
            <tempDestinationAuthorizationEntry admin="tempDestinationAdmins" read="tempDestinationAdmins" write="tempDestinationAdmins"/>
          </tempDestinationAuthorizationEntry>
        </authorizationMap>
      </map>
    </authorizationPlugin>
    <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
    <statisticsBrokerPlugin/>
    <timeStampingBrokerPlugin ttlCeiling="86400000" zeroExpirationOverride="86400000"/>
  </plugins>
</broker>
DATA
}`, configurationName)
}

func testAccMqConfigurationConfig_updateTags1(configurationName string) string {
	return fmt.Sprintf(`
resource "aws_mq_configuration" "test" {
  description = "TfAccTest MQ Configuration"
  name = "%s"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>
DATA

	tags {
		env = "test"
	}
}`, configurationName)
}

func testAccMqConfigurationConfig_updateTags2(configurationName string) string {
	return fmt.Sprintf(`
resource "aws_mq_configuration" "test" {
  description = "TfAccTest MQ Configuration"
  name = "%s"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>
DATA

	tags {
		env = "test2"
		role = "test-role"
	}
}`, configurationName)
}

func testAccMqConfigurationConfig_updateTags3(configurationName string) string {
	return fmt.Sprintf(`
resource "aws_mq_configuration" "test" {
  description = "TfAccTest MQ Configuration"
  name = "%s"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>
DATA

	tags {
		role = "test-role"
	}
}`, configurationName)
}
