package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"strings"
)

func resourceAwsLightsailDomainEntry() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLightsailDomainEntryCreate,
		Read:   resourceAwsLightsailDomainEntryRead,
		Delete: resourceAwsLightsailDomainEntryDelete,

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"is_alias": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"A",
					"CNAME",
					"MX",
					"NS",
					"SOA",
					"SRV",
					"TXT",
				}, false),
				ForceNew: true,
			},
		},
	}
}

func resourceAwsLightsailDomainEntryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lightsailconn
	req := &lightsail.CreateDomainEntryInput{
		DomainName: aws.String(d.Get("domain_name").(string)),

		DomainEntry: &lightsail.DomainEntry{
			IsAlias: aws.Bool(d.Get("is_alias").(bool)),
			Name:    aws.String(d.Get("name").(string)),
			Target:  aws.String(d.Get("target").(string)),
			Type:    aws.String(d.Get("type").(string)),
		},
	}

	_, err := conn.CreateDomainEntry(req)

	if err != nil {
		return err
	}

	// Generate an ID
	vars := []string{
		d.Get("name").(string),
		d.Get("domain_name").(string),
		d.Get("type").(string),
		d.Get("target").(string),
	}

	d.SetId(strings.Join(vars, "_"))

	return resourceAwsLightsailDomainEntryRead(d, meta)
}

func resourceAwsLightsailDomainEntryRead(d *schema.ResourceData, meta interface{}) error {

	id_parts := strings.SplitN(d.Id(), "_", -1)
	if len(id_parts) != 4 {
		return nil
	}
	name := id_parts[0]
	domainname := id_parts[1]
	recordType := id_parts[2]
	recordTarget := id_parts[3]

	conn := meta.(*AWSClient).lightsailconn
	resp, err := conn.GetDomain(&lightsail.GetDomainInput{
		DomainName: aws.String(domainname),
	})

	if err != nil {
		return err
	}

	var entry lightsail.DomainEntry
	entryExists := false

	for _, n := range resp.Domain.DomainEntries {
		if name == *n.Name && recordType == *n.Type && recordTarget == *n.Target {
			entry = *n
			entryExists = true
			break
		}
	}

	if !entryExists {
		d.SetId("")
	}

	d.Set("name", entry.Name)
	d.Set("domain_name", domainname)
	d.Set("type", entry.Type)
	d.Set("is_alias", entry.IsAlias)
	d.Set("target", entry.Target)

	return nil
}

func resourceAwsLightsailDomainEntryDelete(d *schema.ResourceData, meta interface{}) error {

	id_parts := strings.SplitN(d.Id(), "_", -1)
	name := id_parts[0]
	domainname := id_parts[1]
	recordType := id_parts[2]
	recordTarget := id_parts[3]

	conn := meta.(*AWSClient).lightsailconn
	_, err := conn.DeleteDomainEntry(&lightsail.DeleteDomainEntryInput{
		DomainName: aws.String(domainname),
		DomainEntry: &lightsail.DomainEntry{
			Name:    aws.String(name),
			Type:    aws.String(recordType),
			Target:  aws.String(recordTarget),
			IsAlias: aws.Bool(d.Get("is_alias").(bool)),
		},
	})

	return err
}
