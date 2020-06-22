package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsGameliftAlias() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGameliftAliasCreate,
		Read:   resourceAwsGameliftAliasRead,
		Update: resourceAwsGameliftAliasUpdate,
		Delete: resourceAwsGameliftAliasDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"routing_strategy": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fleet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"message": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								gamelift.RoutingStrategyTypeSimple,
								gamelift.RoutingStrategyTypeTerminal,
							}, false),
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsGameliftAliasCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	rs := expandGameliftRoutingStrategy(d.Get("routing_strategy").([]interface{}))
	input := gamelift.CreateAliasInput{
		Name:            aws.String(d.Get("name").(string)),
		RoutingStrategy: rs,
		Tags:            keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GameliftTags(),
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	log.Printf("[INFO] Creating Gamelift Alias: %s", input)
	out, err := conn.CreateAlias(&input)
	if err != nil {
		return err
	}

	d.SetId(*out.Alias.AliasId)

	return resourceAwsGameliftAliasRead(d, meta)
}

func resourceAwsGameliftAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[INFO] Describing Gamelift Alias: %s", d.Id())
	out, err := conn.DescribeAlias(&gamelift.DescribeAliasInput{
		AliasId: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, gamelift.ErrCodeNotFoundException, "") {
			d.SetId("")
			log.Printf("[WARN] Gamelift Alias (%s) not found, removing from state", d.Id())
			return nil
		}
		return err
	}
	a := out.Alias

	arn := aws.StringValue(a.AliasArn)
	d.Set("arn", arn)
	d.Set("description", a.Description)
	d.Set("name", a.Name)
	d.Set("routing_strategy", flattenGameliftRoutingStrategy(a.RoutingStrategy))
	tags, err := keyvaluetags.GameliftListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Game Lift Alias (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsGameliftAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	log.Printf("[INFO] Updating Gamelift Alias: %s", d.Id())
	_, err := conn.UpdateAlias(&gamelift.UpdateAliasInput{
		AliasId:         aws.String(d.Id()),
		Name:            aws.String(d.Get("name").(string)),
		Description:     aws.String(d.Get("description").(string)),
		RoutingStrategy: expandGameliftRoutingStrategy(d.Get("routing_strategy").([]interface{})),
	})
	if err != nil {
		return err
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.GameliftUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Game Lift Alias (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsGameliftAliasRead(d, meta)
}

func resourceAwsGameliftAliasDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	log.Printf("[INFO] Deleting Gamelift Alias: %s", d.Id())
	_, err := conn.DeleteAlias(&gamelift.DeleteAliasInput{
		AliasId: aws.String(d.Id()),
	})
	return err
}

func expandGameliftRoutingStrategy(cfg []interface{}) *gamelift.RoutingStrategy {
	if len(cfg) < 1 {
		return nil
	}

	strategy := cfg[0].(map[string]interface{})

	out := gamelift.RoutingStrategy{
		Type: aws.String(strategy["type"].(string)),
	}

	if v, ok := strategy["fleet_id"].(string); ok && len(v) > 0 {
		out.FleetId = aws.String(v)
	}
	if v, ok := strategy["message"].(string); ok && len(v) > 0 {
		out.Message = aws.String(v)
	}

	return &out
}

func flattenGameliftRoutingStrategy(rs *gamelift.RoutingStrategy) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.FleetId != nil {
		m["fleet_id"] = *rs.FleetId
	}
	if rs.Message != nil {
		m["message"] = *rs.Message
	}
	m["type"] = *rs.Type

	return []interface{}{m}
}
