package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/guardduty/waiter"
)

func resourceAwsGuardDutyPublishingDestination() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGuardDutyPublishingDestinationCreate,
		Read:   resourceAwsGuardDutyPublishingDestinationRead,
		Update: resourceAwsGuardDutyPublishingDestinationUpdate,
		Delete: resourceAwsGuardDutyPublishingDestinationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      guardduty.DestinationTypeS3,
				ValidateFunc: validation.StringInSlice(guardduty.DestinationType_Values(), false),
			},
			"destination_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"kms_key_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsGuardDutyPublishingDestinationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	detectorID := d.Get("detector_id").(string)
	input := guardduty.CreatePublishingDestinationInput{
		DetectorId: aws.String(detectorID),
		DestinationProperties: &guardduty.DestinationProperties{
			DestinationArn: aws.String(d.Get("destination_arn").(string)),
			KmsKeyArn:      aws.String(d.Get("kms_key_arn").(string)),
		},
		DestinationType: aws.String(d.Get("destination_type").(string)),
	}

	log.Printf("[DEBUG] Creating GuardDuty publishing destination: %s", input)
	output, err := conn.CreatePublishingDestination(&input)
	if err != nil {
		return fmt.Errorf("Creating GuardDuty publishing destination failed: %w", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", d.Get("detector_id"), aws.StringValue(output.DestinationId)))

	_, err = waiter.PublishingDestinationCreated(conn, aws.StringValue(output.DestinationId), detectorID)

	if err != nil {
		return fmt.Errorf("Error waiting for GuardDuty PublishingDestination status to be \"%s\": %w",
			guardduty.PublishingStatusPublishing, err)
	}

	return resourceAwsGuardDutyPublishingDestinationRead(d, meta)
}

func resourceAwsGuardDutyPublishingDestinationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	destinationId, detectorId, errStateRead := decodeGuardDutyPublishDestinationID(d.Id())

	if errStateRead != nil {
		return errStateRead
	}

	input := &guardduty.DescribePublishingDestinationInput{
		DetectorId:    aws.String(detectorId),
		DestinationId: aws.String(destinationId),
	}

	log.Printf("[DEBUG] Reading GuardDuty publishing destination: %s", input)
	gdo, err := conn.DescribePublishingDestination(input)
	if err != nil {
		if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the one or more input parameters have invalid values.") {
			log.Printf("[WARN] GuardDuty publishing destination: %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading GuardDuty publishing destination: '%s' failed: %w", d.Id(), err)
	}

	d.Set("detector_id", detectorId)
	d.Set("destination_type", gdo.DestinationType)
	d.Set("kms_key_arn", gdo.DestinationProperties.KmsKeyArn)
	d.Set("destination_arn", gdo.DestinationProperties.DestinationArn)
	return nil
}

func resourceAwsGuardDutyPublishingDestinationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	destinationId, detectorId, errStateRead := decodeGuardDutyPublishDestinationID(d.Id())

	if errStateRead != nil {
		return errStateRead
	}

	input := guardduty.UpdatePublishingDestinationInput{
		DestinationId: aws.String(destinationId),
		DetectorId:    aws.String(detectorId),
		DestinationProperties: &guardduty.DestinationProperties{
			DestinationArn: aws.String(d.Get("destination_arn").(string)),
			KmsKeyArn:      aws.String(d.Get("kms_key_arn").(string)),
		},
	}

	log.Printf("[DEBUG] Update GuardDuty publishing destination: %s", input)
	_, err := conn.UpdatePublishingDestination(&input)
	if err != nil {
		return fmt.Errorf("Updating GuardDuty publishing destination '%s' failed: %w", d.Id(), err)
	}

	return resourceAwsGuardDutyPublishingDestinationRead(d, meta)
}

func resourceAwsGuardDutyPublishingDestinationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	destinationId, detectorId, errStateRead := decodeGuardDutyPublishDestinationID(d.Id())

	if errStateRead != nil {
		return errStateRead
	}

	input := guardduty.DeletePublishingDestinationInput{
		DestinationId: aws.String(destinationId),
		DetectorId:    aws.String(detectorId),
	}

	log.Printf("[DEBUG] Delete GuardDuty publishing destination: %s", input)
	_, err := conn.DeletePublishingDestination(&input)

	if isAWSErr(err, guardduty.ErrCodeBadRequestException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Deleting GuardDuty publishing destination '%s' failed: %w", d.Id(), err)
	}

	return nil
}

func decodeGuardDutyPublishDestinationID(id string) (destinationID, detectorID string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = fmt.Errorf("GuardDuty Publishing Destination ID must be of the form <Detector ID>:<Publishing Destination ID>, was provided: %s", id)
		return
	}
	destinationID = parts[1]
	detectorID = parts[0]
	return
}
