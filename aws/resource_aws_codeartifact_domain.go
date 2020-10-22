package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsCodeArtifactDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeArtifactDomainCreate,
		Read:   resourceAwsCodeArtifactDomainRead,
		Delete: resourceAwsCodeArtifactDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"encryption_key": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asset_size_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"repository_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAwsCodeArtifactDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	log.Print("[DEBUG] Creating CodeArtifact Domain")

	params := &codeartifact.CreateDomainInput{
		Domain:        aws.String(d.Get("domain").(string)),
		EncryptionKey: aws.String(d.Get("encryption_key").(string)),
	}

	domain, err := conn.CreateDomain(params)
	if err != nil {
		return fmt.Errorf("error creating CodeArtifact Domain: %w", err)
	}

	d.SetId(aws.StringValue(domain.Domain.Arn))

	return resourceAwsCodeArtifactDomainRead(d, meta)
}

func resourceAwsCodeArtifactDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn

	log.Printf("[DEBUG] Reading CodeArtifact Domain: %s", d.Id())

	domainOwner, domainName, err := decodeCodeArtifactDomainID(d.Id())
	if err != nil {
		return err
	}

	sm, err := conn.DescribeDomain(&codeartifact.DescribeDomainInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(domainOwner),
	})
	if err != nil {
		if isAWSErr(err, codeartifact.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] CodeArtifact Domain %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading CodeArtifact Domain (%s): %w", d.Id(), err)
	}

	d.Set("domain", sm.Domain.Name)
	d.Set("arn", sm.Domain.Arn)
	d.Set("encryption_key", sm.Domain.EncryptionKey)
	d.Set("owner", sm.Domain.Owner)
	d.Set("asset_size_bytes", sm.Domain.AssetSizeBytes)
	d.Set("repository_count", sm.Domain.RepositoryCount)
	d.Set("created_time", sm.Domain.CreatedTime.Format(time.RFC3339))

	return nil
}

func resourceAwsCodeArtifactDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	log.Printf("[DEBUG] Deleting CodeArtifact Domain: %s", d.Id())

	domainOwner, domainName, err := decodeCodeArtifactDomainID(d.Id())
	if err != nil {
		return err
	}

	input := &codeartifact.DeleteDomainInput{
		Domain:      aws.String(domainName),
		DomainOwner: aws.String(domainOwner),
	}

	_, err = conn.DeleteDomain(input)

	if isAWSErr(err, codeartifact.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CodeArtifact Domain (%s): %w", d.Id(), err)
	}

	return nil
}

func decodeCodeArtifactDomainID(id string) (string, string, error) {
	repoArn, err := arn.Parse(id)
	if err != nil {
		return "", "", err
	}

	domainName := strings.TrimPrefix(repoArn.Resource, "domain/")
	return repoArn.AccountID, domainName, nil
}
