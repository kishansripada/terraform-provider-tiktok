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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func requiredFixture(p property) any {
	if len(p.Enum) > 0 {
		return p.Enum[0]
	}
	switch p.Type {
	case "string":
		return "100"
	case "number", "integer":
		return float64(1)
	case "boolean":
		return false
	case "array":
		return []any{requiredFixture(*p.Items)}
	case "object":
		v := map[string]any{}
		for _, k := range p.Required {
			v[k] = requiredFixture(p.Properties[k])
		}
		return v
	default:
		panic(p.Type)
	}
}
func TestReadOnlyQueryContracts(t *testing.T) {
	for name, op := range querySpecs {
		t.Run(name, func(t *testing.T) {
			if op.Method != "GET" || !strings.HasPrefix(op.Path, "/open_api/v1.3/") || strings.Contains(op.Path, "?") {
				t.Fatal("unsafe route")
			}
			params := requiredFixture(property{Type: "object", Properties: op.Schema.Properties, Required: op.Schema.Required}).(map[string]any)
			b, _ := json.Marshal(params)
			selected, params, err := queryParameters(name, string(b), "100")
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != "GET" || r.URL.Path != op.Path || r.Header.Get("Access-Token") != "fixture" {
					t.Error("incorrect request")
				}
				for k, v := range params {
					want, ok := v.(string)
					if !ok {
						b, _ := json.Marshal(v)
						want = string(b)
					}
					if r.URL.Query().Get(k) != want {
						t.Errorf("parameter %s lost", k)
					}
				}
				fmt.Fprint(w, `{"code":0,"data":{"page_info":{"total_page":7},"list":[]}}`)
			}))
			defer s.Close()
			c := &apiClient{account: "100", token: "fixture", base: s.URL, http: s.Client()}
			result, err := c.request(context.Background(), selected.Method, selected.Path, params)
			if err != nil || calls != 1 {
				t.Fatalf("read-only request: %v", err)
			}
			if !reflect.DeepEqual(result["page_info"], map[string]any{"total_page": float64(7)}) {
				t.Fatal("pagination metadata lost")
			}
			params["unknown_parameter"] = true
			b, _ = json.Marshal(params)
			if _, _, err := queryParameters(name, string(b), "100"); err == nil {
				t.Fatal("unknown parameter accepted")
			}
		})
	}
}
func TestQueryRejectsMutationsAndCrossAccount(t *testing.T) {
	for _, name := range []string{"campaign_create", "pixel_create", "../campaign/create/", "unknown"} {
		if _, _, err := queryParameters(name, "{}", "100"); err == nil {
			t.Fatal(name)
		}
	}
	for _, input := range []string{"null", "[]", "bad", `{"advertiser_ids":["999"]}`, `{"advertiser_ids":["100","999"]}`, `{"unknown":true}`} {
		if _, _, err := queryParameters("advertiser_info_get", input, "100"); err == nil {
			t.Fatal(input)
		}
	}
}
func TestAccQuery(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/open_api/v1.3/advertiser/info/" || r.URL.Query().Get("advertiser_ids") != `["100"]` {
			t.Error("incorrect query")
		}
		fmt.Fprint(w, `{"code":0,"data":{"list":[{"advertiser_id":"100","name":"fixture"}]}}`)
	}))
	defer s.Close()
	client := &apiClient{account: "100", token: "fixture", base: s.URL, http: s.Client()}
	resource.Test(t, resource.TestCase{ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"tiktok": providerserver.NewProtocol6WithError(&tiktokProvider{testClient: client})}, Steps: []resource.TestStep{{Config: `
provider "tiktok" { advertiser_id = "100" }
data "tiktok_query" "account" { operation = "advertiser_info_get" }
`, Check: resource.TestCheckResourceAttr("data.tiktok_query.account", "result_json", `{"list":[{"advertiser_id":"100","name":"fixture"}]}`)}}})
}

func TestQuerySchemaAndRouteProvenance(t *testing.T) {
	var catalog struct {
		Operations []struct {
			Name   string    `json:"name"`
			Schema operation `json:"inputSchema"`
		} `json:"operations"`
	}
	data, err := os.ReadFile("spec/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &catalog); err != nil {
		t.Fatal(err)
	}
	inputs := map[string]operation{}
	for _, op := range catalog.Operations {
		inputs[op.Name] = op.Schema
	}
	var source struct {
		Routes []struct{ Method, Path string }
	}
	data, err = os.ReadFile("spec/routes.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &source); err != nil {
		t.Fatal(err)
	}
	routes := map[string]bool{}
	for _, route := range source.Routes {
		if route.Method == "GET" {
			routes[route.Path] = true
		}
	}
	for name, query := range querySpecs {
		if !routes[query.Path] || !reflect.DeepEqual(query.Schema, inputs[name]) {
			t.Errorf("unverified query %s", name)
		}
	}
}
