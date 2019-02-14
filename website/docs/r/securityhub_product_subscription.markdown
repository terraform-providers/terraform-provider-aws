---
layout: "aws"
page_title: "AWS: aws_securityhub_product_subscription"
sidebar_current: "docs-aws-resource-securityhub-product-subscription"
description: |-
  Subscribes to a Security Hub product.
---

# aws_securityhub_product_subscription

Subscribes to a Security Hub product.

## Example Usage

```hcl
resource "aws_securityhub_account" "example" {}

data "aws_region" "current" {}

resource "aws_securityhub_product_subscription" "example" {
  depends_on  = ["aws_securityhub_account.example"]
  product_arn = "arn:aws:securityhub:${data.aws_region.current.name}:733251395267:product/alertlogic/althreatmanagement"
}
```

## Argument Reference

The following arguments are supported:

* `product_arn` - (Required) The ARN of the product that generates findings that you want to import into Security Hub - see below.

Currently available products (remember to replace `${var.region}` as appropriate):

* `arn:aws:securityhub:${var.region}::product/aws/guardduty`
* `arn:aws:securityhub:${var.region}::product/aws/inspector`
* `arn:aws:securityhub:${var.region}::product/aws/macie`
* `arn:aws:securityhub:${var.region}:733251395267:product/alertlogic/althreatmanagement`
* `arn:aws:securityhub:${var.region}:679703615338:product/armordefense/armoranywhere`
* `arn:aws:securityhub:${var.region}:151784055945:product/barracuda/cloudsecurityguardian`
* `arn:aws:securityhub:${var.region}:758245563457:product/checkpoint/cloudguard-iaas`
* `arn:aws:securityhub:${var.region}:634729597623:product/checkpoint/dome9-arc`
* `arn:aws:securityhub:${var.region}:517716713836:product/crowdstrike/crowdstrike-falcon`
* `arn:aws:securityhub:${var.region}:749430749651:product/cyberark/cyberark-pta`
* `arn:aws:securityhub:${var.region}:250871914685:product/f5networks/f5-advanced-waf`
* `arn:aws:securityhub:${var.region}:123073262904:product/fortinet/fortigate`
* `arn:aws:securityhub:${var.region}:324264561773:product/guardicore/aws-infection-monkey`
* `arn:aws:securityhub:${var.region}:324264561773:product/guardicore/guardicore`
* `arn:aws:securityhub:${var.region}:949680696695:product/ibm/qradar-siem`
* `arn:aws:securityhub:${var.region}:955745153808:product/imperva/imperva-attack-analytics`
* `arn:aws:securityhub:${var.region}:297986523463:product/mcafee-skyhigh/mcafee-mvision-cloud-aws`
* `arn:aws:securityhub:${var.region}:188619942792:product/paloaltonetworks/redlock`
* `arn:aws:securityhub:${var.region}:122442690527:product/paloaltonetworks/vm-series`
* `arn:aws:securityhub:${var.region}:805950163170:product/qualys/qualys-pc`
* `arn:aws:securityhub:${var.region}:805950163170:product/qualys/qualys-vm`
* `arn:aws:securityhub:${var.region}:336818582268:product/rapid7/insightvm`
* `arn:aws:securityhub:${var.region}:062897671886:product/sophos/sophos-server-protection`
* `arn:aws:securityhub:${var.region}:112543817624:product/splunk/splunk-enterprise`
* `arn:aws:securityhub:${var.region}:112543817624:product/splunk/splunk-phantom`
* `arn:aws:securityhub:${var.region}:956882708938:product/sumologicinc/sumologic-mda`
* `arn:aws:securityhub:${var.region}:754237914691:product/symantec-corp/symantec-cwp`
* `arn:aws:securityhub:${var.region}:422820575223:product/tenable/tenable-io`
* `arn:aws:securityhub:${var.region}:679593333241:product/trend-micro/deep-security`
* `arn:aws:securityhub:${var.region}:453761072151:product/turbot/turbot`
* `arn:aws:securityhub:${var.region}:496947949261:product/twistlock/twistlock-enterprise`

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `arn` - The ARN of a resource that represents your subscription to the product that generates the findings that you want to import into Security Hub.

## Import

Security Hub product subscriptions can be imported in the form `product_arn,arn`, e.g.

```sh
$ terraform import aws_securityhub_product_subscription.example arn:aws:securityhub:eu-west-1:733251395267:product/alertlogic/althreatmanagement,arn:aws:securityhub:eu-west-1:123456789012:product-subscription/alertlogic/althreatmanagement
```
