package main

var templateBackupListHtml = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
	<meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate">
	<meta http-equiv="Pragma" content="no-cache">
	<meta http-equiv="Expires" content="0">
	<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
	<meta http-equiv="refresh" content="5">
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
		  	margin-top: 80px;
		}
	</style>
</head>
 
<body onload="restoreScrollPosition()">
	
	<div class="header-div">
		<p>{{.Title}}, Total: {{.TotalCnt}}, Version: {{.Version}},RefreshTime: {{.RefreshTime}}</p>
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
			<td><a style="{{ Style $.Version $Row }}" href="/bk_delete/{{ index $Row 1 }}?{{UUID}}">Delete it</a></td>
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
    </script>
</body>
</html>
`
