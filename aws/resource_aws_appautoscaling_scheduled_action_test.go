package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/applicationautoscaling/finder"
)

func TestAccAWSAppautoscalingScheduledAction_DynamoDB(t *testing.T) {
	var sa1, sa2 applicationautoscaling.ScheduledAction
	rName := acctest.RandomWithPrefix("tf-acc-test")
	schedule1 := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	schedule2 := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05")
	updatedTimezone := "Pacific/Tahiti"
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, applicationautoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_DynamoDB(rName, schedule1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", autoscalingTargetResourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, "schedule", fmt.Sprintf("at(%s)", schedule1)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "10"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckNoResourceAttr(resourceName, "start_time"),
					resource.TestCheckNoResourceAttr(resourceName, "end_time"),
				),
			},
			{
				Config: testAccAppautoscalingScheduledActionConfig_DynamoDB_Updated(rName, schedule2, updatedTimezone),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa2),
					testAccCheckAppautoscalingScheduledActionNotRecreated(&sa1, &sa2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", autoscalingTargetResourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, "schedule", fmt.Sprintf("at(%s)", schedule2)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "9"),
					resource.TestCheckResourceAttr(resourceName, "timezone", updatedTimezone),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckNoResourceAttr(resourceName, "start_time"),
					resource.TestCheckNoResourceAttr(resourceName, "end_time"),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_ECS(t *testing.T) {
	var sa applicationautoscaling.ScheduledAction
	rName := acctest.RandomWithPrefix("tf-acc-test")
	ts := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, applicationautoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_ECS(rName, ts),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", autoscalingTargetResourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, "schedule", fmt.Sprintf("at(%s)", ts)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_EMR(t *testing.T) {
	var sa applicationautoscaling.ScheduledAction
	rName := acctest.RandomWithPrefix("tf-acc-test")
	ts := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, applicationautoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_EMR(rName, ts),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", autoscalingTargetResourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, "schedule", fmt.Sprintf("at(%s)", ts)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "5"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_Name_Duplicate(t *testing.T) {
	var sa1, sa2 applicationautoscaling.ScheduledAction
	resourceName := "aws_appautoscaling_scheduled_action.test"
	resourceName2 := "aws_appautoscaling_scheduled_action.test2"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, applicationautoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_Name_Duplicate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa1),
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName2, &sa2),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_SpotFleet(t *testing.T) {
	var sa applicationautoscaling.ScheduledAction
	rName := acctest.RandomWithPrefix("tf-acc-test")
	ts := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, applicationautoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_SpotFleet(rName, ts, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", autoscalingTargetResourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, "schedule", fmt.Sprintf("at(%s)", ts)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "3"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_Schedule_AtExpression_Timezone(t *testing.T) {
	var sa applicationautoscaling.ScheduledAction
	rName := acctest.RandomWithPrefix("tf-acc-test")
	ts := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	at := fmt.Sprintf("at(%s)", ts)
	timezone := "Pacific/Tahiti"
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	endTime := time.Now().AddDate(0, 0, 8).Format("2006-01-02T15:04:05Z")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, applicationautoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_ScheduleWithTimezone(rName, at, timezone, startTime, endTime),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", autoscalingTargetResourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, "schedule", at),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "10"),
					resource.TestCheckResourceAttr(resourceName, "timezone", timezone),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "start_time", startTime),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTime),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_Schedule_CronExpression_basic(t *testing.T) {
	var sa applicationautoscaling.ScheduledAction
	rName := acctest.RandomWithPrefix("tf-acc-test")
	cron := "cron(0 17 * * ? *)"
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, applicationautoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_Schedule(rName, cron),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", autoscalingTargetResourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, "schedule", cron),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "10"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckNoResourceAttr(resourceName, "start_time"),
					resource.TestCheckNoResourceAttr(resourceName, "end_time"),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_Schedule_CronExpression_Timezone(t *testing.T) {
	var sa applicationautoscaling.ScheduledAction
	rName := acctest.RandomWithPrefix("tf-acc-test")
	cron := "cron(0 17 * * ? *)"
	timezone := "Pacific/Tahiti"
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	endTime := time.Now().AddDate(0, 0, 8).Format("2006-01-02T15:04:05Z")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, applicationautoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_ScheduleWithTimezone(rName, cron, timezone, startTime, endTime),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", autoscalingTargetResourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, "schedule", cron),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "10"),
					resource.TestCheckResourceAttr(resourceName, "timezone", timezone),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "start_time", startTime),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTime),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_Schedule_CronExpression_StartEndTimeTimezone(t *testing.T) {
	var sa applicationautoscaling.ScheduledAction
	rName := acctest.RandomWithPrefix("tf-acc-test")
	cron := "cron(0 17 * * ? *)"
	scheduleTimezone := "Etc/GMT+9"                                    // Z-09:00 (IANA and RFC3339 have inverted signs)
	startTimezone, _ := time.LoadLocation("Antarctica/DumontDUrville") // Z+10:00
	endTimezone, _ := time.LoadLocation("America/Vancouver")           // Z-08:00
	startTime := time.Now().AddDate(0, 0, 2).In(startTimezone)
	startTimeUtc := startTime.UTC()
	endTime := time.Now().AddDate(0, 0, 8).In(endTimezone)
	endTimeUtc := endTime.UTC()
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, applicationautoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_ScheduleWithTimezone(rName, cron, scheduleTimezone, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", autoscalingTargetResourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, "schedule", cron),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "10"),
					resource.TestCheckResourceAttr(resourceName, "timezone", scheduleTimezone),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "start_time", startTimeUtc.Format(time.RFC3339)),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTimeUtc.Format(time.RFC3339)),
				),
			},
			{
				Config: testAccAppautoscalingScheduledActionConfig_Schedule(rName, cron),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", autoscalingTargetResourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, "schedule", cron),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "10"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "start_time", ""),
					resource.TestCheckResourceAttr(resourceName, "end_time", ""),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_Schedule_RateExpression_basic(t *testing.T) {
	var sa applicationautoscaling.ScheduledAction
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rate := "rate(1 day)"
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, applicationautoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_Schedule(rName, rate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", autoscalingTargetResourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, "schedule", rate),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "10"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckNoResourceAttr(resourceName, "start_time"),
					resource.TestCheckNoResourceAttr(resourceName, "end_time"),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_Schedule_RateExpression_Timezone(t *testing.T) {
	var sa applicationautoscaling.ScheduledAction
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rate := "rate(1 day)"
	timezone := "Pacific/Tahiti"
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	endTime := time.Now().AddDate(0, 0, 8).Format("2006-01-02T15:04:05Z")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, applicationautoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_ScheduleWithTimezone(rName, rate, timezone, startTime, endTime),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", autoscalingTargetResourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, "schedule", rate),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "10"),
					resource.TestCheckResourceAttr(resourceName, "timezone", timezone),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "start_time", startTime),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTime),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_MinCapacity(t *testing.T) {
	var sa1, sa2 applicationautoscaling.ScheduledAction
	rName := acctest.RandomWithPrefix("tf-acc-test")
	schedule := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, applicationautoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_MinCapacity(rName, schedule, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", autoscalingTargetResourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, "schedule", fmt.Sprintf("at(%s)", schedule)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckNoResourceAttr(resourceName, "start_time"),
					resource.TestCheckNoResourceAttr(resourceName, "end_time"),
				),
			},
			{
				Config: testAccAppautoscalingScheduledActionConfig_MinCapacity(rName, schedule, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa2),
					testAccCheckAppautoscalingScheduledActionNotRecreated(&sa1, &sa2),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", ""),
				),
			},
			{
				Config: testAccAppautoscalingScheduledActionConfig_MaxCapacity(rName, schedule, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa2),
					testAccCheckAppautoscalingScheduledActionNotRecreated(&sa1, &sa2),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "10"),
				),
			},
		},
	})
}

