package xelastic

import (
	"easymonitor/conf"
	"easymonitor/infra"
	"net/http"
	"strconv"
	"time"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
)

type ElasticClient interface {
	FindByDSL(index string, dsl string, source []string) ([]any, int, int)
	CountByDSL(index string, dsl string) (int, int)
}

func NewElasticClient(esConfig conf.EsConfig, version string) ElasticClient {
	DebugEnable, _ := strconv.ParseBool(conf.AppConf.EsConfig.DebugEnable)
	client, err := elasticsearch7.NewClient(elasticsearch7.Config{
		Addresses:         esConfig.Addresses,
		Username:          esConfig.Username,
		Password:          esConfig.Password,
		MaxRetries:        esConfig.MaxRetries,
		EnableDebugLogger: DebugEnable,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   60,
			ResponseHeaderTimeout: 30 * time.Second,
			DisableCompression:    true,
		},
	})
	if err != nil {
		infra.Logger.Errorln(err)
		return nil
	}
	c := &ElasticClientV7{
		client: client,
	}
	infra.Logger.Infoln("初始化elastic连接成功")
	return c
}
