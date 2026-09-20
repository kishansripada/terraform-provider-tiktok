package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var decimalID = regexp.MustCompile(`^[0-9]+$`)

func validID(s string) bool { return decimalID.MatchString(s) }

type apiClient struct {
	account, token, base string
	writes, allowDelete  bool
	http                 *http.Client
}

func (c *apiClient) call(ctx context.Context, operation string, params map[string]any) (map[string]any, error) {
	method := "POST"
	if strings.HasSuffix(operation, "_get") {
		method = "GET"
	}
	operation = strings.Replace(operation, "smart_plus_", "smart+/", 1)
	endpoint := strings.ReplaceAll(operation, "_", "/")
	endpoint = strings.Replace(endpoint, "smart+/", "smart_plus/", 1)
	return c.request(ctx, method, "/open_api/v1.3/"+endpoint+"/", params)
}

func (c *apiClient) request(ctx context.Context, method, endpoint string, params map[string]any) (map[string]any, error) {
	if c.token == "" {
		return nil, fmt.Errorf("TIKTOK_MARKETING_ACCESS_TOKEN is required")
	}
	if method != "GET" && !c.writes {
		return nil, fmt.Errorf("writes disabled: set provider allow_writes to true")
	}
	u, err := url.Parse(c.base + endpoint)
	if err != nil {
		return nil, err
	}
	var body io.Reader
	if method == "GET" {
		q := u.Query()
		for key, value := range params {
			if s, ok := value.(string); ok {
				q.Set(key, s)
			} else {
				b, err := json.Marshal(value)
				if err != nil {
					return nil, err
				}
				q.Set(key, string(b))
			}
		}
		u.RawQuery = q.Encode()
	} else {
		b, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Access-Token", c.token)
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	// Never retry writes: campaign-create deduplication has a short API window.
	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("TikTok request failed; inspect remote state before retrying: %w", err)
	}
	defer res.Body.Close()
	var payload struct {
		Code      *int           `json:"code"`
		Data      map[string]any `json:"data"`
		RequestID string         `json:"request_id"`
	}
	if err = json.NewDecoder(io.LimitReader(res.Body, 16<<20)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("invalid TikTok JSON response")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 || payload.Code == nil || *payload.Code != 0 || payload.Data == nil {
		code := "missing"
		if payload.Code != nil {
			code = fmt.Sprint(*payload.Code)
		}
		return nil, fmt.Errorf("TikTok rejected request: HTTP %d, code %s, request %s", res.StatusCode, code, payload.RequestID)
	}
	return payload.Data, nil
}

func (c *apiClient) list(ctx context.Context, kind string, campaignID string) ([]map[string]any, error) {
	filter := map[string]any{"primary_status": "STATUS_ALL"}
	if campaignID != "" {
		filter["campaign_ids"] = []string{campaignID}
	}
	return c.listFiltered(ctx, kind, filter)
}

func (c *apiClient) listFiltered(ctx context.Context, kind string, filter map[string]any) ([]map[string]any, error) {
	var rows []map[string]any
	seen := map[string]bool{}
	for page := 1; page <= 10000; page++ {
		data, err := c.call(ctx, kind+"_get", map[string]any{"advertiser_id": c.account, "filtering": filter, "page": page, "page_size": 100})
		if err != nil {
			return nil, err
		}
		items, ok := data["list"].([]any)
		if !ok {
			return nil, fmt.Errorf("invalid %s list", kind)
		}
		field := strings.TrimPrefix(kind, "smart_plus_") + "_id"
		if kind == "smart_plus_ad" {
			field = "smart_plus_ad_id"
		}
		for _, item := range items {
			row, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid row")
			}
			id, ok := row[field].(string)
			if !ok || !validID(id) || seen[id] {
				return nil, fmt.Errorf("invalid or duplicate %s ID", kind)
			}
			if a, exists := row["advertiser_id"]; exists && a != c.account {
				return nil, fmt.Errorf("advertiser mismatch")
			}
			seen[id] = true
			rows = append(rows, row)
		}
		info, ok := data["page_info"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("missing pagination metadata")
		}
		total, ok := info["total_page"].(float64)
		if !ok || total < 0 || total != float64(int(total)) {
			return nil, fmt.Errorf("invalid total_page")
		}
		if p, exists := info["page"]; exists && p != float64(page) {
			return nil, fmt.Errorf("unexpected page")
		}
		if float64(page) >= total {
			if n, exists := info["total_number"]; exists && n != float64(len(rows)) {
				return nil, fmt.Errorf("incomplete pagination")
			}
			return rows, nil
		}
		if len(items) == 0 {
			return nil, fmt.Errorf("empty intermediate page")
		}
	}
	return nil, fmt.Errorf("pagination limit exceeded")
}

func deleted(row map[string]any) bool {
	s, _ := row["secondary_status"].(string)
	return row["operation_status"] == "DELETE" || strings.HasSuffix(s, "_DELETE")
}
func (c *apiClient) read(ctx context.Context, kind, id string) (map[string]any, error) {
	if !validID(id) {
		return nil, fmt.Errorf("invalid campaign ID; reconcile an ambiguous create by importing its verified remote ID")
	}
	rows, err := c.list(ctx, kind, id)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row["campaign_id"] == id {
			return row, nil
		}
	}
	return nil, fmt.Errorf("managed campaign %s missing; refusing automatic recreation", id)
}

func (c *apiClient) verify(ctx context.Context, kind, id string, expected map[string]any, deleting bool) (map[string]any, error) {
	var last map[string]any
	for attempt := 0; attempt < 4; attempt++ {
		row, err := c.read(ctx, kind, id)
		if err != nil {
			return last, err
		}
		last = row
		if deleting && deleted(row) {
			return row, nil
		}
		match := !deleted(row)
		for k, v := range expected {
			if !reflect.DeepEqual(normalize(kind, k, row[k]), v) {
				match = false
			}
		}
		if !deleting && match {
			return row, nil
		}
		if attempt < 3 {
			select {
			case <-ctx.Done():
				return last, ctx.Err()
			case <-time.After(time.Duration(250*(1<<attempt)) * time.Millisecond):
			}
		}
	}
	return last, fmt.Errorf("remote verification did not converge; inspect campaign %s before applying again", id)
}