func TestAccAWSAppautoscalingScheduledAction_MaxCapacity(t *testing.T) {
	var sa1, sa2 applicationautoscaling.ScheduledAction
	rName := acctest.RandomWithPrefix("tf-acc-test")
	schedule := time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05")
	resourceName := "aws_appautoscaling_scheduled_action.test"
	autoscalingTargetResourceName := "aws_appautoscaling_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, applicationautoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAppautoscalingScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppautoscalingScheduledActionConfig_MaxCapacity(rName, schedule, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", autoscalingTargetResourceName, "service_namespace"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", autoscalingTargetResourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", autoscalingTargetResourceName, "scalable_dimension"),
					resource.TestCheckResourceAttr(resourceName, "schedule", fmt.Sprintf("at(%s)", schedule)),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "10"),
					resource.TestCheckResourceAttr(resourceName, "timezone", "UTC"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "autoscaling", regexp.MustCompile(fmt.Sprintf("scheduledAction:.+:scheduledActionName/%s$", rName))),
					resource.TestCheckNoResourceAttr(resourceName, "start_time"),
					resource.TestCheckNoResourceAttr(resourceName, "end_time"),
				),
			},
			{
				Config: testAccAppautoscalingScheduledActionConfig_MaxCapacity(rName, schedule, 8),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa2),
					testAccCheckAppautoscalingScheduledActionNotRecreated(&sa1, &sa2),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", ""),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", "8"),
				),
			},
			{
				Config: testAccAppautoscalingScheduledActionConfig_MinCapacity(rName, schedule, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsAppautoscalingScheduledActionExists(resourceName, &sa2),
					testAccCheckAppautoscalingScheduledActionNotRecreated(&sa1, &sa2),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.min_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "scalable_target_action.0.max_capacity", ""),
				),
			},
		},
	})
}

