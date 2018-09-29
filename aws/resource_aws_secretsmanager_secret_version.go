package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsSecretsManagerSecretVersion() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecretsManagerSecretVersionCreate,
		Read:   resourceAwsSecretsManagerSecretVersionRead,
		Update: resourceAwsSecretsManagerSecretVersionUpdate,
		Delete: resourceAwsSecretsManagerSecretVersionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"secret_string": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  true,
				Sensitive: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if d.HasChange("generate_random_password") {
						return false
					}
					return true
				},
			},
			"version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_stages": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"generate_random_password": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				MaxItems:      1,
				ConflictsWith: []string{"secret_string"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exclude_characters": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"exclude_lowercase": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"exclude_numbers": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"exclude_punctuation": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"exclude_uppercase": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"include_space": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"password_length": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							Default:      32,
							ValidateFunc: validation.IntBetween(1, 4096),
						},
						"require_each_included_type": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
							Default:  true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsSecretsManagerSecretVersionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn
	secretID := d.Get("secret_id").(string)

	input := &secretsmanager.PutSecretValueInput{
		SecretId: aws.String(secretID),
	}

	if v, ok := d.GetOk("secret_string"); ok {
		input.SecretString = aws.String(v.(string))
	} else if v, ok := d.GetOk("generate_random_password"); ok {

		generate_random_password := v.([]interface{})[0].(map[string]interface{})

		param := &secretsmanager.GetRandomPasswordInput{
			RequireEachIncludedType: aws.Bool(generate_random_password["require_each_included_type"].(bool)),
			PasswordLength:          aws.Int64(int64(generate_random_password["password_length"].(int))),
		}

		if v, ok := generate_random_password["exclude_characters"]; ok {
			param.ExcludeCharacters = aws.String(v.(string))
		}
		if v, ok := generate_random_password["exclude_lowercase"]; ok {
			param.ExcludeLowercase = aws.Bool(v.(bool))
		}
		if v, ok := generate_random_password["exclude_numbers"]; ok {
			param.ExcludeNumbers = aws.Bool(v.(bool))
		}
		if v, ok := generate_random_password["exclude_punctuation"]; ok {
			param.ExcludePunctuation = aws.Bool(v.(bool))
		}
		if v, ok := generate_random_password["exclude_uppercase"]; ok {
			param.ExcludeUppercase = aws.Bool(v.(bool))
		}
		if v, ok := generate_random_password["include_space"]; ok {
			param.IncludeSpace = aws.Bool(v.(bool))
		}

		resp, err := conn.GetRandomPassword(param)
		if err != nil {
			return fmt.Errorf("error getting random password: %s", err)
		}
		randomPassword := aws.StringValue(resp.RandomPassword)
		log.Printf("[DEBUG] Generated random password : %s", randomPassword)
		input.SecretString = aws.String(randomPassword)
	}

	if v, ok := d.GetOk("version_stages"); ok {
		input.VersionStages = expandStringList(v.(*schema.Set).List())
	}

	log.Printf("[DEBUG] Putting Secrets Manager Secret %q value", secretID)
	output, err := conn.PutSecretValue(input)
	if err != nil {
		return fmt.Errorf("error putting Secrets Manager Secret value: %s", err)
	}

	d.SetId(fmt.Sprintf("%s|%s", secretID, aws.StringValue(output.VersionId)))

	return resourceAwsSecretsManagerSecretVersionRead(d, meta)
}

func resourceAwsSecretsManagerSecretVersionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn

	secretID, versionID, err := decodeSecretsManagerSecretVersionID(d.Id())
	if err != nil {
		return err
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId:  aws.String(secretID),
		VersionId: aws.String(versionID),
	}

	log.Printf("[DEBUG] Reading Secrets Manager Secret Version: %s", input)
	output, err := conn.GetSecretValue(input)
	if err != nil {
		if isAWSErr(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Secrets Manager Secret Version %q not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		if isAWSErr(err, secretsmanager.ErrCodeInvalidRequestException, "You can’t perform this operation on the secret because it was deleted") {
			log.Printf("[WARN] Secrets Manager Secret Version %q not found - removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Secrets Manager Secret Version: %s", err)
	}

	d.Set("secret_id", secretID)
	d.Set("secret_string", output.SecretString)
	d.Set("version_id", output.VersionId)
	d.Set("arn", output.ARN)

	if err := d.Set("version_stages", flattenStringList(output.VersionStages)); err != nil {
		return fmt.Errorf("error setting version_stages: %s", err)
	}

	return nil
}

func resourceAwsSecretsManagerSecretVersionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn

	secretID, versionID, err := decodeSecretsManagerSecretVersionID(d.Id())
	if err != nil {
		return err
	}

	o, n := d.GetChange("version_stages")
	os := o.(*schema.Set)
	ns := n.(*schema.Set)
	stagesToAdd := ns.Difference(os).List()
	stagesToRemove := os.Difference(ns).List()

	for _, stage := range stagesToAdd {
		input := &secretsmanager.UpdateSecretVersionStageInput{
			MoveToVersionId: aws.String(versionID),
			SecretId:        aws.String(secretID),
			VersionStage:    aws.String(stage.(string)),
		}

		log.Printf("[DEBUG] Updating Secrets Manager Secret Version Stage: %s", input)
		_, err := conn.UpdateSecretVersionStage(input)
		if err != nil {
			return fmt.Errorf("error updating Secrets Manager Secret %q Version Stage %q: %s", secretID, stage.(string), err)
		}
	}

	for _, stage := range stagesToRemove {
		// InvalidParameterException: You can only move staging label AWSCURRENT to a different secret version. It can’t be completely removed.
		if stage.(string) == "AWSCURRENT" {
			log.Printf("[INFO] Skipping removal of AWSCURRENT staging label for secret %q version %q", secretID, versionID)
			continue
		}
		input := &secretsmanager.UpdateSecretVersionStageInput{
			RemoveFromVersionId: aws.String(versionID),
			SecretId:            aws.String(secretID),
			VersionStage:        aws.String(stage.(string)),
		}
		log.Printf("[DEBUG] Updating Secrets Manager Secret Version Stage: %s", input)
		_, err := conn.UpdateSecretVersionStage(input)
		if err != nil {
			return fmt.Errorf("error updating Secrets Manager Secret %q Version Stage %q: %s", secretID, stage.(string), err)
		}
	}

	return resourceAwsSecretsManagerSecretVersionRead(d, meta)
}

func resourceAwsSecretsManagerSecretVersionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).secretsmanagerconn

	secretID, versionID, err := decodeSecretsManagerSecretVersionID(d.Id())
	if err != nil {
		return err
	}

	if v, ok := d.GetOk("version_stages"); ok {
		for _, stage := range v.(*schema.Set).List() {
			// InvalidParameterException: You can only move staging label AWSCURRENT to a different secret version. It can’t be completely removed.
			if stage.(string) == "AWSCURRENT" {
				log.Printf("[WARN] Cannot remove AWSCURRENT staging label, which may leave the secret %q version %q active", secretID, versionID)
				continue
			}
			input := &secretsmanager.UpdateSecretVersionStageInput{
				RemoveFromVersionId: aws.String(versionID),
				SecretId:            aws.String(secretID),
				VersionStage:        aws.String(stage.(string)),
			}
			log.Printf("[DEBUG] Updating Secrets Manager Secret Version Stage: %s", input)
			_, err := conn.UpdateSecretVersionStage(input)
			if err != nil {
				if isAWSErr(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
					return nil
				}
				if isAWSErr(err, secretsmanager.ErrCodeInvalidRequestException, "You can’t perform this operation on the secret because it was deleted") {
					return nil
				}
				return fmt.Errorf("error updating Secrets Manager Secret %q Version Stage %q: %s", secretID, stage.(string), err)
			}
		}
	}

	return nil
}

func decodeSecretsManagerSecretVersionID(id string) (string, string, error) {
	idParts := strings.Split(id, "|")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format SecretID|VersionID, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}
