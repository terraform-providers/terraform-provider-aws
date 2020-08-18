package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediaconnect"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

const (
	AWSMediaConnectFlowRetryTimeout       = 30 * time.Minute
	AWSMediaConnectFlowDeleteRetryTimeout = 60 * time.Minute
	AWSMediaConnectFlowRetryDelay         = 5 * time.Second
	AWSMediaConnectFlowRetryMinTimeout    = 3 * time.Second
)

func resourceAWSMediaConnectFlow() *schema.Resource {
	return &schema.Resource{
		Create: resourceAWSMediaConnectFlowCreate,
		Read:   resourceAWSMediaConnectFlowRead,
		Update: resourceAWSMediaConnectFlowUpdate,
		Delete: resourceAWSMediaConnectFlowDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(AWSMediaConnectFlowRetryTimeout),
			Update: schema.DefaultTimeout(AWSMediaConnectFlowRetryTimeout),
			Delete: schema.DefaultTimeout(AWSMediaConnectFlowDeleteRetryTimeout),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`), "must only include alphanumeric, underscore or hyphen characters"),
				),
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"egress_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(0, 64),
								validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`), "must only include alphanumeric, underscore or hyphen characters"),
							),
							ConflictsWith: []string{
								"source.0.entitlement_arn",
							},
						},
						"decryption": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"algorithm": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											mediaconnect.AlgorithmAes128,
											mediaconnect.AlgorithmAes192,
											mediaconnect.AlgorithmAes256,
										}, false),
									},
									"key_type": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.StringInSlice([]string{
											mediaconnect.KeyTypeStaticKey,
										}, false),
									},
									"role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
									"secret_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "Managed by Terraform",
							ConflictsWith: []string{
								"source.0.entitlement_arn",
							},
						},
						"entitlement_arn": {
							Type:     schema.TypeString,
							Optional: true,
							ConflictsWith: []string{
								"source.0.name",
								"source.0.description",
								"source.0.ingest_ip",
								"source.0.ingest_port",
								"source.0.max_bitrate",
								"source.0.max_latency",
								"source.0.protocol",
								"source.0.stream_id",
								"source.0.whitelist_cidr",
							},
						},
						"ingest_ip": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ingest_port": {
							Type:     schema.TypeInt,
							Optional: true,
							ConflictsWith: []string{
								"source.0.entitlement_arn",
							},
						},
						"max_bitrate": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(1, 200000000),
							ConflictsWith: []string{
								"source.0.entitlement_arn",
							},
						},
						"max_latency": {
							Type:     schema.TypeInt,
							Optional: true,
							ConflictsWith: []string{
								"source.0.entitlement_arn",
							},
						},
						"protocol": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"zixi-push",
								"rtp-fec",
								"rtp",
							}, false),
							ConflictsWith: []string{
								"source.0.entitlement_arn",
							},
						},
						"smoothing_latency": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"stream_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ConflictsWith: []string{
								"source.0.entitlement_arn",
							},
						},
						"whitelist_cidr": {
							Type:     schema.TypeString,
							Optional: true,
							ConflictsWith: []string{
								"source.0.entitlement_arn",
							},
						},
					},
				},
			},
		},
	}
}

func resourceAWSMediaConnectFlowCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediaconnectconn

	createOpts := &mediaconnect.CreateFlowInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		createOpts.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source"); ok {
		createOpts.Source = expandAWSMediaConnectFlowSource(v.([]interface{}))
	}

	log.Printf("[DEBUG] Media Connect Flow create configuration: %#v", createOpts)
	resp, err := conn.CreateFlow(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating elemental media connect flow: %s", err)
	}

	d.SetId(aws.StringValue(resp.Flow.FlowArn))

	return resourceAWSMediaConnectFlowRead(d, meta)
}

func resourceAWSMediaConnectFlowRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediaconnectconn

	descOpts := &mediaconnect.DescribeFlowInput{
		FlowArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Media Connect Flow describe configuration: %#v", descOpts)
	resp, err := conn.DescribeFlow(descOpts)
	if err != nil {
		if isAWSErr(err, mediaconnect.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] Media Connect Flow %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error describing Media Connect Flow %s: %s", d.Id(), err)
	}

	d.Set("arn", resp.Flow.FlowArn)
	d.Set("name", resp.Flow.Name)
	d.Set("description", resp.Flow.Description)
	d.Set("availability_zone", resp.Flow.AvailabilityZone)
	d.Set("egress_ip", aws.StringValue(resp.Flow.EgressIp))
	d.Set("source", flattenAWSMediaConnectFlowSource(resp.Flow.Source))

	return nil
}

func resourceAWSMediaConnectFlowUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediaconnectconn

	if d.HasChange("source") {
		err := updateAWSMediaConnectFlowSource(conn, d.Id(), d.Get("source").([]interface{}))
		if err != nil {
			if isAWSErr(err, mediaconnect.ErrCodeNotFoundException, "") {
				log.Printf("[WARN] Media Connect Flow %s not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}
			return fmt.Errorf("Error updating Media Connect Flow(%s) Source: %s", d.Id(), err)
		}
	}

	stateConf := resource.StateChangeConf{
		Pending: []string{
			mediaconnect.StatusUpdating,
			mediaconnect.StatusStarting,
			mediaconnect.StatusStopping,
		},
		Target: []string{
			mediaconnect.StatusStandby,
			mediaconnect.StatusActive,
		},
		Timeout: d.Timeout(schema.TimeoutUpdate),
		Refresh: refreshMediaConnectFlowStatus(conn, d.Id()),
	}
	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for updating Media Connect Flow(%s) Source: %s", d.Id(), err)
	}

	return resourceAWSMediaConnectFlowRead(d, meta)
}

func resourceAWSMediaConnectFlowDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediaconnectconn

	deleOpts := &mediaconnect.DeleteFlowInput{
		FlowArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Media Connect Flow delete configuration: %#v", deleOpts)
	if _, err := conn.DeleteFlow(deleOpts); err != nil {
		if isAWSErr(err, mediaconnect.ErrCodeNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting Media Connect Flow %s: %s", d.Id(), err)
	}

	return waitForDeleteMediaConnectFlow(conn, d.Id(), d.Timeout(schema.TimeoutDelete))
}

func updateAWSMediaConnectFlowSource(conn *mediaconnect.MediaConnect, flowArn string, configList []interface{}) error {
	config := configList[0].(map[string]interface{})
	updateSourceOpts := &mediaconnect.UpdateFlowSourceInput{
		FlowArn:   aws.String(flowArn),
		SourceArn: aws.String(config["arn"].(string)),
	}
	if v, ok := config["decryption"]; ok {
		updateSourceOpts.Decryption = expandAWSMediaConnectFlowSourceUpdateEncryption(v.([]interface{}))
	}
	if v, ok := config["entitlement_arn"]; ok && v.(string) != "" {
		updateSourceOpts.EntitlementArn = aws.String(v.(string))
	} else {
		if v, ok := config["description"]; ok {
			updateSourceOpts.Description = aws.String(v.(string))
		}
		if v, ok := config["ingest_port"]; ok && v.(int) != 0 {
			updateSourceOpts.IngestPort = aws.Int64(int64(v.(int)))
		}
		if v, ok := config["max_bitrate"]; ok && v.(int) != 0 {
			updateSourceOpts.MaxBitrate = aws.Int64(int64(v.(int)))
		}
		if v, ok := config["max_latency"]; ok && v.(int) != 0 {
			updateSourceOpts.MaxLatency = aws.Int64(int64(v.(int)))
		}
		if v, ok := config["protocol"]; ok && v.(string) != "" {
			updateSourceOpts.Protocol = aws.String(v.(string))
		}
		if v, ok := config["stream_id"]; ok && v.(string) != "" {
			updateSourceOpts.StreamId = aws.String(v.(string))
		}
		if v, ok := config["whitelist_cidr"]; ok && v.(string) != "" {
			updateSourceOpts.WhitelistCidr = aws.String(v.(string))
		}
	}

	log.Printf("[DEBUG] Media Connect Flow Source update configuration: %#v", updateSourceOpts)
	if _, err := conn.UpdateFlowSource(updateSourceOpts); err != nil {
		return err
	}

	return nil
}

func refreshMediaConnectFlowStatus(conn *mediaconnect.MediaConnect, flowArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeFlow(&mediaconnect.DescribeFlowInput{
			FlowArn: aws.String(flowArn),
		})
		if err != nil {
			return 42, "", err
		}
		if output == nil || output.Flow == nil {
			return nil, "", nil
		}
		flow := output.Flow

		return flow, aws.StringValue(flow.Status), nil
	}
}

func waitForDeleteMediaConnectFlow(conn *mediaconnect.MediaConnect, flowArn string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			mediaconnect.StatusStandby,
			mediaconnect.StatusActive,
			mediaconnect.StatusUpdating,
			mediaconnect.StatusDeleting,
			mediaconnect.StatusStarting,
			mediaconnect.StatusStopping,
		},
		Target:     []string{""},
		Refresh:    refreshMediaConnectFlowStatus(conn, flowArn),
		Timeout:    timeout,
		Delay:      AWSMediaConnectFlowRetryDelay,
		MinTimeout: AWSMediaConnectFlowRetryMinTimeout,
	}
	flow, err := stateConf.WaitForState()
	if err != nil {
		if isAWSErr(err, mediaconnect.ErrCodeNotFoundException, "") {
			return nil
		}
	}
	if flow == nil {
		return nil
	}
	return err
}

