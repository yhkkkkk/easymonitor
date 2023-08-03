package conf

import (
	"easymonitor/infra"
	"easymonitor/utils"
	"easymonitor/utils/xtime"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/creasty/defaults"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
)

var AppConf *AppConfig

type EsConfig struct {
	Addresses   []string `toml:"Addr"`
	Username    string   `toml:"Username"`
	Password    string   `toml:"Password"`
	ConnTimeout uint     `toml:"ConnTimeout" default:"10"`
	MaxRetries  int      `toml:"MaxRetries" default:"3"`
	DebugEnable string   `toml:"DebugEnable"`
	Version     string   `toml:"Version" default:"v7"`
}

type AppConfig struct {
	Server struct {
		Port           string `toml:"Port" default:":16060"`
		ReadTimeout    int    `toml:"ReadTimeout" default:"30"`
		WriteTimeout   int    `toml:"WriteTimeout" default:"30"`
		MaxHeaderBytes int    `toml:"MaxHeaderBytes" default:"600"`
		HttpTimeout    int    `toml:"HttpTimeout" default:"300"`
		EnableLimit    string `toml:"EnableLimit" default:"false"`
		LimitSize      int    `toml:"LimitSize" default:"10"`
	} `toml:"Server"`
	Redis struct {
		Addr         string          `toml:"Addr" default:"127.0.0.1"`
		Port         int             `toml:"Port" default:"6379"`
		Password     string          `toml:"Password"`
		Db           int             `toml:"Db" default:"10"`
		ReadTimeout  int             `toml:"ReadTimeout" default:"30"`
		WriteTimeout int             `toml:"WriteTimeout" default:"30"`
		DialTimeout  int             `toml:"DialTimeout" default:"30"`
		PoolSize     int             `toml:"PoolSize" default:"20"`
		MinIdleConns int             `toml:"MinIdleConns" default:"5"`
		MaxRetries   int             `toml:"MaxRetries" default:"3"`
		Expire       xtime.TimeLimit `toml:"Expire"`
	} `toml:"Redis"`
	EsConfig `toml:"Elasticsearch"`
	Pool     struct {
		GorountineSize   int    `toml:"GorountineSize" default:"1500"`
		AlertQueueSize   int    `toml:"AlertQueueSize" default:"1500"`
		MaxBlockingTasks int    `toml:"MaxBlockingTasks" default:"200"`
		NonBlock         string `toml:"NonBlock"`
	} `toml:"Pool"`
	Alert struct {
		AlertUrl string `toml:"AlertUrl"`
		Token    struct {
			Url    string `toml:"Url" default:"https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal/"`
			Secret string `toml:"Secret"`
		} `toml:"Token"`
	} `toml:"Alert"`
	Exclude struct {
		Env []string `toml:"Env"`
		Log []string `toml:"Log"`
	} `toml:"Exclude"`
}

// ElasticJob error日志任务
type ElasticJob struct {
	AlertMessage *LogstashAlert
	StartsAt     *time.Time
	// EndsAt    *time.Time
	// Scheduler *gocron.Scheduler
}

type ElasticAlert struct {
	appConf *AppConfig
	opts    *FlagOption
}

type AlertSampleMessage struct {
	ES        EsConfig   `json:"es"`
	Index     string     `json:"index"`
	Env       string     `json:"env"`
	Log       string     `json:"log"`
	Timestamp *time.Time `json:"timestamp"`
	Message   string     `json:"message"`
	Uuid      string     `json:"uuid"`
}

type FlagOption struct {
	ConfigPath string `short:"c" long:"config" description:"application.toml path" default:"../../application.toml"`
	Debug      bool   `long:"debug" description:"debug log level"`
	Verbose    string `short:"v" long:"verbose" description:"log level: debug、info、warn、error" default:"debug"`
	Rule       string `long:"rule" description:"will only run the given single rule. The rule file may be a complete file path"`
	Zone       string `short:"z" long:"zone" description:"time zone, e.g like PRC、UTC" default:"Asia/Shanghai"`
}

func (f FlagOption) GetLogLevel() log.Level {
	if f.Debug {
		return log.DebugLevel
	}
	switch f.Verbose {
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn", "warning":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.InfoLevel // Return a default level
	}
}

// LogstashAlert 设置日志索引解析模板(解析需要的字段)
type LogstashAlert struct {
	Uuid      string     `json:"uuid"`
	Timestamp *time.Time `json:"@timestamp"`
	IndexName string     `json:"index_name"`
	Host      struct {
		Os struct {
			Codename string `json:"codename"`
			Name     string `json:"name"`
			Type     string `json:"type"`
			Family   string `json:"family"`
			Version  string `json:"version"`
			Kernel   string `json:"kernel"`
			Platform string `json:"platform"`
		} `json:"os"`
		Name          string   `json:"name"`
		Hostname      string   `json:"hostname"`
		Ip            []string `json:"ip"`
		Mac           []string `json:"mac"`
		Id            string   `json:"id"`
		Containerized bool     `json:"containerized"`
		Architecture  string   `json:"architecture"`
	} `json:"host"`
	Log struct {
		Offset int64 `json:"offset"`
		File   struct {
			Path string `json:"path"`
		} `json:"file"`
		Flags []string `json:"flags"`
	} `json:"log"`
	Fields struct {
		Env     string `json:"env"`
		LogType string `json:"log_type"`
	} `json:"fields"`
	LogType     string `json:"log_type"`
	Message     string `json:"message"`
	LogDateTime string `json:"logDateTime"`
	LogLevel    string `json:"logLevel"`
	Pid         string `json:"pid"`
	Thread      string `json:"thread"`
	Logger      string `json:"logger"`
	MessageLess string `json:"message_less"`
	ErrorStack  string `json:"errorStack"`
}

func GetAppConfig(path string) (*AppConfig, error) {
	c := &AppConfig{}
	ok, err := utils.PathExists(path)
	if err != nil {
		return nil, fmt.Errorf("配置文件路径存在错误: %w", err)
	}
	if !ok {
		if err := defaults.Set(c); err != nil {
			return nil, fmt.Errorf("默认设置对象错误: %w", err)
		}
		AppConf = c
		return nil, fmt.Errorf("%s 不存在", path)
	}
	confBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 %s 错误: %w", path, err)
	}
	// 解析TOML文件
	if err := toml.Unmarshal(confBytes, c); err != nil {
		return nil, fmt.Errorf("解析配置文件错误: %w", err)
	}
	var configSchema interface{}
	if err := toml.Unmarshal([]byte(infra.AppTomlSchema), &configSchema); err != nil {
		return nil, fmt.Errorf("解析配置模式错误: %w", err)
	}
	confjson, _ := json.Marshal(c)
	confschema, _ := json.Marshal(configSchema)
	appConfLoader := gojsonschema.NewBytesLoader(confjson)
	appConfSchemaLoader := gojsonschema.NewBytesLoader(confschema)
	res, err := gojsonschema.Validate(appConfSchemaLoader, appConfLoader)
	if err != nil {
		return nil, fmt.Errorf("配置文件模式错误: %w", err)
	}
	if !res.Valid() {
		return nil, fmt.Errorf("配置文件模式错误: %s", res.Errors()[0].String())
	}
	// infra.Logger.Debugf("配置文件解析成功: %v", c)
	AppConf = c
	return c, nil
}
