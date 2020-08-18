package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotevents"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSIoTEventsDetectorModel_basic(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTEventsDetectorModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTEventsDetectorModel_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTEventsDetectorModelExists("aws_iotevents_detector_model.detector"),
					resource.TestCheckResourceAttr("aws_iotevents_detector_model.detector", "name", fmt.Sprintf("test_detector_%s", rString)),
					resource.TestCheckResourceAttr("aws_iotevents_detector_model.detector", "description", "Example detector model"),
					resource.TestCheckResourceAttr("aws_iotevents_detector_model.detector", "tags.tagKey", "tagValue"),
					testAccDetectorModelBasic(rString),
				),
			},
		},
	})
}

func TestAccAWSIoTEventsDetectorModel_full(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTEventsDetectorModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTEventsDetectorModel_full(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTEventsDetectorModelExists("aws_iotevents_detector_model.detector"),
					testAccDetectorModelFull(rString),
				),
			},
		},
	})
}

func TestAccAWSIoTEventsDetectorModel_firehose(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTEventsDetectorModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTEventsDetectorModel_firehose(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTEventsDetectorModelExists("aws_iotevents_detector_model.detector"),
				),
			},
		},
	})
}

func TestAccAWSIoTEventsDetectorModel_iot_events(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTEventsDetectorModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTEventsDetectorModel_iot_events(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTEventsDetectorModelExists("aws_iotevents_detector_model.detector"),
				),
			},
		},
	})
}

func TestAccAWSIoTEventsDetectorModel_iot_topic_publish(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTEventsDetectorModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTEventsDetectorModel_iot_topic_publish(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTEventsDetectorModelExists("aws_iotevents_detector_model.detector"),
				),
			},
		},
	})
}

func TestAccAWSIoTEventsDetectorModel_lambda(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTEventsDetectorModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTEventsDetectorModel_lambda(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTEventsDetectorModelExists("aws_iotevents_detector_model.detector"),
				),
			},
		},
	})
}

func TestAccAWSIoTEventsDetectorModel_reset_timer(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTEventsDetectorModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTEventsDetectorModel_reset_timer(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTEventsDetectorModelExists("aws_iotevents_detector_model.detector"),
				),
			},
		},
	})
}

func TestAccAWSIoTEventsDetectorModel_set_timer(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTEventsDetectorModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTEventsDetectorModel_set_timer(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTEventsDetectorModelExists("aws_iotevents_detector_model.detector"),
				),
			},
		},
	})
}

func TestAccAWSIoTEventsDetectorModel_set_variable(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTEventsDetectorModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTEventsDetectorModel_set_variable(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTEventsDetectorModelExists("aws_iotevents_detector_model.detector"),
				),
			},
		},
	})
}

func TestAccAWSIoTEventsDetectorModel_sns(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTEventsDetectorModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTEventsDetectorModel_sns(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTEventsDetectorModelExists("aws_iotevents_detector_model.detector"),
				),
			},
		},
	})
}

func TestAccAWSIoTEventsDetectorModel_sqs(t *testing.T) {
	rString := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTEventsDetectorModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTEventsDetectorModel_sqs(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTEventsDetectorModelExists("aws_iotevents_detector_model.detector"),
				),
			},
		},
	})
}

func testAccCheckAWSIoTEventsDetectorModelExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testClearTimer(clearTimer *iotevents.ClearTimerAction, expectedClearTimer map[string]interface{}) error {
	if clearTimer == nil {
		return fmt.Errorf("ClearTimer is equal Nil")
	}

	expectedTimerName := expectedClearTimer["TimerName"].(string)
	if *clearTimer.TimerName != expectedTimerName {
		return fmt.Errorf("Clear Timer name %s not equals to %s", *clearTimer.TimerName, expectedTimerName)
	}

	return nil
}

func testEvent(event *iotevents.Event, expectedEvent map[string]interface{}) error {

	expectedEventName := expectedEvent["EventName"].(string)
	expectedEventCondition := expectedEvent["Condition"].(string)
	if *event.EventName != expectedEventName {
		return fmt.Errorf("Event name %s not equals to %s", *event.EventName, expectedEventName)
	}
	if *event.Condition != expectedEventCondition {
		return fmt.Errorf("Event condition %s not equals to %s", *event.Condition, expectedEventCondition)
	}

	return nil
}