func expandAWSMediaConnectFlowSource(configList []interface{}) *mediaconnect.SetSourceRequest {
	if len(configList) == 0 {
		return nil
	}

	config := configList[0].(map[string]interface{})
	sourceRequest := mediaconnect.SetSourceRequest{}
	if v, ok := config["decryption"]; ok {
		sourceRequest.Decryption = expandAWSMediaConnectFlowSourceEncryption(v.([]interface{}))
	}
	if v, ok := config["entitlement_arn"]; ok && v.(string) != "" {
		sourceRequest.EntitlementArn = aws.String(v.(string))
	} else {
		if v, ok := config["name"]; ok {
			sourceRequest.Name = aws.String(v.(string))
		}
		if v, ok := config["description"]; ok {
			sourceRequest.Description = aws.String(v.(string))
		}
		if v, ok := config["ingest_port"]; ok && v.(int) != 0 {
			sourceRequest.IngestPort = aws.Int64(int64(v.(int)))
		}
		if v, ok := config["max_bitrate"]; ok && v.(int) != 0 {
			sourceRequest.MaxBitrate = aws.Int64(int64(v.(int)))
		}
		if v, ok := config["max_latency"]; ok && v.(int) != 0 {
			sourceRequest.MaxLatency = aws.Int64(int64(v.(int)))
		}
		if v, ok := config["protocol"]; ok && v.(string) != "" {
			sourceRequest.Protocol = aws.String(v.(string))
		}
		if v, ok := config["stream_id"]; ok && v.(string) != "" {
			sourceRequest.StreamId = aws.String(v.(string))
		}
		if v, ok := config["whitelist_cidr"]; ok && v.(string) != "" {
			sourceRequest.WhitelistCidr = aws.String(v.(string))
		}
	}

	return &sourceRequest
}

func expandAWSMediaConnectFlowSourceEncryption(configList []interface{}) *mediaconnect.Encryption {
	if len(configList) == 0 {
		return nil
	}
	config := configList[0].(map[string]interface{})
	encryption := mediaconnect.Encryption{}
	if v, ok := config["algorithm"]; ok {
		encryption.Algorithm = aws.String(v.(string))
	}
	if v, ok := config["key_type"]; ok {
		encryption.KeyType = aws.String(v.(string))
	}
	if v, ok := config["role_arn"]; ok {
		encryption.RoleArn = aws.String(v.(string))
	}
	if v, ok := config["secret_arn"]; ok {
		encryption.SecretArn = aws.String(v.(string))
	}

	return &encryption
}

func expandAWSMediaConnectFlowSourceUpdateEncryption(configList []interface{}) *mediaconnect.UpdateEncryption {
	if len(configList) == 0 {
		return nil
	}
	config := configList[0].(map[string]interface{})
	encryption := mediaconnect.UpdateEncryption{}
	if v, ok := config["algorithm"]; ok {
		encryption.Algorithm = aws.String(v.(string))
	}
	if v, ok := config["key_type"]; ok {
		encryption.KeyType = aws.String(v.(string))
	}
	if v, ok := config["role_arn"]; ok {
		encryption.RoleArn = aws.String(v.(string))
	}
	if v, ok := config["secret_arn"]; ok {
		encryption.SecretArn = aws.String(v.(string))
	}

	return &encryption
}

func flattenAWSMediaConnectFlowSource(source *mediaconnect.Source) []map[string]interface{} {
	if source == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"arn":               aws.StringValue(source.SourceArn),
		"name":              aws.StringValue(source.Name),
		"decryption":        flattenAWSMediaConnectFlowSourceEncryption(source.Decryption),
		"description":       aws.StringValue(source.Description),
		"entitlement_arn":   aws.StringValue(source.EntitlementArn),
		"ingest_ip":         aws.StringValue(source.IngestIp),
		"ingest_port":       aws.Int64Value(source.IngestPort),
		"max_bitrate":       aws.Int64Value(source.Transport.MaxBitrate),
		"max_latency":       aws.Int64Value(source.Transport.MaxLatency),
		"protocol":          aws.StringValue(source.Transport.Protocol),
		"smoothing_latency": aws.Int64Value(source.Transport.SmoothingLatency),
		"stream_id":         aws.StringValue(source.Transport.StreamId),
		"whitelist_cidr":    aws.StringValue(source.WhitelistCidr),
	}

	return []map[string]interface{}{m}
}

func flattenAWSMediaConnectFlowSourceEncryption(encryption *mediaconnect.Encryption) []map[string]interface{} {
	if encryption == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"algorithm":  aws.StringValue(encryption.Algorithm),
		"key_type":   aws.StringValue(encryption.KeyType),
		"role_arn":   aws.StringValue(encryption.RoleArn),
		"secret_arn": aws.StringValue(encryption.SecretArn),
	}

	return []map[string]interface{}{m}
}
