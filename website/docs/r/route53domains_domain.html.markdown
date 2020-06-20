---
subcategory: "Route53 Domains"
layout: "aws"
page_title: "AWS: aws_route53domains_domain"
description: |-
  Manages a Route53 Domain
---

# Resource: aws_route53domains_domain

Manages a Route53 Domain.

## Example Usage

```hcl
resource "aws_route53domains_domain" "dev" {
  domain_name = "dev.example.com"

  tags = {
    Environment = "dev"
  }
}
```

## Argument Reference

The following arguments are supported:

* `domain_name` - (Required) The name of a domain.
* `admin_privacy` - (Optional) Specifies whether contact information is concealed from WHOIS queries. If the value is true, WHOIS ("who is") queries return contact information either for Amazon Registrar (for .com, .net, and .org domains) or for our registrar associate, Gandi (for all other TLDs). If the value is false, WHOIS queries return the information that you entered for the admin contact.
* `auto_renew` - (Optional) Specifies whether the domain registration is set to renew automatically.
* `name_servers` - (Optional) The nameservers for the domain.
* `registrant_privacy` - (Optional) Specifies whether contact information is concealed from WHOIS queries. If the value is true, WHOIS ("who is") queries return contact information either for Amazon Registrar (for .com, .net, and .org domains) or for our registrar associate, Gandi (for all other TLDs). If the value is false, WHOIS queries return the information that you entered for the registrant contact (domain owner).
* `tech_privacy` - (Optional) Specifies whether contact information is concealed from WHOIS queries. If the value is true, WHOIS ("who is") queries return contact information either for Amazon Registrar (for .com, .net, and .org domains) or for our registrar associate, Gandi (for all other TLDs). If the value is false, WHOIS queries return the information that you entered for the technical contact.
* `transfer_lock` - (Optional) Enables/disables the transfer lock on the domain.
* `tags` - (Optional) A mapping of tags to assign to the domain.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `abuse_contact_email` - Email address to contact to report incorrect contact information for a domain, to report that the domain is being used to send spam, to report that someone is cybersquatting on a domain name, or report some other type of abuse.
* `abuse_contact_phone` - Phone number for reporting abuse.
* `admin_contact` - Provides details about the domain administrative contact.
* `creation_date` - The date when the domain was created as found in the response to a WHOIS query. The date and time is in Unix time format and Coordinated Universal time (UTC).
* `expiration_date` - The date when the registration for the domain is set to expire. The date and time is in Unix time format and Coordinated Universal time (UTC).
* `registrant_contact` - Provides details about the domain registrant.
* `registrar_name` - Name of the registrar of the domain as identified in the registry. Domains with a .com, .net, or .org TLD are registered by Amazon Registrar. All other domains are registered by our registrar associate, Gandi. The value for domains that are registered by Gandi is "GANDI SAS".
* `registrar_url` - Web address of the registrar.
* `status_list` - An array of domain name status codes, also known as Extensible Provisioning Protocol (EPP) status codes.
* `tech_contact` - Provides details about the domain technical contact.
* `updated_date` - The last updated date of the domain as found in the response to a WHOIS query. The date and time is in Unix time format and Coordinated Universal time (UTC).
* `whois_server` - The fully qualified name of the WHOIS server that can answer the WHOIS query for the domain.

`admin_contact`, `registrant_contact` and `tech_contact` support the following arguments:

* `address_line_1` - First line of the contact's address.
* `address_line_2` - Second line of contact's address, if any.
* `city` - The city of the contact's address.
* `contact_type` - Indicates whether the contact is a person, company, association, or public organization. One of `PERSON`, `COMPANY`, `ASSOCIATION`, `PUBLIC_BODY`, `RESELLER`
* `country_code` - Code for the country of the contact's address.
* `email` - Email address of the contact.
* `fax` - Fax number of the contact. Phone number must be specified in the format "+[country dialing code].[number including any area code]". For example, a US phone number might appear as "+1.1234567890".
* `first_name` - First name of contact.
* `last_name` - Last name of contact.
* `organization_name` - Name of the organization for contact types other than `PERSON`.
* `phone_number` - The phone number of the contact. Phone number must be specified in the format "+[country dialing code].[number including any area code]". For example, a US phone number might appear as "+1.1234567890".
* `state` - The state or province of the contact's city.
* `zip_code` - The zip or postal code of the contact's address.


## Import

Route53 Domains can be imported using the `domain id`, e.g.

```
$ terraform import aws_route53domains_domain.mydomain example.com
```
