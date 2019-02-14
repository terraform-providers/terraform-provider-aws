package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_mq_broker", &resource.Sweeper{
		Name: "aws_mq_broker",
		F:    testSweepMqBrokers,
	})
}

func TestResourceAWSMqBrokerPasswordValidation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "123456789012",
			ErrCount: 0,
		},
		{
			Value:    "12345678901",
			ErrCount: 1,
		},
		{
			Value:    "1234567890" + strings.Repeat("#", 240),
			ErrCount: 0,
		},
		{
			Value:    "1234567890" + strings.Repeat("#", 241),
			ErrCount: 1,
		},
		{
			Value:    "123" + strings.Repeat("#", 9),
			ErrCount: 0,
		},
		{
			Value:    "12" + strings.Repeat("#", 10),
			ErrCount: 1,
		},
		{
			Value:    "12345678901,",
			ErrCount: 1,
		},
		{
			Value:    "1," + strings.Repeat("#", 9),
			ErrCount: 3,
		},
	}

	for _, tc := range cases {
		_, errors := validateMqBrokerPassword(tc.Value, "aws_mq_broker_user_password")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected errors %d for %s while returned errors %d", tc.ErrCount, tc.Value, len(errors))
		}
	}
}

func testSweepMqBrokers(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).mqconn

	resp, err := conn.ListBrokers(&mq.ListBrokersInput{
		MaxResults: aws.Int64(100),
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping MQ Broker sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing MQ brokers: %s", err)
	}

	if len(resp.BrokerSummaries) == 0 {
		log.Print("[DEBUG] No MQ brokers found to sweep")
		return nil
	}
	log.Printf("[DEBUG] %d MQ brokers found", len(resp.BrokerSummaries))

	for _, bs := range resp.BrokerSummaries {
		if !strings.HasPrefix(*bs.BrokerName, "tf-acc-test-") {
			continue
		}

		log.Printf("[INFO] Deleting MQ broker %s", *bs.BrokerId)
		_, err := conn.DeleteBroker(&mq.DeleteBrokerInput{
			BrokerId: bs.BrokerId,
		})
		if err != nil {
			return err
		}
		err = waitForMqBrokerDeletion(conn, *bs.BrokerId)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestDiffAwsMqBrokerUsers(t *testing.T) {
	testCases := []struct {
		OldUsers []interface{}
		NewUsers []interface{}

		Creations []*mq.CreateUserRequest
		Deletions []*mq.DeleteUserInput
		Updates   []*mq.UpdateUserRequest
	}{
		{
			OldUsers: []interface{}{},
			NewUsers: []interface{}{
				map[string]interface{}{
					"console_access": false,
					"username":       "second",
					"password":       "TestTest2222",
					"groups":         schema.NewSet(schema.HashString, []interface{}{"admin"}),
				},
			},
			Creations: []*mq.CreateUserRequest{
				{
					BrokerId:      aws.String("test"),
					ConsoleAccess: aws.Bool(false),
					Username:      aws.String("second"),
					Password:      aws.String("TestTest2222"),
					Groups:        aws.StringSlice([]string{"admin"}),
				},
			},
			Deletions: []*mq.DeleteUserInput{},
			Updates:   []*mq.UpdateUserRequest{},
		},
		{
			OldUsers: []interface{}{
				map[string]interface{}{
					"console_access": true,
					"username":       "first",
					"password":       "TestTest1111",
				},
			},
			NewUsers: []interface{}{
				map[string]interface{}{
					"console_access": false,
					"username":       "second",
					"password":       "TestTest2222",
				},
			},
			Creations: []*mq.CreateUserRequest{
				{
					BrokerId:      aws.String("test"),
					ConsoleAccess: aws.Bool(false),
					Username:      aws.String("second"),
					Password:      aws.String("TestTest2222"),
				},
			},
			Deletions: []*mq.DeleteUserInput{
				{BrokerId: aws.String("test"), Username: aws.String("first")},
			},
			Updates: []*mq.UpdateUserRequest{},
		},
		{
			OldUsers: []interface{}{
				map[string]interface{}{
					"console_access": true,
					"username":       "first",
					"password":       "TestTest1111updated",
				},
				map[string]interface{}{
					"console_access": false,
					"username":       "second",
					"password":       "TestTest2222",
				},
			},
			NewUsers: []interface{}{
				map[string]interface{}{
					"console_access": false,
					"username":       "second",
					"password":       "TestTest2222",
					"groups":         schema.NewSet(schema.HashString, []interface{}{"admin"}),
				},
			},
			Creations: []*mq.CreateUserRequest{},
			Deletions: []*mq.DeleteUserInput{
				{BrokerId: aws.String("test"), Username: aws.String("first")},
			},
			Updates: []*mq.UpdateUserRequest{
				{
					BrokerId:      aws.String("test"),
					ConsoleAccess: aws.Bool(false),
					Username:      aws.String("second"),
					Password:      aws.String("TestTest2222"),
					Groups:        aws.StringSlice([]string{"admin"}),
				},
			},
		},
	}

	for _, tc := range testCases {
		creations, deletions, updates, err := diffAwsMqBrokerUsers("test", tc.OldUsers, tc.NewUsers)
		if err != nil {
			t.Fatal(err)
		}

		expectedCreations := fmt.Sprintf("%s", tc.Creations)
		creationsString := fmt.Sprintf("%s", creations)
		if creationsString != expectedCreations {
			t.Fatalf("Expected creations: %s\nGiven: %s", expectedCreations, creationsString)
		}

		expectedDeletions := fmt.Sprintf("%s", tc.Deletions)
		deletionsString := fmt.Sprintf("%s", deletions)
		if deletionsString != expectedDeletions {
			t.Fatalf("Expected deletions: %s\nGiven: %s", expectedDeletions, deletionsString)
		}

		expectedUpdates := fmt.Sprintf("%s", tc.Updates)
		updatesString := fmt.Sprintf("%s", updates)
		if updatesString != expectedUpdates {
			t.Fatalf("Expected updates: %s\nGiven: %s", expectedUpdates, updatesString)
		}
	}
}

func TestAccAWSMqBroker_basic(t *testing.T) {
	sgName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	brokerName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfig(sgName, brokerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists("aws_mq_broker.test"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "broker_name", brokerName),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "configuration.0.revision", regexp.MustCompile(`^[0-9]+$`)),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "deployment_mode", "SINGLE_INSTANCE"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "engine_version", "5.15.0"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "host_instance_type", "mq.t2.micro"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttrSet("aws_mq_broker.test", "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckResourceAttrSet("aws_mq_broker.test", "maintenance_window_start_time.0.time_of_day"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "logs.#", "1"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "logs.0.general", "true"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "logs.0.audit", "false"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "maintenance_window_start_time.0.time_zone", "UTC"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "publicly_accessible", "false"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "security_groups.#", "1"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.#", "1"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3793764891.console_access", "false"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3793764891.groups.#", "0"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3793764891.username", "Test"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3793764891.password", "TestTest1234"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "arn",
						regexp.MustCompile("^arn:aws:mq:[a-z0-9-]+:[0-9]{12}:broker:[a-z0-9-]+:[a-f0-9-]+$")),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "instances.#", "1"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "instances.0.endpoints.#", "5"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
				),
			},
		},
	})
}

