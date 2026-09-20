package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	framework "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

var objectKinds = []string{"adgroup", "ad", "smart_plus_adgroup", "smart_plus_ad"}

func objectFixture(kind string) map[string]any {
	v := map[string]any{"operation_status": "DISABLE"}
	if strings.HasSuffix(kind, "adgroup") {
		v["campaign_id"] = "300"
		v["adgroup_name"] = "fixture"
		v["schedule_type"] = "SCHEDULE_FROM_NOW"
		v["schedule_start_time"] = "2026-10-01 00:00:00"
		v["optimization_goal"] = "CLICK"
		v["billing_event"] = "CPC"
		if kind == "adgroup" {
			v["budget_mode"] = "BUDGET_MODE_DAY"
			v["budget"] = float64(20)
			v["pacing"] = "PACING_MODE_SMOOTH"
		} else {
			v["bid_type"] = "BID_TYPE_NO_BID"
			v["promotion_type"] = "WEBSITE"
			v["targeting_spec"] = map[string]any{"location_ids": []any{"6252001"}}
		}
	} else {
		v["adgroup_id"] = "300"
		v["ad_name"] = "fixture"
		if kind == "ad" {
			v["ad_format"] = "SINGLE_VIDEO"
			v["identity_type"] = "TT_USER"
			v["identity_id"] = "400"
			v["video_id"] = "video-fixture"
		} else {
			v["creative_list"] = []any{map[string]any{"creative_info": map[string]any{"ad_format": "SINGLE_VIDEO", "identity_type": "TT_USER", "identity_id": "400", "video_info": map[string]any{"video_id": "video-fixture"}}}}
		}
	}
	return v
}
func nameField(kind string) string {
	if strings.HasSuffix(kind, "adgroup") {
		return "adgroup_name"
	}
	return "ad_name"
}

// Fixture validates request envelopes against the pinned input contract, and
// simulates replacement semantics so omitted fields cannot silently pass tests.
type objectServer struct {
	mu                              sync.Mutex
	kind                            string
	row                             map[string]any
	writes                          []map[string]any
	paths                           []string
	failStatus, ambiguous, children bool
}

