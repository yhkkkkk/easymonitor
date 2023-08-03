package alerterserver

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"easymonitor/conf"
	"easymonitor/infra"
	"easymonitor/utils"
	redisx "easymonitor/utils/redis"
	"easymonitor/utils/xelastic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/creasty/defaults"
	"github.com/panjf2000/ants/v2"
)

var (
	httpClient *http.Client
	AlertQueue chan *conf.ElasticJob
	Wg         *sync.WaitGroup
	pool       *ants.Pool
	excludeMap map[string]bool
	tokenObj   *Token
	EsClient   xelastic.ElasticClient
	ctx        = context.Background()
)

type Token struct {
	Url               string `default:"https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal/"`
	Secret            string
	TenantAccessToken string
	TokenExpires      time.Time
}

type Data struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

type Fields struct {
	Env     string `json:"env"`
	LogType string `json:"log_type"`
}

type MessageData struct {
	Timestamp string `json:"timestamp"`
	MsgType   string `json:"msg_type"`
	Sign      string `json:"sign"`
	Content   string `json:"content"`
	// Content AlertContentInfo `json:"content"`
}

type AlertContentInfo struct {
	Text AlertContent `json:"text"`
}

type AlertContent struct {
	Summary   string `json:"summary"`
	Uuid      string `json:"uuid"`
	Timestamp string `json:"timestamp"`
	Fields    Fields `json:"fields"`
	LogType   string `json:"log_type"`
}

//func enqueueLog(log *conf.LogstashAlert, rdb *redis.Client) error {
//	logBytes, err := json.Marshal(log)
//	if err != nil {
//		return fmt.Errorf("failed to marshal log: %w", err)
//	}
//
//	// 使用Redis的LPUSH命令将日志推入队列
//	err = rdb.LPush("log_alert_queue", logBytes).Err()
//	if err != nil {
//		return fmt.Errorf("failed to push log into queue: %w", err)
//	}
//
//	return nil
//}

func InitPool() {
	httpClient = &http.Client{
		Timeout: time.Second * time.Duration(conf.AppConf.Server.HttpTimeout),
	}
	AlertQueue = make(chan *conf.ElasticJob, conf.AppConf.Pool.AlertQueueSize)
	excludeMap = utils.CreateMap(conf.AppConf.Exclude.Env) // 初始化告警排除环境
	tokenObj = NewToken()
	Wg = &sync.WaitGroup{}
	nonBlock, _ := strconv.ParseBool(conf.AppConf.Pool.NonBlock)
	EsClient = xelastic.NewElasticClient(conf.AppConf.EsConfig, conf.AppConf.EsConfig.Version) // 初始化es客户端
	pool, _ = ants.NewPool(
		conf.AppConf.Pool.GorountineSize,
		ants.WithNonblocking(nonBlock),
		ants.WithLogger(infra.Logger),
		ants.WithMaxBlockingTasks(conf.AppConf.Pool.MaxBlockingTasks), // 设置最大等待队列长度
		ants.WithPanicHandler(func(r interface{}) {
			infra.Logger.Debugf("goroutine recovered from panic: %s", r)
			panic(r)
		}),
	)
}

func submitToPool(fn func()) error {
	Wg.Add(1)
	err := pool.Submit(func() {
		defer func() {
			Wg.Done()
			if r := recover(); r != nil {
				infra.Logger.Errorf("Recovered from panic: %v", r)
			}
		}()
		fn()
	})
	return err
}

func parseLog(request *http.Request) (*conf.ElasticJob, error) {
	gz, err := gzip.NewReader(request.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		_ = gz.Close()
	}() // gzip.NewReader返回错误，gz将为nil

	var alert conf.LogstashAlert
	err = json.NewDecoder(gz).Decode(&alert)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("empty input: %w", err)
		}
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}
	now := time.Now()
	job := &conf.ElasticJob{
		AlertMessage: &alert,
		StartsAt:     &now, // 设置 StartsAt 为当前时间的地址
	} // 创建一个ElasticJob对象，并将解析出的LogstashAlert对象设置为ElasticJob的AlertMessage
	sampleMessage := &conf.AlertSampleMessage{
		Index:     alert.IndexName,
		Env:       alert.Fields.Env,
		Log:       alert.Fields.LogType,
		Timestamp: alert.Timestamp,
		Message:   alert.Message,
		Uuid:      alert.Uuid,
	}
	// infra.Logger.Debugf("发送redis: %v", sampleMessage)
	err = submitToPool(func() {
		ProcessMessage(sampleMessage) // 将索引内容发送到Redis中  便于ui查询
	})
	if err != nil {
		infra.Logger.Errorf("redis协程任务创建失败，Cause: %s", err)
		Wg.Done()
	}
	//if _, ok := excludeMap[alert.Fields.Env]; !ok {
	//	// 如果环境不在排除列表中，则将告警消息发送到AlertQueue中
	//	AlertQueue <- job
	//}
	return job, nil
}

