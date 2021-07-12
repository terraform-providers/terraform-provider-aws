package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
)

func resourceAwsEc2ManagedPrefixListEntry() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAwsEc2ManagedPrefixListEntryCreate,
		Read:   resourceAwsEc2ManagedPrefixListEntryRead,
		Delete: resourceAwsEc2ManagedPrefixListEntryDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsEc2ManagedPrefixListEntryImport,
		},

		Schema: map[string]*schema.Schema{
			"cidr": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"prefix_list_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAwsEc2ManagedPrefixListEntryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	pl_id := d.Get("prefix_list_id").(string)
	cidrBlock := d.Get("cidr").(string)
	description := d.Get("description").(string)

	pl, err := finder.ManagedPrefixListByID(conn, pl_id)
	if err != nil {
		return err
	}

	entry, err := expandPrefixEntry(pl, cidrBlock, description)
	if err != nil {
		return err
	}
	input := &ec2.ModifyManagedPrefixListInput{
		PrefixListId:   aws.String(pl_id),
		CurrentVersion: pl.Version,
	}

	input.AddEntries = []*ec2.AddPrefixListEntry{(*ec2.AddPrefixListEntry)(entry)}

	_, err = conn.ModifyManagedPrefixList(input)
	if err != nil {
		return fmt.Errorf("error adding EC2 Managed Prefix List entry (%s): %w", d.Id(), err)
	}
	if _, err := waiter.ManagedPrefixListModified(conn, pl_id); err != nil {
		return fmt.Errorf("error waiting for EC2 Managed Prefix List (%s) update: %w", pl_id, err)
	}

	getEntriesInput := &ec2.GetManagedPrefixListEntriesInput{
		PrefixListId: pl.PrefixListId,
	}
	var entries []*ec2.PrefixListEntry
	id := tfec2.ManagedPrefixListEntryCreateID(pl_id, cidrBlock)
	log.Printf("[DEBUG] Computed EC2 Managed Prefix List entry ID %s", id)

	err = resource.Retry(waiter.ManagedPrefixListEntryCreateTimeout, func() *resource.RetryError {

		entries, err = getEc2ManagedPrefixListEntries(conn, getEntriesInput)

		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error listing EC2 Managed Prefix List (%s) entries: %s", d.Id(), err))
		}

		cidr := findEntryMatch(entry, entries)
		if cidr == nil {
			log.Printf("[DEBUG] Unable to find matching entry (%s) for EC2 Managed Prefix List %s",
				id, pl_id)
			return resource.RetryableError(fmt.Errorf("No match found"))
		}

		log.Printf("[DEBUG] Found entry for EC2 Managed Prefix List entry (%s): %s", id, cidr)
		return nil
	})
	if isResourceTimeoutError(err) {
		entries, err = getEc2ManagedPrefixListEntries(conn, getEntriesInput)

		cidr := findEntryMatch(entry, entries)
		if cidr == nil {
			return fmt.Errorf("Error finding matching EC2 Managed Prefix List entry: %s", err)
		}
	}
	if err != nil {
		return fmt.Errorf("Error finding matching entry (%s) for EC2 Managed Prefix List %s", id, pl_id)
	}

	d.SetId(id)
	return resourceAwsEc2ManagedPrefixListEntryRead(d, meta)
}

