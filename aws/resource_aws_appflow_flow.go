package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appflow"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsAppFlowFlow() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppFlowFlowCreate,
		Read:   resourceAwsAppFlowFlowRead,
		Update: resourceAwsAppFlowFlowUpdate,
		Delete: resourceAwsAppFlowFlowDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 2048),
					validation.StringMatch(regexp.MustCompile(`^[\w!@#\-.?,\s]*`), "must match [\\w!@#\\-.?,\\s]*"),
				),
			},
			"destination_flow_config_list": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connector_profile_name": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 256),
								validation.StringMatch(regexp.MustCompile(`[\w/!@#+=.-]+`), "must match [\\w/!@#+=.-]+"),
							),
						},
						"connector_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appflow.ConnectorType_Values(), false),
						},
						"destination_connector_properties": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"customer_profiles": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"domain_name": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 64),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"object_type_name": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 255),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"event_bridge": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"error_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"bucket_name": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(3, 63),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
															"bucket_prefix": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"fail_on_first_destination_error": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"honeycode": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"error_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"bucket_name": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(3, 63),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
															"bucket_prefix": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"fail_on_first_destination_error": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"redshift": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket_prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"error_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"bucket_name": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(3, 63),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
															"bucket_prefix": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"fail_on_first_destination_error": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"intermediate_bucket_name": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(3, 63),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"s3": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket_name": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(3, 63),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"bucket_prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"s3_output_format_config": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"aggregation_config": {
																Type:     schema.TypeList,
																Optional: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"aggregation_type": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringInSlice(appflow.AggregationType_Values(), false),
																		},
																	},
																},
															},
															"file_type": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringInSlice(appflow.FileType_Values(), false),
															},
															"prefix_config": {
																Type:     schema.TypeList,
																Optional: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"prefix_format": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringInSlice(appflow.PrefixFormat_Values(), false),
																		},
																		"prefix_type": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringInSlice(appflow.PrefixType_Values(), false),
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
									"salesforce": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"error_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"bucket_name": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(3, 63),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
															"bucket_prefix": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"fail_on_first_destination_error": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"id_field_names": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringLenBetween(1, 128),
													},
												},
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"write_operation_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(appflow.WriteOperationType_Values(), false),
												},
											},
										},
									},
									"snowflake": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket_prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"error_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"bucket_name": {
																Type:     schema.TypeString,
																Optional: true,
																ValidateFunc: validation.All(
																	validation.StringLenBetween(3, 63),
																	validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
																),
															},
															"bucket_prefix": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"fail_on_first_destination_error": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"intermediate_bucket_name": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(3, 63),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"upsolver": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket_name": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(16, 63),
														validation.StringMatch(regexp.MustCompile(`^(upsolver-appflow)\S*`), "must match ^(upsolver-appflow)\\S*"),
													),
												},
												"bucket_prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"s3_output_format_config": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"aggregation_config": {
																Type:     schema.TypeList,
																Optional: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"aggregation_type": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringInSlice(appflow.AggregationType_Values(), false),
																		},
																	},
																},
															},
															"file_type": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringInSlice(appflow.FileType_Values(), false),
															},
															"prefix_config": {
																Type:     schema.TypeList,
																Optional: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"prefix_format": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringInSlice(appflow.PrefixFormat_Values(), false),
																		},
																		"prefix_type": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringInSlice(appflow.PrefixType_Values(), false),
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"flow_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"flow_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 256),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][\w!@#.-]+$`), "must match [a-zA-Z0-9][\\w!@#.-]+"),
				),
			},
			"flow_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"source_flow_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connector_profile_name": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 256),
								validation.StringMatch(regexp.MustCompile(`[\w/!@#+=.-]+`), "must match [\\w/!@#+=.-]+"),
							),
						},
						"connector_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appflow.ConnectorType_Values(), false),
						},
						"incremental_pull_config": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"datetime_type_field_name": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
								},
							},
						},
						"source_connector_properties": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"amplitude": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"datadog": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"dynatrace": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"google_analytics": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"infor_nexus": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"marketo": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"s3": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket_name": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(3, 63),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
												"bucket_prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
											},
										},
									},
									"salesforce": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enable_dynamic_field_update": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"include_deleted_records": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"service_now": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"singular": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"slack": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"trendmicro": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"veeva": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
									"zendesk": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 512),
														validation.StringMatch(regexp.MustCompile(`\S+`), "must match \\S+"),
													),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"tags": tagsSchema(),
			"tasks": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connector_operator": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"amplitude": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.AmplitudeConnectorOperator_Values(), false),
									},
									"datadog": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.DatadogConnectorOperator_Values(), false),
									},
									"dynatrace": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.DynatraceConnectorOperator_Values(), false),
									},
									"google_analytics": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.GoogleAnalyticsConnectorOperator_Values(), false),
									},
									"infor_nexus": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.InforNexusConnectorOperator_Values(), false),
									},
									"marketo": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.MarketoConnectorOperator_Values(), false),
									},
									"s3": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.S3ConnectorOperator_Values(), false),
									},
									"salesforce": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.SalesforceConnectorOperator_Values(), false),
									},
									"service_now": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.ServiceNowConnectorOperator_Values(), false),
									},
									"singular": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.SingularConnectorOperator_Values(), false),
									},
									"slack": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.SlackConnectorOperator_Values(), false),
									},
									"trendmicro": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.TrendmicroConnectorOperator_Values(), false),
									},
									"veeva": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.VeevaConnectorOperator_Values(), false),
									},
									"zendesk": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.ZendeskConnectorOperator_Values(), false),
									},
								},
							},
						},
						"destination_field": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
						"source_fields": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 2048),
							},
						},
						"task_properties": {
							Type:     schema.TypeMap,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"task_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appflow.TaskType_Values(), false),
						},
					},
				},
			},
			"trigger_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"trigger_properties": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"scheduled": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"data_pull_mode": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(appflow.DataPullMode_Values(), false),
												},
												"first_execution_from": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"schedule_end_time": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"schedule_expression": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(0, 256),
												},
												"schedule_offset": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(0, 36000),
												},
												"schedule_start_time": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"timezone": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 256),
												},
											},
										},
									},
								},
							},
						},
						"trigger_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appflow.TriggerType_Values(), false),
						},
					},
				},
			},
		},
	}
}

