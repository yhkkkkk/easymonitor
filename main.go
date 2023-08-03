package main

import (
	"easymonitor/alerterserver"
	"easymonitor/boot"
	"easymonitor/conf"
	"easymonitor/infra"
	"easymonitor/utils/redis"
	"easymonitor/utils/xtime"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/panjf2000/ants/v2"
)

func main() {
	fmt.Println(infra.Logo)
	var opts conf.FlagOption
	defer func() {
		if e := recover(); e != nil {
			infra.Logger.Errorf("%s\n", e)
			if opts.Debug {
				debug.PrintStack()
			}
		}
	}()
	// 创建了一个新的命令行参数解析器
	p := flags.NewParser(&opts, flags.HelpFlag)
	_, err := p.ParseArgs(os.Args)
	if err != nil {
		panic(err)
	}

	// 初始化日志文件
	err = infra.InitLog(opts.GetLogLevel())
	if err != nil {
		infra.Logger.Errorln(err)
	}
	xtime.FixedZone(opts.Zone)

	// 初始化配置文件
	_, err = conf.GetAppConfig(opts.ConfigPath)
	if err != nil {
		infra.Logger.Errorln(err)
	}

	// 初始化协程池等
	alerterserver.InitPool()
	infra.Logger.Infof("初始化任务队列, 队列长度: %d", conf.AppConf.Pool.AlertQueueSize)
	infra.Logger.Infof("初始化协程池成功, 协程池大小: %d", conf.AppConf.Pool.GorountineSize)
	// 初始化redis
	redis.Setup()
	infra.Logger.Infof("初始化redis连接成功, 连接池大小: %d", conf.AppConf.Redis.PoolSize)

	// 启动告警日志消费者
	// go startAlertConsumer(rdb)

	http.HandleFunc("/api/alert/log", alerterserver.AlertHandler())
	if conf.AppConf.Server.EnableLimit == "true" {
		http.HandleFunc("/alert/ui", conf.LimitHandler(boot.RenderAlertMessage)) // ui访问加限流处理
	} else {
		http.HandleFunc("/alert/ui", boot.RenderAlertMessage)
	}
	infra.Logger.Infof("alertserver start at port %s", strings.Split(conf.AppConf.Server.Port, ":")[1])
	server := &http.Server{
		Addr:              conf.AppConf.Server.Port,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    conf.AppConf.Server.MaxHeaderBytes,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			infra.Logger.Errorf("error starting alertserver: %v\n", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	// syscall.SIGINT一般是由ctrl+c产生 syscall.SIGTERM一般是由kill产生 syscall.SIGHUP一般是由终端关闭产生
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	s := <-quit
	switch s {
	case syscall.SIGHUP:
		_, err = conf.GetAppConfig(opts.ConfigPath)
		if err != nil {
			infra.Logger.Warnln(err)
		} else {
			infra.Logger.Infoln("reload application config success!")
		}
	case syscall.SIGINT:
		fallthrough
	case syscall.SIGTERM:
		infra.Logger.Infoln("exiting...")
		alerterserver.Wg.Wait() // 等待所有告警处理完成
		ants.Release()
		close(alerterserver.AlertQueue)
		return
	}
}
