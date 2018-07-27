package aws

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSLBListenerRollingUpdate(t *testing.T) {

	var lb elbv2.LoadBalancer
	var poller *rollingVersionPoller

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerRollingUpdate("blue"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists("aws_lb.lb", &lb),
					func(state *terraform.State) error {
						url := fmt.Sprintf("http://%s/", aws.StringValue(lb.DNSName))
						fmt.Printf("\nFinished 'blue' deployment; starting poller for %s...\n", url)
						poller = newRollingVersionPoller(url, 2)
						poller.Poll(2)
						return nil
					}),
			},
			{
				Config: testAccAWSLBListenerRollingUpdate("green"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBExists("aws_lb.lb", &lb),
					func(state *terraform.State) error {
						fmt.Printf("\nFinished 'green' deployment; stopping poller...\n")
						err := poller.Complete()
						return err
					},
				),
			},
		},
	})
}

// rollingVersionPoller exists to continuously ping a particular
// url, testing the string result returned, and treating it the 'version' for
// that endpoint.
//
// Any errors in fetching a string value from the endpoint are reflected as errors
// when the Complete method is called.
//
// The first string value returned from the endpoint is considered the 'blue' version;
// after that, the endpoint is allowed to return one other value, considered the
// 'green' version--once the endpoint has returned a value for the 'green' version,
// it will be considered an error if the endpoint returns any other value.
//
type rollingVersionPoller struct {
	url            string
	killSwitch     chan struct{}
	errors         chan error
	successes      chan string
	pollers        int
	mutex          sync.Mutex
	maxConcurrency int
}

func newRollingVersionPoller(url string, maxConcurrency int) *rollingVersionPoller {
	return &rollingVersionPoller{
		url:            url,
		maxConcurrency: maxConcurrency,
		errors:         make(chan error, maxConcurrency),
		successes:      make(chan string, maxConcurrency),
		killSwitch:     make(chan struct{}, maxConcurrency),
	}
}

// Poll causes the poller to continuously test the output of
// it's configured url with the specified number of concurrent
// clients.
func (p *rollingVersionPoller) Poll(concurrency int) error {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	errCount, err := p.waitForURLReady()
	if err != nil {
		return fmt.Errorf("Timed out waiting for url to become initially ready; %v", err)
	}
	fmt.Printf("Encountered %d errors before first success", errCount)

	for i := 0; i < concurrency; i++ {
		if p.pollers < p.maxConcurrency {
			go p.runPoller()
			p.pollers++
		} else {
			return fmt.Errorf("Already reached a maximum of %d pollers; no more will be created", p.pollers)
		}
	}
	return nil
}

// waitForURLReady continuously pings the configured url until
// it successfully returns a version value.
// The number of failures that occurred before the first success
// is also returned.
func (p *rollingVersionPoller) waitForURLReady() (int, error) {
	errCount := 0
	timeout := time.Now().Add(time.Minute * 5)
	var err error
	for time.Now().Before(timeout) {
		_, err = readVersion(p.url)
		if err == nil {
			return errCount, nil
		}
		errCount++
	}
	return errCount, err
}

func (p *rollingVersionPoller) runPoller() {
	startVersion := ""
	endVersion := ""
	startVersionCount := 0
	endVersionCount := 0
	for {
		select {
		case _ = <-p.killSwitch:
			if len(startVersion) == 0 {
				p.errors <- fmt.Errorf("Never saw a starting version")
			} else if len(endVersion) == 0 {
				p.errors <- fmt.Errorf("Never saw more than one application version")
			} else {
				p.successes <- fmt.Sprintf("Saw '%s' %d times, followed by '%s' %d times.",
					startVersion, startVersionCount, endVersion, endVersionCount)
			}
			return
		default:
			version, err := readVersion(p.url)
			if err != nil {
				err = fmt.Errorf("Error after seeing '%s' %d times; %v",
					startVersion, startVersionCount, err)
				fmt.Printf("%v", err)
				p.errors <- err
				return
			}
			if len(startVersion) == 0 {
				startVersion = version
				startVersionCount++
			} else if version == startVersion {
				startVersionCount++
			} else if len(endVersion) == 0 {
				endVersion = version
				endVersionCount++
			} else if version == endVersion {
				endVersionCount++
			} else if version != endVersion {
				p.errors <- fmt.Errorf("Started with version %s, then switched to %s, then saw unexpected version %s",
					startVersion, endVersion, version)
				return
			}
		}
	}
}

// Complete stops the poller's clients and returns
// an aggregate error describing any of the errors
// which occurred while polling the configured url
func (p *rollingVersionPoller) Complete() error {
	for i := 0; i < p.pollers; i++ {
		p.killSwitch <- struct{}{}
	}
	errors := make([]error, 0, p.maxConcurrency)
	moreResults := true
	for moreResults {
		select {
		case err := <-p.errors:
			errors = append(errors, err)
			break
		default:
			moreResults = false
			break
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("Encountered the following errors: %v", errors)
	}
	for i := 0; i < p.pollers; i++ {
		success := <-p.successes
		fmt.Printf("Poller %d: %s\n", i, success)
	}
	return nil
}

func readVersion(url string) (string, error) {
	resp, err := http.Get(url)
	if err == nil {
		if resp != nil {
			defer resp.Body.Close()
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return "", fmt.Errorf("Failed to read response body: %v", err)
			} else if resp.StatusCode != 200 {
				return "", fmt.Errorf("Received unexpected response status: %s", resp.Status)
			}
			version := string(b)
			return version, nil
		}
	}
	return "", err
}