func TestAccAWSMqBroker_allFieldsDefaultVpc(t *testing.T) {
	sgName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	cfgNameBefore := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	cfgNameAfter := fmt.Sprintf("tf-acc-test-updated-%s", acctest.RandString(5))
	brokerName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	cfgBodyBefore := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>`
	cfgBodyAfter := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
  <plugins>
    <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
    <statisticsBrokerPlugin/>
    <timeStampingBrokerPlugin ttlCeiling="86400000" zeroExpirationOverride="86400000"/>
  </plugins>
</broker>`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfig_allFieldsDefaultVpc(sgName, cfgNameBefore, cfgBodyBefore, brokerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists("aws_mq_broker.test"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "broker_name", brokerName),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "configuration.0.revision", "2"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "deployment_mode", "ACTIVE_STANDBY_MULTI_AZ"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "engine_version", "5.15.0"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "host_instance_type", "mq.t2.micro"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "maintenance_window_start_time.0.day_of_week", "TUESDAY"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "maintenance_window_start_time.0.time_of_day", "02:00"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "maintenance_window_start_time.0.time_zone", "CET"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "logs.#", "1"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "logs.0.general", "false"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "logs.0.audit", "false"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "publicly_accessible", "true"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "security_groups.#", "2"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.#", "2"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1344916805.console_access", "true"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1344916805.groups.#", "3"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1344916805.groups.2456940119", "first"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1344916805.groups.3055489385", "second"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1344916805.groups.607264868", "third"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1344916805.password", "SecondTestTest1234"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1344916805.username", "SecondTest"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3793764891.console_access", "false"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3793764891.groups.#", "0"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3793764891.password", "TestTest1234"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3793764891.username", "Test"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "arn",
						regexp.MustCompile("^arn:aws:mq:[a-z0-9-]+:[0-9]{12}:broker:[a-z0-9-]+:[a-f0-9-]+$")),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "instances.#", "2"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "instances.0.endpoints.#", "5"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.1.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.1.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "instances.1.endpoints.#", "5"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.1.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.1.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.1.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.1.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.1.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
				),
			},
			{
				// Update configuration in-place
				Config: testAccMqBrokerConfig_allFieldsDefaultVpc(sgName, cfgNameBefore, cfgBodyAfter, brokerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists("aws_mq_broker.test"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "broker_name", brokerName),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "configuration.0.revision", "3"),
				),
			},
			{
				// Replace configuration
				Config: testAccMqBrokerConfig_allFieldsDefaultVpc(sgName, cfgNameAfter, cfgBodyAfter, brokerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists("aws_mq_broker.test"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "broker_name", brokerName),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "configuration.0.revision", "2"),
				),
			},
		},
	})
}