func testAccCheckAwsAppautoscalingScheduledActionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).appautoscalingconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appautoscaling_scheduled_action" {
			continue
		}

		input := &applicationautoscaling.DescribeScheduledActionsInput{
			ResourceId:           aws.String(rs.Primary.Attributes["resource_id"]),
			ScheduledActionNames: []*string{aws.String(rs.Primary.Attributes["name"])},
			ServiceNamespace:     aws.String(rs.Primary.Attributes["service_namespace"]),
		}
		resp, err := conn.DescribeScheduledActions(input)
		if err != nil {
			return err
		}
		if len(resp.ScheduledActions) > 0 {
			return fmt.Errorf("Appautoscaling Scheduled Action (%s) not deleted", rs.Primary.Attributes["name"])
		}
	}
	return nil
}

func testAccCheckAwsAppautoscalingScheduledActionExists(name string, obj *applicationautoscaling.ScheduledAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Application Autoscaling scheduled action (%s) ID not set", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).appautoscalingconn

		sa, err := finder.ScheduledAction(conn, rs.Primary.Attributes["name"], rs.Primary.Attributes["service_namespace"], rs.Primary.Attributes["resource_id"])
		if err != nil {
			return err
		}

		*obj = *sa

		return nil
	}
}

func testAccCheckAppautoscalingScheduledActionNotRecreated(i, j *applicationautoscaling.ScheduledAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationTime).Equal(aws.TimeValue(j.CreationTime)) {
			return fmt.Errorf("Application Auto Scaling scheduled action recreated")
		}

		return nil
	}
}

func testAccAppautoscalingScheduledActionConfig_DynamoDB(rName, ts string) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension

  schedule = "at(%[2]s)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 10
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}
`, rName, ts)
}

func testAccAppautoscalingScheduledActionConfig_DynamoDB_Updated(rName, ts, timezone string) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension

  schedule = "at(%[2]s)"
  timezone = %[3]q

  scalable_target_action {
    min_capacity = 2
    max_capacity = 9
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}
`, rName, ts, timezone)
}

