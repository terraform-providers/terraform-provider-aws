package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsGameliftMatchmakingRuleSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGameliftMatchmakingRuleSetCreate,
		Read:   resourceAwsGameliftMatchmakingRuleSetRead,
		Update: resourceAwsGameliftMatchmakingRuleSetUpdate,
		Delete: resourceAwsGameliftMatchmakingRuleSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"rule_set_body": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 65535),
					validation.StringIsJSON,
				),
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsGameliftMatchmakingRuleSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	input := gamelift.CreateMatchmakingRuleSetInput{
		Name:        aws.String(d.Get("name").(string)),
		RuleSetBody: aws.String(d.Get("rule_set_body").(string)),
		Tags:        keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GameliftTags(),
	}
	log.Printf("[INFO] Creating GameLift Matchmaking Rule Set: %s", input)
	out, err := conn.CreateMatchmakingRuleSet(&input)
	if err != nil {
		return fmt.Errorf("error creating GameLift Matchmaking Rule Set: %s", err)
	}

	d.SetId(aws.StringValue(out.RuleSet.RuleSetArn))

	return resourceAwsGameliftMatchmakingRuleSetRead(d, meta)
}

func resourceAwsGameliftMatchmakingRuleSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn
	log.Printf("[INFO] Describing GameLift Matchmaking Rule Set: %s", d.Id())
	out, err := conn.DescribeMatchmakingRuleSets(&gamelift.DescribeMatchmakingRuleSetsInput{
		Names: aws.StringSlice([]string{d.Id()}),
	})
	if err != nil {
		if isAWSErr(err, gamelift.ErrCodeInvalidRequestException, "Failed to find rule set") {
			log.Printf("[WARN] GameLift Matchmaking Rule Set (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading GameLift Matchmaking Rule Set (%s): %s", d.Id(), err)
	}
	ruleSets := out.RuleSets

	if len(ruleSets) < 1 {
		log.Printf("[WARN] GameLift Matchmaking Rule Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if len(ruleSets) != 1 {
		return fmt.Errorf("expected exactly 1 GameLift Matchmaking Rule Set, found %d under %q",
			len(ruleSets), d.Id())
	}
	ruleSet := ruleSets[0]

	arn := aws.StringValue(ruleSet.RuleSetArn)
	d.Set("arn", arn)
	d.Set("name", ruleSet.RuleSetName)
	d.Set("rule_set_body", ruleSet.RuleSetBody)

	tags, err := keyvaluetags.GameliftListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for GameLift Matchmaking Rule Set (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsGameliftMatchmakingRuleSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	log.Printf("[INFO] Updating GameLift Matchmaking Rule Set: %s", d.Id())

	arn := d.Id()
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.GameliftUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating GameLift Matchmaking Rule Set (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsGameliftMatchmakingRuleSetRead(d, meta)
}

func resourceAwsGameliftMatchmakingRuleSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn
	log.Printf("[INFO] Deleting GameLift Matchmaking Rule Set: %s", d.Id())
	_, err := conn.DeleteMatchmakingRuleSet(&gamelift.DeleteMatchmakingRuleSetInput{
		Name: aws.String(d.Get("name").(string)),
	})
	if isAWSErr(err, gamelift.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting GameLift Matchmaking Rule Set (%s): %s", d.Id(), err)
	}

	return nil
}
