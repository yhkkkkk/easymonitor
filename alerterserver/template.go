package alerterserver

//
//import (

//)
//
//// 定义一个 HTML 模板
//const tpl = `
//<div>
//    <h1>{{.Summary}}</h1>
//    <p>UUID: {{.Uuid}}</p>
//    <p>Timestamp: {{.Timestamp}}</p>
//    <p>Fields: {{.Fields.Env}}, {{.Fields.LogType}}</p>
//    <p>LogType: {{.LogType}}</p>
//</div>
//`
//
//func main() {
//	// 创建一个新的模板对象
//	t, err := template.New("content").Parse(tpl)
//	if err != nil {
//		fmt.Errorf("parse template error: %w", err)
//	}
//
//	// 将 content 对象插入到模板中
//	err = t.Execute(&buf, content)
//	if err != nil {
//		fmt.Errorf("execute template error: %w", err)
//	}
//
//}