func testAccAppautoscalingScheduledActionConfig_ECS(rName, ts string) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  schedule           = "at(%[2]s)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 5
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "ecs"
  resource_id        = "service/${aws_ecs_cluster.test.name}/${aws_ecs_service.test.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  min_capacity       = 1
  max_capacity       = 3
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<EOF
[
  {
    "name": "busybox",
    "image": "busybox:latest",
    "cpu": 10,
    "memory": 128,
    "essential": true
  }
]
EOF
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 50
}
`, rName, ts)
}

func testAccAppautoscalingScheduledActionConfig_EMR(rName, ts string) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  schedule           = "at(%[2]s)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 5
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "elasticmapreduce"
  resource_id        = "instancegroup/${aws_emr_cluster.test.id}/${aws_emr_instance_group.test.id}"
  scalable_dimension = "elasticmapreduce:instancegroup:InstanceCount"
  role_arn           = aws_iam_role.autoscale_role.arn
  min_capacity       = 1
  max_capacity       = 5
}

data "aws_availability_zones" "available" {
  # The requested instance type c4.large is not supported in the requested availability zone.
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-5.4.0"
  applications  = ["Spark"]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.instance_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 2
    instance_type  = "c4.large"
  }

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [aws_main_route_table_association.test]

  service_role     = aws_iam_role.emr_role.arn
  autoscaling_role = aws_iam_role.autoscale_role.arn
}

resource "aws_emr_instance_group" "test" {
  cluster_id     = aws_emr_cluster.test.id
  instance_count = 1
  instance_type  = "c4.large"
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  depends_on = [aws_subnet.test]

  lifecycle {
    ignore_changes = [
      ingress,
      egress,
    ]
  }
}

resource "aws_vpc" "test" {
  cidr_block           = "168.31.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-appautoscaling-scheduled-action-emr"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "168.31.0.0/20"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-appautoscaling-scheduled-action"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_main_route_table_association" "test" {
  vpc_id         = aws_vpc.test.id
  route_table_id = aws_route_table.test.id
}

resource "aws_iam_role" "emr_role" {
  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "emr_role" {
  role       = aws_iam_role.emr_role.id
  policy_arn = aws_iam_policy.emr_policy.arn
}

resource "aws_iam_policy" "emr_policy" {
  policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Resource": "*",
    "Action": [
      "ec2:AuthorizeSecurityGroupEgress",
      "ec2:AuthorizeSecurityGroupIngress",
      "ec2:CancelSpotInstanceRequests",
      "ec2:CreateNetworkInterface",
      "ec2:CreateSecurityGroup",
      "ec2:CreateTags",
      "ec2:DeleteNetworkInterface",
      "ec2:DeleteSecurityGroup",
      "ec2:DeleteTags",
      "ec2:DescribeAvailabilityZones",
      "ec2:DescribeAccountAttributes",
      "ec2:DescribeDhcpOptions",
      "ec2:DescribeInstanceStatus",
      "ec2:DescribeInstances",
      "ec2:DescribeKeyPairs",
      "ec2:DescribeNetworkAcls",
      "ec2:DescribeNetworkInterfaces",
      "ec2:DescribePrefixLists",
      "ec2:DescribeRouteTables",
      "ec2:DescribeSecurityGroups",
      "ec2:DescribeSpotInstanceRequests",
      "ec2:DescribeSpotPriceHistory",
      "ec2:DescribeSubnets",
      "ec2:DescribeVpcAttribute",
      "ec2:DescribeVpcEndpoints",
      "ec2:DescribeVpcEndpointServices",
      "ec2:DescribeVpcs",
      "ec2:DetachNetworkInterface",
      "ec2:ModifyImageAttribute",
      "ec2:ModifyInstanceAttribute",
      "ec2:RequestSpotInstances",
      "ec2:RevokeSecurityGroupEgress",
      "ec2:RunInstances",
      "ec2:TerminateInstances",
      "ec2:DeleteVolume",
      "ec2:DescribeVolumeStatus",
      "ec2:DescribeVolumes",
      "ec2:DetachVolume",
      "iam:GetRole",
      "iam:GetRolePolicy",
      "iam:ListInstanceProfiles",
      "iam:ListRolePolicies",
      "iam:PassRole",
      "s3:CreateBucket",
      "s3:Get*",
      "s3:List*",
      "sdb:BatchPutAttributes",
      "sdb:Select",
      "sqs:CreateQueue",
      "sqs:Delete*",
      "sqs:GetQueue*",
      "sqs:PurgeQueue",
      "sqs:ReceiveMessage"
    ]
  }]
}
EOT
}

# IAM Role for EC2 Instance Profile
resource "aws_iam_role" "instance_role" {
  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "instance_profile" {
  name = %[1]q
  role = aws_iam_role.instance_role.name
}

resource "aws_iam_role_policy_attachment" "instance_role" {
  role       = aws_iam_role.instance_role.id
  policy_arn = aws_iam_policy.instance_policy.arn
}

resource "aws_iam_policy" "instance_policy" {
  policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Resource": "*",
    "Action": [
      "cloudwatch:*",
      "dynamodb:*",
      "ec2:Describe*",
      "elasticmapreduce:Describe*",
      "elasticmapreduce:ListBootstrapActions",
      "elasticmapreduce:ListClusters",
      "elasticmapreduce:ListInstanceGroups",
      "elasticmapreduce:ListInstances",
      "elasticmapreduce:ListSteps",
      "kinesis:CreateStream",
      "kinesis:DeleteStream",
      "kinesis:DescribeStream",
      "kinesis:GetRecords",
      "kinesis:GetShardIterator",
      "kinesis:MergeShards",
      "kinesis:PutRecord",
      "kinesis:SplitShard",
      "rds:Describe*",
      "s3:*",
      "sdb:*",
      "sns:*",
      "sqs:*"
    ]
  }]
}
EOT
}

# IAM Role for autoscaling
resource "aws_iam_role" "autoscale_role" {
  assume_role_policy = data.aws_iam_policy_document.autoscale_role.json
}

data "aws_iam_policy_document" "autoscale_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.${data.aws_partition.current.dns_suffix}", "application-autoscaling.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "autoscale_role" {
  role       = aws_iam_role.autoscale_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, rName, ts)
}

func testAccAppautoscalingScheduledActionConfig_Name_Duplicate(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test2" {
  name           = "%[1]s-2"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}

resource "aws_appautoscaling_target" "test2" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test2.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_appautoscaling_scheduled_action" "test2" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test2.service_namespace
  resource_id        = aws_appautoscaling_target.test2.resource_id
  scalable_dimension = aws_appautoscaling_target.test2.scalable_dimension
  schedule           = "cron(0 17 * * ? *)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 10
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  schedule           = "cron(0 17 * * ? *)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 10
  }
}
`, rName)
}