func (f *objectServer) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	prefix := "/open_api/v1.3/" + strings.ReplaceAll(strings.Replace(f.kind, "smart_plus_", "smart+/", 1), "_", "/") + "/"
	prefix = strings.Replace(prefix, "smart+/", "smart_plus/", 1)
	data := map[string]any{}
	if r.Method == "GET" {
		rows := []any{}
		if r.URL.Path == prefix+"get/" && f.row != nil {
			rows = append(rows, f.row)
		} else if f.children && strings.HasSuffix(r.URL.Path, "/ad/get/") {
			rows = append(rows, map[string]any{"ad_id": "500", "smart_plus_ad_id": "500", "adgroup_id": "200"})
		}
		data = map[string]any{"list": rows, "page_info": map[string]any{"page": 1, "total_page": 1, "total_number": len(rows)}}
	} else {
		var params map[string]any
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			panic(err)
		}
		suffix := strings.TrimPrefix(r.URL.Path, prefix)
		op := f.kind + "_" + strings.ReplaceAll(strings.TrimSuffix(suffix, "/"), "/", "_")
		spec, ok := objectSpecs[op]
		if !ok {
			panic("unexpected operation " + op)
		}
		if err := validateValue(property{Type: "object", Properties: spec.Properties, Required: spec.Required}, params); err != nil {
			panic(fmt.Sprintf("%s: %v", op, err))
		}
		f.writes = append(f.writes, params)
		f.paths = append(f.paths, suffix)
		switch suffix {
		case "create/":
			if f.row != nil {
				panic("duplicate create")
			}
			f.row = map[string]any{"advertiser_id": "100", objectID(f.kind): "200"}
			fields := params
			if f.kind == "ad" {
				fields = params["creatives"].([]any)[0].(map[string]any)
				f.row["adgroup_id"] = params["adgroup_id"]
			}
			for k, v := range fields {
				if k != "request_id" {
					f.row[k] = v
				}
			}
			if !f.ambiguous {
				if f.kind == "ad" {
					data["ad_ids"] = []any{"200"}
				} else {
					data[objectID(f.kind)] = "200"
				}
			}
		case "update/":
			fields := params
			if f.kind == "ad" {
				fields = params["creatives"].([]any)[0].(map[string]any)
			}
			if f.kind == "ad" || f.kind == "adgroup" {
				for k := range objectOperation(f.kind, "_update").Properties {
					if k != "adgroup_id" && k != "advertiser_id" && k != objectID(f.kind) {
						delete(f.row, k)
					}
				}
			}
			for k, v := range fields {
				f.row[k] = v
			}
		case "status/update/":
			if f.failStatus {
				fmt.Fprint(w, `{"code":40001,"data":{}}`)
				return
			}
			f.row["operation_status"] = params["operation_status"]
		default:
			panic(suffix)
		}
	}
	if err := json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": data}); err != nil {
		panic(err)
	}
}
func setupObject(t *testing.T, kind string) (*objectResource, *objectServer) {
	t.Helper()
	f := &objectServer{kind: kind}
	s := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(s.Close)
	return &objectResource{kind: kind, client: &apiClient{account: "100", token: "fixture", base: s.URL, http: s.Client(), writes: true, allowDelete: true}}, f
}
func objectState(t *testing.T, r *objectResource, desired map[string]any) tfsdk.State {
	t.Helper()
	state := tfsdk.State{Schema: objectSchema(r.kind)}
	if err := r.save(&state, "200", desired, nil); err != nil {
		t.Fatal(err)
	}
	return state
}
func TestObjectContract(t *testing.T) {
	b, err := os.ReadFile("spec/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	var catalog struct {
		Operations []struct {
			Name   string    `json:"name"`
			Schema operation `json:"inputSchema"`
		} `json:"operations"`
	}
	if err = json.Unmarshal(b, &catalog); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, op := range catalog.Operations {
		if pinned, ok := objectSpecs[op.Name]; ok {
			seen[op.Name] = true
			if !reflect.DeepEqual(pinned, op.Schema) {
				t.Fatalf("schema drift %s", op.Name)
			}
		}
	}
	if len(seen) != len(objectSpecs) {
		t.Fatalf("pinned %d of %d", len(seen), len(objectSpecs))
	}
	for _, kind := range objectKinds {
		t.Run(kind, func(t *testing.T) {
			valid := objectFixture(kind)
			if err := validateObject(kind, valid, true); err != nil {
				t.Fatal(err)
			}
			for _, bad := range []map[string]any{{"operation_status": "ENABLE"}, {"operation_status": "DELETE"}, {"bogus": true}, {"operation_status": false}} {
				v := objectFixture(kind)
				for k, x := range bad {
					v[k] = x
				}
				if validateObject(kind, v, true) == nil {
					t.Fatalf("accepted %v", bad)
				}
			}
			for _, required := range objectOperation(kind, "_create").Required {
				if required == "advertiser_id" || required == "request_id" {
					continue
				}
				v := objectFixture(kind)
				delete(v, required)
				if validateObject(kind, v, true) == nil {
					t.Fatalf("missing %s accepted", required)
				}
			}
		})
	}
}
func TestAccObjectLifecycle(t *testing.T) {
	for _, kind := range objectKinds {
		t.Run(kind, func(t *testing.T) {
			r, f := setupObject(t, kind)
			config := func(name, status string) string {
				v := objectFixture(kind)
				v[nameField(kind)] = name
				v["operation_status"] = status
				config := fmt.Sprintf("provider \"tiktok\" { advertiser_id = \"100\" }\nresource \"tiktok_%s\" \"test\" {\n", kind)
				for k, value := range v {
					b, _ := json.Marshal(value)
					config += fmt.Sprintf("%s = %s\n", k, b)
				}
				return config + "}\n"
			}
			address := "tiktok_" + kind + ".test"
			resource.Test(t, resource.TestCase{CheckDestroy: func(_ *terraform.State) error {
				f.mu.Lock()
				defer f.mu.Unlock()
				if f.row != nil && !deleted(f.row) {
					return fmt.Errorf("remote fixture survived destroy")
				}
				return nil
			}, ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"tiktok": providerserver.NewProtocol6WithError(&tiktokProvider{testClient: r.client})}, Steps: []resource.TestStep{
				{Config: config("fixture", "DISABLE"), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionCreate)}}, Check: resource.TestCheckResourceAttr(address, "id", "200")},
				{Config: config("changed", "ENABLE"), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate)}}, Check: resource.TestCheckResourceAttr(address, nameField(kind), "changed")},
				{ResourceName: address, ImportState: true, ImportStateId: "100/200", ImportStateVerify: true},
				{Config: config("changed", "ENABLE"), PreConfig: func() { f.mu.Lock(); defer f.mu.Unlock(); f.row[nameField(kind)] = "remote drift" }, Check: resource.TestCheckResourceAttr(address, nameField(kind), "changed")},
				{Config: config("paused", "DISABLE")},
			}})
			f.mu.Lock()
			defer f.mu.Unlock()
			want := []string{"create/", "update/", "status/update/", "update/", "status/update/", "update/", "status/update/"}
			if !reflect.DeepEqual(f.paths, want) {
				t.Fatalf("wrong lifecycle order %v", f.paths)
			}
		})
	}
}
func TestObjectRecovery(t *testing.T) {
	for _, kind := range objectKinds {
		for _, scenario := range []string{"stale", "partial", "ambiguous", "protected", "children", "missing", "wrong account", "omit", "readonly"} {
			t.Run(kind+"/"+scenario, func(t *testing.T) {
				r, f := setupObject(t, kind)
				prior := objectFixture(kind)
				state := objectState(t, r, prior)
				f.row = map[string]any{"advertiser_id": "100", objectID(kind): "200"}
				for k, v := range prior {
					f.row[k] = v
				}
				desired := objectFixture(kind)
				desired[nameField(kind)] = "new"
				desired["operation_status"] = "ENABLE"
				ctx := context.Background()
				switch scenario {
				case "ambiguous":
					f.row = nil
					f.ambiguous = true
					res := framework.CreateResponse{State: state}
					r.Create(ctx, framework.CreateRequest{Plan: tfsdk.Plan{Schema: state.Schema, Raw: state.Raw}}, &res)
					values, err := rawMap(res.State.Raw)
					if err != nil {
						t.Fatal(err)
					}
					if !res.Diagnostics.HasError() || !strings.HasPrefix(fmt.Sprint(values["id"]), "pending-") || len(f.writes) != 1 {
						t.Fatal("ambiguous create not blocked")
					}
					if _, _, err := r.identity(res.State.Raw); err == nil {
						t.Fatal("pending identity accepted")
					}
				case "protected", "children":
					if scenario == "protected" {
						r.client.allowDelete = false
					} else {
						if !strings.HasSuffix(kind, "adgroup") {
							return
						}
						f.children = true
					}
					res := framework.DeleteResponse{State: state}
					r.Delete(ctx, framework.DeleteRequest{State: state}, &res)
					if !res.Diagnostics.HasError() || len(f.writes) != 0 {
						t.Fatal("unsafe deletion")
					}
				case "missing", "wrong account":
					if scenario == "missing" {
						f.row = nil
					} else {
						r.client.account = "999"
					}
					res := framework.ReadResponse{State: state}
					r.Read(ctx, framework.ReadRequest{State: state}, &res)
					if !res.Diagnostics.HasError() || len(f.writes) != 0 {
						t.Fatal("unsafe refresh")
					}
				default:
					if scenario == "stale" {
						f.row[nameField(kind)] = "external"
					}
					if scenario == "partial" {
						f.failStatus = true
					}
					if scenario == "readonly" {
						r.client.writes = false
					}
					if scenario == "omit" {
						desired = objectFixture(kind)
						delete(desired, nameField(kind))
					}
					plan := objectState(t, r, desired)
					res := framework.UpdateResponse{State: state}
					r.Update(ctx, framework.UpdateRequest{State: state, Plan: tfsdk.Plan{Schema: plan.Schema, Raw: plan.Raw}}, &res)
					if scenario == "omit" {
						if res.Diagnostics.HasError() || len(f.writes) != 0 {
							t.Fatalf("omission wrote: %v", res.Diagnostics)
						}
						values, _ := rawMap(res.State.Raw)
						if values[nameField(kind)] != nil {
							t.Fatal("retained ownership")
						}
						return
					}
					if !res.Diagnostics.HasError() {
						t.Fatal("expected failure")
					}
					if scenario == "partial" {
						values, _ := rawMap(res.State.Raw)
						if values[nameField(kind)] != "new" || values["operation_status"] != "DISABLE" {
							t.Fatalf("lost partial state: %v", values)
						}
					} else if len(f.writes) != 0 {
						t.Fatal("unexpected write")
					}
				}
			})
		}
	}
}
func TestReplacementPreservesUnmanagedFields(t *testing.T) {
	for _, kind := range []string{"adgroup", "ad"} {
		t.Run(kind, func(t *testing.T) {
			r, _ := setupObject(t, kind)
			live := objectFixture(kind)
			live["share_disabled"] = true
			params, err := r.updateParams("200", live, map[string]any{nameField(kind): "new"})
			if err != nil {
				t.Fatal(err)
			}
			fields := params
			if kind == "ad" {
				fields = params["creatives"].([]any)[0].(map[string]any)
				if fields["adgroup_id"] != nil {
					t.Fatal("parent leaked into creative")
				}
			}
			if _, exists := objectOperation(kind, "_update").Properties["share_disabled"]; exists && fields["share_disabled"] != true {
				t.Fatal("unmanaged setting erased")
			}
		})
	}
}

