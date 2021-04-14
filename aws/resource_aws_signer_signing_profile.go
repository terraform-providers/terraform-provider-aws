package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/signer"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func resourceAwsSignerSigningProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSignerSigningProfileCreate,
		Read:   resourceAwsSignerSigningProfileRead,
		Update: resourceAwsSignerSigningProfileUpdate,
		Delete: resourceAwsSignerSigningProfileDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"platform_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"AWSLambda-SHA384-ECDSA"},
					false),
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_]{0,64}$`), "must be alphanumeric with max length of 64 characters"),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_]{0,38}$`), "must be alphanumeric with max length of 38 characters"),
			},
			"signature_validity_period": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(signer.ValidityType_Values(), false),
						},
					},
				},
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"revocation_record": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"revocation_effective_from": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revoked_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revoked_by": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsSignerSigningProfileCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).signerconn

	log.Printf("[DEBUG] Creating Signer signing profile")

	profileName := naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))
	profileName = strings.Replace(profileName, "-", "_", -1)

	signingProfileInput := &signer.PutSigningProfileInput{
		ProfileName: aws.String(profileName),
		PlatformId:  aws.String(d.Get("platform_id").(string)),
	}

	if v, exists := d.GetOk("signature_validity_period"); exists {
		signatureValidityPeriod := v.([]interface{})[0].(map[string]interface{})
		signingProfileInput.SignatureValidityPeriod = &signer.SignatureValidityPeriod{
			Value: aws.Int64(int64(signatureValidityPeriod["value"].(int))),
			Type:  aws.String(signatureValidityPeriod["type"].(string)),
		}
	}

	if v, exists := d.GetOk("tags"); exists {
		signingProfileInput.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SignerTags()
	}

	_, err := conn.PutSigningProfile(signingProfileInput)
	if err != nil {
		return fmt.Errorf("error creating Signer signing profile: %s", err)
	}

	d.SetId(profileName)

	return resourceAwsSignerSigningProfileRead(d, meta)
}

func resourceAwsSignerSigningProfileRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).signerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	signingProfileOutput, err := conn.GetSigningProfile(&signer.GetSigningProfileInput{
		ProfileName: aws.String(d.Id()),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, signer.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Signer Signing Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Signer signing profile (%s): %s", d.Id(), err)
	}

	if err := d.Set("platform_id", signingProfileOutput.PlatformId); err != nil {
		return fmt.Errorf("error setting signer signing profile platform id: %s", err)
	}

	if err := d.Set("signature_validity_period", []interface{}{
		map[string]interface{}{
			"value": signingProfileOutput.SignatureValidityPeriod.Value,
			"type":  signingProfileOutput.SignatureValidityPeriod.Type,
		},
	}); err != nil {
		return fmt.Errorf("error setting signer signing profile signature validity period: %s", err)
	}

	if err := d.Set("platform_display_name", signingProfileOutput.PlatformDisplayName); err != nil {
		return fmt.Errorf("error setting signer signing profile platform display name: %s", err)
	}

	if err := d.Set("name", signingProfileOutput.ProfileName); err != nil {
		return fmt.Errorf("error setting signer signing profile name: %s", err)
	}

	if err := d.Set("arn", signingProfileOutput.Arn); err != nil {
		return fmt.Errorf("error setting signer signing profile arn: %s", err)
	}

	if err := d.Set("version", signingProfileOutput.ProfileVersion); err != nil {
		return fmt.Errorf("error setting signer signing profile version: %s", err)
	}

	if err := d.Set("version_arn", signingProfileOutput.ProfileVersionArn); err != nil {
		return fmt.Errorf("error setting signer signing profile version arn: %s", err)
	}

	if err := d.Set("status", signingProfileOutput.Status); err != nil {
		return fmt.Errorf("error setting signer signing profile status: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.SignerKeyValueTags(signingProfileOutput.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting signer signing profile tags: %s", err)
	}

	if err := d.Set("revocation_record", flattenSignerSigningProfileRevocationRecord(signingProfileOutput.RevocationRecord)); err != nil {
		return fmt.Errorf("error setting signer signing profile revocation record: %s", err)
	}

	return nil
}

func resourceAwsSignerSigningProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).signerconn

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SignerUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Signer signing profile (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsSignerSigningProfileRead(d, meta)
}

func resourceAwsSignerSigningProfileDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).signerconn

	_, err := conn.CancelSigningProfile(&signer.CancelSigningProfileInput{
		ProfileName: aws.String(d.Id()),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, signer.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error canceling Signer signing profile (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Signer signing profile %q canceled", d.Id())
	return nil
}

func flattenSignerSigningProfileRevocationRecord(apiObject *signer.SigningProfileRevocationRecord) interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.RevocationEffectiveFrom; v != nil {
		tfMap["revocation_effective_from"] = aws.TimeValue(v).Format(time.RFC3339)
	}

	if v := apiObject.RevokedAt; v != nil {
		tfMap["revoked_at"] = aws.TimeValue(v).Format(time.RFC3339)
	}

	if v := apiObject.RevokedBy; v != nil {
		tfMap["revoked_by"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}
