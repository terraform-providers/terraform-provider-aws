package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsFlowLog() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLogFlowCreate,
		Read:   resourceAwsLogFlowRead,
		Update: resourceAwsLogFlowUpdate,
		Delete: resourceAwsLogFlowDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"iam_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"log_destination": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"log_group_name"},
				ValidateFunc:  validateArn,
				StateFunc: func(arn interface{}) string {
					// aws_cloudwatch_log_group arn attribute references contain a trailing `:*`, which breaks functionality
					return strings.TrimSuffix(arn.(string), ":*")
				},
			},

			"log_destination_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  ec2.LogDestinationTypeCloudWatchLogs,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.LogDestinationTypeCloudWatchLogs,
					ec2.LogDestinationTypeS3,
				}, false),
			},

			"log_group_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"log_destination"},
				Deprecated:    "use 'log_destination' argument instead",
			},

			"vpc_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"subnet_id", "eni_id"},
			},

			"subnet_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"eni_id", "vpc_id"},
			},

			"eni_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"subnet_id", "vpc_id"},
			},

			"traffic_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.TrafficTypeAccept,
					ec2.TrafficTypeAll,
					ec2.TrafficTypeReject,
				}, false),
			},

			"log_format": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"max_aggregation_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Default:      600,
				ValidateFunc: validation.IntInSlice([]int{60, 600}),
			},

			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsLogFlowCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	types := []struct {
		ID   string
		Type string
	}{
		{ID: d.Get("vpc_id").(string), Type: "VPC"},
		{ID: d.Get("subnet_id").(string), Type: "Subnet"},
		{ID: d.Get("eni_id").(string), Type: "NetworkInterface"},
	}

	var resourceId string
	var resourceType string
	for _, t := range types {
		if t.ID != "" {
			resourceId = t.ID
			resourceType = t.Type
			break
		}
	}

	if resourceId == "" || resourceType == "" {
		return fmt.Errorf("Error: Flow Logs require either a VPC, Subnet, or ENI ID")
	}

	opts := &ec2.CreateFlowLogsInput{
		LogDestinationType: aws.String(d.Get("log_destination_type").(string)),
		ResourceIds:        []*string{aws.String(resourceId)},
		ResourceType:       aws.String(resourceType),
		TrafficType:        aws.String(d.Get("traffic_type").(string)),
	}

	if v, ok := d.GetOk("iam_role_arn"); ok && v != "" {
		opts.DeliverLogsPermissionArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_destination"); ok && v != "" {
		opts.LogDestination = aws.String(strings.TrimSuffix(v.(string), ":*"))
	}

	if v, ok := d.GetOk("log_group_name"); ok && v != "" {
		opts.LogGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("log_format"); ok && v != "" {
		opts.LogFormat = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_aggregation_interval"); ok {
		opts.MaxAggregationInterval = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("tags"); ok && len(v.(map[string]interface{})) > 0 {
		opts.TagSpecifications = ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), ec2.ResourceTypeVpcFlowLog)
	}

	log.Printf(
		"[DEBUG] Flow Log Create configuration: %s", opts)
	resp, err := conn.CreateFlowLogs(opts)
	if err != nil {
		return fmt.Errorf("Error creating Flow Log for (%s), error: %s", resourceId, err)
	}

	if len(resp.Unsuccessful) > 0 {
		return fmt.Errorf("Error creating Flow Log for (%s), error: %s", resourceId, *resp.Unsuccessful[0].Error.Message)
	}

	if len(resp.FlowLogIds) > 1 {
		return fmt.Errorf("Error: multiple Flow Logs created for (%s)", resourceId)
	}

	d.SetId(aws.StringValue(resp.FlowLogIds[0]))

	return resourceAwsLogFlowRead(d, meta)
}

func resourceAwsLogFlowRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	opts := &ec2.DescribeFlowLogsInput{
		FlowLogIds: []*string{aws.String(d.Id())},
	}

	resp, err := conn.DescribeFlowLogs(opts)
	if err != nil {
		log.Printf("[WARN] Flow Log (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if len(resp.FlowLogs) == 0 {
		log.Printf("[WARN] Flow Log (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	fl := resp.FlowLogs[0]
	d.Set("traffic_type", fl.TrafficType)
	d.Set("log_destination", fl.LogDestination)
	d.Set("log_destination_type", fl.LogDestinationType)
	d.Set("log_group_name", fl.LogGroupName)
	d.Set("iam_role_arn", fl.DeliverLogsPermissionArn)
	d.Set("log_format", fl.LogFormat)
	d.Set("max_aggregation_interval", fl.MaxAggregationInterval)
	var resourceKey string
	if strings.HasPrefix(*fl.ResourceId, "vpc-") {
		resourceKey = "vpc_id"
	} else if strings.HasPrefix(*fl.ResourceId, "subnet-") {
		resourceKey = "subnet_id"
	} else if strings.HasPrefix(*fl.ResourceId, "eni-") {
		resourceKey = "eni_id"
	}
	if resourceKey != "" {
		d.Set(resourceKey, fl.ResourceId)
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(fl.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "ec2",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("vpc-flow-log/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceAwsLogFlowUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsLogFlowRead(d, meta)
}

func resourceAwsLogFlowDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf(
		"[DEBUG] Flow Log Destroy: %s", d.Id())
	_, err := conn.DeleteFlowLogs(&ec2.DeleteFlowLogsInput{
		FlowLogIds: []*string{aws.String(d.Id())},
	})

	if err != nil {
		return fmt.Errorf("Error deleting Flow Log with ID (%s), error: %s", d.Id(), err)
	}

	return nil
}
