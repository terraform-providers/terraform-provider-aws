package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsChimeVoiceConnector() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsChimeVoiceConnectorCreate,
		ReadContext:   resourceAwsChimeVoiceConnectorRead,
		UpdateContext: resourceAwsChimeVoiceConnectorUpdate,
		DeleteContext: resourceAwsChimeVoiceConnectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"outbound_host_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"aws_region": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Default:      chime.VoiceConnectorAwsRegionUsEast1,
				ValidateFunc: validation.StringInSlice([]string{chime.VoiceConnectorAwsRegionUsEast1, chime.VoiceConnectorAwsRegionUsWest2}, false),
			},
			"require_encryption": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceAwsChimeVoiceConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	createInput := &chime.CreateVoiceConnectorInput{
		Name:              aws.String(d.Get("name").(string)),
		RequireEncryption: aws.Bool(d.Get("require_encryption").(bool)),
	}

	if v, ok := d.GetOk("aws_region"); ok {
		createInput.AwsRegion = aws.String(v.(string))
	}

	resp, err := conn.CreateVoiceConnectorWithContext(ctx, createInput)
	if err != nil {
		return diag.Errorf("Error creating Chime Voice connector: %s", err)
	}

	d.SetId(aws.StringValue(resp.VoiceConnector.VoiceConnectorId))

	return resourceAwsChimeVoiceConnectorRead(ctx, d, meta)
}

func resourceAwsChimeVoiceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	getInput := &chime.GetVoiceConnectorInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	resp, err := conn.GetVoiceConnectorWithContext(ctx, getInput)
	if isAWSErr(err, chime.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Chime Voice connector %s not found", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("Error getting Voice connector (%s): %s", d.Id(), err)
	}

	d.Set("aws_region", resp.VoiceConnector.AwsRegion)
	d.Set("outbound_host_name", resp.VoiceConnector.OutboundHostName)
	d.Set("require_encryption", resp.VoiceConnector.RequireEncryption)
	d.Set("name", resp.VoiceConnector.Name)

	return nil
}

func resourceAwsChimeVoiceConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	if d.HasChanges("name", "require_encryption") {
		updateInput := &chime.UpdateVoiceConnectorInput{
			VoiceConnectorId:  aws.String(d.Id()),
			Name:              aws.String(d.Get("name").(string)),
			RequireEncryption: aws.Bool(d.Get("require_encryption").(bool)),
		}

		if _, err := conn.UpdateVoiceConnectorWithContext(ctx, updateInput); err != nil {
			if isAWSErr(err, chime.ErrCodeNotFoundException, "") {
				log.Printf("[WARN] Chime Voice connector %s not found", d.Id())
				d.SetId("")
				return nil
			}
			return diag.Errorf("Error updating Voice connector (%s): %s", d.Id(), err)
		}
	}
	return resourceAwsChimeVoiceConnectorRead(ctx, d, meta)
}

func resourceAwsChimeVoiceConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).chimeconn

	input := &chime.DeleteVoiceConnectorInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	if _, err := conn.DeleteVoiceConnectorWithContext(ctx, input); err != nil {
		return diag.Errorf("Error deleting Voice connector (%s)", d.Id())
	}

	return nil
}
