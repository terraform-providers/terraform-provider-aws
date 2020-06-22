package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	// There is no constant in the SDK for this resource type
	ec2ResourceTypeCapacityReservation = "capacity-reservation"
)

func resourceAwsEc2CapacityReservation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2CapacityReservationCreate,
		Read:   resourceAwsEc2CapacityReservationRead,
		Update: resourceAwsEc2CapacityReservationUpdate,
		Delete: resourceAwsEc2CapacityReservationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ebs_optimized": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"end_date": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"end_date_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ec2.EndDateTypeUnlimited,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.EndDateTypeUnlimited,
					ec2.EndDateTypeLimited,
				}, false),
			},
			"ephemeral_storage": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"instance_count": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"instance_match_criteria": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  ec2.InstanceMatchCriteriaOpen,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.InstanceMatchCriteriaOpen,
					ec2.InstanceMatchCriteriaTargeted,
				}, false),
			},
			"instance_platform": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.CapacityReservationInstancePlatformLinuxUnix,
					ec2.CapacityReservationInstancePlatformRedHatEnterpriseLinux,
					ec2.CapacityReservationInstancePlatformSuselinux,
					ec2.CapacityReservationInstancePlatformWindows,
					ec2.CapacityReservationInstancePlatformWindowswithSqlserver,
					ec2.CapacityReservationInstancePlatformWindowswithSqlserverEnterprise,
					ec2.CapacityReservationInstancePlatformWindowswithSqlserverStandard,
					ec2.CapacityReservationInstancePlatformWindowswithSqlserverWeb,
					ec2.CapacityReservationInstancePlatformLinuxwithSqlserverStandard,
					ec2.CapacityReservationInstancePlatformLinuxwithSqlserverWeb,
					ec2.CapacityReservationInstancePlatformLinuxwithSqlserverEnterprise,
				}, false),
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags": tagsSchema(),
			"tenancy": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  ec2.CapacityReservationTenancyDefault,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.CapacityReservationTenancyDefault,
					ec2.CapacityReservationTenancyDedicated,
				}, false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEc2CapacityReservationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	opts := &ec2.CreateCapacityReservationInput{
		AvailabilityZone:  aws.String(d.Get("availability_zone").(string)),
		EndDateType:       aws.String(d.Get("end_date_type").(string)),
		InstanceCount:     aws.Int64(int64(d.Get("instance_count").(int))),
		InstancePlatform:  aws.String(d.Get("instance_platform").(string)),
		InstanceType:      aws.String(d.Get("instance_type").(string)),
		TagSpecifications: ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), ec2ResourceTypeCapacityReservation),
	}

	if v, ok := d.GetOk("ebs_optimized"); ok {
		opts.EbsOptimized = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("end_date"); ok {
		t, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return fmt.Errorf("Error parsing EC2 Capacity Reservation end date: %s", err.Error())
		}
		opts.EndDate = aws.Time(t)
	}

	if v, ok := d.GetOk("ephemeral_storage"); ok {
		opts.EphemeralStorage = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("instance_match_criteria"); ok {
		opts.InstanceMatchCriteria = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tenancy"); ok {
		opts.Tenancy = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Capacity reservation: %s", opts)

	out, err := conn.CreateCapacityReservation(opts)
	if err != nil {
		return fmt.Errorf("Error creating EC2 Capacity Reservation: %s", err)
	}
	d.SetId(*out.CapacityReservation.CapacityReservationId)
	return resourceAwsEc2CapacityReservationRead(d, meta)
}

func resourceAwsEc2CapacityReservationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeCapacityReservations(&ec2.DescribeCapacityReservationsInput{
		CapacityReservationIds: []*string{aws.String(d.Id())},
	})

	if err != nil {
		if isAWSErr(err, "InvalidCapacityReservationId.NotFound", "") {
			log.Printf("[WARN] EC2 Capacity Reservation (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading EC2 Capacity Reservation %s: %s", d.Id(), err)
	}

	if resp == nil || len(resp.CapacityReservations) == 0 || resp.CapacityReservations[0] == nil {
		return fmt.Errorf("error reading EC2 Capacity Reservation (%s): empty response", d.Id())
	}

	reservation := resp.CapacityReservations[0]

	if aws.StringValue(reservation.State) == ec2.CapacityReservationStateCancelled || aws.StringValue(reservation.State) == ec2.CapacityReservationStateExpired {
		log.Printf("[WARN] EC2 Capacity Reservation (%s) no longer active, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("availability_zone", reservation.AvailabilityZone)
	d.Set("ebs_optimized", reservation.EbsOptimized)

	d.Set("end_date", "")
	if reservation.EndDate != nil {
		d.Set("end_date", aws.TimeValue(reservation.EndDate).Format(time.RFC3339))
	}

	d.Set("end_date_type", reservation.EndDateType)
	d.Set("ephemeral_storage", reservation.EphemeralStorage)
	d.Set("instance_count", reservation.TotalInstanceCount)
	d.Set("instance_match_criteria", reservation.InstanceMatchCriteria)
	d.Set("instance_platform", reservation.InstancePlatform)
	d.Set("instance_type", reservation.InstanceType)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(reservation.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("tenancy", reservation.Tenancy)

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "ec2",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("capacity-reservation/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceAwsEc2CapacityReservationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	opts := &ec2.ModifyCapacityReservationInput{
		CapacityReservationId: aws.String(d.Id()),
		EndDateType:           aws.String(d.Get("end_date_type").(string)),
		InstanceCount:         aws.Int64(int64(d.Get("instance_count").(int))),
	}

	if v, ok := d.GetOk("end_date"); ok {
		t, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return fmt.Errorf("Error parsing EC2 Capacity Reservation end date: %s", err.Error())
		}
		opts.EndDate = aws.Time(t)
	}

	log.Printf("[DEBUG] Capacity reservation: %s", opts)

	_, err := conn.ModifyCapacityReservation(opts)
	if err != nil {
		return fmt.Errorf("Error modifying EC2 Capacity Reservation: %s", err)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsEc2CapacityReservationRead(d, meta)
}

func resourceAwsEc2CapacityReservationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	opts := &ec2.CancelCapacityReservationInput{
		CapacityReservationId: aws.String(d.Id()),
	}

	_, err := conn.CancelCapacityReservation(opts)
	if err != nil {
		return fmt.Errorf("Error cancelling EC2 Capacity Reservation: %s", err)
	}

	return nil
}
