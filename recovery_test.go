package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestResourceFailureRecovery(t *testing.T) {
	for _, scenario := range []string{"stale plan", "partial update", "ambiguous create", "protected delete", "children block delete", "omit budget"} {
		t.Run(scenario, func(t *testing.T) {
			row := map[string]any{"campaign_id": "200", "advertiser_id": "100", "campaign_name": "fixture", "objective_type": "APP_PROMOTION", "app_promotion_type": "APP_INSTALL", "budget_mode": "BUDGET_MODE_DAY", "budget": float64(20), "operation_status": "DISABLE"}
			f := &fakeAccount{kind: "campaign", row: row}
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if scenario == "partial update" && strings.HasSuffix(req.URL.Path, "status/update/") {
					fmt.Fprint(w, `{"code":40001,"data":{}}`)
					return
				}
				if scenario == "ambiguous create" && strings.HasSuffix(req.URL.Path, "create/") {
					recorder := httptest.NewRecorder()
					f.serve(recorder, req)
					fmt.Fprint(w, `{"code":0,"data":{}}`)
					return
				}
				if scenario == "children block delete" && strings.HasSuffix(req.URL.Path, "adgroup/get/") {
					fmt.Fprint(w, `{"code":0,"data":{"list":[{"adgroup_id":"300","campaign_id":"200"}],"page_info":{"total_page":1,"total_number":1}}}`)
					return
				}
				f.serve(w, req)
			}))
			defer s.Close()
			r := &campaignResource{kind: "campaign", client: &apiClient{account: "100", token: "fixture", base: s.URL, http: s.Client(), writes: true, allowDelete: scenario != "protected delete"}}
			state := tfsdk.State{Schema: campaignSchema("campaign")}
			attrs := managed("campaign", row)
			if err := r.save(&state, "200", attrs, nil); err != nil {
				t.Fatal(err)
			}
			planned := tfsdk.State{Schema: state.Schema}
			desired := managed("campaign", row)
			if scenario == "omit budget" {
				delete(desired, "budget")
			} else {
				desired["budget"] = float64(25)
				desired["operation_status"] = "ENABLE"
			}
			if err := r.save(&planned, "200", desired, nil); err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			switch scenario {
			case "stale plan", "partial update", "omit budget":
				if scenario == "stale plan" {
					f.row["budget"] = float64(30)
				}
				res := resource.UpdateResponse{State: state}
				r.Update(ctx, resource.UpdateRequest{State: state, Plan: tfsdk.Plan{Schema: planned.Schema, Raw: planned.Raw}}, &res)
				if scenario == "omit budget" {
					if res.Diagnostics.HasError() || len(f.writes) != 0 || f.row["budget"] != float64(20) {
						t.Fatal("omission changed remote budget")
					}
					observed, err := rawMap(res.State.Raw)
					if err != nil {
						t.Fatal(err)
					}
					if _, ok := observed["budget"]; ok {
						t.Fatal("omission retained ownership")
					}
					return
				}
				if !res.Diagnostics.HasError() {
					t.Fatal("expected failure")
				}
				if scenario == "stale plan" && len(f.writes) != 0 {
					t.Fatal("stale plan mutated account")
				}
				if scenario == "partial update" {
					observed, err := rawMap(res.State.Raw)
					if err != nil {
						t.Fatal(err)
					}
					if observed["budget"] != float64(25) || observed["operation_status"] != "DISABLE" {
						t.Fatalf("lost partial result: %v", observed)
					}
				}
			case "ambiguous create":
				f.row = nil
				res := resource.CreateResponse{State: state}
				r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Schema: state.Schema, Raw: state.Raw}}, &res)
				if !res.Diagnostics.HasError() {
					t.Fatal("ambiguous create succeeded")
				}
				observed, err := rawMap(res.State.Raw)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(observed["id"].(string), "pending-") {
					t.Fatal("missing blocking identity")
				}
				if _, _, err := r.identity(res.State.Raw); err == nil {
					t.Fatal("pending identity accepted")
				}
				if len(f.writes) != 1 {
					t.Fatal("create retried")
				}
			default:
				res := resource.DeleteResponse{State: state}
				r.Delete(ctx, resource.DeleteRequest{State: state}, &res)
				if !res.Diagnostics.HasError() || len(f.writes) != 0 {
					t.Fatal("unsafe delete")
				}
			}
		})
	}
}
