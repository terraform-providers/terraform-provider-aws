package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEc2Tag() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TagCreate,
		Read:   resourceAwsEc2TagRead,
		Update: resourceAwsEc2TagUpdate,
		Delete: resourceAwsEc2TagDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func extractResourceIDAndKeyFromEc2TagID(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid resource ID; cannot look up resource: %s", id)
	}

	return parts[0], parts[1], nil
}

func resourceAwsEc2TagCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	resourceID := d.Get("resource_id").(string)
	key := d.Get("key").(string)
	value := d.Get("value").(string)

	if err := keyvaluetags.Ec2CreateTags(conn, resourceID, map[string]string{key: value}); err != nil {
		return fmt.Errorf("error creating EC2 Tag (%s) for resource (%s): %w", key, resourceID, err)
	}

	d.SetId(fmt.Sprintf("%s,%s", resourceID, key))

	return resourceAwsEc2TagRead(d, meta)
}

func resourceAwsEc2TagRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	resourceID, key, err := extractResourceIDAndKeyFromEc2TagID(d.Id())

	if err != nil {
		return err
	}

	exists, value, err := keyvaluetags.Ec2GetTag(conn, resourceID, key)

	if err != nil {
		return fmt.Errorf("error reading EC2 Tag (%s) for resource (%s): %w", key, resourceID, err)
	}

	if !exists {
		log.Printf("[WARN] EC2 Tag (%s) for resource (%s) not found, removing from state", key, resourceID)
		d.SetId("")
		return nil
	}

	d.Set("key", key)
	d.Set("resource_id", resourceID)
	d.Set("value", value)

	return nil
}

func resourceAwsEc2TagUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	resourceID, key, err := extractResourceIDAndKeyFromEc2TagID(d.Id())

	if err != nil {
		return err
	}

	if err := keyvaluetags.Ec2UpdateTags(conn, resourceID, nil, map[string]string{key: d.Get("value").(string)}); err != nil {
		return fmt.Errorf("error updating EC2 Tag (%s) for resource (%s): %w", key, resourceID, err)
	}

	return resourceAwsEc2TagRead(d, meta)
}

func resourceAwsEc2TagDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	resourceID, key, err := extractResourceIDAndKeyFromEc2TagID(d.Id())

	if err != nil {
		return err
	}

	if err := keyvaluetags.Ec2UpdateTags(conn, resourceID, map[string]string{key: d.Get("value").(string)}, nil); err != nil {
		return fmt.Errorf("error deleting EC2 Tag (%s) for resource (%s): %w", key, resourceID, err)
	}

	return nil
}