func testAccAppautoscalingScheduledActionConfig_SpotFleet(rName, ts, validUntil string) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  schedule           = "at(%[2]s)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 3
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "ec2"
  resource_id        = "spot-fleet-request/${aws_spot_fleet_request.test.id}"
  scalable_dimension = "ec2:spot-fleet-request:TargetCapacity"
  min_capacity       = 1
  max_capacity       = 3
}

data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "fleet_role" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "spotfleet.${data.aws_partition.current.dns_suffix}",
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "fleet_role_policy" {
  role       = aws_iam_role.fleet_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2SpotFleetTaggingRole"
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.fleet_role.arn
  spot_price                          = "0.005"
  target_capacity                     = 2
  valid_until                         = %[3]q
  terminate_instances_with_expiration = true

  launch_specification {
    instance_type = "m3.medium"
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  }
}
`, rName, ts, validUntil)
}

func testAccAppautoscalingScheduledActionConfig_Schedule(rName, schedule string) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  schedule           = %[2]q

  scalable_target_action {
    min_capacity = 1
    max_capacity = 10
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}
`, rName, schedule)
}

func testAccAppautoscalingScheduledActionConfig_ScheduleWithTimezone(rName, schedule, timezone, startTime, endTime string) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension

  timezone   = %[2]q
  schedule   = %[3]q
  start_time = %[4]q
  end_time   = %[5]q

  scalable_target_action {
    min_capacity = 1
    max_capacity = 10
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}
`, rName, timezone, schedule, startTime, endTime)
}

func testAccAppautoscalingScheduledActionConfig_MinCapacity(rName, ts string, minCapacity int) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension

  schedule = "at(%[2]s)"

  scalable_target_action {
    min_capacity = %[3]d
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}
`, rName, ts, minCapacity)
}

func testAccAppautoscalingScheduledActionConfig_MaxCapacity(rName, ts string, maxCapacity int) string {
	return fmt.Sprintf(`
resource "aws_appautoscaling_scheduled_action" "test" {
  name               = %[1]q
  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension

  schedule = "at(%[2]s)"

  scalable_target_action {
    max_capacity = %[3]d
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "dynamodb"
  resource_id        = "table/${aws_dynamodb_table.test.name}"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  min_capacity       = 1
  max_capacity       = 10
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "UserID"

  attribute {
    name = "UserID"
    type = "S"
  }
}
`, rName, ts, maxCapacity)
}