func ProcessMessage(message *conf.AlertSampleMessage) {
	message.ES = conf.AppConf.EsConfig
	messages, err := json.Marshal(message)
	if err != nil {
		infra.Logger.Errorf("failed to marshal message: %s", err)
		return
	}
	_, err = redisx.Client.Set(ctx, message.Uuid, string(messages), conf.AppConf.Redis.Expire.GetTimeDuration()).Result()
	if err != nil {
		infra.Logger.Errorf("redis set失败，Cause: %s", err)
	}
}

func processAlert(ctx context.Context) {
	//logBytes, err := rdb.BLPop(time.Second, "log_alert_queue").Result()
	//if err != nil {
	//	log.Printf("Error while popping from the queue: %v", err)
	//	continue
	//}
	//var logstashAlert conf.LogstashAlert
	//if err := json.Unmarshal([]byte(logBytes[1]), &logstashAlert); err != nil {
	//	log.Printf("Failed to unmarshal log from queue: %v", err)
	//	continue
	//}
	for {
		select {
		case alert := <-AlertQueue:
			_, err := HttpSendAlert(alert)
			if err != nil {
				infra.Logger.Errorf("Sending Error, Cause: %v", err)
				panic(err)
			}
		case <-ctx.Done():
			infra.Logger.Debugf("Context超时 退出当前协程. Cause: %s", ctx.Err())
			return
		}
	}
}

func AlertHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		alert, err := parseLog(request)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		AlertQueue <- alert
		var once sync.Once
		Wg.Add(1)
		err = pool.Submit(func() {
			defer func() {
				once.Do(Wg.Done)
				if r := recover(); r != nil {
					infra.Logger.Errorf("Recovered from panic: %v", r)
				}
			}()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 在协程池任务内部创建超时上下文
			defer cancel()
			processAlert(ctx)
		})
		if err != nil {
			infra.Logger.Errorf("告警协程任务创建失败，Cause: %s", err)
			once.Do(Wg.Done) // 避免每个协程多次调用Done
			return
		}
	}
}

func printCallerInfo() {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		fmt.Println("runtime.Caller failed")
		return
	}
	funcName := runtime.FuncForPC(pc).Name()
	fmt.Printf("Called from %s, line #%d, func: %s\n", file, line, funcName)
}

// NewToken 创建Token
func NewToken() *Token {
	token := &Token{}
	err := defaults.Set(token)
	if err != nil {
		infra.Logger.Errorf("defaults set token error: %s", err)
	}
	return token
}

// GenSign 生成签名
func GenSign(secret string, timestamp int64) (string, error) {
	//timestamp + key 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret

	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}

	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}

// GetTenantAccessToken 获取告警Token
func (y *Token) GetTenantAccessToken() (string, error) {
	if y.TenantAccessToken != "" && time.Now().Before(y.TokenExpires) {
		infra.Logger.Infof("获取未过期token => token: %s", y.TenantAccessToken)
		return y.TenantAccessToken, nil
	} else {
		data := &Data{}
		err := defaults.Set(data)
		if err != nil {
			return "", fmt.Errorf("defaults set data error: %w", err)
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("解析data error: %w", err)
		}

		req, err := http.NewRequest(http.MethodPost, y.Url, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", fmt.Errorf("获取token请求构造失败 error: %w", err)
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		resp, err := httpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("执行请求失败 error: %w", err)
		}
		defer func() {
			if resp != nil {
				_ = resp.Body.Close()
			}
		}()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("获取token响应体失败 error: %w", err)
		}
		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return "", fmt.Errorf("解析token响应体失败 error: %w", err)
		}

		y.TenantAccessToken = result["tenant_access_token"].(string)
		y.TokenExpires = time.Now().Add(time.Second * time.Duration(result["expire"].(float64)))

		infra.Logger.Infof("token过期, 重新获取token => token: %s, expires: %v", y.TenantAccessToken, y.TokenExpires)
		return y.TenantAccessToken, nil
	}
}

