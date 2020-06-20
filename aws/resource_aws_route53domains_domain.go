package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53domains"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsRoute53DomainsDomainContactDetail() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"address_line_1": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"address_line_2": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"city": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"contact_type": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"country_code": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"email": {
					Type:     schema.TypeString,
					Optional: true,
				},
				// "extra_params": {
				// 	Type:     schema.TypeString,
				// 	Optional: true,
				// },
				"fax": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"first_name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"last_name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"organization_name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"phone_number": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"state": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"zip_code": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func resourceAwsRoute53DomainsDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53DomainsDomainCreate,
		Read:   resourceAwsRoute53DomainsDomainRead,
		Update: resourceAwsRoute53DomainsDomainUpdate,
		Delete: resourceAwsRoute53DomainsDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"abuse_contact_email": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"abuse_contact_phone": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"admin_contact": resourceAwsRoute53DomainsDomainContactDetail(),

			"admin_privacy": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"auto_renew": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"expiration_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name_servers": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"glue_ips": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 45),
							},
							Optional: true,
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
				Optional: true,
				Computed: true,
			},

			"registrant_contact": resourceAwsRoute53DomainsDomainContactDetail(),

			"registrant_privacy": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"registrar_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"registrar_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchemaComputed(),

			"tech_contact": resourceAwsRoute53DomainsDomainContactDetail(),

			"tech_privacy": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"transfer_lock": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"status_list": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

			"updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"whois_server": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsRoute53DomainsDomainCreate(d *schema.ResourceData, meta interface{}) error {
	domainName := d.Get("domain_name").(string)
	d.SetId(domainName)
	return resourceAwsRoute53DomainsDomainUpdate(d, meta)
}

