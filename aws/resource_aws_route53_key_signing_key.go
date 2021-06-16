package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfroute53 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53/waiter"
)

func resourceAwsRoute53KeySigningKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53KeySigningKeyCreate,
		Read:   resourceAwsRoute53KeySigningKeyRead,
		Update: resourceAwsRoute53KeySigningKeyUpdate,
		Delete: resourceAwsRoute53KeySigningKeyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"digest_algorithm_mnemonic": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"digest_algorithm_type": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"digest_value": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dnskey_record": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ds_record": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"flag": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key_management_service_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"key_tag": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 128),
					validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9._-]"), "must contain only alphanumeric characters, periods, underscores, or hyphens"),
				),
			},
			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_algorithm_mnemonic": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"signing_algorithm_type": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  tfroute53.KeySigningKeyStatusActive,
				ValidateFunc: validation.StringInSlice([]string{
					tfroute53.KeySigningKeyStatusActive,
					tfroute53.KeySigningKeyStatusInactive,
				}, false),
			},
		},
	}
}

func resourceAwsRoute53KeySigningKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	hostedZoneID := d.Get("hosted_zone_id").(string)
	name := d.Get("name").(string)
	status := d.Get("status").(string)

	input := &route53.CreateKeySigningKeyInput{
		CallerReference: aws.String(resource.UniqueId()),
		HostedZoneId:    aws.String(hostedZoneID),
		Name:            aws.String(name),
		Status:          aws.String(status),
	}

	if v, ok := d.GetOk("key_management_service_arn"); ok {
		input.KeyManagementServiceArn = aws.String(v.(string))
	}

	output, err := conn.CreateKeySigningKey(input)

	if err != nil {
		return fmt.Errorf("error creating Route 53 Key Signing Key: %w", err)
	}

	d.SetId(tfroute53.KeySigningKeyCreateResourceID(hostedZoneID, name))

	if output != nil && output.ChangeInfo != nil {
		if _, err := waiter.ChangeInfoStatusInsync(conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
			return fmt.Errorf("error waiting for Route 53 Key Signing Key (%s) creation: %w", d.Id(), err)
		}
	}

	if _, err := waiter.KeySigningKeyStatusUpdated(conn, hostedZoneID, name, status); err != nil {
		return fmt.Errorf("error waiting for Route 53 Key Signing Key (%s) status (%s): %w", d.Id(), status, err)
	}

	return resourceAwsRoute53KeySigningKeyRead(d, meta)
}

func resourceAwsRoute53KeySigningKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	hostedZoneID, name, err := tfroute53.KeySigningKeyParseResourceID(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing Route 53 Key Signing Key (%s) identifier: %w", d.Id(), err)
	}

	keySigningKey, err := finder.KeySigningKey(conn, hostedZoneID, name)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
		log.Printf("[WARN] Route 53 Key Signing Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchKeySigningKey) {
		log.Printf("[WARN] Route 53 Key Signing Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Route 53 Key Signing Key (%s): %w", d.Id(), err)
	}

	if keySigningKey == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading Route 53 Key Signing Key (%s): not found", d.Id())
		}

		log.Printf("[WARN] Route 53 Key Signing Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("digest_algorithm_mnemonic", keySigningKey.DigestAlgorithmMnemonic)
	d.Set("digest_algorithm_type", keySigningKey.DigestAlgorithmType)
	d.Set("digest_value", keySigningKey.DigestValue)
	d.Set("dnskey_record", keySigningKey.DNSKEYRecord)
	d.Set("ds_record", keySigningKey.DSRecord)
	d.Set("flag", keySigningKey.Flag)
	d.Set("hosted_zone_id", hostedZoneID)
	d.Set("key_management_service_arn", keySigningKey.KmsArn)
	d.Set("key_tag", keySigningKey.KeyTag)
	d.Set("name", keySigningKey.Name)
	d.Set("public_key", keySigningKey.PublicKey)
	d.Set("signing_algorithm_mnemonic", keySigningKey.SigningAlgorithmMnemonic)
	d.Set("signing_algorithm_type", keySigningKey.SigningAlgorithmType)
	d.Set("status", keySigningKey.Status)

	return nil
}

