package aws

import (
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

type EcrLifecyclePolicyDoc struct {
	Rules []*EcrLifecyclePolicyStatement `json:"rules"`
}

type EcrLifecyclePolicyStatement struct {
	RulePriority int                                     `json:"rulePriority,omitempty"`
	Description  string                                  `json:"description,omitempty"`
	Selection    EcrLifecyclePolicyStatementSelectionSet `json:"selection,omitempty"`
	Action       EcrLifecyclePolicyAction                `json:"action"`
}

type EcrLifecyclePolicySelection struct {
	TagStatus     string        `json:"tagStatus,omitempty"`
	TagPrefixList []interface{} `json:"tagPrefixList,omitempty"`
	CountType     string        `json:"countType,omitempty"`
	CountUnit     string        `json:"countUnit,omitempty"`
	CountNumber   int           `json:"countNumber,omitempty"`
}

type EcrLifecyclePolicyAction struct {
	Type string `json:"type"`
}

type EcrLifecyclePolicyStatementSelectionSet EcrLifecyclePolicySelection

func dataSourceAwsEcrLifecyclePolicyDocument() *schema.Resource {

	return &schema.Resource{
		Read: dataSourceAwsEcrLifecyclePolicyDocumentRead,

		Schema: map[string]*schema.Schema{
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"priority": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"selection": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"tag_status": {
										Type:         schema.TypeString,
										Required:     false,
										Optional:     true,
										Default:      "any",
										ValidateFunc: validation.StringInSlice([]string{"tagged", "untagged", "any"}, false),
									},
									"tag_prefix_list": {
										Type:     schema.TypeList,
										Required: false,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"count_type": {
										Type:         schema.TypeString,
										Required:     false,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"imageCountMoreThan", "sinceImagePushed", "any"}, false),
									},
									"count_unit": {
										Type:         schema.TypeString,
										Required:     false,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"days"}, false),
									},
									"count_number": {
										Type:     schema.TypeInt,
										Required: false,
										Optional: true,
									},
								},
							},
						},
						"action": {
							Type:     schema.TypeList,
							Required: false,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "expire",
									},
								},
							},
						},
					},
				},
			},
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEcrLifecyclePolicyDocumentRead(d *schema.ResourceData, meta interface{}) error {
	mergedDoc := &EcrLifecyclePolicyDoc{}

	// process the current document
	doc := &EcrLifecyclePolicyDoc{}

	var cfgStmts = d.Get("rule").([]interface{})
	stmts := make([]*EcrLifecyclePolicyStatement, len(cfgStmts))
	for i, stmtI := range cfgStmts {
		cfgStmt := stmtI.(map[string]interface{})

		stmt := &EcrLifecyclePolicyStatement{
			RulePriority: cfgStmt["priority"].(int),
		}

		if description, ok := cfgStmt["description"]; ok {
			stmt.Description = description.(string)
		}

		if selection := cfgStmt["selection"].(*schema.Set).List(); len(selection) > 0 {
			stmt.Selection = dataSourceAwsEcrLifecyclePolicyDocumentMakeSelection(selection)
		}

		if action := cfgStmt["action"].([]interface{}); len(action) == 0 {
			stmt.Action = EcrLifecyclePolicyAction{
				Type: "expire",
			}
		}

		stmts[i] = stmt
	}

	doc.Rules = stmts

	// merge our current document into mergedDoc
	mergedDoc.Merge(doc)

	jsonDoc, err := json.MarshalIndent(mergedDoc, "", "  ")
	if err != nil {
		// should never happen if the above code is correct
		return err
	}
	jsonString := string(jsonDoc)

	d.Set("json", jsonString)
	d.SetId(strconv.Itoa(hashcode.String(jsonString)))

	return nil
}

func dataSourceAwsEcrLifecyclePolicyDocumentMakeSelection(in []interface{}) EcrLifecyclePolicyStatementSelectionSet {
	item := in[0].(map[string]interface{})
	out := EcrLifecyclePolicySelection{
		TagStatus:     item["tag_status"].(string),
		TagPrefixList: item["tag_prefix_list"].([]interface{}),
		CountType:     item["count_type"].(string),
		CountUnit:     item["count_unit"].(string),
		CountNumber:   item["count_number"].(int),
	}
	return EcrLifecyclePolicyStatementSelectionSet(out)
}

func (self *EcrLifecyclePolicyDoc) Merge(newDoc *EcrLifecyclePolicyDoc) {
	// merge in newDoc's statements, overwriting any existing Sids
	var seen bool
	for _, newRule := range newDoc.Rules {
		if newRule.RulePriority == 0 {
			self.Rules = append(self.Rules, newRule)
			continue
		}
		seen = false
		for i, existingRule := range self.Rules {
			if existingRule.RulePriority == newRule.RulePriority {
				self.Rules[i] = newRule
				seen = true
				break
			}
		}
		if !seen {
			self.Rules = append(self.Rules, newRule)
		}
	}
}
