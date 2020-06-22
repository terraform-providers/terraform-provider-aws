package aws

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsAppmeshRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppmeshRouteCreate,
		Read:   resourceAwsAppmeshRouteRead,
		Update: resourceAwsAppmeshRouteUpdate,
		Delete: resourceAwsAppmeshRouteDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsAppmeshRouteImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"mesh_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"virtual_router_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"spec": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_route": {
							Type:          schema.TypeList,
							Optional:      true,
							MinItems:      0,
							MaxItems:      1,
							ConflictsWith: []string{"spec.0.tcp_route"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"weighted_target": {
													Type:     schema.TypeSet,
													Required: true,
													MinItems: 1,
													MaxItems: 10,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"virtual_node": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 255),
															},

															"weight": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntBetween(0, 100),
															},
														},
													},
													Set: appmeshWeightedTargetHash,
												},
											},
										},
									},

									"match": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"header": {
													Type:     schema.TypeSet,
													Optional: true,
													MinItems: 0,
													MaxItems: 10,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"invert": {
																Type:     schema.TypeBool,
																Optional: true,
																Default:  false,
															},

															"match": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"exact": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},

																		"prefix": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},

																		"range": {
																			Type:     schema.TypeList,
																			Optional: true,
																			MinItems: 0,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"end": {
																						Type:     schema.TypeInt,
																						Required: true,
																					},

																					"start": {
																						Type:     schema.TypeInt,
																						Required: true,
																					},
																				},
																			},
																		},

																		"regex": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},

																		"suffix": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},
																	},
																},
															},

															"name": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 50),
															},
														},
													},
													Set: appmeshHttpRouteHeaderHash,
												},

												"method": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.StringInSlice([]string{
														appmesh.HttpMethodConnect,
														appmesh.HttpMethodDelete,
														appmesh.HttpMethodGet,
														appmesh.HttpMethodHead,
														appmesh.HttpMethodOptions,
														appmesh.HttpMethodPatch,
														appmesh.HttpMethodPost,
														appmesh.HttpMethodPut,
														appmesh.HttpMethodTrace,
													}, false),
												},

												"prefix": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/`), "must start with /"),
												},

												"scheme": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.StringInSlice([]string{
														appmesh.HttpSchemeHttp,
														appmesh.HttpSchemeHttps,
													}, false),
												},
											},
										},
									},
								},
							},
						},

						"priority": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 1000),
						},

						"tcp_route": {
							Type:          schema.TypeList,
							Optional:      true,
							MinItems:      0,
							MaxItems:      1,
							ConflictsWith: []string{"spec.0.http_route"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"weighted_target": {
													Type:     schema.TypeSet,
													Required: true,
													MinItems: 1,
													MaxItems: 10,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"virtual_node": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 255),
															},

															"weight": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntBetween(0, 100),
															},
														},
													},
													Set: appmeshWeightedTargetHash,
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

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsAppmeshRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	req := &appmesh.CreateRouteInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		RouteName:         aws.String(d.Get("name").(string)),
		VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
		Spec:              expandAppmeshRouteSpec(d.Get("spec").([]interface{})),
		Tags:              keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AppmeshTags(),
	}

	log.Printf("[DEBUG] Creating App Mesh route: %#v", req)
	resp, err := conn.CreateRoute(req)
	if err != nil {
		return fmt.Errorf("error creating App Mesh route: %s", err)
	}

	d.SetId(aws.StringValue(resp.Route.Metadata.Uid))

	return resourceAwsAppmeshRouteRead(d, meta)
}

func resourceAwsAppmeshRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeRoute(&appmesh.DescribeRouteInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		RouteName:         aws.String(d.Get("name").(string)),
		VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
	})
	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] App Mesh route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading App Mesh route: %s", err)
	}
	if aws.StringValue(resp.Route.Status.Status) == appmesh.RouteStatusCodeDeleted {
		log.Printf("[WARN] App Mesh route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(resp.Route.Metadata.Arn)
	d.Set("name", resp.Route.RouteName)
	d.Set("mesh_name", resp.Route.MeshName)
	d.Set("virtual_router_name", resp.Route.VirtualRouterName)
	d.Set("arn", arn)
	d.Set("created_date", resp.Route.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.Route.Metadata.LastUpdatedAt.Format(time.RFC3339))
	err = d.Set("spec", flattenAppmeshRouteSpec(resp.Route.Spec))
	if err != nil {
		return fmt.Errorf("error setting spec: %s", err)
	}

	tags, err := keyvaluetags.AppmeshListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for App Mesh route (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsAppmeshRouteUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	if d.HasChange("spec") {
		_, v := d.GetChange("spec")
		req := &appmesh.UpdateRouteInput{
			MeshName:          aws.String(d.Get("mesh_name").(string)),
			RouteName:         aws.String(d.Get("name").(string)),
			VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
			Spec:              expandAppmeshRouteSpec(v.([]interface{})),
		}

		log.Printf("[DEBUG] Updating App Mesh route: %#v", req)
		_, err := conn.UpdateRoute(req)
		if err != nil {
			return fmt.Errorf("error updating App Mesh route: %s", err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.AppmeshUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating App Mesh route (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsAppmeshRouteRead(d, meta)
}

func resourceAwsAppmeshRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	log.Printf("[DEBUG] Deleting App Mesh route: %s", d.Id())
	_, err := conn.DeleteRoute(&appmesh.DeleteRouteInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		RouteName:         aws.String(d.Get("name").(string)),
		VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
	})
	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting App Mesh route: %s", err)
	}

	return nil
}

func resourceAwsAppmeshRouteImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'mesh-name/virtual-router-name/route-name'", d.Id())
	}

	mesh := parts[0]
	vrName := parts[1]
	name := parts[2]
	log.Printf("[DEBUG] Importing App Mesh route %s from mesh %s/virtual router %s ", name, mesh, vrName)

	conn := meta.(*AWSClient).appmeshconn

	resp, err := conn.DescribeRoute(&appmesh.DescribeRouteInput{
		MeshName:          aws.String(mesh),
		RouteName:         aws.String(name),
		VirtualRouterName: aws.String(vrName),
	})
	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(resp.Route.Metadata.Uid))
	d.Set("name", resp.Route.RouteName)
	d.Set("mesh_name", resp.Route.MeshName)
	d.Set("virtual_router_name", resp.Route.VirtualRouterName)

	return []*schema.ResourceData{d}, nil
}

func appmeshHttpRouteHeaderHash(vHttpRouteHeader interface{}) int {
	var buf bytes.Buffer
	mHttpRouteHeader := vHttpRouteHeader.(map[string]interface{})
	if v, ok := mHttpRouteHeader["invert"].(bool); ok {
		buf.WriteString(fmt.Sprintf("%t-", v))
	}
	if vMatch, ok := mHttpRouteHeader["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
		mMatch := vMatch[0].(map[string]interface{})
		if v, ok := mMatch["exact"].(string); ok {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
		if v, ok := mMatch["prefix"].(string); ok {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
		if vRange, ok := mMatch["range"].([]interface{}); ok && len(vRange) > 0 && vRange[0] != nil {
			mRange := vRange[0].(map[string]interface{})
			if v, ok := mRange["end"].(int); ok {
				buf.WriteString(fmt.Sprintf("%d-", v))
			}
			if v, ok := mRange["start"].(int); ok {
				buf.WriteString(fmt.Sprintf("%d-", v))
			}
		}
		if v, ok := mMatch["regex"].(string); ok {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
		if v, ok := mMatch["suffix"].(string); ok {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	if v, ok := mHttpRouteHeader["name"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	return hashcode.String(buf.String())
}

func appmeshWeightedTargetHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["virtual_node"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := m["weight"].(int); ok {
		buf.WriteString(fmt.Sprintf("%d-", v))
	}
	return hashcode.String(buf.String())
}