func resourceAwsRoute53KeySigningKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	if d.HasChange("status") {
		status := d.Get("status").(string)

		switch status {
		default:
			return fmt.Errorf("error updating Route 53 Key Signing Key (%s) status: unknown status (%s)", d.Id(), status)
		case tfroute53.KeySigningKeyStatusActive:
			input := &route53.ActivateKeySigningKeyInput{
				HostedZoneId: aws.String(d.Get("hosted_zone_id").(string)),
				Name:         aws.String(d.Get("name").(string)),
			}

			output, err := conn.ActivateKeySigningKey(input)

			if err != nil {
				return fmt.Errorf("error updating Route 53 Key Signing Key (%s) status (%s): %w", d.Id(), status, err)
			}

			if output != nil && output.ChangeInfo != nil {
				if _, err := waiter.ChangeInfoStatusInsync(conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
					return fmt.Errorf("error waiting for Route 53 Key Signing Key (%s) status (%s) update: %w", d.Id(), status, err)
				}
			}
		case tfroute53.KeySigningKeyStatusInactive:
			input := &route53.DeactivateKeySigningKeyInput{
				HostedZoneId: aws.String(d.Get("hosted_zone_id").(string)),
				Name:         aws.String(d.Get("name").(string)),
			}

			output, err := conn.DeactivateKeySigningKey(input)

			if err != nil {
				return fmt.Errorf("error updating Route 53 Key Signing Key (%s) status (%s): %w", d.Id(), status, err)
			}

			if output != nil && output.ChangeInfo != nil {
				if _, err := waiter.ChangeInfoStatusInsync(conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
					return fmt.Errorf("error waiting for Route 53 Key Signing Key (%s) status (%s) update: %w", d.Id(), status, err)
				}
			}
		}

		if _, err := waiter.KeySigningKeyStatusUpdated(conn, d.Get("hosted_zone_id").(string), d.Get("name").(string), status); err != nil {
			return fmt.Errorf("error waiting for Route 53 Key Signing Key (%s) status (%s): %w", d.Id(), status, err)
		}
	}

	return resourceAwsRoute53KeySigningKeyRead(d, meta)
}

func resourceAwsRoute53KeySigningKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53conn

	status := d.Get("status").(string)

	if status == tfroute53.KeySigningKeyStatusActive {
		input := &route53.DeactivateKeySigningKeyInput{
			HostedZoneId: aws.String(d.Get("hosted_zone_id").(string)),
			Name:         aws.String(d.Get("name").(string)),
		}

		output, err := conn.DeactivateKeySigningKey(input)

		if err != nil {
			return fmt.Errorf("error updating Route 53 Key Signing Key (%s) status (%s): %w", d.Id(), status, err)
		}

		if output != nil && output.ChangeInfo != nil {
			if _, err := waiter.ChangeInfoStatusInsync(conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
				return fmt.Errorf("error waiting for Route 53 Key Signing Key (%s) status (%s) update: %w", d.Id(), status, err)
			}
		}
	}

	input := &route53.DeleteKeySigningKeyInput{
		HostedZoneId: aws.String(d.Get("hosted_zone_id").(string)),
		Name:         aws.String(d.Get("name").(string)),
	}

	output, err := conn.DeleteKeySigningKey(input)

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchKeySigningKey) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route 53 Key Signing Key (%s): %w", d.Id(), err)
	}

	if output != nil && output.ChangeInfo != nil {
		if _, err := waiter.ChangeInfoStatusInsync(conn, aws.StringValue(output.ChangeInfo.Id)); err != nil {
			return fmt.Errorf("error waiting for Route 53 Key Signing Key (%s) deletion: %w", d.Id(), err)
		}
	}

	return nil
}
