package aws

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/mq/waiter"
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

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(resp.BrokerSummaries), func(i, j int) {
		resp.BrokerSummaries[i], resp.BrokerSummaries[j] = resp.BrokerSummaries[j], resp.BrokerSummaries[i]
	})

	for _, bs := range resp.BrokerSummaries {
		log.Printf("[INFO] Deleting MQ broker %s", aws.StringValue(bs.BrokerId))
		_, err := conn.DeleteBroker(&mq.DeleteBrokerInput{
			BrokerId: bs.BrokerId,
		})
		if tfawserr.ErrMessageContains(err, mq.ErrCodeBadRequestException, "while in state [CREATION_IN_PROGRESS") {
			log.Printf("[WARN] Broker in state CREATION_IN_PROGRESS and must complete creation before deletion")
			if _, err = waiter.BrokerCreated(conn, aws.StringValue(bs.BrokerId)); err != nil {
				return err
			}

			log.Printf("[WARN] Retrying deletion of broker %s", aws.StringValue(bs.BrokerId))
			_, err = conn.DeleteBroker(&mq.DeleteBrokerInput{
				BrokerId: bs.BrokerId,
			})
		}
		if err != nil {
			return err
		}
		if _, err = waiter.BrokerDeleted(conn, aws.StringValue(bs.BrokerId)); err != nil {
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
	var broker mq.DescribeBrokerResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(mq.EndpointsID, t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttr(resourceName, "authentication_strategy", "simple"),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.revision", regexp.MustCompile(`^[0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", "SINGLE_INSTANCE"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", "true"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.15.0"),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.time_of_day"),
					resource.TestCheckResourceAttr(resourceName, "logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", "true"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", "false"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "efs"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "mq", regexp.MustCompile(`broker:+.`)),
					resource.TestCheckResourceAttr(resourceName, "instances.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccAWSMqBroker_throughputOptimized(t *testing.T) {
	var broker mq.DescribeBrokerResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(mq.EndpointsID, t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerEBSConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.revision", regexp.MustCompile(`^[0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", "SINGLE_INSTANCE"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", "true"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.15.9"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "ebs"),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.time_of_day"),
					resource.TestCheckResourceAttr(resourceName, "logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", "true"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", "false"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "mq", regexp.MustCompile(`broker:+.`)),
					resource.TestCheckResourceAttr(resourceName, "instances.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
				),
			},
		},
	})
}

func TestAccAWSMqBroker_allFieldsDefaultVpc(t *testing.T) {
	var broker mq.DescribeBrokerResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := acctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_mq_broker.test"

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
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(mq.EndpointsID, t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfig_allFieldsDefaultVpc(rName, rName, cfgBodyBefore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", "ACTIVE_STANDBY_MULTI_AZ"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.15.0"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "efs"),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.day_of_week", "TUESDAY"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_of_day", "02:00"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "CET"),
					resource.TestCheckResourceAttr(resourceName, "logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", "false"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", "false"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "user.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "true",
						"groups.#":       "3",
						"username":       "SecondTest",
						"password":       "SecondTestTest1234",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "first"),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "second"),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "third"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "mq", regexp.MustCompile(`broker:+.`)),
					resource.TestCheckResourceAttr(resourceName, "instances.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.1.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
			{
				// Update configuration in-place
				Config: testAccMqBrokerConfig_allFieldsDefaultVpc(rName, rName, cfgBodyAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", "3"),
				),
			},
			{
				// Replace configuration
				Config: testAccMqBrokerConfig_allFieldsDefaultVpc(rName, rNameUpdated, cfgBodyAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", "2"),
				),
			},
		},
	})
}

