package aws

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	StorageGatewayGatewayConnected = "GatewayConnected"
)

func resourceAwsStorageGatewayGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsStorageGatewayGatewayCreate,
		Read:   resourceAwsStorageGatewayGatewayRead,
		Update: resourceAwsStorageGatewayGatewayUpdate,
		Delete: resourceAwsStorageGatewayGatewayDelete,
		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("smb_active_directory_settings", func(old, new, meta interface{}) bool {
				return len(old.([]interface{})) == 1 && len(new.([]interface{})) == 0
			}),
		),
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
				ConflictsWith: []string{"gateway_ip_address"},
			},
			"gateway_vpc_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"gateway_ip_address": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"activation_key"},
			},
			"gateway_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"gateway_timezone": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"gateway_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "STORED",
				ValidateFunc: validation.StringInSlice([]string{
					"CACHED",
					"FILE_S3",
					"STORED",
					"VTL",
				}, false),
			},
			"medium_changer_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"AWS-Gateway-VTL",
					"STK-L700",
				}, false),
			},
			"smb_active_directory_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"smb_guest_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"tape_drive_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"IBM-ULT3580-TD5",
				}, false),
			},
			"tags": tagsSchema(),
			"cloudwatch_log_group_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsStorageGatewayGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn
	region := meta.(*AWSClient).region

	activationKey := d.Get("activation_key").(string)
	gatewayIpAddress := d.Get("gateway_ip_address").(string)

	// Perform one time fetch of activation key from gateway IP address
	if activationKey == "" {
		if gatewayIpAddress == "" {
			return fmt.Errorf("either activation_key or gateway_ip_address must be provided")
		}

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: time.Second * 10,
		}

		requestURL := fmt.Sprintf("http://%s/?activationRegion=%s", gatewayIpAddress, region)
		if v, ok := d.GetOk("gateway_vpc_endpoint"); ok {
			requestURL = fmt.Sprintf("%s&vpcEndpoint=%s", requestURL, v.(string))
		}

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

			for _, retryableStatusCode := range []int{504} {
				if response.StatusCode == retryableStatusCode {
					errMessage := fmt.Errorf("status code in HTTP response: %d", response.StatusCode)
					log.Printf("[DEBUG] retryable %s", errMessage)
					return resource.RetryableError(errMessage)
				}
			}

			return nil
		})
		if isResourceTimeoutError(err) {
			response, err = client.Do(request)
		}
		if err != nil {
			return fmt.Errorf("error retrieving activation key from IP Address (%s): %s", gatewayIpAddress, err)
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
			return fmt.Errorf("empty activationKey received from IP Address: %s", gatewayIpAddress)
		}
	}

	input := &storagegateway.ActivateGatewayInput{
		ActivationKey:   aws.String(activationKey),
		GatewayRegion:   aws.String(region),
		GatewayName:     aws.String(d.Get("gateway_name").(string)),
		GatewayTimezone: aws.String(d.Get("gateway_timezone").(string)),
		GatewayType:     aws.String(d.Get("gateway_type").(string)),
		Tags:            keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().StoragegatewayTags(),
	}

	if v, ok := d.GetOk("medium_changer_type"); ok {
		input.MediumChangerType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tape_drive_type"); ok {
		input.TapeDriveType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Activating Storage Gateway Gateway: %s", input)
	output, err := conn.ActivateGateway(input)
	if err != nil {
		return fmt.Errorf("error activating Storage Gateway Gateway: %s", err)
	}

	d.SetId(aws.StringValue(output.GatewayARN))

	if _, err := WaitForStorageGatewayGatewayConnected(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Storage Gateway Gateway activation: %s", err)
	}

	if v, ok := d.GetOk("smb_active_directory_settings"); ok && len(v.([]interface{})) > 0 {
		m := v.([]interface{})[0].(map[string]interface{})

		input := &storagegateway.JoinDomainInput{
			DomainName: aws.String(m["domain_name"].(string)),
			GatewayARN: aws.String(d.Id()),
			Password:   aws.String(m["password"].(string)),
			UserName:   aws.String(m["username"].(string)),
		}

		log.Printf("[DEBUG] Storage Gateway Gateway %q joining Active Directory domain: %s", d.Id(), m["domain_name"].(string))
		_, err := conn.JoinDomain(input)
		if err != nil {
			return fmt.Errorf("error joining Active Directory domain: %s", err)
		}
	}

	if v, ok := d.GetOk("smb_guest_password"); ok && v.(string) != "" {
		input := &storagegateway.SetSMBGuestPasswordInput{
			GatewayARN: aws.String(d.Id()),
			Password:   aws.String(v.(string)),
		}

		log.Printf("[DEBUG] Storage Gateway Gateway %q setting SMB guest password", d.Id())
		_, err := conn.SetSMBGuestPassword(input)
		if err != nil {
			return fmt.Errorf("error setting SMB guest password: %s", err)
		}
	}

	if v, ok := d.GetOk("cloudwatch_log_group_arn"); ok && v.(string) != "" {
		input := &storagegateway.UpdateGatewayInformationInput{
			GatewayARN:            aws.String(d.Id()),
			CloudWatchLogGroupARN: aws.String(v.(string)),
		}

		log.Printf("[DEBUG] Storage Gateway Gateway %q setting CloudWatch Log Group", input)
		_, err := conn.UpdateGatewayInformation(input)
		if err != nil {
			return fmt.Errorf("error setting CloudWatch Log Group: %s", err)
		}
	}

	return resourceAwsStorageGatewayGatewayRead(d, meta)
}

func resourceAwsStorageGatewayGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &storagegateway.DescribeGatewayInformationInput{
		GatewayARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Storage Gateway Gateway: %s", input)

	output, err := conn.DescribeGatewayInformation(input)

	if err != nil {
		if isAWSErrStorageGatewayGatewayNotFound(err) {
			log.Printf("[WARN] Storage Gateway Gateway %q not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Storage Gateway Gateway: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.StoragegatewayKeyValueTags(output.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	smbSettingsInput := &storagegateway.DescribeSMBSettingsInput{
		GatewayARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Storage Gateway SMB Settings: %s", smbSettingsInput)
	smbSettingsOutput, err := conn.DescribeSMBSettings(smbSettingsInput)
	if err != nil && !isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "This operation is not valid for the specified gateway") {
		if isAWSErrStorageGatewayGatewayNotFound(err) {
			log.Printf("[WARN] Storage Gateway Gateway %q not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Storage Gateway SMB Settings: %s", err)
	}

	// The Storage Gateway API currently provides no way to read this value
	// We allow Terraform to passthrough the configuration value into the state
	d.Set("activation_key", d.Get("activation_key").(string))

	d.Set("arn", output.GatewayARN)
	d.Set("gateway_id", output.GatewayId)

	// The Storage Gateway API currently provides no way to read this value
	// We allow Terraform to passthrough the configuration value into the state
	d.Set("gateway_ip_address", d.Get("gateway_ip_address").(string))

	d.Set("gateway_name", output.GatewayName)
	d.Set("gateway_timezone", output.GatewayTimezone)
	d.Set("gateway_type", output.GatewayType)
	d.Set("gateway_vpc_endpoint", output.VPCEndpoint)

	// The Storage Gateway API currently provides no way to read this value
	// We allow Terraform to passthrough the configuration value into the state
	d.Set("medium_changer_type", d.Get("medium_changer_type").(string))

	// Treat the entire nested argument as a whole, based on domain name
	// to simplify schema and difference logic
	if smbSettingsOutput == nil || aws.StringValue(smbSettingsOutput.DomainName) == "" {
		if err := d.Set("smb_active_directory_settings", []interface{}{}); err != nil {
			return fmt.Errorf("error setting smb_active_directory_settings: %s", err)
		}
	} else {
		m := map[string]interface{}{
			"domain_name": aws.StringValue(smbSettingsOutput.DomainName),
			// The Storage Gateway API currently provides no way to read these values
			// "password": ...,
			// "username": ...,
		}
		// We must assemble these into the map from configuration or Terraform will enter ""
		// into state and constantly show a difference (also breaking downstream references)
		//  UPDATE: aws_storagegateway_gateway.test
		//    smb_active_directory_settings.0.password: "<sensitive>" => "<sensitive>" (attribute changed)
		//    smb_active_directory_settings.0.username: "" => "Administrator"
		if v, ok := d.GetOk("smb_active_directory_settings"); ok && len(v.([]interface{})) > 0 {
			configM := v.([]interface{})[0].(map[string]interface{})
			m["password"] = configM["password"]
			m["username"] = configM["username"]
		}
		if err := d.Set("smb_active_directory_settings", []map[string]interface{}{m}); err != nil {
			return fmt.Errorf("error setting smb_active_directory_settings: %s", err)
		}
	}

	// The Storage Gateway API currently provides no way to read this value
	// We allow Terraform to _automatically_ passthrough the configuration value into the state here
	// as the API does clue us in whether or not its actually set at all,
	// which can be used to tell Terraform to show a difference in this case
	// as well as ensuring there is some sort of attribute value (unlike the others)
	if smbSettingsOutput == nil || !aws.BoolValue(smbSettingsOutput.SMBGuestPasswordSet) {
		d.Set("smb_guest_password", "")
	}

	// The Storage Gateway API currently provides no way to read this value
	// We allow Terraform to passthrough the configuration value into the state
	d.Set("tape_drive_type", d.Get("tape_drive_type").(string))
	d.Set("cloudwatch_log_group_arn", output.CloudWatchLogGroupARN)

	return nil
}

func resourceAwsStorageGatewayGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	if d.HasChanges("gateway_name", "gateway_timezone", "cloudwatch_log_group_arn") {
		input := &storagegateway.UpdateGatewayInformationInput{
			GatewayARN:            aws.String(d.Id()),
			GatewayName:           aws.String(d.Get("gateway_name").(string)),
			GatewayTimezone:       aws.String(d.Get("gateway_timezone").(string)),
			CloudWatchLogGroupARN: aws.String(d.Get("cloudwatch_log_group_arn").(string)),
		}

		log.Printf("[DEBUG] Updating Storage Gateway Gateway: %s", input)
		_, err := conn.UpdateGatewayInformation(input)
		if err != nil {
			return fmt.Errorf("error updating Storage Gateway Gateway: %s", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.StoragegatewayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	if d.HasChange("smb_active_directory_settings") {
		l := d.Get("smb_active_directory_settings").([]interface{})
		m := l[0].(map[string]interface{})

		input := &storagegateway.JoinDomainInput{
			DomainName: aws.String(m["domain_name"].(string)),
			GatewayARN: aws.String(d.Id()),
			Password:   aws.String(m["password"].(string)),
			UserName:   aws.String(m["username"].(string)),
		}

		log.Printf("[DEBUG] Storage Gateway Gateway %q joining Active Directory domain: %s", d.Id(), m["domain_name"].(string))
		_, err := conn.JoinDomain(input)
		if err != nil {
			return fmt.Errorf("error joining Active Directory domain: %s", err)
		}
	}

	if d.HasChange("smb_guest_password") {
		input := &storagegateway.SetSMBGuestPasswordInput{
			GatewayARN: aws.String(d.Id()),
			Password:   aws.String(d.Get("smb_guest_password").(string)),
		}

		log.Printf("[DEBUG] Storage Gateway Gateway %q setting SMB guest password", d.Id())
		_, err := conn.SetSMBGuestPassword(input)
		if err != nil {
			return fmt.Errorf("error setting SMB guest password: %s", err)
		}
	}

	return resourceAwsStorageGatewayGatewayRead(d, meta)
}

func resourceAwsStorageGatewayGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	input := &storagegateway.DeleteGatewayInput{
		GatewayARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Storage Gateway Gateway: %s", input)
	_, err := conn.DeleteGateway(input)
	if err != nil {
		if isAWSErrStorageGatewayGatewayNotFound(err) {
			return nil
		}
		return fmt.Errorf("error deleting Storage Gateway Gateway: %s", err)
	}

	return nil
}

// The API returns multiple responses for a missing gateway
func isAWSErrStorageGatewayGatewayNotFound(err error) bool {
	if isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified gateway was not found.") {
		return true
	}
	if isAWSErr(err, storagegateway.ErrorCodeGatewayNotFound, "") {
		return true
	}
	return false
}

func StorageGatewayGatewayConnectedStatus(conn *storagegateway.StorageGateway, gatewayARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeGatewayInformationInput{
			GatewayARN: aws.String(gatewayARN),
		}

		output, err := conn.DescribeGatewayInformation(input)

		if isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified gateway is not connected") {
			return output, storagegateway.ErrorCodeGatewayNotConnected, nil
		}

		if err != nil {
			return output, "", err
		}

		return output, StorageGatewayGatewayConnected, nil
	}
}

func WaitForStorageGatewayGatewayConnected(conn *storagegateway.StorageGateway, gatewayARN string, timeout time.Duration) (*storagegateway.DescribeGatewayInformationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{storagegateway.ErrorCodeGatewayNotConnected},
		Target:                    []string{StorageGatewayGatewayConnected},
		Refresh:                   StorageGatewayGatewayConnectedStatus(conn, gatewayARN),
		Timeout:                   timeout,
		MinTimeout:                10 * time.Second,
		ContinuousTargetOccurence: 6, // Gateway activations can take a few seconds and can trigger a reboot of the Gateway
	}

	outputRaw, err := stateConf.WaitForState()

	switch output := outputRaw.(type) {
	case *storagegateway.DescribeGatewayInformationOutput:
		return output, err
	default:
		return nil, err
	}
}