func CreateAlertMessage(data *conf.ElasticJob) (*AlertContentInfo, string, string, string, error) {
	token, err := tokenObj.GetTenantAccessToken()
	if err != nil {
		return nil, "", "", "", fmt.Errorf("获取租户访问令牌失败 error: %w", err)
	}
	t, _ := time.Parse(time.RFC3339, data.AlertMessage.Timestamp.Format(time.RFC3339))
	loc, _ := time.LoadLocation(infra.Location)
	t = t.In(loc)
	timestamp := time.Now().Unix()
	sign, _ := GenSign(tokenObj.Secret, timestamp)
	infra.Logger.Infof("收到异常日志 <= uuid: %s, 日志时间: %v", data.AlertMessage.Uuid, t)
	content := &AlertContentInfo{
		Text: AlertContent{
			Summary:   "error日志告警",
			Uuid:      data.AlertMessage.Uuid,
			Timestamp: t.Format(time.RFC3339),
			Fields: Fields{
				Env:     data.AlertMessage.Fields.Env,
				LogType: data.AlertMessage.Fields.LogType,
			},
			LogType: data.AlertMessage.LogType,
		},
	}
	jsonContent, err := json.Marshal(content)
	messageData := &MessageData{
		Timestamp: strconv.FormatInt(timestamp, 10),
		MsgType:   "text",
		Sign:      sign,
		Content:   string(jsonContent),
	}
	jsonData, err := json.Marshal(messageData)
	if err != nil {
		return nil, "", "", "", fmt.Errorf("解析数据失败, error: %w", err)
	} else {
		infra.Logger.Infof("解析数据成功: %s", jsonData)
	}
	return content, token, strconv.FormatInt(timestamp, 10), sign, nil
}

func SendAlert(contentData *AlertContentInfo, token, timestamp, sign string) (bool, error) {
	var almessage string
	if contentData.Text.LogType != "" {
		almessage = fmt.Sprintf(
			"告警类型: %s \n日志uuid: %s \n日志时间: %s \n日志环境: %s \n日志类型: %s \n基本内容: %s",
			contentData.Text.Summary, contentData.Text.Uuid, contentData.Text.Timestamp, contentData.Text.Fields.Env, contentData.Text.Fields.LogType, contentData.Text.LogType)
	} else {
		almessage = fmt.Sprintf(
			"告警类型: %s \n日志uuid: %s \n日志时间: %s \n日志环境: %s \n日志类型: %s",
			contentData.Text.Summary, contentData.Text.Uuid, contentData.Text.Timestamp, contentData.Text.Fields.Env, contentData.Text.Fields.LogType)
	}
	formattedJsonData := fmt.Sprintf(
		`{"msg_type": "text", "content": {"text": %q}}`, almessage) //使用签名有点问题
	req, err := http.NewRequest(http.MethodPost, conf.AppConf.Alert.AlertUrl, strings.NewReader(formattedJsonData))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	var resp *http.Response
	if contentData.Text.Fields.Env != "" {
		if !utils.StringInMap(contentData.Text.Fields.Env, excludeMap) {
			resp, err = httpClient.Do(req)
			if err != nil {
				return false, fmt.Errorf("send alert to feishu error: %w", err)
			}
			defer func() {
				_ = resp.Body.Close()
			}()
			infra.Logger.Infof("发送告警 => 响应状态: %s, 告警内容: %s", resp.Status, formattedJsonData)
			return true, nil
		} else {
			infra.Logger.Debugf("环境 %s 在排除列表中，跳过发送alert请求", contentData.Text.Fields.Env)
			return false, nil
		}
	} else {
		infra.Logger.Warnln("环境变量字段为空，跳过发送alert请求")
		return false, nil
	}
}

func HttpSendAlert(data *conf.ElasticJob) (bool, error) {
	messageData, token, timestamp, sign, err := CreateAlertMessage(data)
	if err != nil {
		return false, err
	}
	return SendAlert(messageData, token, timestamp, sign)
}
