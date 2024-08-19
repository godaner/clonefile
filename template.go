package main

var templateBackupListHtml = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
	<meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate">
	<meta http-equiv="Pragma" content="no-cache">
	<meta http-equiv="Expires" content="0">
	<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
	//<meta http-equiv="refresh" content="60">
	<title>{{.Title}}</title>
	<style>
        body {
            font-family: Arial, sans-serif;
            font-size: 14px;
            line-height: 1.4;
            margin: 0;
            padding: 0;
        }
        table {
			width: 100%;
			margin-left: calc(100vw - 100%);
            border-collapse: collapse;
            border: 1px solid #ddd;
        }
        
        th, td {
            padding: 8px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        
        th {
            background-color: #f2f2f2;
        }
        
        td {
            max-width: 200px; /* 限制单元格最大宽度 */
            word-wrap: break-word; /* 自动换行 */
            overflow-wrap: break-word; /* 同上,不同浏览器兼容 */
        }
		
		.header-div {
		  position: fixed;
		  top: 0;
		  left: 0;
		  width: 100%;
		  height: auto;
		  background-color: #333;
		  color: #fff;
		  display: flex;
		  flex-direction: column;
		  align-items: flex-start;
		  justify-content: center;
		  padding: 10px 20px;
		  z-index: 1000;
		}
	
		.header-div p {
		  margin: 5px 0;
		  line-height: 1.5;
		  text-align: left;
		}
	
		.content {
		  	margin-top: 130px;
		}
	</style>
</head>
 
<body onload="restoreScrollPosition()">
	
	<div class="header-div">
		<p>{{.SfVersion}}</p>
		<p>{{.Title}}, Total: {{.TotalCnt}}, Version: {{.Version}},RefreshTime: {{.RefreshTime}}</p>
		<form action="/cf_set?uuid={{UUID}}" method="post">
			<input type="text" name='s' placeholder="Src dir" required value="{{.Conf.S}}">
			<input type="text" name='d' placeholder="Dst dir" required value="{{.Conf.D}}">
			<input type="number" name='i' placeholder="Interval" required value="{{.Conf.I}}">
			<input type="number" name='m' placeholder="Max count" required value="{{.Conf.M}}">
			<input type="text" name='p' placeholder="Prefix" required value="{{.Conf.P}}">
			<input type="text" name='e' placeholder="Exclude file, split by ," required value="{{.Conf.E}}">
			<button type="submit">Submit</button>
			<button style="{{StateStyle .State}}" type="button" id="start-btn">{{.State}}</button>
		</form>
		<p id="msg"></p>
	  </div>
	
	  <div class="content">
		<table>
		  <tr>
			{{- range .Header }}
			<th>{{.}}</th>
			{{- end}}
			<th>Use</th>
			<th>Delete</th>
		  </tr>
		  {{ range $index, $Row := .Rows }}
		  <tr style="{{ Style $.Version $Row }}">
			{{- range $Row }}
			<td>{{.}}</td>
			{{- end}}
			<td><a style="{{ Style $.Version $Row }}" href="/bk_use/{{ index $Row 1 }}?{{UUID}}">Use it</a></td>
			<td><a style="{{ Style $.Version $Row }}" href="/bk_delete/{{ index $Row 1 }}?uuid={{UUID}}">Delete it</a></td>
		  </tr>
		  {{ end }}
		</table>
	  </div>
	<script>
        const paramsStr = window.location.search
		const params = new URLSearchParams(paramsStr)
		var errMsg=params.get('errMsg')
		console.log('errMsg:', errMsg);
		var msgObj = document.getElementById("msg");
		if (errMsg!=null&&errMsg=="Success"){
    		msgObj.innerText= "Success"; 
			msgObj.style.color = "green";
		}
		if (errMsg!=null&&errMsg!="Success"){
    		msgObj.innerText= "Error: "+errMsg; 
			msgObj.style.color = "red";
		}
		function saveScrollPosition() {
		  localStorage.setItem('scrollPosition', window.pageYOffset);
		}
	
		function restoreScrollPosition() {
		  const scrollPosition = localStorage.getItem('scrollPosition');
		  if (scrollPosition) {
			window.scrollTo(0, scrollPosition);
		  }
		}
		addEventListener("wheel", (event) => {
			saveScrollPosition()
		});


		// 获取 "Start" 按钮元素
		const startBtn = document.getElementById('start-btn');
	
		// 为按钮添加点击事件监听器
		startBtn.addEventListener('click', function() {
			// 执行您的自定义操作
			// 例如,跳转到另一个页面
			// 获取按钮的文本内容
			const buttonText = this.textContent;
	
			// 根据按钮的文本内容做不同的操作
			if (buttonText === 'Start') {
				window.location.href = '/start?{{UUID}}';
			} else if (buttonText === 'Stop') {
				window.location.href = '/stop?{{UUID}}';
			}
		});
    </script>
</body>
</html>
`
