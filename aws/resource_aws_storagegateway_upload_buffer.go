package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/storagegateway/finder"
)

func resourceAwsStorageGatewayUploadBuffer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsStorageGatewayUploadBufferCreate,
		Read:   resourceAwsStorageGatewayUploadBufferRead,
		Delete: schema.Noop,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"disk_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"disk_id", "disk_path"},
			},
			"disk_path": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"disk_id", "disk_path"},
			},
			"gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsStorageGatewayUploadBufferCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	input := &storagegateway.AddUploadBufferInput{}

	if v, ok := d.GetOk("disk_id"); ok {
		input.DiskIds = aws.StringSlice([]string{v.(string)})
	}

	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17809
	if v, ok := d.GetOk("disk_path"); ok {
		input.DiskIds = aws.StringSlice([]string{v.(string)})
	}

	if v, ok := d.GetOk("gateway_arn"); ok {
		input.GatewayARN = aws.String(v.(string))
	}

	output, err := conn.AddUploadBuffer(input)

	if err != nil {
		return fmt.Errorf("error adding Storage Gateway upload buffer: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error adding Storage Gateway upload buffer: empty response")
	}

	if v, ok := d.GetOk("disk_id"); ok {
		d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(output.GatewayARN), v.(string)))

		return resourceAwsStorageGatewayUploadBufferRead(d, meta)
	}

	disk, err := finder.LocalDiskByDiskPath(conn, aws.StringValue(output.GatewayARN), aws.StringValue(input.DiskIds[0]))

	if err != nil {
		return fmt.Errorf("error listing Storage Gateway Local Disks after creating Upload Buffer: %w", err)
	}

	if disk == nil {
		return fmt.Errorf("error listing Storage Gateway Local Disks after creating Upload Buffer: disk not found")
	}

	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(output.GatewayARN), aws.StringValue(disk.DiskId)))

	return resourceAwsStorageGatewayUploadBufferRead(d, meta)
}

func resourceAwsStorageGatewayUploadBufferRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).storagegatewayconn

	gatewayARN, diskID, err := decodeStorageGatewayUploadBufferID(d.Id())
	if err != nil {
		return err
	}

	foundDiskID, err := finder.UploadBufferDisk(conn, gatewayARN, diskID)

	if !d.IsNewResource() && isAWSErrStorageGatewayGatewayNotFound(err) {
		log.Printf("[WARN] Storage Gateway Upload Buffer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Storage Gateway Upload Buffer (%s): %w", d.Id(), err)
	}

	if foundDiskID == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading Storage Gateway Upload Buffer (%s): not found", d.Id())
		}

		log.Printf("[WARN] Storage Gateway Upload Buffer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("disk_id", foundDiskID)
	d.Set("gateway_arn", gatewayARN)

	if _, ok := d.GetOk("disk_path"); !ok {
		disk, err := finder.LocalDiskByDiskId(conn, gatewayARN, aws.StringValue(foundDiskID))

		if err != nil {
			return fmt.Errorf("error listing Storage Gateway Local Disks: %w", err)
		}

		if disk == nil {
			return fmt.Errorf("error listing Storage Gateway Local Disks: disk not found")
		}

		d.Set("disk_path", disk.DiskPath)
	}

	return nil
}

func decodeStorageGatewayUploadBufferID(id string) (string, string, error) {
	// id = arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0
	idFormatErr := fmt.Errorf("expected ID in form of GatewayARN:DiskId, received: %s", id)
	gatewayARNAndDisk, err := arn.Parse(id)
	if err != nil {
		return "", "", idFormatErr
	}
	// gatewayARNAndDisk.Resource = gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0
	resourceParts := strings.SplitN(gatewayARNAndDisk.Resource, ":", 2)
	if len(resourceParts) != 2 {
		return "", "", idFormatErr
	}
	// resourceParts = ["gateway/sgw-12345678", "pci-0000:03:00.0-scsi-0:0:0:0"]
	gatewayARN := &arn.ARN{
		AccountID: gatewayARNAndDisk.AccountID,
		Partition: gatewayARNAndDisk.Partition,
		Region:    gatewayARNAndDisk.Region,
		Service:   gatewayARNAndDisk.Service,
		Resource:  resourceParts[0],
	}
	return gatewayARN.String(), resourceParts[1], nil
}