func TestAccAWSMqBroker_allFieldsCustomVpc(t *testing.T) {
	sgName := fmt.Sprintf("tf-acc-test-vpc-%s", acctest.RandString(5))
	cfgNameBefore := fmt.Sprintf("tf-acc-test-vpc-%s", acctest.RandString(5))
	cfgNameAfter := fmt.Sprintf("tf-acc-test-vpc-updated-%s", acctest.RandString(5))
	brokerName := fmt.Sprintf("tf-acc-test-vpc-%s", acctest.RandString(5))

	cfgBodyBefore := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>`
	cfgBodyAfter := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
  <plugins>
    <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
    <statisticsBrokerPlugin/>
    <timeStampingBrokerPlugin ttlCeiling="86400000" zeroExpirationOverride="86400000"/>
  </plugins>
</broker>`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfig_allFieldsCustomVpc(sgName, cfgNameBefore, cfgBodyBefore, brokerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists("aws_mq_broker.test"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "broker_name", brokerName),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "configuration.0.revision", "2"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "deployment_mode", "ACTIVE_STANDBY_MULTI_AZ"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "engine_version", "5.15.0"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "host_instance_type", "mq.t2.micro"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "maintenance_window_start_time.0.day_of_week", "TUESDAY"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "maintenance_window_start_time.0.time_of_day", "02:00"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "maintenance_window_start_time.0.time_zone", "CET"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "logs.#", "1"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "logs.0.general", "true"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "logs.0.audit", "true"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "publicly_accessible", "true"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "security_groups.#", "2"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.#", "2"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1344916805.console_access", "true"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1344916805.groups.#", "3"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1344916805.groups.2456940119", "first"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1344916805.groups.3055489385", "second"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1344916805.groups.607264868", "third"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1344916805.password", "SecondTestTest1234"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1344916805.username", "SecondTest"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3793764891.console_access", "false"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3793764891.groups.#", "0"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3793764891.password", "TestTest1234"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3793764891.username", "Test"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "arn",
						regexp.MustCompile("^arn:aws:mq:[a-z0-9-]+:[0-9]{12}:broker:[a-z0-9-]+:[a-f0-9-]+$")),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "instances.#", "2"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "instances.0.endpoints.#", "5"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.0.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.1.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.1.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "instances.1.endpoints.#", "5"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.1.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.1.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.1.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.1.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "instances.1.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
				),
			},
			{
				// Update configuration in-place
				Config: testAccMqBrokerConfig_allFieldsCustomVpc(sgName, cfgNameBefore, cfgBodyAfter, brokerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists("aws_mq_broker.test"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "broker_name", brokerName),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "configuration.0.revision", "3"),
				),
			},
			{
				// Replace configuration
				Config: testAccMqBrokerConfig_allFieldsCustomVpc(sgName, cfgNameAfter, cfgBodyAfter, brokerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists("aws_mq_broker.test"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "broker_name", brokerName),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "configuration.#", "1"),
					resource.TestMatchResourceAttr("aws_mq_broker.test", "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "configuration.0.revision", "2"),
				),
			},
		},
	})
}

