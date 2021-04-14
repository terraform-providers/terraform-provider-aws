package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/msk/finder"
)

const (
	AssociatingSecret    = "associating"
	DisassociatingSecret = "disassociating"
	ScramSecretBatchSize = 10
)

func resourceAwsMskScramSecretAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMskScramSecretAssociationCreate,
		Read:   resourceAwsMskScramSecretAssociationRead,
		Update: resourceAwsMskScramSecretAssociationUpdate,
		Delete: resourceAwsMskScramSecretAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"cluster_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"secret_arn_list": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
			},
		},
	}
}

func resourceAwsMskScramSecretAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	clusterArn := d.Get("cluster_arn").(string)
	secretArnList := expandStringSet(d.Get("secret_arn_list").(*schema.Set))

	output, err := associateMSKClusterSecrets(conn, clusterArn, secretArnList)
	if err != nil {
		return fmt.Errorf("error associating scram secret(s) to MSK cluster (%s): %w", clusterArn, err)
	}

	d.SetId(aws.StringValue(output.ClusterArn))

	if len(output.UnprocessedScramSecrets) != 0 {
		return unprocessedScramSecretsError(output.ClusterArn, output.UnprocessedScramSecrets, AssociatingSecret)
	}

	return resourceAwsMskScramSecretAssociationRead(d, meta)
}

func resourceAwsMskScramSecretAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	secretArnList, err := finder.ScramSecrets(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, kafka.ErrCodeNotFoundException) {
		log.Printf("[WARN] Scram secret(s) for MSK cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading MSK cluster (%s) scram secret(s): %w", d.Id(), err)
	}

	d.Set("cluster_arn", d.Id())
	if err := d.Set("secret_arn_list", flattenStringSet(secretArnList)); err != nil {
		return fmt.Errorf("error setting secret_arn_list: %w", err)
	}

	return nil
}

func resourceAwsMskScramSecretAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	o, n := d.GetChange("secret_arn_list")
	oldSet, newSet := o.(*schema.Set), n.(*schema.Set)

	if newSet.Len() > 0 {
		if newSecrets := newSet.Difference(oldSet); newSecrets.Len() > 0 {
			output, err := associateMSKClusterSecrets(conn, d.Id(), expandStringSet(newSecrets))
			if err != nil {
				return fmt.Errorf("error associating scram secret(s) with MSK cluster (%s): %w", d.Id(), err)
			}

			if len(output.UnprocessedScramSecrets) != 0 {
				return unprocessedScramSecretsError(output.ClusterArn, output.UnprocessedScramSecrets, AssociatingSecret)
			}
		}
	}

	if oldSet.Len() > 0 {
		if deleteSecrets := oldSet.Difference(newSet); deleteSecrets.Len() > 0 {
			output, err := disassociateMSKClusterSecrets(conn, d.Id(), expandStringSet(deleteSecrets))
			if err != nil {
				return fmt.Errorf("error disassociating scram secret(s) from MSK cluster (%s): %w", d.Id(), err)
			}

			if len(output.UnprocessedScramSecrets) != 0 {
				return unprocessedScramSecretsError(output.ClusterArn, output.UnprocessedScramSecrets, DisassociatingSecret)
			}
		}
	}

	return resourceAwsMskScramSecretAssociationRead(d, meta)
}

func resourceAwsMskScramSecretAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	secretArnList, err := finder.ScramSecrets(conn, d.Id())

	if err != nil {
		if tfawserr.ErrCodeEquals(err, kafka.ErrCodeNotFoundException) {
			return nil
		}
		return fmt.Errorf("error reading scram secret(s) for MSK cluster (%s): %w", d.Id(), err)
	}

	if len(secretArnList) > 0 {
		output, err := disassociateMSKClusterSecrets(conn, d.Id(), secretArnList)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, kafka.ErrCodeNotFoundException) {
				return nil
			}
			return fmt.Errorf("error disassociating scram secret(s) from MSK cluster (%s): %w", d.Id(), err)
		}
		if len(output.UnprocessedScramSecrets) != 0 {
			return unprocessedScramSecretsError(output.ClusterArn, output.UnprocessedScramSecrets, DisassociatingSecret)
		}
	}

	return nil
}

func associateMSKClusterSecrets(conn *kafka.Kafka, clusterArn string, secretArnList []*string) (*kafka.BatchAssociateScramSecretOutput, error) {
	output := &kafka.BatchAssociateScramSecretOutput{}

	for i := 0; i < len(secretArnList); i += ScramSecretBatchSize {
		end := i + ScramSecretBatchSize
		if end > len(secretArnList) {
			end = len(secretArnList)
		}

		resp, err := conn.BatchAssociateScramSecret(&kafka.BatchAssociateScramSecretInput{
			ClusterArn:    aws.String(clusterArn),
			SecretArnList: secretArnList[i:end],
		})
		if err != nil {
			return nil, err
		}

		output.ClusterArn = resp.ClusterArn
		output.UnprocessedScramSecrets = append(output.UnprocessedScramSecrets, resp.UnprocessedScramSecrets...)
	}
	return output, nil
}

func disassociateMSKClusterSecrets(conn *kafka.Kafka, clusterArn string, secretArnList []*string) (*kafka.BatchDisassociateScramSecretOutput, error) {
	output := &kafka.BatchDisassociateScramSecretOutput{}

	for i := 0; i < len(secretArnList); i += ScramSecretBatchSize {
		end := i + ScramSecretBatchSize
		if end > len(secretArnList) {
			end = len(secretArnList)
		}

		resp, err := conn.BatchDisassociateScramSecret(&kafka.BatchDisassociateScramSecretInput{
			ClusterArn:    aws.String(clusterArn),
			SecretArnList: secretArnList[i:end],
		})
		if err != nil {
			return nil, err
		}

		output.ClusterArn = resp.ClusterArn
		output.UnprocessedScramSecrets = append(output.UnprocessedScramSecrets, resp.UnprocessedScramSecrets...)
	}
	return output, nil
}

func unprocessedScramSecretsError(clusterArn *string, secrets []*kafka.UnprocessedScramSecret, action string) error {
	var errors *multierror.Error

	for _, s := range secrets {
		secretArn, errMsg := aws.StringValue(s.SecretArn), aws.StringValue(s.ErrorMessage)
		errors = multierror.Append(errors, fmt.Errorf("error %s MSK cluster (%s) with scram secret (%s): %s", action, aws.StringValue(clusterArn), secretArn, errMsg))
	}

	return errors.ErrorOrNil()
}
