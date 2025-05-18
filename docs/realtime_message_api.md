
# 实时消息推送API文档

## 功能概述
当用户收到新消息时，系统会通过Redis Pub/Sub机制实时推送消息到前端。前端需要订阅用户专属频道接收实时消息。

## 订阅方式
推荐两种实现方式：

### 1. WebSocket连接
```javascript
// 建立WebSocket连接
const socket = new WebSocket(`wss://your-api-domain.com/ws?token=${userToken}`);

// 监听消息
socket.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('收到新消息:', message);

  // 更新未读计数
  updateUnreadCount(message.from);

  // 显示新消息
  displayNewMessage(message);
};

// 错误处理
socket.onerror = (error) => {
  console.error('WebSocket错误:', error);
};
```

### 2. 长轮询(兼容性更好)
```javascript
function pollMessages() {
  fetch('/api/messages/poll')
    .then(response => response.json())
    .then(messages => {
      messages.forEach(processNewMessage);
      pollMessages(); // 继续轮询
    })
    .catch(error => {
      console.error('轮询错误:', error);
      setTimeout(pollMessages, 5000); // 5秒后重试
    });
}

// 初始化轮询
pollMessages();
```

## 消息格式
```json
{
  "id": "消息ID",
  "from": 发送者ID,
  "to": 接收者ID,
  "content": "消息内容",
  "type": 1, // 1-文本 2-图片 3-文件
  "created_at": "2023-01-01T00:00:00Z",
  "file_url": "文件URL(如果是文件消息)"
}
```

## 未读计数更新
当收到新消息时，前端应:
1. 更新对应联系人的未读计数
2. 播放新消息提示音(可选)
3. 显示桌面通知(可选)

## 消息确认
建议前端在成功显示消息后发送确认:
```javascript
fetch('/api/messages/ack', {
  method: 'POST',
  body: JSON.stringify({message_id: message.id}),
  headers: {'Content-Type': 'application/json'}
});
```

## 错误处理
- 网络中断时自动重连
- 消息去重处理
- 本地缓存未确认消息