func TestAccAWSMqBroker_updateUsers(t *testing.T) {
	sgName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	brokerName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfig_updateUsers1(sgName, brokerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists("aws_mq_broker.test"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.#", "1"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3400735725.console_access", "false"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3400735725.groups.#", "0"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3400735725.password", "TestTest1111"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.3400735725.username", "first"),
				),
			},
			// Adding new user + modify existing
			{
				Config: testAccMqBrokerConfig_updateUsers2(sgName, brokerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists("aws_mq_broker.test"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.#", "2"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1074486012.console_access", "false"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1074486012.groups.#", "0"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1074486012.password", "TestTest2222"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1074486012.username", "second"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1166726986.console_access", "true"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1166726986.groups.#", "0"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1166726986.password", "TestTest1111updated"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.1166726986.username", "first"),
				),
			},
			// Deleting user + modify existing
			{
				Config: testAccMqBrokerConfig_updateUsers3(sgName, brokerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists("aws_mq_broker.test"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.#", "1"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.2244717082.console_access", "false"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.2244717082.groups.#", "1"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.2244717082.groups.2282622326", "admin"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.2244717082.password", "TestTest2222"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "user.2244717082.username", "second"),
				),
			},
		},
	})
}

func TestAccAWSMqBroker_updateTags(t *testing.T) {
	sgName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	brokerName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfig_updateTags1(sgName, brokerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists("aws_mq_broker.test"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "tags.env", "test"),
				),
			},
			// Adding new user + modify existing
			{
				Config: testAccMqBrokerConfig_updateTags2(sgName, brokerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists("aws_mq_broker.test"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "tags.env", "test2"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "tags.role", "test-role"),
				),
			},
			// Deleting user + modify existing
			{
				Config: testAccMqBrokerConfig_updateTags3(sgName, brokerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists("aws_mq_broker.test"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_mq_broker.test", "tags.role", "test-role"),
				),
			},
		},
	})
}

func testAccCheckAwsMqBrokerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).mqconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_mq_broker" {
			continue
		}

		input := &mq.DescribeBrokerInput{
			BrokerId: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeBroker(input)
		if err != nil {
			if isAWSErr(err, "NotFoundException", "") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected MQ Broker to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsMqBrokerExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccMqBrokerConfig(sgName, brokerName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = "%s"
}

resource "aws_mq_broker" "test" {
  broker_name = "%s"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  host_instance_type = "mq.t2.micro"
  security_groups = ["${aws_security_group.test.id}"]
  logs {
    general = true
  }
  user {
    username = "Test"
    password = "TestTest1234"
  }
}`, sgName, brokerName)
}

func testAccMqBrokerConfig_allFieldsDefaultVpc(sgName, cfgName, cfgBody, brokerName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "mq1" {
  name = "%s-1"
}

resource "aws_security_group" "mq2" {
  name = "%s-2"
}

resource "aws_mq_configuration" "test" {
  name = "%s"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  data = <<DATA
%s
DATA
}

resource "aws_mq_broker" "test" {
  auto_minor_version_upgrade = true
  apply_immediately = true
  broker_name = "%s"
  configuration {
    id = "${aws_mq_configuration.test.id}"
    revision = "${aws_mq_configuration.test.latest_revision}"
  }
  deployment_mode = "ACTIVE_STANDBY_MULTI_AZ"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  host_instance_type = "mq.t2.micro"
  maintenance_window_start_time {
    day_of_week = "TUESDAY"
    time_of_day = "02:00"
    time_zone = "CET"
  }
  publicly_accessible = true
  security_groups = ["${aws_security_group.mq1.id}", "${aws_security_group.mq2.id}"]
  user {
    username = "Test"
    password = "TestTest1234"
  }
  user {
    username = "SecondTest"
    password = "SecondTestTest1234"
    console_access = true
    groups = ["first", "second", "third"]
  }
}`, sgName, sgName, cfgName, cfgBody, brokerName)
}

func testAccMqBrokerConfig_allFieldsCustomVpc(sgName, cfgName, cfgBody, brokerName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_vpc" "main" {
  cidr_block = "10.11.0.0/16"
  tags = {
    Name = "terraform-testacc-mq-broker-all-fields-custom-vpc"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_route_table" "test" {
  vpc_id = "${aws_vpc.main.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.test.id}"
  }
}

resource "aws_subnet" "private" {
  count = 2
  cidr_block = "10.11.${count.index}.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  vpc_id = "${aws_vpc.main.id}"
  tags = {
    Name = "tf-acc-mq-broker-all-fields-custom-vpc-${count.index}"
  }
}

