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
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPinnedSchemas(t *testing.T) {
	b, err := os.ReadFile("spec/operations.json")
	if err != nil {
		t.Fatal(err)
	}
	var original struct {
		Operations []struct {
			Name   string          `json:"name"`
			Schema json.RawMessage `json:"inputSchema"`
		} `json:"operations"`
	}
	if err = json.Unmarshal(b, &original); err != nil {
		t.Fatal(err)
	}
	for _, op := range original.Operations {
		if embedded, ok := specs[op.Name]; ok {
			var original operation
			if err := json.Unmarshal(op.Schema, &original); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(embedded, original) {
				t.Fatalf("schema drift: %s", op.Name)
			}
		}
	}
}

func TestSafetyRules(t *testing.T) {
	for _, kind := range []string{"campaign", "smart_plus_campaign"} {
		base := map[string]any{"campaign_name": "test", "objective_type": "APP_PROMOTION", "app_promotion_type": "APP_INSTALL", "budget_mode": "BUDGET_MODE_DAY", "budget": float64(20), "operation_status": "DISABLE"}
		if err := validateCreate(kind, base); err != nil {
			t.Fatal(err)
		}
		for _, change := range []map[string]any{{"operation_status": "ENABLE"}, {"budget": float64(0)}, {"objective_type": "RF_REACH"}, {"unknown": "x"}, {"budget": "20"}, {"operation_status": "DELETE"}, {"postback_window_mode": "POSTBACK_WINDOW_MODE1"}} {
			copy := map[string]any{}
			for k, v := range base {
				copy[k] = v
			}
			for k, v := range change {
				copy[k] = v
			}
			if err := validateCreate(kind, copy); err == nil {
				t.Fatalf("accepted invalid create %v", change)
			}
		}
		for _, change := range []map[string]any{{"objective_type": "TRAFFIC"}, {"budget": float64(0)}, {"special_industries": []any{"HOUSING"}}, {"budget_auto_adjust_strategy": "AUTO_BUDGET_INCREASE"}} {
			if err := validateUpdate(kind, base, change); err == nil {
				t.Fatalf("accepted unsafe update %v", change)
			}
		}
		if len(delta(kind, map[string]any{"budget": "20"}, map[string]any{"budget": float64(20)})) != 0 {
			t.Fatal("numeric normalization")
		}
		if len(delta(kind, base, map[string]any{})) != 0 {
			t.Fatal("omission must relinquish ownership")
		}
	}
}

