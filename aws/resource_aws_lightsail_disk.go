package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"time"
)

func resourceAwsLightsailDisk() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLightsailDiskCreate,
		Read:   resourceAwsLightsailDiskRead,
		Delete: resourceAwsLightsailDiskDelete,
		Update: resourceAwsLightsailDiskUpdate,

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsLightsailDiskCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn
	req := &lightsail.CreateDiskInput{
		DiskName:         aws.String(d.Get("name").(string)),
		AvailabilityZone: aws.String(d.Get("availability_zone").(string)),
		SizeInGb:         aws.Int64(int64(d.Get("size").(int))),
	}

	tags := tagsFromMapLightsail(d.Get("tags").(map[string]interface{}))

	if len(tags) != 0 {
		req.Tags = tags
	}

	_, err := conn.CreateDisk(req)

	if err != nil {
		return err
	}
	// The name is unique for all resource per az so we can use it as the id
	d.SetId(d.Get("name").(string))

	return resourceAwsLightsailDiskRead(d, meta)
}

func resourceAwsLightsailDiskRead(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*AWSClient).lightsailconn
	resp, err := conn.GetDisk(&lightsail.GetDiskInput{
		DiskName: aws.String(d.Id()),
	})

	if err != nil {
		return err
	}

	if resp == nil {
		log.Printf("[WARN] Lightsail Disk (%s) not found, nil response from server, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	i := resp.Disk

	d.Set("name", i.Name)
	d.Set("availability_zone", i.Location)
	d.Set("size", i.SizeInGb)
	d.Set("created_at", i.CreatedAt.Format(time.RFC3339))
	if err := d.Set("tags", tagsToMapLightsail(i.Tags)); err != nil {
		return fmt.Errorf("Error setting tags: %s", err)
	}

	return nil
}

func resourceAwsLightsailDiskDelete(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*AWSClient).lightsailconn
	_, err := conn.DeleteDisk(&lightsail.DeleteDiskInput{
		DiskName: aws.String(d.Id()),
	})

	return err
}

func resourceAwsLightsailDiskUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn

	if d.HasChange("tags") {
		if err := setTagsLightsail(conn, d); err != nil {
			return err
		}
		d.SetPartial("tags")
	}

	return resourceAwsLightsailDiskRead(d, meta)
}
