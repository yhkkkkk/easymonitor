package xelastic

import (
	"bytes"
	"context"
	"easymonitor/infra"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type ElasticClientV7 struct {
	client *elasticsearch7.Client
}

func (ec *ElasticClientV7) FindByDSL(index string, dsl string, source []string) ([]any, int, int) {
	req := esapi.SearchRequest{
		Index:        []string{index},
		DocumentType: []string{"_doc"},
		Body:         strings.NewReader(dsl),
		Pretty:       true,
	}
	if source != nil {
		req.Source = source
	}
	dst := &bytes.Buffer{}
	_ = json.Compact(dst, []byte(dsl))
	var ctx = context.Background()
	res, e := req.Do(ctx, ec.client)
	var hits []any
	totalValue := 0
	if e != nil {
		infra.Logger.Errorln(fmt.Sprintf("%s : %s", index, e.Error()))
		return hits, totalValue, res.StatusCode
	} else {
		m := ec.parseResponseBody(res)
		infra.Logger.Debugln(fmt.Sprintf("%s : %s", index, dst.String()))
		j, ok := m["hits"]
		if ok {
			hitsVal := j.(map[string]any)
			hits = hitsVal["hits"].([]any)
			total := hitsVal["total"].(map[string]any)
			totalFloat := total["value"].(float64)
			totalValue = int(totalFloat)
		}
		return hits, totalValue, res.StatusCode
	}
}

func (ec *ElasticClientV7) CountByDSL(index string, dsl string) (int, int) {
	req := esapi.CountRequest{
		Index:        []string{index},
		DocumentType: []string{"_doc"},
		Body:         strings.NewReader(dsl),
		Pretty:       true,
	}
	var ctx = context.Background()
	res, e := req.Do(ctx, ec.client)
	if e != nil {
		t := fmt.Sprintf("%s : %s", index, e.Error())
		infra.Logger.Errorln(t)
		return 0, res.StatusCode
	} else {
		m := ec.parseResponseBody(res)
		c, ok := m["count"]
		if ok {
			countFloat := c.(float64)
			return int(countFloat), res.StatusCode
		} else {
			return 0, res.StatusCode
		}
	}
}

func (ec *ElasticClientV7) parseResponseBody(resp *esapi.Response) map[string]any {
	s := map[string]any{}
	if !resp.IsError() {
		bs, _ := io.ReadAll(resp.Body)
		if !json.Valid(bs) {
			return s
		} else {
			_ = json.Unmarshal(bs, &s)
		}
	}
	return s
}

func FindByUuidDSLBody(uuids string, size int) string {
	m := map[string]any{
		"size":             size,
		"track_total_hits": true,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"match": map[string]interface{}{
							"uuid.keyword": uuids,
						},
					},
				},
			},
		},
	}
	bs, _ := json.Marshal(m)
	return string(bs)
}

func FindTermByUuidDSLBody(uuids string, size int) string {
	m := map[string]any{
		"size":             size,
		"track_total_hits": true,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"uuid.keyword": map[string]interface{}{
								"value": uuids,
							},
						},
					},
				},
			},
		},
	}
	bs, _ := json.Marshal(m)
	return string(bs)
}

func FindTimeByUuidDSLBody(timestamp *time.Time, size int) string {
	m := map[string]any{
		"size":             size,
		"track_total_hits": true,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"range": map[string]interface{}{
							"@timestamp": map[string]interface{}{
								"lte": timestamp,
							},
						},
					},
				},
			},
		},
		"sort": []map[string]interface{}{
			{
				"@timestamp": map[string]string{
					"order": "desc",
				},
			},
		},
	}
	bs, _ := json.Marshal(m)
	return string(bs)
}

type FieldCapsResponse struct {
	Indices map[string]struct {
		Fields map[string]map[string]struct {
			Type string `json:"type"`
		} `json:"fields"`
	} `json:"indices"`
}

func (ec *ElasticClientV7) GetFieldCapabilities(index string, fields string) (*FieldCapsResponse, int) {
	unmapped := false
	req := esapi.FieldCapsRequest{
		Index:           []string{index},
		Fields:          []string{fields},
		IncludeUnmapped: &unmapped,
		Pretty:          true,
	}

	res, err := req.Do(context.Background(), ec.client)
	if err != nil {
		fmt.Println("Error:", err)
		return &FieldCapsResponse{}, res.StatusCode
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return &FieldCapsResponse{}, res.StatusCode
	}

	var fieldCapsResponse FieldCapsResponse
	err = json.Unmarshal(body, &fieldCapsResponse)
	if err != nil {
		fmt.Println("Error:", err)
		return &FieldCapsResponse{}, res.StatusCode
	}

	return &fieldCapsResponse, res.StatusCode
}
