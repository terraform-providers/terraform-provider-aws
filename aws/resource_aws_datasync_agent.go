package aws

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDataSyncAgent() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDataSyncAgentCreate,
		Read:   resourceAwsDataSyncAgentRead,
		Update: resourceAwsDataSyncAgentUpdate,
		Delete: resourceAwsDataSyncAgentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"activation_key": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"ip_address"},
			},
			"ip_address": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"activation_key"},
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsDataSyncAgentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn
	region := meta.(*AWSClient).region

	activationKey := d.Get("activation_key").(string)
	agentIpAddress := d.Get("ip_address").(string)

	// Perform one time fetch of activation key from gateway IP address
	if activationKey == "" {
		if agentIpAddress == "" {
			return fmt.Errorf("either activation_key or ip_address must be provided")
		}

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: time.Second * 10,
		}

		requestURL := fmt.Sprintf("http://%s/?gatewayType=SYNC&activationRegion=%s", agentIpAddress, region)
		log.Printf("[DEBUG] Creating HTTP request: %s", requestURL)
		request, err := http.NewRequest("GET", requestURL, nil)
		if err != nil {
			return fmt.Errorf("error creating HTTP request: %s", err)
		}

		var response *http.Response
		err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			log.Printf("[DEBUG] Making HTTP request: %s", request.URL.String())
			response, err = client.Do(request)
			if err != nil {
				if err, ok := err.(net.Error); ok {
					errMessage := fmt.Errorf("error making HTTP request: %s", err)
					log.Printf("[DEBUG] retryable %s", errMessage)
					return resource.RetryableError(errMessage)
				}
				return resource.NonRetryableError(fmt.Errorf("error making HTTP request: %s", err))
			}
			return nil
		})
		if isResourceTimeoutError(err) {
			response, err = client.Do(request)
		}
		if err != nil {
			return fmt.Errorf("error retrieving activation key from IP Address (%s): %s", agentIpAddress, err)
		}
		if response == nil {
			return fmt.Errorf("Error retrieving response for activation key request: %s", err)
		}

		log.Printf("[DEBUG] Received HTTP response: %#v", response)
		if response.StatusCode != 302 {
			return fmt.Errorf("expected HTTP status code 302, received: %d", response.StatusCode)
		}

		redirectURL, err := response.Location()
		if err != nil {
			return fmt.Errorf("error extracting HTTP Location header: %s", err)
		}

		activationKey = redirectURL.Query().Get("activationKey")

		if activationKey == "" {
			return fmt.Errorf("empty activationKey received from IP Address: %s", agentIpAddress)
		}
	}

	input := &datasync.CreateAgentInput{
		ActivationKey: aws.String(activationKey),
		Tags:          keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().DatasyncTags(),
	}

	if v, ok := d.GetOk("name"); ok {
		input.AgentName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating DataSync Agent: %s", input)
	output, err := conn.CreateAgent(input)
	if err != nil {
		return fmt.Errorf("error creating DataSync Agent: %s", err)
	}

	d.SetId(aws.StringValue(output.AgentArn))

	// Agent activations can take a few minutes
	descAgentInput := &datasync.DescribeAgentInput{
		AgentArn: aws.String(d.Id()),
	}
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := conn.DescribeAgent(descAgentInput)

		if isAWSErr(err, "InvalidRequestException", "does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DescribeAgent(descAgentInput)
	}
	if err != nil {
		return fmt.Errorf("error waiting for DataSync Agent (%s) creation: %s", d.Id(), err)
	}

	return resourceAwsDataSyncAgentRead(d, meta)
}

func resourceAwsDataSyncAgentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &datasync.DescribeAgentInput{
		AgentArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading DataSync Agent: %s", input)
	output, err := conn.DescribeAgent(input)

	if isAWSErr(err, "InvalidRequestException", "does not exist") {
		log.Printf("[WARN] DataSync Agent %q not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Agent (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.AgentArn)
	d.Set("name", output.Name)

	tags, err := keyvaluetags.DatasyncListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for DataSync Agent (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDataSyncAgentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	if d.HasChange("name") {
		input := &datasync.UpdateAgentInput{
			AgentArn: aws.String(d.Id()),
			Name:     aws.String(d.Get("name").(string)),
		}

		log.Printf("[DEBUG] Updating DataSync Agent: %s", input)
		_, err := conn.UpdateAgent(input)
		if err != nil {
			return fmt.Errorf("error updating DataSync Agent (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.DatasyncUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating DataSync Agent (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsDataSyncAgentRead(d, meta)
}

func resourceAwsDataSyncAgentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	input := &datasync.DeleteAgentInput{
		AgentArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Agent: %s", input)
	_, err := conn.DeleteAgent(input)

	if isAWSErr(err, "InvalidRequestException", "does not exist") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Agent (%s): %s", d.Id(), err)
	}

	return nil
}
