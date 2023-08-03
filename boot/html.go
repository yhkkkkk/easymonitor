package boot

import (
	"context"
	"easymonitor/alerterserver"
	"easymonitor/conf"
	"easymonitor/infra"
	redisx "easymonitor/utils/redis"
	"easymonitor/utils/xelastic"
	"easymonitor/utils/xtime"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
)

func RenderAlertMessage(writer http.ResponseWriter, request *http.Request) {
	q := request.URL.Query()
	key := q.Get("uuid")
	sizeStr := q.Get("size")
	dcontext := q.Get("context")
	var size int
	if sizeStr != "" {
		var err error
		size, err = strconv.Atoi(sizeStr)
		if err != nil {
			http.Error(writer, "无效的size参数: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		size = 1
	}
	if key == "" {
		_, _ = writer.Write([]byte("未获取到查询参数!"))
		infra.Logger.Infoln("查询日志uuid => 未获取到查询参数!")
		return
	}
	infra.Logger.Infof("查询日志uuid => %s", key)
	var message conf.AlertSampleMessage
	var ctx = context.Background()
	v, e := redisx.Client.Get(ctx, key).Result()
	infra.Logger.Debugf("获取redis数据 => %s", v)
	if e != nil {
		http.Error(writer, "服务器内部错误! Cause: "+e.Error(), http.StatusInternalServerError)
		return
	}
	err := json.Unmarshal([]byte(v), &message)
	if err != nil {
		http.Error(writer, "", http.StatusInternalServerError)
		return
	}
	// 为模板添加两个自定义函数：json 和 showTime。这两个函数可以在模板中使用
	t, _ := template.New("index.html").Funcs(template.FuncMap{
		"json": func(v any) string {
			res, _ := json.Marshal(v)
			return string(res)
		},
		"showTime": func(v map[string]any) string {
			m := v["_source"].(map[string]any)
			return xtime.TimeFormatISO8601(xtime.Parse(m["@timestamp"].(string)))
		},
	}).Parse(infra.HtmlPage)
	var hits []any
	var hitsStr []byte
	var body string
	if dcontext == "true" {
		body = xelastic.FindTimeByUuidDSLBody(xtime.DtimeParse(message.Timestamp), size)
	} else {
		body = xelastic.FindTermByUuidDSLBody(message.Uuid, size)
	}
	hits, _, _ = alerterserver.EsClient.FindByDSL(message.Index, body, nil)
	hitsStr, _ = json.Marshal(hits)
	// 将查出来的数据渲染到html页面
	_ = t.Execute(writer, map[string]any{
		"hitsStr": string(hitsStr),
		"hits":    hits,
	})
}