func testTransitionEvent(transitionEvent *iotevents.TransitionEvent, expectedTransitionEvent map[string]interface{}) error {
	expectedTransitionEventName := expectedTransitionEvent["EventName"].(string)
	expectedTransitionEventCondition := expectedTransitionEvent["Condition"].(string)
	expectedTransitionEventNextState := expectedTransitionEvent["NextState"].(string)

	if *transitionEvent.EventName != expectedTransitionEventName {
		return fmt.Errorf("Transition Event name %s not equals to %s", *transitionEvent.EventName, expectedTransitionEventName)
	}
	if *transitionEvent.Condition != expectedTransitionEventCondition {
		return fmt.Errorf("Transition Event condition %s not equals to %s", *transitionEvent.Condition, expectedTransitionEventCondition)
	}
	if *transitionEvent.NextState != expectedTransitionEventNextState {
		return fmt.Errorf("Transition Event Next state %s not equals to %s", *transitionEvent.NextState, expectedTransitionEventNextState)
	}

	return nil
}

func testAccDetectorModelBasic(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ioteventsconn

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotevents_detector_model" {
				continue
			}

			params := &iotevents.DescribeDetectorModelInput{
				DetectorModelName: aws.String(rs.Primary.ID),
			}

			out, err := conn.DescribeDetectorModel(params)

			if err != nil {
				return err
			}

			detectorDefinition := out.DetectorModel.DetectorModelDefinition
			testInitialStateName := "first_state"
			if *detectorDefinition.InitialStateName != testInitialStateName {
				return fmt.Errorf("Initial state name %s not equals to %s", *detectorDefinition.InitialStateName, testInitialStateName)
			}

			states := detectorDefinition.States
			testStatesLen := 1
			if len(states) != testStatesLen {
				return fmt.Errorf("States len %d not equals to %d", len(states), testStatesLen)
			}

			// Test first state
			firstState := states[0]
			testFirstStateName := "first_state"
			if *firstState.StateName != testFirstStateName {
				return fmt.Errorf("State name %s not equals to %s", *firstState.StateName, testFirstStateName)
			}

			// Test OnEnter
			onEnter := firstState.OnEnter
			events := onEnter.Events
			testEventsLen := 1
			if len(events) != testEventsLen {
				return fmt.Errorf("Events len %d not equals to %d", len(events), testEventsLen)
			}

			event := events[0]
			expectedEventData := map[string]interface{}{
				"EventName": "test_event_name",
				"Condition": fmt.Sprintf("convert(Decimal, $input.test_input_%s.temperature) > 20", rString),
			}
			if err := testEvent(event, expectedEventData); err != nil {
				return err
			}

			actions := event.Actions
			testActionsLen := 1
			if len(actions) != testActionsLen {
				return fmt.Errorf("Actions len %d not equals to %d", len(actions), testActionsLen)
			}

			clearTimerAction := actions[0]
			clearTimerExpectedData := map[string]interface{}{
				"TimerName": "test_timer_name",
			}
			if err := testClearTimer(clearTimerAction.ClearTimer, clearTimerExpectedData); err != nil {
				return err
			}
		}

		return nil
	}
}

