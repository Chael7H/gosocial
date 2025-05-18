
# Postman测试指南 - 发送消息功能

## 测试准备
1. 获取有效的JWT token(通过登录接口)
2. 准备两个测试用户账号(互为好友)

## 测试步骤

### 1. 配置请求
- **方法**: POST
- **URL**: `http://your-domain.com/api/v1/messages`

#### 设置Headers步骤：
1. 在Postman中打开请求的"Headers"选项卡
2. 在Key输入框中输入: `Content-Type`
3. 在Value输入框中输入: `application/json`
4. 点击"Add"按钮添加该Header

#### 设置Authorization步骤：
1. 在Key输入框中输入: `Authorization`
2. 在Value输入框中输入: `Bearer <your-jwt-token>`
   - 将`<your-jwt-token>`替换为实际获取的JWT令牌
3. 点击"Add"按钮添加该Header

#### 可视化操作指引：
1. 点击Postman界面上的"Headers"选项卡
2. 你会看到两个列：Key和Value
3. 按照上述步骤添加两个必需的Header
4. 确保Header添加后左侧的复选框被勾选

### 2. 请求体示例
```json
{
  "to": 123456,
  "content": "这是一条测试消息",
  "type": 1
}
```
字段说明：
- `to`: 接收者用户ID
- `content`: 消息内容
- `type`: 消息类型(1-文本, 2-图片, 3-文件)

### 3. 预期成功响应(200)
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message_id": "abc123",
    "timestamp": 1672531200,
    "receiver_id": 123456,
    "status": "sent"
  }
}
```

### 4. 测试用例

#### 用例1: 发送文本消息
- 请求体:
```json
{
  "to": 123456,
  "content": "你好！",
  "type": 1
}
```
- 验证点:
  - 返回状态码200
  - 响应包含message_id和timestamp

#### 用例2: 发送给非好友用户
- 请求体:
```json
{
  "to": 999999,
  "content": "非法消息",
  "type": 1
}
```
- 预期响应(403):
```json
{
  "code": 403,
  "message": "不是好友关系，不能发送消息"
}
```

#### 用例3: 无效token测试
- 不设置Authorization头或使用无效token
- 预期响应(401):
```json
{
  "code": 401,
  "message": "未授权"
}
```

## 验证实时接收
1. 在另一个Postman窗口创建WebSocket连接:
   ```
   ws://your-domain.com/ws?token=<receiver-token>
   ```
2. 发送消息后，在WebSocket连接中应实时收到新消息通知