func TestAccAWSMqBroker_allFieldsCustomVpc(t *testing.T) {
	var broker mq.DescribeBrokerResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := acctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_mq_broker.test"

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
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(mq.EndpointsID, t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfig_allFieldsCustomVpc(rName, rName, cfgBodyBefore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "true"),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", "ACTIVE_STANDBY_MULTI_AZ"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "ActiveMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.15.0"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "efs"),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.day_of_week", "TUESDAY"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_of_day", "02:00"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "CET"),
					resource.TestCheckResourceAttr(resourceName, "logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", "true"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.audit", "true"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "true"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "user.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "true",
						"groups.#":       "3",
						"username":       "SecondTest",
						"password":       "SecondTestTest1234",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "first"),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "second"),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "third"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "mq", regexp.MustCompile(`broker:+.`)),
					resource.TestCheckResourceAttr(resourceName, "instances.#", "2"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com:8162$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.ip_address",
						regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.1.endpoints.#", "5"),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.0", regexp.MustCompile(`^ssl://[a-z0-9-\.]+:61617$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.1", regexp.MustCompile(`^amqp\+ssl://[a-z0-9-\.]+:5671$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.2", regexp.MustCompile(`^stomp\+ssl://[a-z0-9-\.]+:61614$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.3", regexp.MustCompile(`^mqtt\+ssl://[a-z0-9-\.]+:8883$`)),
					resource.TestMatchResourceAttr(resourceName, "instances.1.endpoints.4", regexp.MustCompile(`^wss://[a-z0-9-\.]+:61619$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
			{
				// Update configuration in-place
				Config: testAccMqBrokerConfig_allFieldsCustomVpc(rName, rName, cfgBodyAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", "3"),
				),
			},
			{
				// Replace configuration
				Config: testAccMqBrokerConfig_allFieldsCustomVpc(rName, rNameUpdated, cfgBodyAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "configuration.0.id", regexp.MustCompile(`^c-[a-z0-9-]+$`)),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.revision", "2"),
				),
			},
		},
	})
}

func TestAccAWSMqBroker_EncryptionOptions_KmsKeyId(t *testing.T) {
	var broker mq.DescribeBrokerResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(mq.EndpointsID, t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfigEncryptionOptionsKmsKeyId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_options.0.kms_key_id", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccAWSMqBroker_EncryptionOptions_UseAwsOwnedKey_Disabled(t *testing.T) {
	var broker mq.DescribeBrokerResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(mq.EndpointsID, t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfigEncryptionOptionsUseAwsOwnedKey(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccAWSMqBroker_EncryptionOptions_UseAwsOwnedKey_Enabled(t *testing.T) {
	var broker mq.DescribeBrokerResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(mq.EndpointsID, t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfigEncryptionOptionsUseAwsOwnedKey(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccAWSMqBroker_updateUsers(t *testing.T) {
	var broker mq.DescribeBrokerResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(mq.EndpointsID, t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfig_updateUsers1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "user.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "first",
						"password":       "TestTest1111",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
			// Adding new user + modify existing
			{
				Config: testAccMqBrokerConfig_updateUsers2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "user.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "second",
						"password":       "TestTest2222",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "true",
						"groups.#":       "0",
						"username":       "first",
						"password":       "TestTest1111updated",
					}),
				),
			},
			// Deleting user + modify existing
			{
				Config: testAccMqBrokerConfig_updateUsers3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "user.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "1",
						"username":       "second",
						"password":       "TestTest2222",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "user.*.groups.*", "admin"),
				),
			},
		},
	})
}

func TestAccAWSMqBroker_tags(t *testing.T) {
	var broker mq.DescribeBrokerResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(mq.EndpointsID, t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
			{
				Config: testAccMqBrokerConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccMqBrokerConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSMqBroker_updateSecurityGroup(t *testing.T) {
	var broker mq.DescribeBrokerResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(mq.EndpointsID, t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
			{
				Config: testAccMqBrokerConfig_updateSecurityGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "2"),
				),
			},
			// Trigger a reboot and ensure the password change was applied
			// User hashcode can be retrieved by calling resourceAwsMqUserHash
			{
				Config: testAccMqBrokerConfig_updateUsersSecurityGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "user.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"username": "Test",
						"password": "TestTest9999",
					}),
				),
			},
		},
	})
}

func TestAccAWSMqBroker_updateEngineVersion(t *testing.T) {
	var broker mq.DescribeBrokerResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(mq.EndpointsID, t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.15.0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
			{
				Config: testAccMqBrokerEngineVersionUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "5.15.9"),
				),
			},
		},
	})
}

func TestAccAWSMqBroker_disappears(t *testing.T) {
	var broker mq.DescribeBrokerResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(mq.EndpointsID, t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsMqBroker(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSMqBroker_rabbitmq(t *testing.T) {
	var broker mq.DescribeBrokerResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(mq.EndpointsID, t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfigRabbit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "RabbitMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.8.6"),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.t3.micro"),
					resource.TestCheckResourceAttr(resourceName, "instances.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexp.MustCompile(`^amqps://[a-z0-9-\.]+:5671$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccAWSMqBroker_clusterRabbitMQ(t *testing.T) {
	var broker mq.DescribeBrokerResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(mq.EndpointsID, t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRabbitMqClusterBrokerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_mode", "CLUSTER_MULTI_AZ"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_options.0.use_aws_owned_key", "true"),
					resource.TestCheckResourceAttr(resourceName, "engine_type", "RabbitMQ"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.8.6"),
					resource.TestCheckResourceAttr(resourceName, "host_instance_type", "mq.m5.large"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.time_of_day"),
					resource.TestCheckResourceAttr(resourceName, "logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logs.0.general", "false"),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window_start_time.0.time_zone", "UTC"),
					resource.TestCheckResourceAttr(resourceName, "publicly_accessible", "false"),
					resource.TestCheckResourceAttr(resourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "user.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "mq", regexp.MustCompile(`broker:+.`)),
					resource.TestCheckResourceAttr(resourceName, "instances.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.console_url",
						regexp.MustCompile(`^https://[a-f0-9-]+\.mq.[a-z0-9-]+.amazonaws.com$`)),
					resource.TestCheckResourceAttr(resourceName, "instances.0.endpoints.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "instances.0.endpoints.0", regexp.MustCompile(`^amqps://[a-z0-9-\.]+:5671$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccAWSMqBroker_ldap(t *testing.T) {
	var broker mq.DescribeBrokerResponse
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_mq_broker.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(mq.EndpointsID, t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMqBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMqBrokerConfig_ldap(rName, "anyusername"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqBrokerExists(resourceName, &broker),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttr(resourceName, "broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "authentication_strategy", "ldap"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.hosts.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.hosts.0", "my.ldap.server-1.com"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.hosts.1", "my.ldap.server-2.com"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.role_base", "role.base"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.role_name", "role.name"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.role_search_matching", "role.search.matching"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.role_search_subtree", "true"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.service_account_username", "anyusername"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.user_base", "user.base"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.user_role_name", "user.role.name"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.user_search_matching", "user.search.matching"),
					resource.TestCheckResourceAttr(resourceName, "ldap_server_metadata.0.user_search_subtree", "true"),
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
			if isAWSErr(err, mq.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected MQ Broker to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsMqBrokerExists(name string, broker *mq.DescribeBrokerResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MQ Broker is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).mqconn
		resp, err := conn.DescribeBroker(&mq.DescribeBrokerInput{
			BrokerId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("Error describing MQ Broker: %s", err.Error())
		}

		*broker = *resp

		return nil
	}
}

func testAccPreCheckAWSMq(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).mqconn

	input := &mq.ListBrokersInput{}

	_, err := conn.ListBrokers(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccMqBrokerConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  broker_name             = %[1]q
  engine_type             = "ActiveMQ"
  engine_version          = "5.15.0"
  host_instance_type      = "mq.t2.micro"
  security_groups         = [aws_security_group.test.id]
  authentication_strategy = "simple"
  storage_type            = "efs"

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName)
}

func testAccMqBrokerEBSConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = "5.15.9"
  storage_type       = "ebs"
  host_instance_type = "mq.m5.large"
  security_groups    = [aws_security_group.test.id]

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName)
}

func testAccMqBrokerEngineVersionUpdateConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  apply_immediately  = true
  engine_type        = "ActiveMQ"
  engine_version     = "5.15.9"
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName)
}

func testAccMqBrokerConfig_allFieldsDefaultVpc(rName, cfgName, cfgBody string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "mq1" {
  name = %[1]q
}

resource "aws_security_group" "mq2" {
  name = "%[1]s-2"
}

resource "aws_mq_configuration" "test" {
  name           = %[2]q
  engine_type    = "ActiveMQ"
  engine_version = "5.15.0"

  data = <<DATA
%[3]s
DATA
}

resource "aws_mq_broker" "test" {
  auto_minor_version_upgrade = true
  apply_immediately          = true
  broker_name                = %[1]q

  configuration {
    id       = aws_mq_configuration.test.id
    revision = aws_mq_configuration.test.latest_revision
  }

  deployment_mode    = "ACTIVE_STANDBY_MULTI_AZ"
  engine_type        = "ActiveMQ"
  engine_version     = "5.15.0"
  storage_type       = "efs"
  host_instance_type = "mq.t2.micro"

  maintenance_window_start_time {
    day_of_week = "TUESDAY"
    time_of_day = "02:00"
    time_zone   = "CET"
  }

  publicly_accessible = true
  security_groups     = [aws_security_group.mq1.id, aws_security_group.mq2.id]

  user {
    username = "Test"
    password = "TestTest1234"
  }

  user {
    username       = "SecondTest"
    password       = "SecondTestTest1234"
    console_access = true
    groups         = ["first", "second", "third"]
  }
}
`, rName, cfgName, cfgBody)
}

func testAccMqBrokerConfig_allFieldsCustomVpc(rName, cfgName, cfgBody string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.11.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "private" {
  count             = 2
  cidr_block        = "10.11.${count.index}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.main.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  count          = 2
  subnet_id      = aws_subnet.private.*.id[count.index]
  route_table_id = aws_route_table.test.id
}

resource "aws_security_group" "mq1" {
  name   = "%[1]s-1"
  vpc_id = aws_vpc.main.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "mq2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.main.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_mq_configuration" "test" {
  name           = %[2]q
  engine_type    = "ActiveMQ"
  engine_version = "5.15.0"

  data = <<DATA
%[3]s
DATA
}

resource "aws_mq_broker" "test" {
  auto_minor_version_upgrade = true
  apply_immediately          = true
  broker_name                = %[1]q

  configuration {
    id       = aws_mq_configuration.test.id
    revision = aws_mq_configuration.test.latest_revision
  }

  deployment_mode    = "ACTIVE_STANDBY_MULTI_AZ"
  engine_type        = "ActiveMQ"
  engine_version     = "5.15.0"
  storage_type       = "efs"
  host_instance_type = "mq.t2.micro"

  logs {
    general = true
    audit   = true
  }

  maintenance_window_start_time {
    day_of_week = "TUESDAY"
    time_of_day = "02:00"
    time_zone   = "CET"
  }

  publicly_accessible = true
  security_groups     = [aws_security_group.mq1.id, aws_security_group.mq2.id]
  subnet_ids          = aws_subnet.private[*].id

  user {
    username = "Test"
    password = "TestTest1234"
  }

  user {
    username       = "SecondTest"
    password       = "SecondTestTest1234"
    console_access = true
    groups         = ["first", "second", "third"]
  }

  depends_on = [aws_internet_gateway.test]
}
`, rName, cfgName, cfgBody)
}

func testAccMqBrokerConfigEncryptionOptionsKmsKeyId(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = "5.15.0"
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  encryption_options {
    kms_key_id        = aws_kms_key.test.arn
    use_aws_owned_key = false
  }

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName)
}

func testAccMqBrokerConfigEncryptionOptionsUseAwsOwnedKey(rName string, useAwsOwnedKey bool) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = "5.15.0"
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  encryption_options {
    use_aws_owned_key = %[2]t
  }

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, useAwsOwnedKey)
}

func testAccMqBrokerConfig_updateUsers1(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  apply_immediately  = true
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = "5.15.0"
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  user {
    username = "first"
    password = "TestTest1111"
  }
}
`, rName)
}

func testAccMqBrokerConfig_updateUsers2(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  apply_immediately  = true
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = "5.15.0"
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  user {
    console_access = true
    username       = "first"
    password       = "TestTest1111updated"
  }

  user {
    username = "second"
    password = "TestTest2222"
  }
}
`, rName)
}