func resourceAwsEc2ManagedPrefixListEntryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	pl_id := d.Get("prefix_list_id").(string)
	cidrBlock := d.Get("cidr").(string)
	description := d.Get("description").(string)
	pl, err := finder.ManagedPrefixListByID(conn, pl_id)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidPrefixListIDNotFound) {
		// The EC2 Managed Prefix List containing this entry no longer exists.
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error finding EC2 Managed Prefix List (%s) for entry (%s): %s", pl_id, d.Id(), err)
	}

	input := &ec2.GetManagedPrefixListEntriesInput{
		PrefixListId: pl.PrefixListId,
	}
	var entry *ec2.PrefixListEntry
	var entries []*ec2.PrefixListEntry
	entries, err = getEc2ManagedPrefixListEntries(conn, input)

	if err != nil {
		return fmt.Errorf("error listing entries of EC2 Managed Prefix List (%s): %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Entries %v", entries)

	p, err := expandPrefixEntry(pl, cidrBlock, description)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		log.Printf("[WARN] No entries were found for EC2 Managed Prefix List (%s) looking for entry (%s)",
			*pl.PrefixListName, d.Id())
		d.SetId("")
		return nil
	}

	entry = findEntryMatch(p, entries)

	if entry == nil {
		log.Printf("[DEBUG] Unable to find matching entry (%s) for EC2 Managed Prefix List %s",
			d.Id(), pl_id)
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] Found entry for EC2 Managed Prefix List (%s): %s", pl_id, entry)

	d.Set("cidr", entry.Cidr)
	d.Set("description", entry.Description)
	d.Set("version", pl.Version)

	return nil
}

func resourceAwsEc2ManagedPrefixListEntryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	pl_id := d.Get("prefix_list_id").(string)
	cidrBlock := d.Get("cidr").(string)
	description := d.Get("description").(string)

	pl, err := finder.ManagedPrefixListByID(conn, pl_id)
	if err != nil {
		return err
	}

	entry, err := expandPrefixEntry(pl, cidrBlock, description)
	if err != nil {
		return err
	}
	input := &ec2.ModifyManagedPrefixListInput{
		CurrentVersion: pl.Version,
		PrefixListId:   pl.PrefixListId,
		RemoveEntries:  []*ec2.RemovePrefixListEntry{{Cidr: entry.Cidr}},
	}

	_, err = conn.ModifyManagedPrefixList(input)

	if err != nil {
		return fmt.Errorf("error deleting EC2 Managed Prefix List entry (%s): %w", d.Id(), err)
	}

	_, err = waiter.ManagedPrefixListModified(conn, pl_id)

	if err != nil {
		return fmt.Errorf("error waiting for EC2 Managed Prefix List entry (%s) deletion: %w", d.Id(), err)
	}

	return nil
}

func findEntryMatch(e *ec2.PrefixListEntry, entries []*ec2.PrefixListEntry) *ec2.PrefixListEntry {
	var entry *ec2.PrefixListEntry
	for _, r := range entries {
		if e.Cidr != nil && r.Cidr != nil && *e.Cidr != *r.Cidr {
			continue
		}

		entry = r
	}
	return entry
}

func expandPrefixEntry(pl *ec2.ManagedPrefixList, cidrBlock string, description string) (*ec2.PrefixListEntry, error) {
	var apiObject ec2.PrefixListEntry

	if description != "" {
		apiObject.Description = aws.String(description)
	}

	var err error
	switch *pl.AddressFamily {
	case "IPv4":
		err = validateIpv4CIDRBlock(cidrBlock)
	case "IPv6":
		err = validateIpv6CIDRBlock(cidrBlock)
	}
	if err != nil {
		return nil, err
	}
	apiObject.Cidr = aws.String(cidrBlock)

	return &apiObject, nil
}

func resourceAwsEc2ManagedPrefixListEntryImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	prefixListID, cidrBlock, err := tfec2.ManagedPrefixListEntryParseID(d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("prefix_list_id", prefixListID)
	d.Set("cidr", cidrBlock)
	d.SetId(tfec2.ManagedPrefixListEntryCreateID(prefixListID, cidrBlock))

	return []*schema.ResourceData{d}, nil
}

func getEc2ManagedPrefixListEntries(conn *ec2.EC2, input *ec2.GetManagedPrefixListEntriesInput) ([]*ec2.PrefixListEntry, error) {
	var entries []*ec2.PrefixListEntry
	err := conn.GetManagedPrefixListEntriesPages(input, func(page *ec2.GetManagedPrefixListEntriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		entries = append(entries, page.Entries...)

		return !lastPage
	})
	return entries, err
}