func TestSmartPlusNestedReplacement(t *testing.T) {
	r, _ := setupObject(t, "smart_plus_ad")
	live := map[string]any{"ad_configuration": map[string]any{"product_info": map[string]any{"product_titles": []any{"old"}, "selling_points": []any{"preserve"}, "server_metadata": "discard"}}}
	params, err := r.updateParams("200", live, map[string]any{"ad_configuration": map[string]any{"product_info": map[string]any{"product_titles": []any{"new"}}}})
	if err != nil {
		t.Fatal(err)
	}
	product := params["ad_configuration"].(map[string]any)["product_info"].(map[string]any)
	if !reflect.DeepEqual(product["product_titles"], []any{"new"}) || !reflect.DeepEqual(product["selling_points"], []any{"preserve"}) || product["server_metadata"] != nil {
		t.Fatalf("bad nested replacement %v", product)
	}
	a := map[string]any{"creative_info": map[string]any{"ad_format": "SINGLE_VIDEO", "video_info": map[string]any{"video_id": "a"}}}
	b := map[string]any{"creative_info": map[string]any{"ad_format": "SINGLE_VIDEO", "video_info": map[string]any{"video_id": "b"}}}
	remote := []any{mergeObject(a, map[string]any{"ad_material_id": "800"}), mergeObject(b, map[string]any{"ad_material_id": "900"})}
	result := reuseMaterials(remote, []any{b, a}).([]any)
	if result[0].(map[string]any)["ad_material_id"] != "900" || result[1].(map[string]any)["ad_material_id"] != "800" {
		t.Fatal("material identity lost on reorder")
	}
	if a["ad_material_id"] != nil {
		t.Fatal("mutated desired configuration")
	}
}