func expandTasks(lst []interface{}) []*appflow.Task {
	var items []*appflow.Task

	for _, v := range lst {
		items = append(items, expandTask(v.(map[string]interface{})))
	}

	return items
}

func expandTask(m map[string]interface{}) *appflow.Task {
	t := &appflow.Task{
		TaskProperties: expandStringMap(m["task_properties"].(map[string]interface{})),
		TaskType:       aws.String(m["task_type"].(string)),
		SourceFields:   expandStringList(m["source_fields"].([]interface{})),
	}

	if v, ok := m["destination_field"].(string); ok && v != "" {
		t.DestinationField = aws.String(v)
	}

	if v, ok := m["connector_operator"].([]interface{}); ok && len(v) > 0 {
		t.ConnectorOperator = expandConnectorOperator(v[0].(map[string]interface{}))
	}

	return t
}

func expandConnectorOperator(m map[string]interface{}) *appflow.ConnectorOperator {
	co := appflow.ConnectorOperator{}

	if v, ok := m["amplitude"].(string); ok && v != "" {
		co.Amplitude = aws.String(v)
	}
	if v, ok := m["datadog"].(string); ok && v != "" {
		co.Datadog = aws.String(v)
	}
	if v, ok := m["dynatrace"].(string); ok && v != "" {
		co.Dynatrace = aws.String(v)
	}
	if v, ok := m["google_analytics"].(string); ok && v != "" {
		co.GoogleAnalytics = aws.String(v)
	}
	if v, ok := m["infor_nexus"].(string); ok && v != "" {
		co.InforNexus = aws.String(v)
	}
	if v, ok := m["marketo"].(string); ok && v != "" {
		co.Marketo = aws.String(v)
	}
	if v, ok := m["s3"].(string); ok && v != "" {
		co.S3 = aws.String(v)
	}
	if v, ok := m["salesforce"].(string); ok && v != "" {
		co.Salesforce = aws.String(v)
	}
	if v, ok := m["service_now"].(string); ok && v != "" {
		co.ServiceNow = aws.String(v)
	}
	if v, ok := m["singular"].(string); ok && v != "" {
		co.Singular = aws.String(v)
	}
	if v, ok := m["slack"].(string); ok && v != "" {
		co.Slack = aws.String(v)
	}
	if v, ok := m["trendmicro"].(string); ok && v != "" {
		co.Trendmicro = aws.String(v)
	}
	if v, ok := m["veeva"].(string); ok && v != "" {
		co.Veeva = aws.String(v)
	}
	if v, ok := m["zendesk"].(string); ok && v != "" {
		co.Zendesk = aws.String(v)
	}

	return &co
}

func expandTriggerConfig(m map[string]interface{}) *appflow.TriggerConfig {
	tc := &appflow.TriggerConfig{
		TriggerType: aws.String(m["trigger_type"].(string)),
	}

	if v, ok := m["trigger_properties"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		tp := expandTriggerProperties(v[0].(map[string]interface{}))

		if tp != nil {
			tc.TriggerProperties = tp
		}
	}

	return tc
}

func expandTriggerProperties(m map[string]interface{}) *appflow.TriggerProperties {
	if v, ok := m["scheduled"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		tp := appflow.TriggerProperties{}
		tp.Scheduled = expandScheduledTriggerProperties(v[0].(map[string]interface{}))
		return &tp
	}

	return nil
}

func expandScheduledTriggerProperties(m map[string]interface{}) *appflow.ScheduledTriggerProperties {
	stp := appflow.ScheduledTriggerProperties{
		ScheduleExpression: aws.String(m["schedule_expression"].(string)),
	}

	if v, ok := m["data_pull_mode"].(string); ok && v != "" {
		stp.DataPullMode = aws.String(v)
	}
	if v, ok := m["first_execution_from"].(string); ok && v != "" {
		fef, _ := time.Parse(time.RFC3339, v)
		stp.FirstExecutionFrom = aws.Time(fef)
	}
	if v, ok := m["schedule_end_time"].(string); ok && v != "" {
		set, _ := time.Parse(time.RFC3339, v)
		stp.ScheduleEndTime = aws.Time(set)
	}
	if v, ok := m["schedule_offset"].(int64); ok && v > 0 {
		stp.ScheduleOffset = aws.Int64(v)
	}
	if v, ok := m["schedule_start_time"].(string); ok && v != "" {
		sst, _ := time.Parse(time.RFC3339, v)
		stp.ScheduleStartTime = aws.Time(sst)
	}
	if v, ok := m["timezone"].(string); ok && v != "" {
		stp.Timezone = aws.String(v)
	}

	return &stp
}