func TestAPIRejectsErrorsAndIncompleteReads(t *testing.T) {
	for name, body := range map[string]string{
		"business error":     `{"code":40001,"data":{}}`,
		"missing code":       `{"data":{}}`,
		"missing data":       `{"code":0}`,
		"invalid json":       `not json`,
		"missing pagination": `{"code":0,"data":{"list":[]}}`,
		"wrong account":      `{"code":0,"data":{"list":[{"campaign_id":"2","advertiser_id":"9"}],"page_info":{"total_page":1}}}`,
		"duplicate":          `{"code":0,"data":{"list":[{"campaign_id":"2"},{"campaign_id":"2"}],"page_info":{"total_page":1}}}`,
		"wrong count":        `{"code":0,"data":{"list":[],"page_info":{"total_page":1,"total_number":1}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, body) }))
			defer s.Close()
			c := &apiClient{account: "1", token: "fixture", base: s.URL, http: s.Client()}
			if _, err := c.list(context.Background(), "campaign", ""); err == nil {
				t.Fatal("invalid response accepted")
			}
		})
	}
}

type fakeAccount struct {
	mu     sync.Mutex
	row    map[string]any
	writes []string
	kind   string
}

func (f *fakeAccount) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.Header.Get("Access-Token") != "fixture" {
		panic("missing token")
	}
	prefix := "/open_api/v1.3/" + strings.ReplaceAll(strings.Replace(f.kind, "smart_plus_", "smart+/", 1), "_", "/") + "/"
	prefix = strings.Replace(prefix, "smart+/", "smart_plus/", 1)
	data := map[string]any{}
	if r.Method == "GET" {
		rows := []any{}
		if r.URL.Path == prefix+"get/" && f.row != nil {
			rows = append(rows, f.row)
		}
		data = map[string]any{"list": rows, "page_info": map[string]any{"page": 1, "total_page": 1, "total_number": len(rows)}}
	} else {
		var p map[string]any
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			panic(err)
		}
		f.writes = append(f.writes, r.URL.Path)
		switch r.URL.Path {
		case prefix + "create/":
			if f.row != nil {
				panic("duplicate create")
			}
			f.row = map[string]any{"campaign_id": "200", "advertiser_id": "100"}
			for k, v := range p {
				if k != "request_id" {
					f.row[k] = v
				}
			}
			data["campaign_id"] = "200"
		case prefix + "update/":
			if _, ok := p["operation_status"]; ok {
				panic("status leaked into settings update")
			}
			for k, v := range p {
				f.row[k] = v
			}
		case prefix + "status/update/":
			f.row["operation_status"] = p["operation_status"]
		default:
			panic("unexpected write " + r.URL.Path)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": data}); err != nil {
		panic(err)
	}
}

// Runs the real Terraform executable against an in-process provider and HTTP
// fixture. No advertiser credentials or live writes are used.
func TestAccCampaignLifecycle(t *testing.T) {
	for _, kind := range []string{"campaign", "smart_plus_campaign"} {
		t.Run(kind, func(t *testing.T) {
			account := &fakeAccount{kind: kind}
			s := httptest.NewServer(http.HandlerFunc(account.serve))
			defer s.Close()
			client := &apiClient{account: "100", token: "fixture", base: s.URL, http: s.Client(), writes: true, allowDelete: true}
			config := func(budget int, status string) string {
				return fmt.Sprintf(`
provider "tiktok" { advertiser_id = "100" }
resource "tiktok_%s" "test" {
 campaign_name = "fixture"
 objective_type = "APP_PROMOTION"
 app_promotion_type = "APP_INSTALL"
 budget_mode = "BUDGET_MODE_DAY"
 budget = %d
 operation_status = %q
}

`, kind, budget, status)
			}
			address := "tiktok_" + kind + ".test"
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"tiktok": providerserver.NewProtocol6WithError(&tiktokProvider{testClient: client})},
				Steps: []resource.TestStep{
					{Config: config(20, "DISABLE"), Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(address, "id", "200"), resource.TestCheckResourceAttr(address, "budget", "20"))},
					{Config: config(25, "ENABLE"), Check: resource.TestCheckResourceAttr(address, "operation_status", "ENABLE")},
					{ResourceName: address, ImportState: true, ImportStateId: "100/200", ImportStateVerify: true},
					{Config: config(25, "ENABLE"), PreConfig: func() { account.mu.Lock(); defer account.mu.Unlock(); account.row["budget"] = float64(30) }, Check: resource.TestCheckResourceAttr(address, "budget", "25")},
				},
			})
			account.mu.Lock()
			defer account.mu.Unlock()
			if account.row["operation_status"] != "DELETE" {
				t.Fatal("cleanup did not delete")
			}
			if len(account.writes) != 5 {
				t.Fatalf("unexpected requests: %v", account.writes)
			}
		})
	}
}

func TestAccImportBlockAdoptsEnabledCampaign(t *testing.T) {
	f := &fakeAccount{kind: "smart_plus_campaign", row: map[string]any{
		"advertiser_id": "100", "campaign_id": "200", "campaign_name": "existing campaign",
		"objective_type": "APP_PROMOTION", "app_promotion_type": "APP_INSTALL",
		"budget_mode": "BUDGET_MODE_DYNAMIC_DAILY_BUDGET", "budget": float64(25), "operation_status": "ENABLE",
	}}
	s := httptest.NewServer(http.HandlerFunc(f.serve))
	defer s.Close()
	client := &apiClient{account: "100", token: "fixture", base: s.URL, http: s.Client(), writes: true, allowDelete: true}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"tiktok": providerserver.NewProtocol6WithError(&tiktokProvider{testClient: client})},
		Steps: []resource.TestStep{{Config: `
provider "tiktok" { advertiser_id = "100" }
resource "tiktok_smart_plus_campaign" "existing" {
 campaign_name = "existing campaign"
 objective_type = "APP_PROMOTION"
 app_promotion_type = "APP_INSTALL"
 budget_mode = "BUDGET_MODE_DYNAMIC_DAILY_BUDGET"
 budget = 25
 operation_status = "ENABLE"
}
import {
 to = tiktok_smart_plus_campaign.existing
 id = "100/200"
}
`, Check: resource.TestCheckResourceAttr("tiktok_smart_plus_campaign.existing", "id", "200")}},
	})
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.writes) != 1 || !strings.HasSuffix(f.writes[0], "status/update/") || f.row["operation_status"] != "DELETE" {
		t.Fatalf("import wrote to the account; only fixture cleanup should write: %v", f.writes)
	}
}