resource "aws_route_table_association" "test" {
  count          = 2
  subnet_id      = "${aws_subnet.private.*.id[count.index]}"
  route_table_id = "${aws_route_table.test.id}"
}

resource "aws_security_group" "mq1" {
  name = "%s-1"
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_security_group" "mq2" {
  name = "%s-2"
  vpc_id = "${aws_vpc.main.id}"
}

resource "aws_mq_configuration" "test" {
  name = "%s"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  data = <<DATA
%s
DATA
}

resource "aws_mq_broker" "test" {
  auto_minor_version_upgrade = true
  apply_immediately = true
  broker_name = "%s"
  configuration {
    id = "${aws_mq_configuration.test.id}"
    revision = "${aws_mq_configuration.test.latest_revision}"
  }
  deployment_mode = "ACTIVE_STANDBY_MULTI_AZ"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  host_instance_type = "mq.t2.micro"
  logs {
    general = true
    audit = true
  }
  maintenance_window_start_time {
    day_of_week = "TUESDAY"
    time_of_day = "02:00"
    time_zone = "CET"
  }
  publicly_accessible = true
  security_groups = ["${aws_security_group.mq1.id}", "${aws_security_group.mq2.id}"]
  subnet_ids = ["${aws_subnet.private.*.id[0]}", "${aws_subnet.private.*.id[1]}"]
  user {
    username = "Test"
    password = "TestTest1234"
  }
  user {
    username = "SecondTest"
    password = "SecondTestTest1234"
    console_access = true
    groups = ["first", "second", "third"]
  }

  depends_on = ["aws_internet_gateway.test"]
}`, sgName, sgName, cfgName, cfgBody, brokerName)
}

func testAccMqBrokerConfig_updateUsers1(sgName, brokerName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = "%s"
}

resource "aws_mq_broker" "test" {
  apply_immediately = true
  broker_name = "%s"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  host_instance_type = "mq.t2.micro"
  security_groups = ["${aws_security_group.test.id}"]
  user {
    username = "first"
    password = "TestTest1111"
  }
}`, sgName, brokerName)
}

func testAccMqBrokerConfig_updateUsers2(sgName, brokerName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = "%s"
}

resource "aws_mq_broker" "test" {
  apply_immediately = true
  broker_name = "%s"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  host_instance_type = "mq.t2.micro"
  security_groups = ["${aws_security_group.test.id}"]
  user {
    console_access = true
    username = "first"
    password = "TestTest1111updated"
  }
  user {
    username = "second"
    password = "TestTest2222"
  }
}`, sgName, brokerName)
}

func testAccMqBrokerConfig_updateUsers3(sgName, brokerName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = "%s"
}

resource "aws_mq_broker" "test" {
  apply_immediately = true
  broker_name = "%s"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  host_instance_type = "mq.t2.micro"
  security_groups = ["${aws_security_group.test.id}"]
  user {
    username = "second"
    password = "TestTest2222"
    groups = ["admin"]
  }
}`, sgName, brokerName)
}

func testAccMqBrokerConfig_updateTags1(sgName, brokerName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = "%s"
}

resource "aws_mq_broker" "test" {
  apply_immediately = true
  broker_name = "%s"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  host_instance_type = "mq.t2.micro"
	security_groups = ["${aws_security_group.test.id}"]
  user {
    username = "Test"
    password = "TestTest1234"
	}

  tags {
    env = "test"
  }
}`, sgName, brokerName)
}

func testAccMqBrokerConfig_updateTags2(sgName, brokerName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = "%s"
}

resource "aws_mq_broker" "test" {
  apply_immediately = true
  broker_name = "%s"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  host_instance_type = "mq.t2.micro"
	security_groups = ["${aws_security_group.test.id}"]
  user {
    username = "Test"
    password = "TestTest1234"
	}

  tags {
		env = "test2"
		role = "test-role"
  }
}`, sgName, brokerName)
}

func testAccMqBrokerConfig_updateTags3(sgName, brokerName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = "%s"
}

resource "aws_mq_broker" "test" {
  apply_immediately = true
  broker_name = "%s"
  engine_type = "ActiveMQ"
  engine_version = "5.15.0"
  host_instance_type = "mq.t2.micro"
	security_groups = ["${aws_security_group.test.id}"]
  user {
    username = "Test"
    password = "TestTest1234"
	}

  tags {
		role = "test-role"
  }
}`, sgName, brokerName)
}