func expandDestinationFlowConfigList(lst []interface{}) []*appflow.DestinationFlowConfig {
	var items []*appflow.DestinationFlowConfig

	for _, v := range lst {
		items = append(items, expandDestinationFlowConfig(v.(map[string]interface{})))
	}

	return items
}

func expandDestinationFlowConfig(m map[string]interface{}) *appflow.DestinationFlowConfig {
	dfc := appflow.DestinationFlowConfig{
		ConnectorType:                  aws.String(m["connector_type"].(string)),
		DestinationConnectorProperties: expandDestinationConnectorProperties(m["destination_connector_properties"].([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := m["connector_profile_name"].(string); ok && v != "" {
		dfc.ConnectorProfileName = aws.String(v)
	}

	return &dfc
}

func expandDestinationConnectorProperties(m map[string]interface{}) *appflow.DestinationConnectorProperties {
	dcp := appflow.DestinationConnectorProperties{}

	if v, ok := m["customer_profiles"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		dcp.CustomerProfiles = expandCustomerProfilesDestinationProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["event_bridge"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		dcp.EventBridge = expandEventBridgeDestinationProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["honeycode"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		dcp.Honeycode = expandHoneycodeDestinationProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["redshift"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		dcp.Redshift = expandRedshiftDestinationProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["s3"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		dcp.S3 = expandS3DestinationProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["salesforce"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		dcp.Salesforce = expandSalesforceDestinationProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["snowflake"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		dcp.Snowflake = expandSnowflakeDestinationProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["upsolver"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		dcp.Upsolver = expandUpsolverDestinationProperties(v[0].(map[string]interface{}))
	}

	return &dcp
}

func expandCustomerProfilesDestinationProperties(m map[string]interface{}) *appflow.CustomerProfilesDestinationProperties {
	cpdp := appflow.CustomerProfilesDestinationProperties{
		DomainName: aws.String(m["domain_name"].(string)),
	}

	if v, ok := m["object_type_name"].(string); ok && v != "" {
		cpdp.ObjectTypeName = aws.String(v)
	}

	return &cpdp
}

func expandEventBridgeDestinationProperties(m map[string]interface{}) *appflow.EventBridgeDestinationProperties {
	ebdp := appflow.EventBridgeDestinationProperties{
		Object: aws.String(m["object"].(string)),
	}

	if v, ok := m["error_handling_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		ebdp.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	return &ebdp
}

func expandHoneycodeDestinationProperties(m map[string]interface{}) *appflow.HoneycodeDestinationProperties {
	hdp := appflow.HoneycodeDestinationProperties{
		Object: aws.String(m["object"].(string)),
	}

	if v, ok := m["error_handling_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		hdp.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	return &hdp
}

func expandRedshiftDestinationProperties(m map[string]interface{}) *appflow.RedshiftDestinationProperties {
	rdp := appflow.RedshiftDestinationProperties{
		IntermediateBucketName: aws.String(m["intermediate_bucket_name"].(string)),
		Object:                 aws.String(m["object"].(string)),
	}

	if v, ok := m["bucket_prefix"].(string); ok && v != "" {
		rdp.BucketPrefix = aws.String(v)
	}
	if v, ok := m["error_handling_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		rdp.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	return &rdp
}

func expandS3DestinationProperties(m map[string]interface{}) *appflow.S3DestinationProperties {
	s3dp := appflow.S3DestinationProperties{
		BucketName: aws.String(m["bucket_name"].(string)),
	}

	if v, ok := m["bucket_prefix"].(string); ok && v != "" {
		s3dp.BucketPrefix = aws.String(v)
	}

	if v, ok := m["s3_output_format_config"].([]interface{}); ok && len(v) > 0 {
		s3dp.S3OutputFormatConfig = expandS3OutputFormatConfig(v[0].(map[string]interface{}))
	}

	return &s3dp
}

func expandSalesforceDestinationProperties(m map[string]interface{}) *appflow.SalesforceDestinationProperties {
	sdp := appflow.SalesforceDestinationProperties{
		Object: aws.String(m["object"].(string)),
	}

	if v, ok := m["error_handling_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		sdp.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}
	if v, ok := m["id_field_names"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		sdp.IdFieldNames = expandStringList(v)
	}
	if v, ok := m["write_operation_type"].(string); ok && v != "" {
		sdp.WriteOperationType = aws.String(v)
	}

	return &sdp
}

func expandSnowflakeDestinationProperties(m map[string]interface{}) *appflow.SnowflakeDestinationProperties {
	sdp := appflow.SnowflakeDestinationProperties{
		IntermediateBucketName: aws.String(m["intermediate_bucket_name"].(string)),
		Object:                 aws.String(m["object"].(string)),
	}

	if v, ok := m["bucket_prefix"].(string); ok && v != "" {
		sdp.BucketPrefix = aws.String(v)
	}
	if v, ok := m["error_handling_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		sdp.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	return &sdp
}

func expandUpsolverDestinationProperties(m map[string]interface{}) *appflow.UpsolverDestinationProperties {
	udp := appflow.UpsolverDestinationProperties{
		BucketName:           aws.String(m["bucket_name"].(string)),
		S3OutputFormatConfig: expandUpsolverS3OutputFormatConfig(m["s3_output_format_config"].([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := m["bucket_prefix"].(string); ok && v != "" {
		udp.BucketPrefix = aws.String(v)
	}

	return &udp
}

func expandErrorHandlingConfig(m map[string]interface{}) *appflow.ErrorHandlingConfig {
	ehc := appflow.ErrorHandlingConfig{}

	if v, ok := m["bucket_name"].(string); ok && v != "" {
		ehc.BucketName = aws.String(v)
	}
	if v, ok := m["bucket_prefix"].(string); ok && v != "" {
		ehc.BucketPrefix = aws.String(v)
	}
	if v, ok := m["fail_on_first_destination_error"].(bool); ok {
		ehc.FailOnFirstDestinationError = aws.Bool(v)
	}

	return &ehc
}

func expandS3OutputFormatConfig(m map[string]interface{}) *appflow.S3OutputFormatConfig {
	s3ofc := appflow.S3OutputFormatConfig{}

	if v, ok := m["aggregation_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		s3ofc.AggregationConfig = expandAggregationConfig(v[0].(map[string]interface{}))
	}

	if v, ok := m["file_type"].(string); ok && v != "" {
		s3ofc.FileType = aws.String(v)
	}

	if v, ok := m["prefix_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		s3ofc.PrefixConfig = expandPrefixConfig(v[0].(map[string]interface{}))
	}

	return &s3ofc
}

func expandUpsolverS3OutputFormatConfig(m map[string]interface{}) *appflow.UpsolverS3OutputFormatConfig {
	us3ofc := appflow.UpsolverS3OutputFormatConfig{
		PrefixConfig: expandPrefixConfig(m["prefix_config"].([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := m["aggregation_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		us3ofc.AggregationConfig = expandAggregationConfig(v[0].(map[string]interface{}))
	}
	if v, ok := m["file_type"].(string); ok && v != "" {
		us3ofc.FileType = aws.String(v)
	}

	return &us3ofc
}

func expandPrefixConfig(m map[string]interface{}) *appflow.PrefixConfig {
	pc := appflow.PrefixConfig{}

	if v, ok := m["prefix_format"].(string); ok && v != "" {
		pc.PrefixFormat = aws.String(v)
	}

	if v, ok := m["prefix_type"].(string); ok && v != "" {
		pc.PrefixType = aws.String(v)
	}

	return &pc
}

func expandAggregationConfig(m map[string]interface{}) *appflow.AggregationConfig {
	ac := appflow.AggregationConfig{}

	if v, ok := m["aggregation_type"].(string); ok && v != "" {
		ac.AggregationType = aws.String(v)
	}

	return &ac
}

func expandSourceFlowConfig(m map[string]interface{}) *appflow.SourceFlowConfig {
	sfc := appflow.SourceFlowConfig{
		ConnectorType:             aws.String(m["connector_type"].(string)),
		SourceConnectorProperties: expandSourceConnectorProperties(m["source_connector_properties"].([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := m["connector_profile_name"].(string); ok && v != "" {
		sfc.ConnectorProfileName = aws.String(v)
	}

	if v, ok := m["incremental_pull_config"].([]interface{}); ok && len(v) > 0 {
		sfc.IncrementalPullConfig = expandIncrementalPullConfig(v[0].(map[string]interface{}))
	}

	return &sfc
}

func expandIncrementalPullConfig(m map[string]interface{}) *appflow.IncrementalPullConfig {
	ipc := appflow.IncrementalPullConfig{}

	if v, ok := m["datetime_type_field_name"].(string); ok && v != "" {
		ipc.DatetimeTypeFieldName = aws.String(v)
	}

	return &ipc
}

func expandSourceConnectorProperties(m map[string]interface{}) *appflow.SourceConnectorProperties {
	scp := appflow.SourceConnectorProperties{}

	if v, ok := m["amplitude"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		scp.Amplitude = expandAmplitudeSourceProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["datadog"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		scp.Datadog = expandDatadogSourceProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["dynatrace"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		scp.Dynatrace = expandDynatraceSourceProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["google_analytics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		scp.GoogleAnalytics = expandGoogleAnalyticsSourceProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["infor_nexus"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		scp.InforNexus = expandInforNexusSourceProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["marketo"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		scp.Marketo = expandMarketoSourceProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["s3"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		scp.S3 = expandS3SourceProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["salesforce"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		scp.Salesforce = expandSalesforceSourceProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["service_now"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		scp.ServiceNow = expandServiceNowSourceProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["singular"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		scp.Singular = expandSingularSourceProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["slack"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		scp.Slack = expandSlackSourceProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["trendmicro"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		scp.Trendmicro = expandTrendmicroSourceProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["veeva"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		scp.Veeva = expandVeevaSourceProperties(v[0].(map[string]interface{}))
	}
	if v, ok := m["zendesk"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		scp.Zendesk = expandZendeskSourceProperties(v[0].(map[string]interface{}))
	}

	return &scp
}

func expandAmplitudeSourceProperties(m map[string]interface{}) *appflow.AmplitudeSourceProperties {
	return &appflow.AmplitudeSourceProperties{
		Object: aws.String(m["object"].(string)),
	}
}

func expandDatadogSourceProperties(m map[string]interface{}) *appflow.DatadogSourceProperties {
	return &appflow.DatadogSourceProperties{
		Object: aws.String(m["object"].(string)),
	}
}

func expandDynatraceSourceProperties(m map[string]interface{}) *appflow.DynatraceSourceProperties {
	return &appflow.DynatraceSourceProperties{
		Object: aws.String(m["object"].(string)),
	}
}

func expandGoogleAnalyticsSourceProperties(m map[string]interface{}) *appflow.GoogleAnalyticsSourceProperties {
	return &appflow.GoogleAnalyticsSourceProperties{
		Object: aws.String(m["object"].(string)),
	}
}

func expandInforNexusSourceProperties(m map[string]interface{}) *appflow.InforNexusSourceProperties {
	return &appflow.InforNexusSourceProperties{
		Object: aws.String(m["object"].(string)),
	}
}

func expandMarketoSourceProperties(m map[string]interface{}) *appflow.MarketoSourceProperties {
	return &appflow.MarketoSourceProperties{
		Object: aws.String(m["object"].(string)),
	}
}

func expandS3SourceProperties(m map[string]interface{}) *appflow.S3SourceProperties {
	return &appflow.S3SourceProperties{
		BucketName:   aws.String(m["bucket_name"].(string)),
		BucketPrefix: aws.String(m["bucket_prefix"].(string)),
	}
}

func expandSalesforceSourceProperties(m map[string]interface{}) *appflow.SalesforceSourceProperties {
	ssp := &appflow.SalesforceSourceProperties{
		EnableDynamicFieldUpdate: aws.Bool(m["enable_dynamic_field_update"].(bool)),
		IncludeDeletedRecords:    aws.Bool(m["include_deleted_records"].(bool)),
		Object:                   aws.String(m["object"].(string)),
	}

	return ssp
}

func expandServiceNowSourceProperties(m map[string]interface{}) *appflow.ServiceNowSourceProperties {
	return &appflow.ServiceNowSourceProperties{
		Object: aws.String(m["object"].(string)),
	}
}

func expandSingularSourceProperties(m map[string]interface{}) *appflow.SingularSourceProperties {
	return &appflow.SingularSourceProperties{
		Object: aws.String(m["object"].(string)),
	}
}

func expandSlackSourceProperties(m map[string]interface{}) *appflow.SlackSourceProperties {
	return &appflow.SlackSourceProperties{
		Object: aws.String(m["object"].(string)),
	}
}

func expandTrendmicroSourceProperties(m map[string]interface{}) *appflow.TrendmicroSourceProperties {
	return &appflow.TrendmicroSourceProperties{
		Object: aws.String(m["object"].(string)),
	}
}

func expandVeevaSourceProperties(m map[string]interface{}) *appflow.VeevaSourceProperties {
	return &appflow.VeevaSourceProperties{
		Object: aws.String(m["object"].(string)),
	}
}

func expandZendeskSourceProperties(m map[string]interface{}) *appflow.ZendeskSourceProperties {
	return &appflow.ZendeskSourceProperties{
		Object: aws.String(m["object"].(string)),
	}
}

func resourceAwsAppFlowFlowCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appflowconn
	flowName := d.Get("flow_name").(string)

	createFlowInput := appflow.CreateFlowInput{
		Description:               aws.String(d.Get("description").(string)),
		DestinationFlowConfigList: expandDestinationFlowConfigList(d.Get("destination_flow_config_list").([]interface{})),
		FlowName:                  aws.String(flowName),
		SourceFlowConfig:          expandSourceFlowConfig(d.Get("source_flow_config").([]interface{})[0].(map[string]interface{})),
		Tags:                      keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AppflowTags(),
		Tasks:                     expandTasks(d.Get("tasks").([]interface{})),
		TriggerConfig:             expandTriggerConfig(d.Get("trigger_config").([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := d.Get("kms_arn").(string); ok && len(v) > 0 {
		createFlowInput.KmsArn = aws.String(v)
	}

	_, err := conn.CreateFlow(&createFlowInput)

	if err != nil {
		return fmt.Errorf("error creating AppFlow flow: %w", err)
	}

	d.SetId(flowName)

	return resourceAwsAppFlowFlowRead(d, meta)
}

func resourceAwsAppFlowFlowRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appflowconn

	resp, err := conn.DescribeFlow(&appflow.DescribeFlowInput{
		FlowName: aws.String(d.Id()),
	})

	if !d.IsNewResource() && isAWSErr(err, appflow.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] AppFlow flow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("created_at", aws.TimeValue(resp.CreatedAt).Format(time.RFC3339))
	d.Set("created_by", resp.CreatedBy)
	d.Set("description", resp.Description)
	d.Set("destination_flow_config_list", flattenDestinationFlowConfigList(resp.DestinationFlowConfigList))
	d.Set("flow_arn", resp.FlowArn)
	d.Set("flow_name", resp.FlowName)
	d.Set("flow_status", resp.FlowStatus)
	d.Set("kms_arn", resp.KmsArn)
	d.Set("source_flow_config", flattenSourceFlowConfig(resp.SourceFlowConfig))
	d.Set("tasks", flattenTasks(resp.Tasks))
	d.Set("trigger_config", flattenTriggerConfig(resp.TriggerConfig))

	if err != nil {
		return err
	}

	return nil
}

func flattenDestinationFlowConfigList(dfc []*appflow.DestinationFlowConfig) []interface{} {
	result := make([]interface{}, len(dfc))

	for i, e := range dfc {
		m := make(map[string]interface{})

		if e.ConnectorProfileName != nil {
			m["connector_profile_name"] = e.ConnectorProfileName
		}
		m["connector_type"] = e.ConnectorType
		m["destination_connector_properties"] = flattenDestinationConnectorProperties(e.DestinationConnectorProperties)

		result[i] = m
	}

	return result
}

func flattenDestinationConnectorProperties(dcp *appflow.DestinationConnectorProperties) []interface{} {
	result := make(map[string]interface{})

	if dcp.CustomerProfiles != nil {
		result["customer_profiles"] = flattenCustomerProfilesDestinationProperties(dcp.CustomerProfiles)
	}
	if dcp.EventBridge != nil {
		result["event_bridge"] = flattenEventBridgeDestinationProperties(dcp.EventBridge)
	}
	if dcp.Honeycode != nil {
		result["honeycode"] = flattenHoneycodeDestinationProperties(dcp.Honeycode)
	}
	if dcp.Redshift != nil {
		result["redshift"] = flattenRedshiftDestinationProperties(dcp.Redshift)
	}
	if dcp.S3 != nil {
		result["s3"] = flattenS3DestinationProperties(dcp.S3)
	}
	if dcp.Salesforce != nil {
		result["salesforce"] = flattenSalesforceDestinationProperties(dcp.Salesforce)
	}
	if dcp.Snowflake != nil {
		result["snowflake"] = flattenSnowflakeDestinationProperties(dcp.Snowflake)
	}
	if dcp.Upsolver != nil {
		result["upsolver"] = flattenUpsolverDestinationProperties(dcp.Upsolver)
	}

	return []interface{}{result}
}

func flattenCustomerProfilesDestinationProperties(cpdp *appflow.CustomerProfilesDestinationProperties) []interface{} {
	m := make(map[string]interface{})

	m["domain_name"] = cpdp.DomainName
	if cpdp.ObjectTypeName != nil {
		m["object_type_name"] = cpdp.ObjectTypeName
	}

	return []interface{}{m}
}

func flattenEventBridgeDestinationProperties(ebdp *appflow.EventBridgeDestinationProperties) []interface{} {
	m := make(map[string]interface{})

	m["object"] = ebdp.Object

	if ebdp.ErrorHandlingConfig != nil {
		m["error_handling_config"] = flattenErrorHandlingConfig(ebdp.ErrorHandlingConfig)
	}

	return []interface{}{m}
}

func flattenHoneycodeDestinationProperties(hdp *appflow.HoneycodeDestinationProperties) []interface{} {
	m := make(map[string]interface{})

	m["object"] = hdp.Object

	if hdp.ErrorHandlingConfig != nil {
		m["error_handling_config"] = flattenErrorHandlingConfig(hdp.ErrorHandlingConfig)
	}

	return []interface{}{m}
}

func flattenRedshiftDestinationProperties(rdp *appflow.RedshiftDestinationProperties) []interface{} {
	m := make(map[string]interface{})

	if rdp.BucketPrefix != nil {
		m["bucket_prefix"] = rdp.BucketPrefix
	}
	if rdp.ErrorHandlingConfig != nil {
		m["error_handling_config"] = flattenErrorHandlingConfig(rdp.ErrorHandlingConfig)
	}
	m["intermediate_bucket_name"] = rdp.IntermediateBucketName
	m["object"] = rdp.Object

	return []interface{}{m}
}

func flattenS3DestinationProperties(s3dp *appflow.S3DestinationProperties) []interface{} {
	m := make(map[string]interface{})

	m["bucket_name"] = s3dp.BucketName
	if s3dp.BucketPrefix != nil {
		m["bucket_prefix"] = s3dp.BucketPrefix
	}
	if s3dp.S3OutputFormatConfig != nil {
		m["s3_output_format_config"] = flatenS3OutputFormatConfig(s3dp.S3OutputFormatConfig)
	}

	return []interface{}{m}
}

func flattenSalesforceDestinationProperties(sdp *appflow.SalesforceDestinationProperties) []interface{} {
	m := make(map[string]interface{})

	if sdp.ErrorHandlingConfig != nil {
		m["error_handling_config"] = flattenErrorHandlingConfig(sdp.ErrorHandlingConfig)
	}
	if sdp.IdFieldNames != nil {
		m["id_field_names"] = sdp.IdFieldNames
	}
	m["object"] = sdp.Object
	if sdp.WriteOperationType != nil {
		m["write_operation_type"] = sdp.WriteOperationType
	}

	return []interface{}{m}
}

func flattenSnowflakeDestinationProperties(sdp *appflow.SnowflakeDestinationProperties) []interface{} {
	m := make(map[string]interface{})

	if sdp.BucketPrefix != nil {
		m["bucket_prefix"] = sdp.BucketPrefix
	}
	if sdp.ErrorHandlingConfig != nil {
		m["error_handling_config"] = flattenErrorHandlingConfig(sdp.ErrorHandlingConfig)
	}
	m["intermediate_bucket_name"] = sdp.IntermediateBucketName
	m["object"] = sdp.Object

	return []interface{}{m}
}

func flattenUpsolverDestinationProperties(udp *appflow.UpsolverDestinationProperties) []interface{} {
	m := make(map[string]interface{})

	m["bucket_name"] = udp.BucketName
	if udp.BucketPrefix != nil {
		m["bucket_prefix"] = udp.BucketPrefix
	}
	m["s3_output_format_config"] = flattenUpsolverS3OutputFormatConfig(udp.S3OutputFormatConfig)

	return []interface{}{m}
}

func flatenS3OutputFormatConfig(s3ofc *appflow.S3OutputFormatConfig) []interface{} {
	m := make(map[string]interface{})

	if s3ofc.AggregationConfig != nil {
		m["aggregation_config"] = flattenAggregationConfig(s3ofc.AggregationConfig)
	}
	if s3ofc.FileType != nil {
		m["file_type"] = s3ofc.FileType
	}
	if s3ofc.PrefixConfig != nil {
		m["prefix_config"] = flattenPrefixConfig(s3ofc.PrefixConfig)
	}

	return []interface{}{m}
}

func flattenUpsolverS3OutputFormatConfig(us3ofc *appflow.UpsolverS3OutputFormatConfig) []interface{} {
	m := make(map[string]interface{})

	if us3ofc.AggregationConfig != nil {
		m["aggregation_config"] = flattenAggregationConfig(us3ofc.AggregationConfig)
	}
	if us3ofc.FileType != nil {
		m["file_type"] = us3ofc.FileType
	}
	m["prefix_config"] = flattenPrefixConfig(us3ofc.PrefixConfig)

	return []interface{}{m}
}

func flattenAggregationConfig(ac *appflow.AggregationConfig) []interface{} {
	m := make(map[string]interface{})

	if ac.AggregationType != nil {
		m["aggregation_type"] = ac.AggregationType
	}

	return []interface{}{m}
}

func flattenPrefixConfig(pc *appflow.PrefixConfig) []interface{} {
	m := make(map[string]interface{})

	if pc.PrefixFormat != nil {
		m["prefix_format"] = pc.PrefixFormat
	}
	if pc.PrefixType != nil {
		m["prefix_type"] = pc.PrefixType
	}

	if len(m) > 0 {
		return []interface{}{m}
	}

	return []interface{}{}
}

func flattenErrorHandlingConfig(ehc *appflow.ErrorHandlingConfig) []interface{} {
	m := make(map[string]interface{})

	if ehc.BucketName != nil {
		m["bucket_name"] = ehc.BucketName
	}
	if ehc.BucketPrefix != nil {
		m["bucket_prefix"] = ehc.BucketPrefix
	}
	if ehc.FailOnFirstDestinationError != nil {
		m["fail_on_first_destination_error"] = ehc.FailOnFirstDestinationError
	}

	return []interface{}{m}
}

func flattenSourceFlowConfig(sfc *appflow.SourceFlowConfig) []interface{} {
	m := make(map[string]interface{})

	if sfc.ConnectorProfileName != nil {
		m["connector_profile_name"] = sfc.ConnectorProfileName
	}
	m["connector_type"] = sfc.ConnectorType
	if sfc.IncrementalPullConfig != nil {
		m["incremental_pull_config"] = flattenIncrementalPullConfig(sfc.IncrementalPullConfig)
	}
	m["source_connector_properties"] = flattenSourceConnectorProperties(sfc.SourceConnectorProperties)

	return []interface{}{m}
}

func flattenIncrementalPullConfig(ipc *appflow.IncrementalPullConfig) []interface{} {
	m := make(map[string]interface{})

	if ipc.DatetimeTypeFieldName != nil {
		m["datetime_type_field_name"] = aws.StringValue(ipc.DatetimeTypeFieldName)
	}

	return []interface{}{m}
}

func flattenSourceConnectorProperties(scp *appflow.SourceConnectorProperties) []interface{} {
	result := make(map[string]interface{})

	if scp.Amplitude != nil {
		m := make(map[string]interface{})
		m["object"] = aws.StringValue(scp.Amplitude.Object)
		result["amplitude"] = []interface{}{m}
	}

	if scp.Datadog != nil {
		m := make(map[string]interface{})
		m["object"] = aws.StringValue(scp.Datadog.Object)
		result["datadog"] = []interface{}{m}
	}

	if scp.Dynatrace != nil {
		m := make(map[string]interface{})
		m["object"] = aws.StringValue(scp.Dynatrace.Object)
		result["dynatrace"] = []interface{}{m}
	}

	if scp.GoogleAnalytics != nil {
		m := make(map[string]interface{})
		m["object"] = aws.StringValue(scp.GoogleAnalytics.Object)
		result["google_analytics"] = []interface{}{m}
	}

	if scp.InforNexus != nil {
		m := make(map[string]interface{})
		m["object"] = aws.StringValue(scp.InforNexus.Object)
		result["infor_nexus"] = []interface{}{m}
	}

	if scp.Marketo != nil {
		m := make(map[string]interface{})
		m["object"] = aws.StringValue(scp.Marketo.Object)
		result["marketo"] = []interface{}{m}
	}

	if scp.S3 != nil {
		m := make(map[string]interface{})
		m["bucket_name"] = aws.StringValue(scp.S3.BucketName)
		m["bucket_prefix"] = aws.StringValue(scp.S3.BucketPrefix)
		result["s3"] = []interface{}{m}
	}

	if scp.Salesforce != nil {
		m := make(map[string]interface{})
		m["enable_dynamic_field_update"] = aws.BoolValue(scp.Salesforce.EnableDynamicFieldUpdate)
		m["include_deleted_records"] = aws.BoolValue(scp.Salesforce.IncludeDeletedRecords)
		m["object"] = aws.StringValue(scp.Salesforce.Object)
		result["salesforce"] = []interface{}{m}
	}

	if scp.ServiceNow != nil {
		m := make(map[string]interface{})
		m["object"] = scp.ServiceNow.Object
		m["service_now"] = []interface{}{m}
	}

	if scp.Singular != nil {
		m := make(map[string]interface{})
		m["object"] = scp.Singular.Object
		result["singular"] = []interface{}{m}
	}

	if scp.Slack != nil {
		m := make(map[string]interface{})
		m["object"] = scp.Slack.Object
		result["slack"] = []interface{}{m}
	}

	if scp.Trendmicro != nil {
		m := make(map[string]interface{})
		m["object"] = scp.Trendmicro.Object
		result["trendmicro"] = []interface{}{m}
	}

	if scp.Veeva != nil {
		m := make(map[string]interface{})
		m["object"] = scp.Veeva.Object
		result["veeva"] = []interface{}{m}
	}

	if scp.Zendesk != nil {
		m := make(map[string]interface{})
		m["object"] = scp.Zendesk.Object
		result["zendesk"] = []interface{}{m}
	}

	return []interface{}{result}
}

func flattenTriggerConfig(tc *appflow.TriggerConfig) []interface{} {
	m := make(map[string]interface{})

	if tc.TriggerProperties != nil {
		tp := flattenTriggerProperties(tc.TriggerProperties)

		if len(tp) > 0 {
			m["trigger_properties"] = tp
		}
	}
	m["trigger_type"] = tc.TriggerType

	return []interface{}{m}
}

func flattenTriggerProperties(tp *appflow.TriggerProperties) []interface{} {
	m := make(map[string]interface{})

	if tp.Scheduled != nil {
		m["scheduled"] = flattenScheduledTriggerProperties(tp.Scheduled)
	}

	if len(m) > 0 {
		return []interface{}{m}
	}

	return []interface{}{}
}

func flattenScheduledTriggerProperties(stp *appflow.ScheduledTriggerProperties) []interface{} {
	m := make(map[string]interface{})

	if stp.DataPullMode != nil {
		m["data_pull_mode"] = stp.DataPullMode
	}
	if stp.FirstExecutionFrom != nil {
		m["first_execution_from"] = aws.TimeValue(stp.FirstExecutionFrom).Format(time.RFC3339)
	}
	if stp.ScheduleEndTime != nil {
		m["schedule_end_time"] = aws.TimeValue(stp.ScheduleEndTime).Format(time.RFC3339)
	}
	m["schedule_expression"] = stp.ScheduleExpression
	if stp.ScheduleOffset != nil {
		m["schedule_offset"] = stp.ScheduleOffset
	}
	if stp.ScheduleStartTime != nil {
		m["schedule_start_time"] = aws.TimeValue(stp.ScheduleStartTime).Format(time.RFC3339)
	}
	if stp.Timezone != nil {
		m["timezone"] = stp.Timezone
	}

	return []interface{}{m}
}

func flattenTasks(tasks []*appflow.Task) []interface{} {
	result := []interface{}{}
	for _, task := range tasks {
		t := make(map[string]interface{})

		if task.ConnectorOperator != nil {
			t["connector_operator"] = flattenConnectorOperator(task.ConnectorOperator)
		}
		if task.DestinationField != nil {
			t["destination_field"] = task.DestinationField
		}
		t["source_fields"] = aws.StringValueSlice(task.SourceFields)
		t["task_properties"] = aws.StringValueMap(task.TaskProperties)
		t["task_type"] = task.TaskType

		result = append(result, t)
	}
	return result
}

func flattenConnectorOperator(co *appflow.ConnectorOperator) []interface{} {
	m := make(map[string]string)

	if co.Amplitude != nil {
		m["amplitude"] = aws.StringValue(co.Amplitude)
	}
	if co.Datadog != nil {
		m["datadog"] = aws.StringValue(co.Datadog)
	}
	if co.Dynatrace != nil {
		m["dynatrace"] = aws.StringValue(co.Dynatrace)
	}
	if co.GoogleAnalytics != nil {
		m["google_analytics"] = aws.StringValue(co.GoogleAnalytics)
	}
	if co.InforNexus != nil {
		m["infor_nexus"] = aws.StringValue(co.InforNexus)
	}
	if co.Marketo != nil {
		m["marketo"] = aws.StringValue(co.Marketo)
	}
	if co.S3 != nil {
		m["s3"] = aws.StringValue(co.S3)
	}
	if co.Salesforce != nil {
		m["salesforce"] = aws.StringValue(co.Salesforce)
	}
	if co.ServiceNow != nil {
		m["service_now"] = aws.StringValue(co.ServiceNow)
	}
	if co.Singular != nil {
		m["singular"] = aws.StringValue(co.Singular)
	}
	if co.Slack != nil {
		m["slack"] = aws.StringValue(co.Slack)
	}
	if co.Trendmicro != nil {
		m["trendmicro"] = aws.StringValue(co.Trendmicro)
	}
	if co.Veeva != nil {
		m["veeva"] = aws.StringValue(co.Veeva)
	}
	if co.Zendesk != nil {
		m["zendesk"] = aws.StringValue(co.Zendesk)
	}

	return []interface{}{m}
}

func resourceAwsAppFlowFlowUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appflowconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.AppflowUpdateTags(conn, d.Get("flow_arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating AppFlow flow (%s) tags: %s", d.Get("flow_arn").(string), err)
		}
	}

	if d.HasChanges("description", "destination_flow_config_list", "flow_name", "source_flow_config", "tasks", "trigger_config") {
		updateFlowInput := &appflow.UpdateFlowInput{
			Description:               aws.String(d.Get("description").(string)),
			DestinationFlowConfigList: expandDestinationFlowConfigList(d.Get("destination_flow_config_list").([]interface{})),
			FlowName:                  aws.String(d.Get("flow_name").(string)),
			SourceFlowConfig:          expandSourceFlowConfig(d.Get("source_flow_config").([]interface{})[0].(map[string]interface{})),
			Tasks:                     expandTasks(d.Get("tasks").([]interface{})),
			TriggerConfig:             expandTriggerConfig(d.Get("trigger_config").([]interface{})[0].(map[string]interface{})),
		}

		_, err := conn.UpdateFlow(updateFlowInput)

		if err != nil {
			return fmt.Errorf("error updating AppFlow flow (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsAppFlowFlowRead(d, meta)
}

func resourceAwsAppFlowFlowDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appflowconn

	_, err := conn.DeleteFlow(&appflow.DeleteFlowInput{
		FlowName: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error deleting AppFlow flow (%s): %w", d.Id(), err)
	}

	return nil
}
