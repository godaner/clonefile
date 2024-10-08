package main

var templateBackupListHtml = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
    <meta http-equiv="Cache-Control" content="no-cache, no-store, must-revalidate">
    <meta http-equiv="Pragma" content="no-cache">
    <meta http-equiv="Expires" content="0">
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
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
			margin-top: 188px;
        }
		.popup {
		  display: none;
		  position: fixed;
		  z-index: 2000;
		  left: 0;
		  top: 0;
		  width: 100%;
		  height: 100%;
		  overflow: auto;
		  background-color: rgba(0, 0, 0, 0.4);
		}
	
		.popup-content {
		  background-color: #fefefe;
		  margin: 15% auto;
		  padding: 20px;
		  border: 1px solid #888;
		  width: 30%;
		}
    </style>
</head>
 
<body onload="restoreScrollPosition()">
	<div id="myPopup" class="popup">
		<div class="popup-content">
		  <h2>Waiting...</h2>
		  <p>Please wait while we process your request.</p>
		</div>
	</div>
    <div class="header-div">
        <p>{{.SfVersion}}</p>
        <p>{{.Title}}, Total: {{.TotalCnt}}, Used version: {{.Version}}</p>
		{{ if eq .NextState "Stop" }}
        	<p id='next_refresh_in'>Next refresh in: {{.Conf.Interval}}</p>
		{{end}}
		&#10;
        <form action="/cf_set?uuid={{UUID}}" method="post">
            Src:<input type="text" name='s' placeholder="Src dir" required value="{{.Conf.Src}}">
            Dst:<input type="text" name='d' placeholder="Dst dir" required value="{{.Conf.Dst}}">
            Interval:<input type="number" name='i' placeholder="Interval" required value="{{.Conf.Interval}}">
            MaxCount:<input type="number" name='m' placeholder="Max count" required value="{{.Conf.MaxCount}}">
            Prefix:<input type="text" name='p' placeholder="Prefix" required value="{{.Conf.Prefix}}">
            Exclude:<input type="text" name='e' placeholder="Exclude file, split by ," required value="{{.Conf.Exclude}}">
            <button type="submit" class="popup-trigger">Save</button>
            <button style="{{StateStyle .NextState}}" type="button" id="start-btn" class="popup-trigger">{{.NextState}}</button>
            <button type="button" class="popup-trigger" data-href="/bk_list?{{UUID}}">Refresh</button>
            <button type="button" class="popup-trigger" data-href="/clone?{{UUID}}">Clone</button>
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
            <th>Browser</th>
          </tr>
          {{ range $index, $Row := .Rows }}
          <tr style="{{ Style $.Version $Row }}">
            {{- range $Row }}
            <td>{{.}}</td>
            {{- end}}
            <td><a style="{{ Style $.Version $Row }}" class="popup-trigger" href="#" data-href="/bk_use/{{ index $Row 1 }}?{{UUID}}">Use it</a></td>
            <td><a style="{{ Style $.Version $Row }}" class="popup-trigger" href="#" data-href="/bk_delete/{{ index $Row 1 }}?uuid={{UUID}}">Delete it</a></td>
            <td><a style="{{ Style $.Version $Row }}" class="popup-trigger" href="#" data-href="/browser_file/{{ index $Row 1 }}?uuid={{UUID}}">Browser it</a></td>
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
        startBtn.addEventListener('click', function() {
            const buttonText = this.textContent;
            if (buttonText === 'Start') {
                window.location.href = '/start?{{UUID}}';
            } else if (buttonText === 'Stop') {
                window.location.href = '/stop?{{UUID}}';
            }
        });


        function updateCountdown() {
            let nextRefreshInSeconds = {{.Conf.Interval}};
            const countdownDisplay = document.getElementById('next_refresh_in');
            const countdownInterval = setInterval(() => {
				if (nextRefreshInSeconds>0){
					nextRefreshInSeconds--;
				}
                if (nextRefreshInSeconds === 0) {
                    clearInterval(countdownInterval);
					window.location.href = '/bk_list?{{UUID}}';
					return
                }
                countdownDisplay.textContent = "Next refresh in: "+nextRefreshInSeconds
            }, 1000);
		}
		if ({{.NextState}} === "Stop"){
			updateCountdown();
		}
		const popupTriggers = document.querySelectorAll('.popup-trigger');
		const popupClose = document.getElementById('popup-close');
		const popup = document.getElementById('myPopup');
	
		popupTriggers.forEach(trigger => {
		  trigger.addEventListener('click', (event) => {
			//event.preventDefault();
			const href = event.target.getAttribute('data-href');
			if (href) {
			  // Simulate a delayed action
			  //setTimeout(() => {
				window.location.href = href;
			  //}, 2000);
			}
			popup.style.display = 'block';
		  });
		});
	
		popupClose.addEventListener('click', () => {
		  popup.style.display = 'none';
		});
	
		window.addEventListener('click', (event) => {
		  if (event.target == popup) {
			popup.style.display = 'none';
		  }
		});
    </script>
</body>
</html>
`