func checkAccFirstStateFull(state *iotevents.State, rString string) error {
	testFirstStateName := "first_state"
	if *state.StateName != testFirstStateName {
		return fmt.Errorf("State name %s not equals to %s", *state.StateName, testFirstStateName)
	}

	// Check OnEnter
	onEnter := state.OnEnter
	events := onEnter.Events
	testEventsLen := 1
	if len(events) != testEventsLen {
		return fmt.Errorf("Events len %d not equals to %d", len(events), testEventsLen)
	}

	event := events[0]
	expectedEventData := map[string]interface{}{
		"EventName": "test_event_name",
		"Condition": fmt.Sprintf("convert(Decimal, $input.test_input_%s.temperature) > 20", rString),
	}
	if err := testEvent(event, expectedEventData); err != nil {
		return err
	}

	actions := event.Actions
	testActionsLen := 1
	if len(actions) != testActionsLen {
		return fmt.Errorf("Actions len %d not equals to %d", len(actions), testActionsLen)
	}

	clearTimerAction := actions[0]
	clearTimerExpectedData := map[string]interface{}{
		"TimerName": "test_timer_name",
	}
	if err := testClearTimer(clearTimerAction.ClearTimer, clearTimerExpectedData); err != nil {
		return err
	}

	// Test OnExit
	onExit := state.OnExit
	events = onExit.Events
	testEventsLen = 1
	if len(events) != testEventsLen {
		return fmt.Errorf("Events len %d not equals to %d", len(events), testEventsLen)
	}

	event = events[0]
	expectedEventData = map[string]interface{}{
		"EventName": "test_event_name",
		"Condition": fmt.Sprintf("convert(Decimal, $input.test_input_%s.temperature) > 20", rString),
	}
	if err := testEvent(event, expectedEventData); err != nil {
		return err
	}

	actions = event.Actions
	testActionsLen = 1
	if len(actions) != testActionsLen {
		return fmt.Errorf("Actions len %d not equals to %d", len(actions), testActionsLen)
	}

	clearTimerAction = actions[0]
	clearTimerExpectedData = map[string]interface{}{
		"TimerName": "test_timer_name",
	}
	if err := testClearTimer(clearTimerAction.ClearTimer, clearTimerExpectedData); err != nil {
		return err
	}

	// Test OnInput
	onInput := state.OnInput
	events = onInput.Events
	testEventsLen = 1
	if len(events) != testEventsLen {
		return fmt.Errorf("Events len %d not equals to %d", len(events), testEventsLen)
	}

	event = events[0]
	expectedEventData = map[string]interface{}{
		"EventName": "test_event_name",
		"Condition": fmt.Sprintf("convert(Decimal, $input.test_input_%s.temperature) > 20", rString),
	}
	if err := testEvent(event, expectedEventData); err != nil {
		return err
	}

	actions = event.Actions
	testActionsLen = 1
	if len(actions) != testActionsLen {
		return fmt.Errorf("Actions len %d not equals to %d", len(actions), testActionsLen)
	}

	clearTimerAction = actions[0]
	clearTimerExpectedData = map[string]interface{}{
		"TimerName": "test_timer_name",
	}
	if err := testClearTimer(clearTimerAction.ClearTimer, clearTimerExpectedData); err != nil {
		return err
	}

	transitionEvents := onInput.TransitionEvents
	testTransitionEventsLen := 1
	if len(transitionEvents) != testTransitionEventsLen {
		return fmt.Errorf("Transition Events len %d not equals to %d", len(transitionEvents), testTransitionEventsLen)
	}

	transitionEvent := transitionEvents[0]
	expectedTransitionEventData := map[string]interface{}{
		"EventName": "test_transition_event_name",
		"Condition": fmt.Sprintf("convert(Decimal, $input.test_input_%s.temperature) > 20", rString),
		"NextState": "second_state",
	}

	if err := testTransitionEvent(transitionEvent, expectedTransitionEventData); err != nil {
		return err
	}

	actions = transitionEvent.Actions
	clearTimerAction = actions[0]
	clearTimerExpectedData = map[string]interface{}{
		"TimerName": "test_timer_name",
	}
	if err := testClearTimer(clearTimerAction.ClearTimer, clearTimerExpectedData); err != nil {
		return err
	}
	return nil
}

func checkAccSecondStateFull(state *iotevents.State, rString string) error {
	testStateName := "second_state"
	if *state.StateName != testStateName {
		return fmt.Errorf("State name %s not equals to %s", *state.StateName, testStateName)
	}

	// Check OnEnter
	onEnter := state.OnEnter
	events := onEnter.Events
	testEventsLen := 1
	if len(events) != testEventsLen {
		return fmt.Errorf("Events len %d not equals to %d", len(events), testEventsLen)
	}

	event := events[0]
	expectedEventData := map[string]interface{}{
		"EventName": "test_event_name",
		"Condition": fmt.Sprintf("convert(Decimal, $input.test_input_%s.temperature) > 20", rString),
	}
	if err := testEvent(event, expectedEventData); err != nil {
		return err
	}

	actions := event.Actions
	testActionsLen := 1
	if len(actions) != testActionsLen {
		return fmt.Errorf("Actions len %d not equals to %d", len(actions), testActionsLen)
	}

	clearTimerAction := actions[0]
	clearTimerExpectedData := map[string]interface{}{
		"TimerName": "test_timer_name",
	}
	if err := testClearTimer(clearTimerAction.ClearTimer, clearTimerExpectedData); err != nil {
		return err
	}

	return nil

}

