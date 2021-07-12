package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsLocationTracker() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLocationTrackerCreate,
		Read:   resourceAwsLocationTrackerRead,
		Update: resourceAwsLocationTrackerUpdate,
		Delete: resourceAwsLocationTrackerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexp.MustCompile(`[-._\w]+`), "must contain only alphanumeric characters , hyphens, periods, and underscores"),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"pricing_plan": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(locationservice.PricingPlan_Values(), false),
			},
			"pricing_plan_data_source": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Esri",
					"Here",
				}, false),
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsLocationTrackerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).locationconn
	name := d.Get("name").(string)

	req, err := getAwsLocationCreateTrackerInput(d, meta)
	log.Printf("[DEBUG] Creating Location tracker: %#v", req)
	if err != nil {
		return err
	}

	resp, err := conn.CreateTracker(&req)
	if err != nil {
		return fmt.Errorf("Error creating Location tracker (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(resp.TrackerName))

	return resourceAwsLocationTrackerRead(d, meta)
}

func resourceAwsLocationTrackerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).locationconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeTracker(&locationservice.DescribeTrackerInput{
		TrackerName: aws.String(d.Id()),
	})

	if err != nil {
		if isAWSErr(err, locationservice.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Location tracker (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Location tracker (%s): %s", d.Id(), err)
	}

	d.Set("name", resp.TrackerName)
	d.Set("description", resp.Description)
	d.Set("pricing_plan", resp.PricingPlan)
	d.Set("pricing_plan_data_source", resp.PricingPlanDataSource)
	d.Set("arn", resp.TrackerArn)
	d.Set("kms_key_id", resp.KmsKeyId)
	tags := keyvaluetags.LocationserviceKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsLocationTrackerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).locationconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.LocationserviceUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsLocationTrackerRead(d, meta)
}

func resourceAwsLocationTrackerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).locationconn

	req := &locationservice.DeleteTrackerInput{
		TrackerName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteTracker(req); err != nil {
		return fmt.Errorf("Error deleting Location tracker %s: %s", d.Id(), err)
	}

	return nil
}

func getAwsLocationCreateTrackerInput(d *schema.ResourceData, meta interface{}) (locationservice.CreateTrackerInput, error) {
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	result := locationservice.CreateTrackerInput{
		TrackerName: aws.String(d.Get("name").(string)),
		Description: aws.String(d.Get("description").(string)),
		Tags:        tags.IgnoreAws().LocationserviceTags(),
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		result.KmsKeyId = aws.String(v.(string))
	}

	pricingPlan := d.Get("pricing_plan").(string)
	result.PricingPlan = aws.String(pricingPlan)

	pricingPlanDataSource, ok := d.GetOk("pricing_plan_data_source")
	if (pricingPlan == locationservice.PricingPlanMobileAssetTracking || pricingPlan == locationservice.PricingPlanMobileAssetManagement) && !ok {
		return result, fmt.Errorf("pricing_plan_data_source is required for pricing_plan %s", pricingPlan)
	}
	if ok {
		result.PricingPlanDataSource = aws.String(pricingPlanDataSource.(string))
	}

	return result, nil
}