func testAccAWSLBListenerRollingUpdate(version string) string {
	return fmt.Sprintf(`
variable "subnets" {
	default = ["10.0.1.0/24", "10.0.2.0/24"]
	type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "vpc" {
	cidr_block = "10.0.0.0/16"

	tags {
		TestName = "TestAccAWSLBRollingUpdate"
	}
}

resource "aws_internet_gateway" "gw" {
		vpc_id = "${aws_vpc.vpc.id}"
		tags {
				TestName = "TestAccAWSLBRollingUpdate"
		}
}
resource "aws_subnet" "subnets" {
	count                   = 2
	vpc_id                  = "${aws_vpc.vpc.id}"
	cidr_block              = "${element(var.subnets, count.index)}"
	map_public_ip_on_launch = true
	availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"
	tags {
		TestName = "TestAccAWSLBRollingUpdate"
	}
}

resource "aws_launch_configuration" "lc" {
	name_prefix   = "rolling-update-test-"
	image_id      = "ami-0b09cf73"
	instance_type = "t2.micro"
	user_data     = <<EOF
#cloud-config
write-files:
- path: /opt/www/index.html
	permissions: 0644
	content: "%s"
coreos:
	update:
		reboot-strategy: off
	units:
	- name: nginx.service
		enable: true
		command: start
		content: |
			[Unit]
			Description=Test app
			Requires=docker.socket
			After=docker.socket
			[Service]
			Restart=on-failure
			RestartSec=10
			Environment=DOCKER_IMAGE=nginx:1.12-alpine
			ExecStartPre=-/usr/bin/docker kill app
			ExecStartPre=-/usr/bin/docker rm app
			ExecStartPre=/usr/bin/docker pull $DOCKER_IMAGE
			ExecStart=/usr/bin/docker run --name app -v /opt/www:/usr/share/nginx/html -p 80:80 $DOCKER_IMAGE
			[Install]
			WantedBy=multi-user.target
EOF

	security_groups = [
		"${aws_security_group.sg.id}",
	]

	lifecycle {
		create_before_destroy = true
	}
}
		
resource "aws_autoscaling_group" "asg" {
	vpc_zone_identifier   = ["${aws_subnet.subnets.*.id}"]
	# using the name of the launch config causes 
	# the ASG to update each time the lc changes
	name                  = "${aws_launch_configuration.lc.name}"
	launch_configuration  = "${aws_launch_configuration.lc.name}"
	target_group_arns     = ["${aws_lb_target_group.tg.id}"]
	min_size              = "2"
	desired_capacity      = "2"
	max_size              = "2"
	health_check_type     = "ELB"

	lifecycle {
		create_before_destroy = true
	}

	tag {
		key                 = "Name"
		value               = "TestAccAWSLBRollingUpdate"
		propagate_at_launch = true
	}
	tag {
		key                 = "TestName"
		value               = "TestAccAWSLBRollingUpdate"
		propagate_at_launch = true
	}
}

resource "aws_lb_target_group" "tg" {
	name     = "${substr(aws_launch_configuration.lc.name, 0, 32)}"
	port     = 80
	protocol = "HTTP"
	vpc_id   = "${aws_vpc.vpc.id}"

	health_check {
		healthy_threshold   = 2
		unhealthy_threshold = 2
		timeout             = 5
		path                = "/"
		port                = "80"
		interval            = 10
	}

	tags {
		Name     = "TestAccAWSLBRollingUpdate"
		TestName = "TestAccAWSLBRollingUpdate"
	}

	lifecycle {
		create_before_destroy = true
	}
}

resource "aws_lb_listener" "listener" {
		load_balancer_arn = "${aws_lb.lb.id}"
		protocol          = "HTTP"
		port              = 80

		default_action {
			target_group_arn = "${aws_lb_target_group.tg.id}"
			type = "forward"
		}
}

resource "aws_lb" "lb" {
	name            = "${substr("rolling-update-test-%d", 0, 32)}"
	internal        = false
	security_groups = ["${aws_security_group.sg.id}"]
	subnets         = ["${aws_subnet.subnets.*.id}"]

	idle_timeout = 30
	enable_deletion_protection = false

	tags {
		TestName = "TestAccAWSLBRollingUpdate"
	}
}

resource "aws_security_group" "sg" {
	name        = "allow_all_lb_rolling_update"
	description = "Used for Terraform LB Testing"
	vpc_id      = "${var.vpc_id}"

	ingress {
		from_port   = 0
		to_port     = 0
		protocol    = "-1"
		cidr_blocks = ["0.0.0.0/0"]
	}

	egress {
		from_port   = 0
		to_port     = 0
		protocol    = "-1"
		cidr_blocks = ["0.0.0.0/0"]
	}

	tags {
		Name = "TestAccAWSLBRollingUpdate"
		TestName = "TestAccAWSLBRollingUpdate"
	}
}
				
`, version, time.Now().UnixNano())
}