func testAccDetectorModelFull(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ioteventsconn

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotevents_detector_model" {
				continue
			}

			params := &iotevents.DescribeDetectorModelInput{
				DetectorModelName: aws.String(rs.Primary.ID),
			}

			out, err := conn.DescribeDetectorModel(params)

			if err != nil {
				return err
			}

			detectorDefinition := out.DetectorModel.DetectorModelDefinition
			testInitialStateName := "first_state"
			if *detectorDefinition.InitialStateName != testInitialStateName {
				return fmt.Errorf("Initial state name %s not equals to %s", *detectorDefinition.InitialStateName, testInitialStateName)
			}

			states := detectorDefinition.States
			testStatesLen := 2
			if len(states) != testStatesLen {
				return fmt.Errorf("States len %d not equals to %d", len(states), testStatesLen)
			}

			// Test first State
			firstState := states[1]
			err = checkAccFirstStateFull(firstState, rString)
			if err != nil {
				return err
			}

			// Test second State
			secondState := states[0]
			err = checkAccSecondStateFull(secondState, rString)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func checkAWSResponseError(responseErr error, checkErr string, responseErrNilValueMessage string) error {
	// responseErr: error part of response to AWS API.
	// checkErr: expected error
	// responseErrNilValueMessage: text of error that should be returned if responseErr equals nil.
	// return: nil if responseErr equals to checkErr
	//		   error if responseErr is nil or does not equal to checkErr

	if responseErr != nil {
		if isAWSErr(responseErr, checkErr, "") {
			return nil
		} else {
			return responseErr
		}
	} else {
		return fmt.Errorf(responseErrNilValueMessage)
	}
}

func testAccCheckAWSIoTEventsDetectorModelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ioteventsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotevents_detector_model" {
			continue
		}

		params := &iotevents.DescribeDetectorModelInput{
			DetectorModelName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeDetectorModel(params)

		errNilMessage := fmt.Sprintf("Expected IoTEvents Detector Model to be destroyed, %s found", rs.Primary.ID)
		return checkAWSResponseError(err, iotevents.ErrCodeResourceNotFoundException, errNilMessage)
	}

	return nil
}

const testAccAWSIoTEventsDetectorModelBasicConfig = `
resource "aws_iam_role" "iotevents_role" {
    name = "test_role_%[1]s"
    assume_role_policy = <<EOF
{
    "Version":"2012-10-17",
    "Statement":[{
        "Effect": "Allow",
        "Principal": {
            "Service": "iotevents.amazonaws.com"
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
    roles = ["${aws_iam_role.iotevents_role.name}"]
    policy_arn = "${aws_iam_policy.policy.arn}"
}

resource "aws_iotevents_input" "test_input" {
	name = "test_input_%[1]s"

	definition {
	  attribute {
		json_path = "temperature"
	  }
	}
}
`

func testAccAWSIoTEventsDetectorModel_basic(dName string) string {
	return fmt.Sprintf(testAccAWSIoTEventsDetectorModelBasicConfig+`
resource "aws_iotevents_detector_model" "detector" {
  name = "test_detector_%[1]s"
  description = "Example detector model"
  role_arn = "${aws_iam_role.iotevents_role.arn}"

  depends_on = [
	aws_iotevents_input.test_input,
	aws_iam_policy_attachment.attach_policy,
  ]

	tags = {
		"tagKey" = "tagValue",
	}

  definition {
    initial_state_name = "first_state"

    state {
      name = "first_state"
      on_enter {
        event {
          name      = "test_event_name"
          condition = "convert(Decimal, $input.test_input_%[1]s.temperature) > 20"

          action {
            clear_timer {
              name = "test_timer_name"
            }
          }
        }
			}
			on_exit {}
			on_input {}
    }
  }
}
`, dName)
}

func testAccAWSIoTEventsDetectorModel_full(dName string) string {
	return fmt.Sprintf(testAccAWSIoTEventsDetectorModelBasicConfig+`
resource "aws_iotevents_detector_model" "detector" {
  name = "test_detector_%[1]s"
  role_arn = "${aws_iam_role.iotevents_role.arn}"

  depends_on = [
	aws_iotevents_input.test_input,
	aws_iam_policy_attachment.attach_policy,
  ]

  definition {
    initial_state_name = "first_state"

    state {
      name = "first_state"
      on_enter {
				event {
					name      = "test_event_name"
					condition = "convert(Decimal, $input.test_input_%[1]s.temperature) > 20"
			
					action {
						clear_timer {
						name = "test_timer_name"
						}
					}
				}
	  	}
      on_exit {
				event {
					name      = "test_event_name"
					condition = "convert(Decimal, $input.test_input_%[1]s.temperature) > 20"
			
					action {
						clear_timer {
						name = "test_timer_name"
						}
					}
				}	
			}
			on_input {
				event {
					name      = "test_event_name"
					condition = "convert(Decimal, $input.test_input_%[1]s.temperature) > 20"
			
					action {
						clear_timer {
							name = "test_timer_name"
						}
					}
				}
				transition_event {
					name      = "test_transition_event_name"
					condition = "convert(Decimal, $input.test_input_%[1]s.temperature) > 20"
					next_state = "second_state"

					action {
						clear_timer {
							name = "test_timer_name"
						}
					}
				}
			}
		}
	
		state {
			name = "second_state"
			on_enter {
				event {
					name      = "test_event_name"
					condition = "convert(Decimal, $input.test_input_%[1]s.temperature) > 20"
		
					action {
						clear_timer {
							name = "test_timer_name"
						}
					}
				}
			}
			on_exit {}
			on_input {}
		}
  }
}
`, dName)
}

