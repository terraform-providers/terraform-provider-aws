package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsSsmMaintenanceWindowTarget() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsmMaintenanceWindowTargetCreate,
		Read:   resourceAwsSsmMaintenanceWindowTargetRead,
		Update: resourceAwsSsmMaintenanceWindowTargetUpdate,
		Delete: resourceAwsSsmMaintenanceWindowTargetDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected WINDOW_ID/WINDOW_TARGET_ID", d.Id())
				}
				d.Set("window_id", idParts[0])
				d.SetId(idParts[1])
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"window_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ssm.MaintenanceWindowResourceType_Values(), true),
			},

			"targets": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexp.MustCompile(`^[\p{L}\p{Z}\p{N}_.:/=\-@]*$|resource-groups:ResourceTypeFilters|resource-groups:Name$`), ""),
								validation.StringLenBetween(1, 163),
							),
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]{3,128}$`), "Only alphanumeric characters, hyphens, dots & underscores allowed"),
			},

			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 128),
			},

			"owner_information": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func resourceAwsSsmMaintenanceWindowTargetCreate(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Registering SSM Maintenance Window Target")

	params := &ssm.RegisterTargetWithMaintenanceWindowInput{
		WindowId:     aws.String(d.Get("window_id").(string)),
		ResourceType: aws.String(d.Get("resource_type").(string)),
		Targets:      expandAwsSsmTargets(d.Get("targets").([]interface{})),
	}

	if v, ok := d.GetOk("name"); ok {
		params.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("owner_information"); ok {
		params.OwnerInformation = aws.String(v.(string))
	}

	resp, err := ssmconn.RegisterTargetWithMaintenanceWindow(params)
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(resp.WindowTargetId))

	return resourceAwsSsmMaintenanceWindowTargetRead(d, meta)
}

func resourceAwsSsmMaintenanceWindowTargetRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	windowID := d.Get("window_id").(string)
	params := &ssm.DescribeMaintenanceWindowTargetsInput{
		WindowId: aws.String(windowID),
		Filters: []*ssm.MaintenanceWindowFilter{
			{
				Key:    aws.String("WindowTargetId"),
				Values: []*string{aws.String(d.Id())},
			},
		},
	}

	resp, err := ssmconn.DescribeMaintenanceWindowTargets(params)
	if isAWSErr(err, ssm.ErrCodeDoesNotExistException, "") {
		log.Printf("[WARN] Maintenance Window (%s) Target (%s) not found, removing from state", windowID, d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error getting Maintenance Window (%s) Target (%s): %w", windowID, d.Id(), err)
	}

	found := false
	for _, t := range resp.Targets {
		if aws.StringValue(t.WindowTargetId) == d.Id() {
			found = true

			d.Set("owner_information", t.OwnerInformation)
			d.Set("window_id", t.WindowId)
			d.Set("resource_type", t.ResourceType)
			d.Set("name", t.Name)
			d.Set("description", t.Description)

			if err := d.Set("targets", flattenAwsSsmTargets(t.Targets)); err != nil {
				return fmt.Errorf("Error setting targets: %w", err)
			}
		}
	}

	if !found {
		log.Printf("[INFO] Maintenance Window Target not found. Removing from state")
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsSsmMaintenanceWindowTargetUpdate(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Updating SSM Maintenance Window Target: %s", d.Id())

	params := &ssm.UpdateMaintenanceWindowTargetInput{
		Targets:        expandAwsSsmTargets(d.Get("targets").([]interface{})),
		WindowId:       aws.String(d.Get("window_id").(string)),
		WindowTargetId: aws.String(d.Id()),
	}

	if d.HasChange("name") {
		params.Name = aws.String(d.Get("name").(string))
	}

	if d.HasChange("description") {
		params.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("owner_information") {
		params.OwnerInformation = aws.String(d.Get("owner_information").(string))
	}

	_, err := ssmconn.UpdateMaintenanceWindowTarget(params)
	if err != nil {
		return fmt.Errorf("error updating SSM Maintenance Window Target (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsSsmMaintenanceWindowTargetDelete(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Deregistering SSM Maintenance Window Target: %s", d.Id())

	params := &ssm.DeregisterTargetFromMaintenanceWindowInput{
		WindowId:       aws.String(d.Get("window_id").(string)),
		WindowTargetId: aws.String(d.Id()),
	}

	_, err := ssmconn.DeregisterTargetFromMaintenanceWindow(params)
	if isAWSErr(err, ssm.ErrCodeDoesNotExistException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deregistering SSM Maintenance Window Target (%s): %w", d.Id(), err)
	}

	return nil
}
