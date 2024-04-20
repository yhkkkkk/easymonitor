## 监控系统架构

[](https://github.com/HobbyBear/easymonitor/tree/main#%E7%9B%91%E6%8E%A7%E7%B3%BB%E7%BB%9F%E6%9E%B6%E6%9E%84)

[![image.png](https://camo.githubusercontent.com/883df4f6e200cf9119bc77099e83370deac0c724d1bf77c602950c57351b0213/68747470733a2f2f73322e6c6f6c692e6e65742f323032332f30372f32342f734170514e766f646b456157364a782e706e67)](https://camo.githubusercontent.com/883df4f6e200cf9119bc77099e83370deac0c724d1bf77c602950c57351b0213/68747470733a2f2f73322e6c6f6c692e6e65742f323032332f30372f32342f734170514e766f646b456157364a782e706e67)
![image](https://github.com/yhkkkkk/easymonitor/assets/125347894/4cea2a5b-5f33-483d-a7e2-453502658b1b)

## 目录结构

[](https://github.com/yhkkkkk/easymonitor/tree/main#%E7%9B%AE%E5%BD%95%E7%BB%93%E6%9E%84)

```shell
(base) ➜  easymonitor git:(main) ✗ tree -L 1
.
├── README.md
├── build.sh // 对webhookserver 以及 webapp 项目进行编译 ，然后放到program文件夹里
├── docker-compose.yaml // 启动各个监控系统组件
├── go.mod
├── go.sum
├── grafanadashbord // 放置grafana的监控面板导出的json文件，可直接启动项目，然后导入json文件，即可构建监控面板  目前暂未使用
├── infra // 项目基础组件的代码
├── logconf // 放置主机上的日志采集配置文件，filebeat.yml 中会引入这个文件夹下的配置规则做不同的采集策略
├── program // 放置alertserver项目编译好的二进制文件
├── utils // 一些redis、es、time工具代码
└── alerterserver // 模拟自研日志报警系统代码
```

## 启动步骤

[](https://github.com/yhkkkkk/easymonitor/tree/main#%E5%90%AF%E5%8A%A8%E6%AD%A5%E9%AA%A4)

```shell
cd easymonitor
go build -0 alertserver main.go
mv alertserver program/
使用supervisor启动 后续会改为容器化 supervisor配置放在了logconf目录下面
```

## 感谢

感谢蓝胖子大佬的项目给予的灵感 该程序绝大部分是在基础上做的二开

https://github.com/HobbyBear/easymonitor

也感谢prom-elastic-alert这个项目给予的部分灵感

https://github.com/dream-mo/prom-elastic-alert