func testAccAWSIoTEventsDetectorModel_firehose(dName string) string {
	return fmt.Sprintf(testAccAWSIoTEventsDetectorModelBasicConfig+`
resource "aws_iotevents_detector_model" "detector" {
  name = "test_detector_%[1]s"
  description = "Example detector model"
  role_arn = "${aws_iam_role.iotevents_role.arn}"

  depends_on = [
	aws_iotevents_input.test_input,
	aws_iam_policy_attachment.attach_policy,
  ]

  definition {
    initial_state_name = "first_state"

    state {
      name = "first_state"
      on_enter {
        event {
          name      = "test_event_name"
          condition = "convert(Decimal, $input.test_input_%[1]s.temperature) > 20"

          action {
            firehose {
				delivery_stream_name = "test_stream_name"
				separator = "\n"
            }
          }
        }
			}
			on_exit {}
			on_input {}
    }
  }
}
`, dName)
}

func testAccAWSIoTEventsDetectorModel_iot_events(dName string) string {
	return fmt.Sprintf(testAccAWSIoTEventsDetectorModelBasicConfig+`
resource "aws_iotevents_detector_model" "detector" {
  name = "test_detector_%[1]s"
  description = "Example detector model"
  role_arn = "${aws_iam_role.iotevents_role.arn}"

  depends_on = [
	aws_iotevents_input.test_input,
	aws_iam_policy_attachment.attach_policy,
  ]

  definition {
    initial_state_name = "first_state"

    state {
      name = "first_state"
      on_enter {
        event {
          name      = "test_event_name"
          condition = "convert(Decimal, $input.test_input_%[1]s.temperature) > 20"

          action {
            iot_events {
				      input_name = "test_input_%[1]s"
            }
          }
        }
			}
			on_exit {}
			on_input {}
    }
  }

}
`, dName)
}

func testAccAWSIoTEventsDetectorModel_iot_topic_publish(dName string) string {
	return fmt.Sprintf(testAccAWSIoTEventsDetectorModelBasicConfig+`
resource "aws_iotevents_detector_model" "detector" {
  name = "test_detector_%[1]s"
  description = "Example detector model"
  role_arn = "${aws_iam_role.iotevents_role.arn}"

  depends_on = [
	aws_iotevents_input.test_input,
	aws_iam_policy_attachment.attach_policy,
  ]

  definition {
    initial_state_name = "first_state"

    state {
      name = "first_state"
      on_enter {
        event {
          name      = "test_event_name"
          condition = "convert(Decimal, $input.test_input_%[1]s.temperature) > 20"

          action {
            iot_topic_publish {
				mqtt_topic = "test_mqtt_topic"
            }
          }
        }
			}
			on_exit {}
			on_input {}
    }
  }

}
`, dName)
}

func testAccAWSIoTEventsDetectorModel_lambda(dName string) string {
	return fmt.Sprintf(testAccAWSIoTEventsDetectorModelBasicConfig+`
resource "aws_iotevents_detector_model" "detector" {
  name = "test_detector_%[1]s"
  description = "Example detector model"
  role_arn = "${aws_iam_role.iotevents_role.arn}"

  depends_on = [
	aws_iotevents_input.test_input,
	aws_iam_policy_attachment.attach_policy,
  ]

  definition {
    initial_state_name = "first_state"

    state {
      name = "first_state"
      on_enter {
        event {
          name      = "test_event_name"
          condition = "convert(Decimal, $input.test_input_%[1]s.temperature) > 20"

          action {
            lambda {
							function_arn = "arn:aws:lambda:us-east-1:123456789012:function:ProcessKinesisRecords"
            }
          }
        }
			}
			on_exit {}
			on_input {}
    }
  }

}
`, dName)
}