func TestAccDependentAdResources(t *testing.T) {
	group := &objectServer{kind: "adgroup"}
	ad := &objectServer{kind: "ad"}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/open_api/v1.3/ad/") {
			ad.serve(w, r)
		} else {
			group.serve(w, r)
		}
	}))
	defer s.Close()
	client := &apiClient{account: "100", token: "fixture", base: s.URL, http: s.Client(), writes: true, allowDelete: true}
	config := `provider "tiktok" { advertiser_id = "100" }` + "\n"
	for _, kind := range []string{"adgroup", "ad"} {
		config += fmt.Sprintf("resource %q %q {\n", "tiktok_"+kind, "test")
		for k, v := range objectFixture(kind) {
			if kind == "ad" && k == "adgroup_id" {
				config += "adgroup_id = tiktok_adgroup.test.id\n"
			} else {
				b, _ := json.Marshal(v)
				config += fmt.Sprintf("%s = %s\n", k, b)
			}
		}
		config += "}\n"
	}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"tiktok": providerserver.NewProtocol6WithError(&tiktokProvider{testClient: client})}, CheckDestroy: func(_ *terraform.State) error {
		group.mu.Lock()
		defer group.mu.Unlock()
		ad.mu.Lock()
		defer ad.mu.Unlock()
		if !deleted(group.row) || !deleted(ad.row) {
			return fmt.Errorf("dependent resources survived destroy")
		}
		return nil
	}, Steps: []resource.TestStep{{Config: config, Check: resource.TestCheckResourceAttr("tiktok_ad.test", "adgroup_id", "200")}}})
}

func TestObjectMissingAndDeletedFailClosed(t *testing.T) {
	for _, kind := range objectKinds {
		t.Run(kind, func(t *testing.T) {
			r, f := setupObject(t, kind)
			state := objectState(t, r, objectFixture(kind))
			f.row = map[string]any{objectID(kind): "200", "operation_status": "DELETE"}
			res := framework.ReadResponse{State: state}
			r.Read(context.Background(), framework.ReadRequest{State: state}, &res)
			if !res.Diagnostics.HasError() || res.State.Raw.IsNull() {
				t.Fatal("deleted resource silently scheduled for recreation")
			}
			imported := framework.ImportStateResponse{State: state}
			r.ImportState(context.Background(), framework.ImportStateRequest{ID: "999/200"}, &imported)
			if !imported.Diagnostics.HasError() {
				t.Fatal("cross-account import accepted")
			}
		})
	}
}
