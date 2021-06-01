package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/lex/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	LexBotAliasCreateTimeout = 1 * time.Minute
	LexBotAliasUpdateTimeout = 1 * time.Minute
	LexBotAliasDeleteTimeout = 5 * time.Minute
)

func resourceAwsLexBotAlias() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLexBotAliasCreate,
		Read:   resourceAwsLexBotAliasRead,
		Update: resourceAwsLexBotAliasUpdate,
		Delete: resourceAwsLexBotAliasDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsLexBotAliasImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(LexBotAliasCreateTimeout),
			Update: schema.DefaultTimeout(LexBotAliasUpdateTimeout),
			Delete: schema.DefaultTimeout(LexBotAliasDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bot_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateLexBotName,
			},
			"bot_version": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateLexBotVersion,
			},
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"conversation_logs": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iam_role_arn": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(20, 2048),
								validateArn,
							),
						},
						// Currently the API docs do not list a min and max for this list.
						// https://docs.aws.amazon.com/lex/latest/dg/API_PutBotAlias.html#lex-PutBotAlias-request-conversationLogs
						"log_settings": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     lexLogSettings,
						},
					},
				},
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateLexBotAliasName,
			},
		},
	}
}

var validateLexBotAliasName = validation.All(
	validation.StringLenBetween(1, 100),
	validation.StringMatch(regexp.MustCompile(`^([A-Za-z]_?)+$`), ""),
)

func resourceAwsLexBotAliasCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	botName := d.Get("bot_name").(string)
	botAliasName := d.Get("name").(string)
	id := fmt.Sprintf("%s:%s", botName, botAliasName)

	input := &lexmodelbuildingservice.PutBotAliasInput{
		BotName:     aws.String(botName),
		BotVersion:  aws.String(d.Get("bot_version").(string)),
		Description: aws.String(d.Get("description").(string)),
		Name:        aws.String(botAliasName),
	}

	if v, ok := d.GetOk("conversation_logs"); ok {
		conversationLogs, err := expandLexConversationLogs(v)
		if err != nil {
			return err
		}
		input.ConversationLogs = conversationLogs
	}

	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		output, err := conn.PutBotAlias(input)

		input.Checksum = output.Checksum
		// IAM eventual consistency
		if tfawserr.ErrMessageContains(err, lexmodelbuildingservice.ErrCodeBadRequestException, "Lex can't access your IAM role") {
			return resource.RetryableError(err)
		}
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			return resource.RetryableError(fmt.Errorf("%q bot alias still creating, another operation is pending: %w", id, err))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutBotAlias(input)
	}

	if err != nil {
		return fmt.Errorf("error creating bot alias '%s': %w", id, err)
	}

	d.SetId(id)

	return resourceAwsLexBotAliasRead(d, meta)
}

func resourceAwsLexBotAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	resp, err := conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
		BotName: aws.String(d.Get("bot_name").(string)),
		Name:    aws.String(d.Get("name").(string)),
	})
	if isAWSErr(err, lexmodelbuildingservice.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Bot alias (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error getting bot alias '%s': %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "lex",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("bot:%s", d.Id()),
	}
	d.Set("arn", arn.String())

	d.Set("bot_name", resp.BotName)
	d.Set("bot_version", resp.BotVersion)
	d.Set("checksum", resp.Checksum)
	d.Set("created_date", resp.CreatedDate.Format(time.RFC3339))
	d.Set("description", resp.Description)
	d.Set("last_updated_date", resp.LastUpdatedDate.Format(time.RFC3339))
	d.Set("name", resp.Name)

	if resp.ConversationLogs != nil {
		d.Set("conversation_logs", flattenLexConversationLogs(resp.ConversationLogs))
	}

	return nil
}

func resourceAwsLexBotAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	input := &lexmodelbuildingservice.PutBotAliasInput{
		BotName:    aws.String(d.Get("bot_name").(string)),
		BotVersion: aws.String(d.Get("bot_version").(string)),
		Checksum:   aws.String(d.Get("checksum").(string)),
		Name:       aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("conversation_logs"); ok {
		conversationLogs, err := expandLexConversationLogs(v)
		if err != nil {
			return err
		}
		input.ConversationLogs = conversationLogs
	}

	err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err := conn.PutBotAlias(input)

		// IAM eventual consistency
		if tfawserr.ErrMessageContains(err, lexmodelbuildingservice.ErrCodeBadRequestException, "Lex can't access your IAM role") {
			return resource.RetryableError(err)
		}
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			return resource.RetryableError(fmt.Errorf("%q bot alias still updating", d.Id()))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutBotAlias(input)
	}

	if err != nil {
		return fmt.Errorf("error updating bot alias '%s': %w", d.Id(), err)
	}

	return resourceAwsLexBotAliasRead(d, meta)
}

func resourceAwsLexBotAliasDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	botName := d.Get("bot_name").(string)
	botAliasName := d.Get("name").(string)

	input := &lexmodelbuildingservice.DeleteBotAliasInput{
		BotName: aws.String(botName),
		Name:    aws.String(botAliasName),
	}

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.DeleteBotAlias(input)

		if isAWSErr(err, lexmodelbuildingservice.ErrCodeConflictException, "") {
			return resource.RetryableError(fmt.Errorf("'%q': bot alias still deleting", d.Id()))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteBotAlias(input)
	}

	if err != nil {
		return fmt.Errorf("error deleting bot alias '%s': %w", d.Id(), err)
	}

	_, err = waiter.LexBotAliasDeleted(conn, botAliasName, botName)

	return err
}

func resourceAwsLexBotAliasImport(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid Lex Bot Alias resource id '%s', expected BOT_NAME:BOT_ALIAS_NAME", d.Id())
	}

	d.Set("bot_name", parts[0])
	d.Set("name", parts[1])

	return []*schema.ResourceData{d}, nil
}

var lexLogSettings = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"destination": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.Destination_Values(), false),
		},
		"kms_key_arn": {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(20, 2048),
				validateArn,
			),
		},
		"log_type": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.LogType_Values(), false),
		},
		"resource_arn": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.All(
				validation.StringLenBetween(1, 2048),
				validateArn,
			),
		},
		"resource_prefix": {
			Type:     schema.TypeString,
			Computed: true,
		},
	},
}

func flattenLexConversationLogs(response *lexmodelbuildingservice.ConversationLogsResponse) (flattened []map[string]interface{}) {
	return []map[string]interface{}{
		{
			"iam_role_arn": aws.StringValue(response.IamRoleArn),
			"log_settings": flattenLexLogSettings(response.LogSettings),
		},
	}
}

func expandLexConversationLogs(rawObject interface{}) (*lexmodelbuildingservice.ConversationLogsRequest, error) {
	request := rawObject.([]interface{})[0].(map[string]interface{})

	logSettings, err := expandLexLogSettings(request["log_settings"].(*schema.Set).List())
	if err != nil {
		return nil, err
	}
	return &lexmodelbuildingservice.ConversationLogsRequest{
		IamRoleArn:  aws.String(request["iam_role_arn"].(string)),
		LogSettings: logSettings,
	}, nil
}

func flattenLexLogSettings(responses []*lexmodelbuildingservice.LogSettingsResponse) (flattened []map[string]interface{}) {
	for _, response := range responses {
		flattened = append(flattened, map[string]interface{}{
			"destination":     response.Destination,
			"kms_key_arn":     response.KmsKeyArn,
			"log_type":        response.LogType,
			"resource_arn":    response.ResourceArn,
			"resource_prefix": response.ResourcePrefix,
		})
	}
	return
}

func expandLexLogSettings(rawValues []interface{}) ([]*lexmodelbuildingservice.LogSettingsRequest, error) {
	requests := make([]*lexmodelbuildingservice.LogSettingsRequest, 0, len(rawValues))

	for _, rawValue := range rawValues {
		value, ok := rawValue.(map[string]interface{})
		if !ok {
			continue
		}
		destination := value["destination"].(string)
		request := &lexmodelbuildingservice.LogSettingsRequest{
			Destination: aws.String(destination),
			LogType:     aws.String(value["log_type"].(string)),
			ResourceArn: aws.String(value["resource_arn"].(string)),
		}

		if v, ok := value["kms_key_arn"]; ok && v != "" {
			if destination != lexmodelbuildingservice.DestinationS3 {
				return nil, fmt.Errorf("`kms_key_arn` cannot be specified when `destination` is %q", destination)
			}
			request.KmsKeyArn = aws.String(value["kms_key_arn"].(string))
		}

		requests = append(requests, request)
	}

	return requests, nil
}
