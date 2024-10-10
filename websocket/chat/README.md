这部分代码是`gorilla/websocket`的官方样例， 在`websocket/chat/`目录下执`go run .`就可以启动这个示例。

这个例子展示了一个基于websocket的聊天界面， 启动服务后主页加载`home.html`页面建立websocket连接， 通过send按钮发送请求到服务端处理消息。

用户访问web页面的时候就会建立websocket连接，通过回调函数处理连接发过来的消息和连接中断后的操作：

```
    conn = new WebSocket("ws://" + document.location.host + "/ws");
    conn.onclose = function (evt) {
        var item = document.createElement("div");
        item.innerHTML = "<b>Connection closed.</b>";
        appendLog(item);
    };
    conn.onmessage = function (evt) {
        var messages = evt.data.split('\n');
        for (var i = 0; i < messages.length; i++) {
            var item = document.createElement("div");
            item.innerText = messages[i];
            appendLog(item);
        }
    };
```

在服务端， 可以从`serveWs`函数看到通过upgrade将http请求升级成websocket连接并构建对应的client对象。 并将client对象加入到`hub`对象的控制中。 启动两个goroutine分别处理读消息和写消息。

`readPump` 中循环从连接中读消息， 并写入到broadcast channel中, hub会将broadcast的消息发送到关联的各个client的send channel中进行分发。

`writePump` 中启动一个定时器，如果时间内没有消息发送过来，则发送一条ping message来保持连接。 否则从client的send channel中获取message并通过连接发送到端。



