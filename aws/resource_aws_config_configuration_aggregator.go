package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsConfigConfigurationAggregator() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsConfigConfigurationAggregatorPut,
		Read:   resourceAwsConfigConfigurationAggregatorRead,
		Update: resourceAwsConfigConfigurationAggregatorPut,
		Delete: resourceAwsConfigConfigurationAggregatorDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.Sequence(
			// This is to prevent this error:
			// All fields are ForceNew or Computed w/out Optional, Update is superfluous
			customdiff.ForceNewIfChange("account_aggregation_source", func(_ context.Context, old, new, meta interface{}) bool {
				return len(old.([]interface{})) == 0 && len(new.([]interface{})) > 0
			}),
			customdiff.ForceNewIfChange("organization_aggregation_source", func(_ context.Context, old, new, meta interface{}) bool {
				return len(old.([]interface{})) == 0 && len(new.([]interface{})) > 0
			}),
			SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"account_aggregation_source": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"organization_aggregation_source"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_ids": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateAwsAccountId,
							},
						},
						"all_regions": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
						"regions": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"organization_aggregation_source": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"account_aggregation_source"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"all_regions": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
						"regions": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func resourceAwsConfigConfigurationAggregatorPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	req := &configservice.PutConfigurationAggregatorInput{
		ConfigurationAggregatorName: aws.String(d.Get("name").(string)),
		Tags:                        tags.IgnoreAws().ConfigserviceTags(),
	}

	if v, ok := d.GetOk("account_aggregation_source"); ok && len(v.([]interface{})) > 0 {
		req.AccountAggregationSources = expandConfigAccountAggregationSources(v.([]interface{}))
	}

	if v, ok := d.GetOk("organization_aggregation_source"); ok && len(v.([]interface{})) > 0 {
		req.OrganizationAggregationSource = expandConfigOrganizationAggregationSource(v.([]interface{})[0].(map[string]interface{}))
	}

	resp, err := conn.PutConfigurationAggregator(req)
	if err != nil {
		return fmt.Errorf("error creating aggregator: %w", err)
	}

	configAgg := resp.ConfigurationAggregator
	d.SetId(aws.StringValue(configAgg.ConfigurationAggregatorName))

	if !d.IsNewResource() && d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		arn := aws.StringValue(configAgg.ConfigurationAggregatorArn)
		if err := keyvaluetags.ConfigserviceUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Config Configuration Aggregator (%s) tags: %w", arn, err)
		}
	}

	return resourceAwsConfigConfigurationAggregatorRead(d, meta)
}

func resourceAwsConfigConfigurationAggregatorRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &configservice.DescribeConfigurationAggregatorsInput{
		ConfigurationAggregatorNames: []*string{aws.String(d.Id())},
	}

	res, err := conn.DescribeConfigurationAggregators(req)
	if err != nil {
		if isAWSErr(err, configservice.ErrCodeNoSuchConfigurationAggregatorException, "") {
			log.Printf("[WARN] No such configuration aggregator (%s), removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if res == nil || len(res.ConfigurationAggregators) == 0 {
		log.Printf("[WARN] No aggregators returned (%s), removing from state", d.Id())
		d.SetId("")
		return nil
	}

	aggregator := res.ConfigurationAggregators[0]
	arn := aws.StringValue(aggregator.ConfigurationAggregatorArn)
	d.Set("arn", arn)
	d.Set("name", aggregator.ConfigurationAggregatorName)

	if err := d.Set("account_aggregation_source", flattenConfigAccountAggregationSources(aggregator.AccountAggregationSources)); err != nil {
		return fmt.Errorf("error setting account_aggregation_source: %s", err)
	}

	if err := d.Set("organization_aggregation_source", flattenConfigOrganizationAggregationSource(aggregator.OrganizationAggregationSource)); err != nil {
		return fmt.Errorf("error setting organization_aggregation_source: %s", err)
	}

	tags, err := keyvaluetags.ConfigserviceListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Config Configuration Aggregator (%s): %w", arn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsConfigConfigurationAggregatorDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).configconn

	req := &configservice.DeleteConfigurationAggregatorInput{
		ConfigurationAggregatorName: aws.String(d.Id()),
	}
	_, err := conn.DeleteConfigurationAggregator(req)

	if isAWSErr(err, configservice.ErrCodeNoSuchConfigurationAggregatorException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Config Configuration Aggregator (%s): %w", d.Id(), err)
	}

	return nil
}
