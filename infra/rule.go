package infra

var AppTomlSchema = `
[Redis]
Addr = "type: string"
Port = "type: number"
Password = "type: string"
Db = "type: number"
ReadTimeout = "type: number"
WriteTimeout = "type: number"
DialTimeout = "type: number"
PoolSize = "type: number"
MaxRetries = "type: number"
[Redis.Expire]
Days = "type: number"

[Elasticsearch]
Addr = "type: []string"
Username = "type: string"
Password = "type: string"
ConnTimeout = "type: number"
Version = "type: string"

[Alert]
AlertUrl = "type: string"
`
