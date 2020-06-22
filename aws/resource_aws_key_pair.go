package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsKeyPair() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAwsKeyPairCreate,
		Read:   resourceAwsKeyPairRead,
		Update: resourceAwsKeyPairUpdate,
		Delete: resourceAwsKeyPairDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,
		MigrateState:  resourceAwsKeyPairMigrateState,

		Schema: map[string]*schema.Schema{
			"key_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"key_name_prefix"},
				ValidateFunc:  validation.StringLenBetween(0, 255),
			},
			"key_name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"key_name"},
				ValidateFunc:  validation.StringLenBetween(0, 255-resource.UniqueIDSuffixLength),
			},
			"public_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					switch v := v.(type) {
					case string:
						return strings.TrimSpace(v)
					default:
						return ""
					}
				},
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_pair_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsKeyPairCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	var keyName string
	if v, ok := d.GetOk("key_name"); ok {
		keyName = v.(string)
	} else if v, ok := d.GetOk("key_name_prefix"); ok {
		keyName = resource.PrefixedUniqueId(v.(string))
		d.Set("key_name", keyName)
	} else {
		keyName = resource.UniqueId()
		d.Set("key_name", keyName)
	}

	publicKey := d.Get("public_key").(string)
	req := &ec2.ImportKeyPairInput{
		KeyName:           aws.String(keyName),
		PublicKeyMaterial: []byte(publicKey),
		TagSpecifications: ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), ec2.ResourceTypeKeyPair),
	}
	resp, err := conn.ImportKeyPair(req)
	if err != nil {
		return fmt.Errorf("Error import KeyPair: %s", err)
	}

	d.SetId(aws.StringValue(resp.KeyName))

	return resourceAwsKeyPairRead(d, meta)
}

func resourceAwsKeyPairRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeKeyPairsInput{
		KeyNames: []*string{aws.String(d.Id())},
	}
	resp, err := conn.DescribeKeyPairs(req)
	if err != nil {
		if isAWSErr(err, "InvalidKeyPair.NotFound", "") {
			log.Printf("[WARN] Key Pair (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving KeyPair: %s", err)
	}

	if len(resp.KeyPairs) == 0 {
		log.Printf("[WARN] Key Pair (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	kp := resp.KeyPairs[0]

	if aws.StringValue(kp.KeyName) != d.Id() {
		return fmt.Errorf("Unable to find key pair within: %#v", resp.KeyPairs)
	}

	d.Set("key_name", kp.KeyName)
	d.Set("fingerprint", kp.KeyFingerprint)
	d.Set("key_pair_id", kp.KeyPairId)
	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(kp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "ec2",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("key-pair/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceAwsKeyPairUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Ec2UpdateTags(conn, d.Get("key_pair_id").(string), o, n); err != nil {
			return fmt.Errorf("error adding tags: %s", err)
		}
	}

	return resourceAwsKeyPairRead(d, meta)
}

func resourceAwsKeyPairDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	_, err := conn.DeleteKeyPair(&ec2.DeleteKeyPairInput{
		KeyName: aws.String(d.Id()),
	})
	return err
}
