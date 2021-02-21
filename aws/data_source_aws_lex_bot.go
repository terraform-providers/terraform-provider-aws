package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsLexBot() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLexBotRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"child_directed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"detect_sentiment": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_model_improvements": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"failure_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"idle_session_ttl_in_seconds": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"locale": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateLexBotName,
			},
			"nlu_intent_confidence_threshold": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      LexBotVersionLatest,
				ValidateFunc: validateLexBotVersion,
			},
			"voice_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsLexBotRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	botName := d.Get("name").(string)
	resp, err := conn.GetBot(&lexmodelbuildingservice.GetBotInput{
		Name:           aws.String(botName),
		VersionOrAlias: aws.String(d.Get("version").(string)),
	})
	if err != nil {
		return fmt.Errorf("error reading bot %s: %w", botName, err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "lex",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("bot:%s", d.Get("name").(string)),
	}
	d.Set("arn", arn.String())

	d.Set("checksum", resp.Checksum)
	d.Set("child_directed", resp.ChildDirected)
	d.Set("created_date", resp.CreatedDate.Format(time.RFC3339))
	d.Set("description", resp.Description)
	d.Set("detect_sentiment", resp.DetectSentiment)
	d.Set("enable_model_improvements", resp.EnableModelImprovements)
	d.Set("failure_reason", resp.FailureReason)
	d.Set("idle_session_ttl_in_seconds", resp.IdleSessionTTLInSeconds)
	d.Set("last_updated_date", resp.LastUpdatedDate.Format(time.RFC3339))
	d.Set("locale", resp.Locale)
	d.Set("name", resp.Name)
	d.Set("nlu_intent_confidence_threshold", resp.NluIntentConfidenceThreshold)
	d.Set("status", resp.Status)
	d.Set("version", resp.Version)
	d.Set("voice_id", resp.VoiceId)

	d.SetId(botName)

	return nil
}