func testAccMqBrokerConfig_updateUsers3(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  apply_immediately  = true
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = "5.15.0"
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  user {
    username = "second"
    password = "TestTest2222"
    groups   = ["admin"]
  }
}
`, rName)
}

func testAccMqBrokerConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  apply_immediately  = true
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = "5.15.0"
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  user {
    username = "Test"
    password = "TestTest1234"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccMqBrokerConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  apply_immediately  = true
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = "5.15.0"
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id]

  user {
    username = "Test"
    password = "TestTest1234"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccMqBrokerConfig_updateSecurityGroups(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_security_group" "test2" {
  name = "%[1]s-2"
}

resource "aws_mq_broker" "test" {
  apply_immediately  = true
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = "5.15.0"
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test.id, aws_security_group.test2.id]

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName)
}

func testAccMqBrokerConfig_updateUsersSecurityGroups(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_security_group" "test2" {
  name = "%[1]s-2"
}

resource "aws_mq_broker" "test" {
  apply_immediately  = true
  broker_name        = %[1]q
  engine_type        = "ActiveMQ"
  engine_version     = "5.15.0"
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.test2.id]

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest9999"
  }
}
`, rName)
}

func testAccMqBrokerConfigRabbit(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  engine_type        = "RabbitMQ"
  engine_version     = "3.8.6"
  host_instance_type = "mq.t3.micro"
  security_groups    = [aws_security_group.test.id]

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName)
}

func testAccRabbitMqClusterBrokerConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  broker_name        = %[1]q
  engine_type        = "RabbitMQ"
  engine_version     = "3.8.6"
  host_instance_type = "mq.m5.large"
  security_groups    = [aws_security_group.test.id]
  storage_type       = "ebs"
  deployment_mode    = "CLUSTER_MULTI_AZ"

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName)
}

func testAccMqBrokerConfig_ldap(rName, ldapUsername string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_mq_broker" "test" {
  apply_immediately       = true
  authentication_strategy = "ldap"
  broker_name             = %[1]q
  engine_type             = "ActiveMQ"
  engine_version          = "5.15.0"
  host_instance_type      = "mq.t2.micro"
  security_groups         = [aws_security_group.test.id]

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }

  ldap_server_metadata {
    hosts                    = ["my.ldap.server-1.com", "my.ldap.server-2.com"]
    role_base                = "role.base"
    role_name                = "role.name"
    role_search_matching     = "role.search.matching"
    role_search_subtree      = true
    service_account_password = "supersecret"
    service_account_username = %[2]q
    user_base                = "user.base"
    user_role_name           = "user.role.name"
    user_search_matching     = "user.search.matching"
    user_search_subtree      = true
  }
}
`, rName, ldapUsername)
}
