package aws

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/storagegateway/waiter"
)

func resourceAwsStorageGatewayGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsStorageGatewayGatewayCreate,
		Read:   resourceAwsStorageGatewayGatewayRead,
		Update: resourceAwsStorageGatewayGatewayUpdate,
		Delete: resourceAwsStorageGatewayGatewayDelete,
		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("smb_active_directory_settings", func(_ context.Context, old, new, meta interface{}) bool {
				return len(old.([]interface{})) == 1 && len(new.([]interface{})) == 0
			}),
		),
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
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
				ValidateFunc:  validation.IsIPv4Address,
				ConflictsWith: []string{"activation_key"},
			},
			"gateway_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^[ -\.0-\[\]-~]*[!-\.0-\[\]-~][ -\.0-\[\]-~]*$`), ""),
					validation.StringLenBetween(2, 255),
				),
			},
			"gateway_timezone": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.Any(
					validation.StringMatch(regexp.MustCompile(`^GMT[+-][0-9]{1,2}:[0-9]{2}$`), ""),
					validation.StringMatch(regexp.MustCompile(`^GMT$`), ""),
				),
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
					"IBM-03584L32-0402",
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
							ValidateFunc: validation.All(
								validation.StringMatch(regexp.MustCompile(`^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`), ""),
								validation.StringLenBetween(1, 1024),
							),
						},
						"timeout_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      20,
							ValidateFunc: validation.IntBetween(0, 3600),
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexp.MustCompile(`^[ -~]+$`), ""),
								validation.StringLenBetween(1, 1024),
							),
						},
						"username": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexp.MustCompile(`^\w[\w\.\- ]*$`), ""),
								validation.StringLenBetween(1, 1024),
							),
						},
						"organizational_unit": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"domain_controllers": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.All(
									validation.StringMatch(regexp.MustCompile(`^(([a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9\-]*[A-Za-z0-9])(:(\d+))?$`), ""),
									validation.StringLenBetween(6, 1024),
								),
							},
						},
						"active_directory_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"smb_guest_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^[ -~]+$`), ""),
					validation.StringLenBetween(6, 512),
				),
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
			"smb_security_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(storagegateway.SMBSecurityStrategy_Values(), false),
			},
			"average_download_rate_limit_in_bits_per_sec": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(102400),
			},
			"average_upload_rate_limit_in_bits_per_sec": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(51200),
			},
			"ec2_instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_environment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"gateway_network_interface": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ipv4_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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
			return fmt.Errorf("error creating HTTP request: %w", err)
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

				return resource.NonRetryableError(fmt.Errorf("error making HTTP request: %w", err))
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
			return fmt.Errorf("error retrieving activation key from IP Address (%s): %w", gatewayIpAddress, err)
		}

		log.Printf("[DEBUG] Received HTTP response: %#v", response)
		if response.StatusCode != 302 {
			return fmt.Errorf("expected HTTP status code 302, received: %d", response.StatusCode)
		}

		redirectURL, err := response.Location()
		if err != nil {
			return fmt.Errorf("error extracting HTTP Location header: %w", err)
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
		return fmt.Errorf("error activating Storage Gateway Gateway: %w", err)
	}

	d.SetId(aws.StringValue(output.GatewayARN))

	if _, err = waiter.StorageGatewayGatewayConnected(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Storage Gateway Gateway (%q) to be Connected: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("smb_active_directory_settings"); ok && len(v.([]interface{})) > 0 {
		input := expandStorageGatewayGatewayDomain(v.([]interface{}), d.Id())
		log.Printf("[DEBUG] Storage Gateway Gateway %q joining Active Directory domain: %s", d.Id(), aws.StringValue(input.DomainName))
		_, err := conn.JoinDomain(input)
		if err != nil {
			return fmt.Errorf("error joining Active Directory domain: %w", err)
		}

		if _, err = waiter.StorageGatewayGatewayJoinDomainJoined(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for Storage Gateway Gateway (%q) to join domain (%s): %w", d.Id(), aws.StringValue(input.DomainName), err)
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
			return fmt.Errorf("error setting SMB guest password: %w", err)
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
			return fmt.Errorf("error setting CloudWatch Log Group: %w", err)
		}
	}

	if v, ok := d.GetOk("smb_security_strategy"); ok {
		input := &storagegateway.UpdateSMBSecurityStrategyInput{
			GatewayARN:          aws.String(d.Id()),
			SMBSecurityStrategy: aws.String(v.(string)),
		}

		log.Printf("[DEBUG] Storage Gateway Gateway %q setting SMB Security Strategy", input)
		_, err := conn.UpdateSMBSecurityStrategy(input)
		if err != nil {
			return fmt.Errorf("error setting SMB Security Strategy: %w", err)
		}
	}

	bandwidthInput := &storagegateway.UpdateBandwidthRateLimitInput{
		GatewayARN: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("average_download_rate_limit_in_bits_per_sec"); ok {
		bandwidthInput.AverageDownloadRateLimitInBitsPerSec = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("average_upload_rate_limit_in_bits_per_sec"); ok {
		bandwidthInput.AverageUploadRateLimitInBitsPerSec = aws.Int64(int64(v.(int)))
	}

	if bandwidthInput.AverageDownloadRateLimitInBitsPerSec != nil || bandwidthInput.AverageUploadRateLimitInBitsPerSec != nil {
		log.Printf("[DEBUG] Storage Gateway Gateway %q setting Bandwidth Rate Limit: %#v", d.Id(), bandwidthInput)
		_, err := conn.UpdateBandwidthRateLimit(bandwidthInput)
		if err != nil {
			return fmt.Errorf("error setting Bandwidth Rate Limit: %s", err)
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
		return fmt.Errorf("error reading Storage Gateway Gateway: %w", err)
	}

	if err := d.Set("tags", keyvaluetags.StoragegatewayKeyValueTags(output.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
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
		return fmt.Errorf("error reading Storage Gateway SMB Settings: %w", err)
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
			return fmt.Errorf("error setting smb_active_directory_settings: %w", err)
		}
	} else {
		m := map[string]interface{}{
			"domain_name":             aws.StringValue(smbSettingsOutput.DomainName),
			"active_directory_status": aws.StringValue(smbSettingsOutput.ActiveDirectoryStatus),
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
			m["timeout_in_seconds"] = configM["timeout_in_seconds"]

			if v, ok := configM["organizational_unit"]; ok {
				m["organizational_unit"] = v
			}

			if v, ok := configM["domain_controllers"]; ok {
				m["domain_controllers"] = v
			}
		}
		if err := d.Set("smb_active_directory_settings", []map[string]interface{}{m}); err != nil {
			return fmt.Errorf("error setting smb_active_directory_settings: %w", err)
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
	d.Set("smb_security_strategy", smbSettingsOutput.SMBSecurityStrategy)
	d.Set("ec2_instance_id", output.Ec2InstanceId)
	d.Set("endpoint_type", output.EndpointType)
	d.Set("host_environment", output.HostEnvironment)

	if err := d.Set("gateway_network_interface", flattenStorageGatewayGatewayNetworkInterfaces(output.GatewayNetworkInterfaces)); err != nil {
		return fmt.Errorf("error setting gateway_network_interface: %w", err)
	}

	bandwidthInput := &storagegateway.DescribeBandwidthRateLimitInput{
		GatewayARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Storage Gateway Bandwidth rate limit: %s", bandwidthInput)
	bandwidthOutput, err := conn.DescribeBandwidthRateLimit(bandwidthInput)
	if err != nil && !isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified operation is not supported") {
		return fmt.Errorf("error reading Storage Gateway Bandwidth rate limit: %s", err)
	}
	if err == nil {
		d.Set("average_download_rate_limit_in_bits_per_sec", aws.Int64Value(bandwidthOutput.AverageDownloadRateLimitInBitsPerSec))
		d.Set("average_upload_rate_limit_in_bits_per_sec", aws.Int64Value(bandwidthOutput.AverageUploadRateLimitInBitsPerSec))
	}

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
			return fmt.Errorf("error updating Storage Gateway Gateway: %w", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.StoragegatewayUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	if d.HasChange("smb_active_directory_settings") {
		input := expandStorageGatewayGatewayDomain(d.Get("smb_active_directory_settings").([]interface{}), d.Id())
		log.Printf("[DEBUG] Storage Gateway Gateway %q joining Active Directory domain: %s", d.Id(), aws.StringValue(input.DomainName))
		_, err := conn.JoinDomain(input)
		if err != nil {
			return fmt.Errorf("error joining Active Directory domain: %w", err)
		}

		if _, err = waiter.StorageGatewayGatewayJoinDomainJoined(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for Storage Gateway Gateway (%q) to be Join domain (%s): %w", d.Id(), aws.StringValue(input.DomainName), err)
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
			return fmt.Errorf("error setting SMB guest password: %w", err)
		}
	}

	if d.HasChange("smb_security_strategy") {
		input := &storagegateway.UpdateSMBSecurityStrategyInput{
			GatewayARN:          aws.String(d.Id()),
			SMBSecurityStrategy: aws.String(d.Get("smb_security_strategy").(string)),
		}

		log.Printf("[DEBUG] Storage Gateway Gateway %q updating SMB Security Strategy", input)
		_, err := conn.UpdateSMBSecurityStrategy(input)
		if err != nil {
			return fmt.Errorf("error updating SMB Security Strategy: %w", err)
		}
	}

	if d.HasChanges("average_download_rate_limit_in_bits_per_sec",
		"average_upload_rate_limit_in_bits_per_sec") {

		deleteInput := &storagegateway.DeleteBandwidthRateLimitInput{
			GatewayARN: aws.String(d.Id()),
		}
		updateInput := &storagegateway.UpdateBandwidthRateLimitInput{
			GatewayARN: aws.String(d.Id()),
		}
		needsDelete := false
		needsUpdate := false

		if v, ok := d.GetOk("average_download_rate_limit_in_bits_per_sec"); ok {
			updateInput.AverageDownloadRateLimitInBitsPerSec = aws.Int64(int64(v.(int)))
			needsUpdate = true
		} else if d.HasChange("average_download_rate_limit_in_bits_per_sec") {
			deleteInput.BandwidthType = aws.String("DOWNLOAD")
			needsDelete = true
		}

		if v, ok := d.GetOk("average_upload_rate_limit_in_bits_per_sec"); ok {
			updateInput.AverageUploadRateLimitInBitsPerSec = aws.Int64(int64(v.(int)))
			needsUpdate = true
		} else if d.HasChange("average_upload_rate_limit_in_bits_per_sec") {
			if needsDelete {
				deleteInput.BandwidthType = aws.String("ALL")
			} else {
				deleteInput.BandwidthType = aws.String("UPLOAD")
				needsDelete = true
			}
		}

		if needsUpdate {
			log.Printf("[DEBUG] Storage Gateway Gateway (%q) updating Bandwidth Rate Limit: %#v", d.Id(), updateInput)
			_, err := conn.UpdateBandwidthRateLimit(updateInput)
			if err != nil {
				return fmt.Errorf("error updating Bandwidth Rate Limit: %w", err)
			}
		}

		if needsDelete {
			log.Printf("[DEBUG] Storage Gateway Gateway (%q) unsetting Bandwidth Rate Limit: %#v", d.Id(), deleteInput)
			_, err := conn.DeleteBandwidthRateLimit(deleteInput)
			if err != nil {
				return fmt.Errorf("error unsetting Bandwidth Rate Limit: %w", err)
			}
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
		return fmt.Errorf("error deleting Storage Gateway Gateway: %w", err)
	}

	return nil
}

func expandStorageGatewayGatewayDomain(l []interface{}, gatewayArn string) *storagegateway.JoinDomainInput {
	if l == nil || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	domain := &storagegateway.JoinDomainInput{
		DomainName:       aws.String(tfMap["domain_name"].(string)),
		GatewayARN:       aws.String(gatewayArn),
		Password:         aws.String(tfMap["password"].(string)),
		UserName:         aws.String(tfMap["username"].(string)),
		TimeoutInSeconds: aws.Int64(int64(tfMap["timeout_in_seconds"].(int))),
	}

	if v, ok := tfMap["organizational_unit"].(string); ok && v != "" {
		domain.OrganizationalUnit = aws.String(v)
	}

	if v, ok := tfMap["domain_controllers"].(*schema.Set); ok && v.Len() > 0 {
		domain.DomainControllers = expandStringSet(v)
	}

	return domain
}

func flattenStorageGatewayGatewayNetworkInterfaces(nis []*storagegateway.NetworkInterface) []interface{} {
	if len(nis) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, ni := range nis {
		if ni == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"ipv4_address": aws.StringValue(ni.Ipv4Address),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
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