func testAccAWSIoTEventsDetectorModel_reset_timer(dName string) string {
	return fmt.Sprintf(testAccAWSIoTEventsDetectorModelBasicConfig+`
resource "aws_iotevents_detector_model" "detector" {
  name = "test_detector_%[1]s"
  description = "Example detector model"
  role_arn = "${aws_iam_role.iotevents_role.arn}"

  depends_on = [
	aws_iotevents_input.test_input,
	aws_iam_policy_attachment.attach_policy,
  ]

  definition {
    initial_state_name = "first_state"

    state {
      name = "first_state"
      on_enter {
        event {
          name      = "test_event_name"
          condition = "convert(Decimal, $input.test_input_%[1]s.temperature) > 20"

          action {
            reset_timer {
							name = "test_timer_name"
            }
          }
        }
			}
			on_exit {}
			on_input {}
    }
  }

}
`, dName)
}

func testAccAWSIoTEventsDetectorModel_set_timer(dName string) string {
	return fmt.Sprintf(testAccAWSIoTEventsDetectorModelBasicConfig+`
resource "aws_iotevents_detector_model" "detector" {
  name = "test_detector_%[1]s"
  description = "Example detector model"
  role_arn = "${aws_iam_role.iotevents_role.arn}"

  depends_on = [
	aws_iotevents_input.test_input,
	aws_iam_policy_attachment.attach_policy,
  ]

  definition {
    initial_state_name = "first_state"

    state {
      name = "first_state"
      on_enter {
        event {
          name      = "test_event_name"
          condition = "convert(Decimal, $input.test_input_%[1]s.temperature) > 20"

          action {
            set_timer {
							name = "test_timer_name"
							seconds = 60
            }
          }
        }
			}
			on_exit {}
			on_input {}
    }
  }

}
`, dName)
}

func testAccAWSIoTEventsDetectorModel_set_variable(dName string) string {
	return fmt.Sprintf(testAccAWSIoTEventsDetectorModelBasicConfig+`
resource "aws_iotevents_detector_model" "detector" {
  name = "test_detector_%[1]s"
  description = "Example detector model"
  role_arn = "${aws_iam_role.iotevents_role.arn}"

  depends_on = [
	aws_iotevents_input.test_input,
	aws_iam_policy_attachment.attach_policy,
  ]

  definition {
    initial_state_name = "first_state"

    state {
      name = "first_state"
      on_enter {
        event {
          name      = "test_event_name"
          condition = "convert(Decimal, $input.test_input_%[1]s.temperature) > 20"

          action {
            set_variable {
							name = "test_variable_name"
							value = "\"val\""
            }
          }
        }
			}
			on_exit {}
			on_input {}
    }
  }

}
`, dName)
}

func testAccAWSIoTEventsDetectorModel_sns(dName string) string {
	return fmt.Sprintf(testAccAWSIoTEventsDetectorModelBasicConfig+`
resource "aws_iotevents_detector_model" "detector" {
  name = "test_detector_%[1]s"
  description = "Example detector model"
  role_arn = "${aws_iam_role.iotevents_role.arn}"

  depends_on = [
	aws_iotevents_input.test_input,
	aws_iam_policy_attachment.attach_policy,
  ]

  definition {
    initial_state_name = "first_state"

    state {
      name = "first_state"
      on_enter {
        event {
          name      = "test_event_name"
          condition = "convert(Decimal, $input.test_input_%[1]s.temperature) > 20"

          action {
            sns {
							target_arn = "arn:aws:sns:us-east-1:123456789012:my_corporate_topic"
            }
          }
        }
			}
			on_exit {}
			on_input {}
    }
  }

}
`, dName)
}

func testAccAWSIoTEventsDetectorModel_sqs(dName string) string {
	return fmt.Sprintf(testAccAWSIoTEventsDetectorModelBasicConfig+`
resource "aws_iotevents_detector_model" "detector" {
  name = "test_detector_%[1]s"
  description = "Example detector model"
  role_arn = "${aws_iam_role.iotevents_role.arn}"

  depends_on = [
	aws_iotevents_input.test_input,
	aws_iam_policy_attachment.attach_policy,
  ]

  definition {
    initial_state_name = "first_state"

    state {
      name = "first_state"
      on_enter {
        event {
          name      = "test_event_name"
          condition = "convert(Decimal, $input.test_input_%[1]s.temperature) > 20"

          action {
            sqs {
							queue_url = "fakedata"
							use_base64 = false
						}
          }
        }
			}		
			on_exit {}
			on_input {}
    }
  }
}
`, dName)
}
