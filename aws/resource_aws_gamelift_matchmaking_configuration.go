package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsGameliftMatchmakingConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGameliftMatchmakingConfigurationCreate,
		Read:   resourceAwsGameliftMatchmakingConfigurationRead,
		Update: resourceAwsGameliftMatchmakingConfigurationUpdate,
		Delete: resourceAwsGameliftMatchmakingConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"acceptance_required": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"acceptance_timeout_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 600),
			},
			"additional_player_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"backfill_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{gamelift.BackfillModeAutomatic, gamelift.BackfillModeManual}, false),
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_event_data": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"game_property": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 16,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 32),
						},
						"value": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 96),
						},
					},
				},
			},
			"game_session_data": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 4096),
			},
			"game_session_queue_arns": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 256),
						validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9:/-]+$`), "must contain only alphanumeric characters, colon, slash and hyphens"),
					),
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-\.]*$`), "must contain only alphanumeric characters, hyphens and periods"),
				),
			},
			"notification_target": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 300),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9:_/-]*$`), "must contain only alphanumeric characters, colons, underscores, slashes and hyphens"),
				),
			},
			"request_timeout_seconds": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 43200),
			},
			"rule_set_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rule_set_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-\.]*$`), "must contain only alphanumeric characters, hyphens and periods"),
				),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsGameliftMatchmakingConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	input := gamelift.CreateMatchmakingConfigurationInput{
		AcceptanceRequired:    aws.Bool(d.Get("acceptance_required").(bool)),
		Name:                  aws.String(d.Get("name").(string)),
		RequestTimeoutSeconds: aws.Int64(int64(d.Get("request_timeout_seconds").(int))),
		RuleSetName:           aws.String(d.Get("rule_set_name").(string)),
		Tags:                  keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GameliftTags(),
	}

	if v, ok := d.GetOk("acceptance_timeout_seconds"); ok {
		input.AcceptanceTimeoutSeconds = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("additional_player_count"); ok {
		input.AdditionalPlayerCount = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("backfill_mode"); ok {
		input.BackfillMode = aws.String(v.(string))
	}
	if v, ok := d.GetOk("custom_event_data"); ok {
		input.CustomEventData = aws.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("game_property"); ok {
		set := v.(*schema.Set)
		input.GameProperties = expandGameliftGameProperties(set.List())
	}
	if v, ok := d.GetOk("game_session_data"); ok {
		input.GameSessionData = aws.String(v.(string))
	}
	if v, ok := d.GetOk("game_session_queue_arns"); ok {
		input.GameSessionQueueArns = expandStringList(v.([]interface{}))
	}
	if v, ok := d.GetOk("notification_target"); ok {
		input.NotificationTarget = aws.String(v.(string))
	}

	log.Printf("[INFO] Creating GameLift Matchmaking Configuration: %s", input)
	out, err := conn.CreateMatchmakingConfiguration(&input)
	if err != nil {
		return fmt.Errorf("error creating GameLift Matchmaking Configuration: %s", err)
	}

	d.SetId(aws.StringValue(out.Configuration.ConfigurationArn))
	return resourceAwsGameliftMatchmakingConfigurationRead(d, meta)
}

func resourceAwsGameliftMatchmakingConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn
	log.Printf("[INFO] Describing GameLift Matchmaking Configuration: %s", d.Id())
	out, err := conn.DescribeMatchmakingConfigurations(&gamelift.DescribeMatchmakingConfigurationsInput{
		Names: aws.StringSlice([]string{d.Id()}),
	})
	if err != nil {
		if isAWSErr(err, gamelift.ErrCodeInvalidRequestException, "Configuration not found") {
			log.Printf("[WARN] GameLift Matchmaking Configuration (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading GameLift Matchmaking Configuration (%s): %s", d.Id(), err)
	}
	configurations := out.Configurations

	if len(configurations) < 1 {
		log.Printf("[WARN] GameLift Matchmaking Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if len(configurations) != 1 {
		return fmt.Errorf("expected exactly 1 GameLift Matchmaking Configuration, found %d under %q",
			len(configurations), d.Id())
	}
	configuration := configurations[0]

	arn := aws.StringValue(configuration.ConfigurationArn)
	d.Set("acceptance_required", configuration.AcceptanceRequired)
	d.Set("acceptance_timeout_seconds", configuration.AcceptanceTimeoutSeconds)
	d.Set("additional_player_count", configuration.AdditionalPlayerCount)
	d.Set("arn", arn)
	d.Set("backfill_mode", configuration.BackfillMode)
	d.Set("creation_time", configuration.CreationTime.Format("2006-01-02 15:04:05"))
	d.Set("custom_event_data", configuration.CustomEventData)
	d.Set("description", configuration.Description)
	d.Set("game_property", flattenGameliftGameProperties(configuration.GameProperties))
	d.Set("game_session_data", configuration.GameSessionData)
	d.Set("game_session_queue_arns", configuration.GameSessionQueueArns)
	d.Set("name", configuration.Name)
	d.Set("notification_target", configuration.NotificationTarget)
	d.Set("request_timeout_seconds", configuration.RequestTimeoutSeconds)
	d.Set("rule_set_arn", configuration.RuleSetArn)
	d.Set("rule_set_name", configuration.RuleSetName)

	tags, err := keyvaluetags.GameliftListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for GameLift Matchmaking Configuration (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsGameliftMatchmakingConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	log.Printf("[INFO] Updating GameLift Matchmaking Configuration: %s", d.Id())

	input := gamelift.UpdateMatchmakingConfigurationInput{
		Name:                  aws.String(d.Id()),
		AcceptanceRequired:    aws.Bool(d.Get("acceptance_required").(bool)),
		RequestTimeoutSeconds: aws.Int64(int64(d.Get("request_timeout_seconds").(int))),
		RuleSetName:           aws.String(d.Get("rule_set_name").(string)),
	}

	if v, ok := d.GetOk("acceptance_timeout_seconds"); ok {
		input.AcceptanceTimeoutSeconds = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("additional_player_count"); ok {
		input.AdditionalPlayerCount = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("backfill_mode"); ok {
		input.BackfillMode = aws.String(v.(string))
	}
	if v, ok := d.GetOk("custom_event_data"); ok {
		input.CustomEventData = aws.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("game_property"); ok {
		set := v.(*schema.Set)
		input.GameProperties = expandGameliftGameProperties(set.List())
	}
	if v, ok := d.GetOk("game_session_data"); ok {
		input.GameSessionData = aws.String(v.(string))
	}
	if v, ok := d.GetOk("game_session_queue_arns"); ok {
		input.GameSessionQueueArns = expandStringList(v.([]interface{}))
	}
	if v, ok := d.GetOk("notification_target"); ok {
		input.NotificationTarget = aws.String(v.(string))
	}

	_, err := conn.UpdateMatchmakingConfiguration(&input)
	if err != nil {
		return fmt.Errorf("error updating Gamelift Matchmaking Configuration (%s): %s", d.Id(), err)
	}

	arn := d.Id()
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.GameliftUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating GameLift Matchmaking Configuration (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsGameliftMatchmakingConfigurationRead(d, meta)
}

func resourceAwsGameliftMatchmakingConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn
	log.Printf("[INFO] Deleting GameLift Matchmaking Configuration: %s", d.Id())
	_, err := conn.DeleteMatchmakingConfiguration(&gamelift.DeleteMatchmakingConfigurationInput{
		Name: aws.String(d.Id()),
	})
	if isAWSErr(err, gamelift.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting GameLift Matchmaking Configuration (%s): %s", d.Id(), err)
	}

	return nil
}

func expandGameliftGameProperties(cfg []interface{}) []*gamelift.GameProperty {
	properties := make([]*gamelift.GameProperty, len(cfg))
	for i, property := range cfg {
		prop := property.(map[string]interface{})
		properties[i] = &gamelift.GameProperty{
			Key:   aws.String(prop["key"].(string)),
			Value: aws.String(prop["value"].(string)),
		}
	}
	return properties
}

func flattenGameliftGameProperties(awsProperties []*gamelift.GameProperty) []interface{} {
	properties := []interface{}{}
	for _, awsProperty := range awsProperties {
		property := map[string]string{
			"key":   *awsProperty.Key,
			"value": *awsProperty.Value,
		}
		properties = append(properties, property)
	}
	return properties
}
