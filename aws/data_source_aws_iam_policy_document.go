package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
)

var dataSourceAwsIamPolicyDocumentVarReplacer = strings.NewReplacer("&{", "${")

func dataSourceAwsIamPolicyDocument() *schema.Resource {
	setOfString := &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	return &schema.Resource{
		Read: dataSourceAwsIamPolicyDocumentRead,

		Schema: map[string]*schema.Schema{
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"override_json": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"override_policy_documents": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"policy_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"source_json": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"source_policy_documents": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"statement": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"actions": setOfString,
						"condition": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"test": {
										Type:     schema.TypeString,
										Required: true,
									},
									"values": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"variable": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"effect": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "Allow",
							ValidateFunc: validation.StringInSlice([]string{"Allow", "Deny"}, false),
						},
						"not_actions":    setOfString,
						"not_principals": dataSourceAwsIamPolicyPrincipalSchema(),
						"not_resources":  setOfString,
						"principals":     dataSourceAwsIamPolicyPrincipalSchema(),
						"resources":      setOfString,
						"sid": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2012-10-17",
				ValidateFunc: validation.StringInSlice([]string{
					"2008-10-17",
					"2012-10-17",
				}, false),
			},
		},
	}
}

func dataSourceAwsIamPolicyDocumentRead(d *schema.ResourceData, meta interface{}) error {
	mergedDoc := &IAMPolicyDoc{}

	if v, ok := d.GetOk("source_json"); ok {
		if err := json.Unmarshal([]byte(v.(string)), mergedDoc); err != nil {
			return err
		}
	}

	if v, ok := d.GetOk("source_policy_documents"); ok && len(v.([]interface{})) > 0 {
		// generate sid map to assure there are no duplicates in source jsons
		sidMap := make(map[string]struct{})
		for _, stmt := range mergedDoc.Statements {
			if stmt.Sid != "" {
				sidMap[stmt.Sid] = struct{}{}
			}
		}

		// merge sourceDocs in order specified
		for sourceJSONIndex, sourceJSON := range v.([]interface{}) {
			sourceDoc := &IAMPolicyDoc{}
			if err := json.Unmarshal([]byte(sourceJSON.(string)), sourceDoc); err != nil {
				return err
			}

			// assure all statements in sourceDoc are unique before merging
			for stmtIndex, stmt := range sourceDoc.Statements {
				if stmt.Sid != "" {
					if _, sidExists := sidMap[stmt.Sid]; sidExists {
						return fmt.Errorf("duplicate Sid (%s) in source_policy_documents (item %d; statement %d). Remove the Sid or ensure Sids are unique.", stmt.Sid, sourceJSONIndex, stmtIndex)
					}
					sidMap[stmt.Sid] = struct{}{}
				}
			}

			mergedDoc.Merge(sourceDoc)
		}

	}

	// process the current document
	doc := &IAMPolicyDoc{
		Version: d.Get("version").(string),
	}

	if policyID, hasPolicyID := d.GetOk("policy_id"); hasPolicyID {
		doc.Id = policyID.(string)
	}

	if cfgStmts, hasCfgStmts := d.GetOk("statement"); hasCfgStmts {
		var cfgStmtIntf = cfgStmts.([]interface{})
		stmts := make([]*IAMPolicyStatement, len(cfgStmtIntf))
		sidMap := make(map[string]struct{})

		for i, stmtI := range cfgStmtIntf {
			cfgStmt := stmtI.(map[string]interface{})
			stmt := &IAMPolicyStatement{
				Effect: cfgStmt["effect"].(string),
			}

			if sid, ok := cfgStmt["sid"]; ok {
				if _, ok := sidMap[sid.(string)]; ok {
					return fmt.Errorf("duplicate Sid (%s). Remove the Sid or ensure the Sid is unique.", sid.(string))
				}
				stmt.Sid = sid.(string)
				if len(stmt.Sid) > 0 {
					sidMap[stmt.Sid] = struct{}{}
				}
			}

			if actions := cfgStmt["actions"].(*schema.Set).List(); len(actions) > 0 {
				stmt.Actions = iamPolicyDecodeConfigStringList(actions)
			}
			if actions := cfgStmt["not_actions"].(*schema.Set).List(); len(actions) > 0 {
				stmt.NotActions = iamPolicyDecodeConfigStringList(actions)
			}

			if resources := cfgStmt["resources"].(*schema.Set).List(); len(resources) > 0 {
				var err error
				stmt.Resources, err = dataSourceAwsIamPolicyDocumentReplaceVarsInList(
					iamPolicyDecodeConfigStringList(resources), doc.Version,
				)
				if err != nil {
					return fmt.Errorf("error reading resources: %w", err)
				}
			}
			if notResources := cfgStmt["not_resources"].(*schema.Set).List(); len(notResources) > 0 {
				var err error
				stmt.NotResources, err = dataSourceAwsIamPolicyDocumentReplaceVarsInList(
					iamPolicyDecodeConfigStringList(notResources), doc.Version,
				)
				if err != nil {
					return fmt.Errorf("error reading not_resources: %w", err)
				}
			}

			if principals := cfgStmt["principals"].(*schema.Set).List(); len(principals) > 0 {
				var err error
				stmt.Principals, err = dataSourceAwsIamPolicyDocumentMakePrincipals(principals, doc.Version)
				if err != nil {
					return fmt.Errorf("error reading principals: %w", err)
				}
			}

			if notPrincipals := cfgStmt["not_principals"].(*schema.Set).List(); len(notPrincipals) > 0 {
				var err error
				stmt.NotPrincipals, err = dataSourceAwsIamPolicyDocumentMakePrincipals(notPrincipals, doc.Version)
				if err != nil {
					return fmt.Errorf("error reading not_principals: %w", err)
				}
			}

			if conditions := cfgStmt["condition"].(*schema.Set).List(); len(conditions) > 0 {
				var err error
				stmt.Conditions, err = dataSourceAwsIamPolicyDocumentMakeConditions(conditions, doc.Version)
				if err != nil {
					return fmt.Errorf("error reading condition: %w", err)
				}
			}

			stmts[i] = stmt
		}

		doc.Statements = stmts

	}

	// merge our current document into mergedDoc
	mergedDoc.Merge(doc)

	// merge override_policy_documents policies into mergedDoc in order specified
	if v, ok := d.GetOk("override_policy_documents"); ok && len(v.([]interface{})) > 0 {
		for _, overrideJSON := range v.([]interface{}) {
			overrideDoc := &IAMPolicyDoc{}
			if err := json.Unmarshal([]byte(overrideJSON.(string)), overrideDoc); err != nil {
				return err
			}

			mergedDoc.Merge(overrideDoc)
		}

	}

	// merge in override_json
	if v, ok := d.GetOk("override_json"); ok {
		overrideDoc := &IAMPolicyDoc{}
		if err := json.Unmarshal([]byte(v.(string)), overrideDoc); err != nil {
			return err
		}

		mergedDoc.Merge(overrideDoc)
	}

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

func dataSourceAwsIamPolicyDocumentReplaceVarsInList(in interface{}, version string) (interface{}, error) {
	switch v := in.(type) {
	case string:
		if version == "2008-10-17" && strings.Contains(v, "&{") {
			return nil, fmt.Errorf("found &{ sequence in (%s), which is not supported in document version 2008-10-17", v)
		}
		return dataSourceAwsIamPolicyDocumentVarReplacer.Replace(v), nil
	case []string:
		out := make([]string, len(v))
		for i, item := range v {
			if version == "2008-10-17" && strings.Contains(item, "&{") {
				return nil, fmt.Errorf("found &{ sequence in (%s), which is not supported in document version 2008-10-17", item)
			}
			out[i] = dataSourceAwsIamPolicyDocumentVarReplacer.Replace(item)
		}
		return out, nil
	default:
		return nil, errors.New("dataSourceAwsIamPolicyDocumentReplaceVarsInList: input not string nor []string")
	}
}

func dataSourceAwsIamPolicyDocumentMakeConditions(in []interface{}, version string) (IAMPolicyStatementConditionSet, error) {
	out := make([]IAMPolicyStatementCondition, len(in))
	for i, itemI := range in {
		var err error
		item := itemI.(map[string]interface{})
		out[i] = IAMPolicyStatementCondition{
			Test:     item["test"].(string),
			Variable: item["variable"].(string),
		}
		out[i].Values, err = dataSourceAwsIamPolicyDocumentReplaceVarsInList(
			aws.StringValueSlice(expandStringListKeepEmpty(item["values"].([]interface{}))),
			version,
		)
		if err != nil {
			return nil, fmt.Errorf("error reading values: %w", err)
		}
	}
	return IAMPolicyStatementConditionSet(out), nil
}

func dataSourceAwsIamPolicyDocumentMakePrincipals(in []interface{}, version string) (IAMPolicyStatementPrincipalSet, error) {
	out := make([]IAMPolicyStatementPrincipal, len(in))
	for i, itemI := range in {
		var err error
		item := itemI.(map[string]interface{})
		out[i] = IAMPolicyStatementPrincipal{
			Type: item["type"].(string),
		}
		out[i].Identifiers, err = dataSourceAwsIamPolicyDocumentReplaceVarsInList(
			iamPolicyDecodeConfigStringList(
				item["identifiers"].(*schema.Set).List(),
			), version,
		)
		if err != nil {
			return nil, fmt.Errorf("error reading identifiers: %w", err)
		}
	}
	return IAMPolicyStatementPrincipalSet(out), nil
}

func dataSourceAwsIamPolicyPrincipalSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Type:     schema.TypeString,
					Required: true,
				},
				"identifiers": {
					Type:     schema.TypeSet,
					Required: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}
