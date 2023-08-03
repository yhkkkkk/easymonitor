package infra

const (
	Cost       = "cost"
	MetricType = "metricType"
	Stack      = "stack"
	Location   = "Asia/Shanghai"
)

var Logo = `
 ________  ___       _______   ________  _________  ________  _______   ________       
|\   __  \|\  \     |\  ___ \ |\   __  \|\___   ___\\   ____\|\  ___ \ |\   __  \
\ \  \|\  \ \  \    \ \   __/|\ \  \|\  \|___ \  \_\ \  \___|\ \   __/|\ \  \|\  \ 
 \ \   __  \ \  \    \ \  \_|/_\ \   _  _\   \ \  \ \ \_____  \ \  \_|/_\ \   _  _\ 
  \ \  \ \  \ \  \____\ \  \_|\ \ \  \\  \|   \ \  \ \|____|\  \ \  \_|\ \ \  \\  \
   \ \__\ \__\ \_______\ \_______\ \__\\ _\    \ \__\  ____\_\  \ \_______\ \__\\ _\
    \|__|\|__|\|_______|\|_______|\|__|\|__|    \|__| |\_________\|_______|\|__|\|__|
                                                      \|_________|                                                    
                                                                                    `

var HtmlPage = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width,initial-scale=1,user-scalable=0">
    <title>日志告警详细信息</title>
    <link type="text/css" href="http://summerstyle.github.io/jsonTreeViewer/libs/jsonTree/jsonTree.css" rel="stylesheet"/>
    <link type="text/css" href="https://www.layuicdn.com/layui-v2.7.4/css/layui.css" rel="stylesheet"/>
</head>
<style type="text/css">
    table {
        table-layout: fixed;
    }
    table td {
        word-wrap: break-word;
    }
    .jsontree_node {
        width: 95%;
        word-wrap: break-word;
        display: inline-block;
    }
	.notice-box {
        border: 1px solid #ccc;
        padding: 10px;
        margin: 10px 0;
        background-color: #f9f9f9;
    }
</style>
<body>
<div style="margin: 10px">
    <blockquote class="layui-elem-quote">告警日志</blockquote>
	<div class="notice-box">
        <p>URL的三个请求参数解释：</p>
        <ul>
            <li><strong>uuid</strong>: 日志唯一标识</li>
            <li><strong>context</strong>: true: 开启日志上下文, false: 关闭 (默认关闭)</li>
            <li><strong>size</strong>: 查询日志个数</li>
			<li><strong>url类似于: http://47.99.144.15:16060/alert/ui?uuid=cf0684f3-08bf-4c4b-bf6e-6ce1220ad734&size=5&context=true </strong></li>
        </ul>
    </div>
    <table class="layui-table" lay-size="sm">
        <colgroup>
            <col width="110">
            <col width="700">
            <col width="60">
        </colgroup>
        <thead>
        <tr>
            <th>日志时间</th>
            <th>索引内容</th>
            <th style="text-align: center">操作</th>
        </tr>
        </thead>
        <tbody>
        {{range $i, $v := .hits}}
        <tr>
            <td>{{showTime $v}}</td>
            <td>{{json $v}}</td>
            <td align="center">
                <button onclick='jsonViewer({{$i}})' class="layui-btn layui-btn-xs">查看格式化日志</button>
            </td>
        </tr>
        {{end}}
        </tbody>
    </table>
</div>
</body>
<script type="text/javascript" src="http://summerstyle.github.io/jsonTreeViewer/libs/jsonTree/jsonTree.js"></script>
<script type="text/javascript" src="https://www.layuicdn.com/layui-v2.7.4/layui.js"></script>
<script>

    let listString = {{.hitsStr}}
    function renderJson(jsonContent, wrapperId) {
        let wrapper = document.getElementById(wrapperId);
        if (typeof jsonContent === 'string') {
            try {
                console.log(jsonContent)
                var data = JSON.parse(jsonContent)
            } catch (e) {
                console.log(e)
                var data = {"message": jsonContent}
            }
        } else {
            data = jsonContent;
        }
        let tree = jsonTree.create(data, wrapper);
        tree.expand()
    }

    function jsonViewer(index) {
        let lists = JSON.parse(listString)
		let jsonContent = lists[index];
        let indexVal = index + 1
        let title = '【第' + indexVal + '条】日志内容'
        let warp = 'wrapper_' + indexVal
        layer.open({
            id: warp,
            type: 1,
            title: title,
            skin: 'layui-layer-rim',
            area: ['80%', '80%'],
            content: '<div style="overflow-x: scroll" ' + 'id="' + warp + '"></div>',
            shadeClose: true,
            maxmin: true,
            fixed: false
        });
        renderJson(jsonContent, warp)
    }
</script>
</html>
`

// height: 50px;
