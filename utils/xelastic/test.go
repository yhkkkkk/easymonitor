package xelastic

import (
	"context"
	"easymonitor/conf"
	"fmt"
	"log"
	"os"

	olelasticv7 "github.com/olivere/elastic/v7"
)

var Client *olelasticv7.Client

// Init 初始化链接
func Init() {
	var err error

	//初始化Client
	Client, err = olelasticv7.NewClient(
		//设置es的url，支持多节点
		olelasticv7.SetURL(conf.AppConf.EsConfig.Addresses...),
		//允许指定弹性是否应该定期检查集群（默认为true），在使用docker部署时，应该设置为false，否则检查集群会获取其节点内网地址，导致健康检查失败，导致错误
		olelasticv7.SetSniff(false),
		//基于http base auth 验证机制的账号密码。
		olelasticv7.SetBasicAuth("elastic", "pasword"),
		//设置日志输出，传入实现elastic.logger接口的日志对象
		olelasticv7.SetErrorLog(log.New(os.Stderr, "ELASTIC_ERROR ", log.LstdFlags)),
		olelasticv7.SetInfoLog(log.New(os.Stdout, "ELASTIC ", log.LstdFlags)),
	)
	if err != nil {
		panic(err)
	}
}

// 建立测试数据模型
// User 假定有一个user数据，字段内容为id,name,age,city,tags

// 对应的index名称为new_es_user
const userIndex = "new_es_user"

// mapping json 字符串
var userMapping = `{
  "mappings": {
    "properties": {
      "id": {
        "type": "keyword"
      },
      "name": {
        "type": "text"
      },
      "gender": {
        "type": "keyword"
      },
      "age": {
        "type": "integer"
      },
      "City": {
        "type": "text"
      },
      "Tags": {
        "type": "keyword"
      }
    }
  }
}`

// UserModel 与mapping对应的model，用于插入、修改、查询
type UserModel struct {
	Id     int      `json:"id"`
	Name   string   `json:"name"`
	Gender int      `json:"gender"` //1-男 2-女
	Age    int      `json:"age"`
	City   string   `json:"city"`
	Tags   []string `json:"tags"`
}

// UserIndex 对user信息进行操作对象
type UserIndex struct {
	index   string
	mapping string
}

// 新建一个新的操作对象
func NewUserIndex() (*UserIndex, error) {
	user := &UserIndex{
		index:   userIndex,
		mapping: userMapping,
	}
	err := user.init()
	if err != nil {
		return nil, err
	}
	return user, nil
}

// 初始化index，保证对应的index在Es中存在，并定义mapping，方便后续操作
func (u *UserIndex) init() error {
	ctx := context.Background()
	//查询指定index是否存在，返回bool
	exist, err := Client.IndexExists(u.index).Do(ctx)
	if err != nil {
		fmt.Println("index check exist failed", err)
		return err
	}
	if !exist {
		//创建index，并在body中指定mapping。
		//在elasticsearch7中不再区分type，直接默认为_doc，所以此处的mapping及代码中均不用指定type
		_, err = Client.CreateIndex(u.index).Body(u.mapping).Do(ctx)
		if err != nil {
			fmt.Println("create index failed", err)
			return err
		}
	}
	return nil
}