func resourceAwsRoute53DomainsDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53domainsconn

	d.Set("domain_name", d.Id())

	log.Printf("[DEBUG] Get domain details for Route 53 Domain: %s", d.Id())
	out, err := conn.GetDomainDetail(&route53domains.GetDomainDetailInput{
		DomainName: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, route53domains.ErrCodeInvalidInput, "not found") {
			d.SetId("")
			log.Printf("[DEBUG] Route 53 Domain %s not found", d.Id())
			return nil
		}
		return err
	}
	log.Printf("[DEBUG] Route 53 Domain received: %#v", out)

	// Unpack response
	d.Set("abuse_contact_email", out.AbuseContactEmail)
	d.Set("abuse_contact_phone", out.AbuseContactPhone)
	d.Set("admin_contact", resourceAwsRoute53DomainsDomainFlattenContactDetail(out.AdminContact))
	d.Set("admin_privacy", out.AdminPrivacy)
	d.Set("auto_renew", out.AutoRenew)
	d.Set("creation_date", out.CreationDate.String())
	d.Set("expiration_date", out.ExpirationDate.String())
	d.Set("name_servers", resourceAwsRoute53DomainsDomainFlattenNameservers(out.Nameservers))
	d.Set("registrant_contact", resourceAwsRoute53DomainsDomainFlattenContactDetail(out.RegistrantContact))
	d.Set("registrant_privacy", out.RegistrantPrivacy)
	d.Set("registrar_name", out.RegistrarName)
	d.Set("registrar_url", out.RegistrarUrl)
	d.Set("tech_contact", resourceAwsRoute53DomainsDomainFlattenContactDetail(out.TechContact))
	d.Set("tech_privacy", out.TechPrivacy)
	// d.Set("updated_date", out.UpdatedDate.String())
	d.Set("whois_server", out.WhoIsServer)

	if err := d.Set("status_list", flattenStringList(out.StatusList)); err != nil {
		return fmt.Errorf("Error setting status_list error: %s", err)
	}

	// Only way to check if a domain is locked
	transferLock := false
	for _, status := range out.StatusList {
		if *status == "clientTransferProhibited" {
			transferLock = true
			break
		}
	}
	d.Set("transfer_lock", transferLock)

	tags, err := keyvaluetags.Route53domainsListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for Route53 Domains domain (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsRoute53DomainsDomainExpandNameservers(in []interface{}) []*route53domains.Nameserver {
	nameservers := []*route53domains.Nameserver{}
	for _, mRaw := range in {
		m := mRaw.(map[string]interface{})
		nameserver := &route53domains.Nameserver{
			GlueIps: expandStringList(m["glue_ips"].([]interface{})),
			Name:    aws.String(m["name"].(string)),
		}
		nameservers = append(nameservers, nameserver)
	}
	log.Printf("[DEBUG] Route 53 Domain expanded name servers: %#v", nameservers)
	return nameservers
}

func resourceAwsRoute53DomainsDomainFlattenNameservers(in []*route53domains.Nameserver) []map[string]interface{} {
	var out = make([]map[string]interface{}, len(in))
	for i, v := range in {
		m := make(map[string]interface{})
		m["glue_ips"] = flattenStringList(v.GlueIps)
		m["name"] = v.Name
		out[i] = m
	}
	return out
}

// func resourceAwsRoute53DomainsDomainFlattenExtraParams(in []*route53domains.ExtraParam) []map[string]interface{} {
// 	var out = make([]map[string]interface{}, len(in))
// 	for i, v := range in {
// 		m := make(map[string]interface{})
// 		m["name"] = v.Name
// 		m["value"] = v.Value
// 		out[i] = m
// 	}
// 	return out
// }

func resourceAwsRoute53DomainsDomainFlattenContactDetail(in *route53domains.ContactDetail) []interface{} {
	m := make(map[string]interface{})
	if in.AddressLine1 != nil {
		m["address_line_1"] = *in.AddressLine1
	}
	if in.AddressLine2 != nil {
		m["address_line_2"] = *in.AddressLine2
	}
	if in.City != nil {
		m["city"] = *in.City
	}
	if in.ContactType != nil {
		m["contact_type"] = *in.ContactType
	}
	if in.CountryCode != nil {
		m["country_code"] = *in.CountryCode
	}
	if in.Email != nil {
		m["email"] = *in.Email
	}
	// if in.ExtraParams != nil {
	// 	m["extra_params"] = resourceAwsRoute53DomainsDomainFlattenExtraParams(in.ExtraParams)
	// }
	if in.Fax != nil {
		m["fax"] = *in.Fax
	}
	if in.FirstName != nil {
		m["first_name"] = *in.FirstName
	}
	if in.LastName != nil {
		m["last_name"] = *in.LastName
	}
	if in.OrganizationName != nil {
		m["organization_name"] = *in.OrganizationName
	}
	if in.PhoneNumber != nil {
		m["phone_number"] = *in.PhoneNumber
	}
	if in.State != nil {
		m["state"] = *in.State
	}
	if in.ZipCode != nil {
		m["zip_code"] = *in.ZipCode
	}
	out := []interface{}{m}
	log.Printf("[DEBUG] Contact details: %v", out)
	return out
}

func resourceAwsRoute53DomainsDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53domainsconn

	// Changes to tags
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Route53domainsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	// Changes to auto renew
	if d.HasChange("auto_renew") {
		if d.Get("auto_renew").(bool) {
			log.Printf("[DEBUG] Enabling domain auto renew for Route 53 domain (%s)", d.Id())
			_, err := conn.EnableDomainAutoRenew(&route53domains.EnableDomainAutoRenewInput{
				DomainName: aws.String(d.Id()),
			})
			if err != nil {
				return fmt.Errorf("Error enabling domain auto renew for Route 53 domain (%s), error: %s", d.Id(), err)
			}
		} else {
			log.Printf("[DEBUG] Disabling domain auto renew for Route 53 domain (%s)", d.Id())
			_, err := conn.DisableDomainAutoRenew(&route53domains.DisableDomainAutoRenewInput{
				DomainName: aws.String(d.Id()),
			})
			if err != nil {
				return fmt.Errorf("Error disabling domain auto renew for Route 53 domain (%s), error: %s", d.Id(), err)
			}
		}
	}

	// Changes to transfer lock
	if d.HasChange("transfer_lock") {
		if d.Get("transfer_lock").(bool) {
			log.Printf("[DEBUG] Enabling domain transfer lock for Route 53 domain (%s)", d.Id())
			out, err := conn.EnableDomainTransferLock(&route53domains.EnableDomainTransferLockInput{
				DomainName: aws.String(d.Id()),
			})
			if err != nil {
				return fmt.Errorf("Error enabling domain transfer lock for Route 53 domain (%s), error: %s", d.Id(), err)
			}
			if err := resourceAwsRoute53DomainsDomainWaitForOperation(conn, *out.OperationId, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return fmt.Errorf("Error waiting for Route 53 domain operation (%s) on domain (%s) to complete: %s", *out.OperationId, d.Id(), err)
			}
		} else {
			log.Printf("[DEBUG] Disabling domain transfer lock for Route 53 domain (%s)", d.Id())
			out, err := conn.DisableDomainTransferLock(&route53domains.DisableDomainTransferLockInput{
				DomainName: aws.String(d.Id()),
			})
			if err != nil {
				return fmt.Errorf("Error disabling domain transfer lock for Route 53 domain (%s), error: %s", d.Id(), err)
			}
			if err := resourceAwsRoute53DomainsDomainWaitForOperation(conn, *out.OperationId, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return fmt.Errorf("Error waiting for Route 53 domain operation (%s) on domain (%s) to complete: %s", *out.OperationId, d.Id(), err)
			}
		}
	}

	// Changes to domain contact privacy
	if d.HasChange("admin_privacy") || d.HasChange("registrant_privacy") || d.HasChange("tech_privacy") {
		log.Printf("[DEBUG] Updating domain contact privacy settings for Route 53 domain (%s)", d.Id())
		out, err := conn.UpdateDomainContactPrivacy(&route53domains.UpdateDomainContactPrivacyInput{
			DomainName:        aws.String(d.Id()),
			AdminPrivacy:      aws.Bool(d.Get("admin_privacy").(bool)),
			RegistrantPrivacy: aws.Bool(d.Get("registrant_privacy").(bool)),
			TechPrivacy:       aws.Bool(d.Get("tech_privacy").(bool)),
		})
		if err != nil {
			return fmt.Errorf("Error updating domain contact privacy settings for Route 53 domain (%s), error: %s", d.Id(), err)
		}
		if err := resourceAwsRoute53DomainsDomainWaitForOperation(conn, *out.OperationId, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("Error waiting for Route 53 domain operation (%s) on domain (%s) to complete: %s", *out.OperationId, d.Id(), err)
		}
	}

	// Changes to domain nameservers
	if d.HasChange("name_servers") {
		log.Printf("[DEBUG] Updating domain name servers for Route 53 domain (%s)", d.Id())
		out, err := conn.UpdateDomainNameservers(&route53domains.UpdateDomainNameserversInput{
			DomainName:  aws.String(d.Id()),
			Nameservers: resourceAwsRoute53DomainsDomainExpandNameservers(d.Get("name_servers").([]interface{})),
		})
		if err != nil {
			return fmt.Errorf("Error updating domain name servers for Route 53 domain (%s), error: %s", d.Id(), err)
		}
		if err := resourceAwsRoute53DomainsDomainWaitForOperation(conn, *out.OperationId, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("Error waiting for Route 53 domain operation (%s) on domain (%s) to complete: %s", *out.OperationId, d.Id(), err)
		}
	}

	return resourceAwsRoute53DomainsDomainRead(d, meta)
}

func resourceAwsRoute53DomainsDomainDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Deletion of Route 53 Domain is a no-op")
	return nil
}

func resourceAwsRoute53DomainsDomainWaitForOperation(conn *route53domains.Route53Domains, operationId string, timeout time.Duration) error {
	log.Printf("Waiting for Route 53 domain operation (%s)...", operationId)

	stateConf := &resource.StateChangeConf{
		Pending: []string{route53domains.OperationStatusSubmitted, route53domains.OperationStatusInProgress},
		Target:  []string{route53domains.OperationStatusSuccessful},
		Refresh: resourceAwsRoute53DomainsDomainOperationStateRefreshFunc(conn, operationId),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Route 53 domain operation (%s): %v", operationId, err)
	}

	return nil
}

func resourceAwsRoute53DomainsDomainOperationStateRefreshFunc(conn *route53domains.Route53Domains, operationId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := conn.GetOperationDetail(&route53domains.GetOperationDetailInput{
			OperationId: aws.String(operationId),
		})
		if err != nil {
			return nil, "", fmt.Errorf("Error on refresh: %+v", err)
		}
		return out, *out.Status, nil
	}
}
