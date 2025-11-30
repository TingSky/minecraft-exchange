# 修复AJAX提交失败问题的实施计划

## 1. 问题分析

通过检查AJAX处理函数和后端响应逻辑，发现以下问题：

### 1.1 AJAX错误处理逻辑问题

在`ajaxFormSubmit`函数中，错误处理逻辑存在问题：

```javascript
.then(response => {
    // 检查响应状态
    if (!response.ok) {
        // 尝试解析错误响应
        return response.json().then(errData => {
            throw new Error(errData.message || `HTTP错误 ${response.status}`);
        }).catch(() => {
            throw new Error(`HTTP错误 ${response.status}`);
        });
    }
    // 尝试解析JSON响应，如果失败则返回空对象
    return response.json().catch(() => ({}));
})
```

当响应状态不是200 OK时，函数会抛出错误，然后进入catch块，显示"操作失败: " + error.message。

### 1.2 后端响应状态问题

后端处理函数在某些情况下返回非200 OK状态，比如：

- 登录失败时返回400 Bad Request
- 参数验证失败时返回400 Bad Request
- 服务器错误时返回500 Internal Server Error

### 1.3 响应处理逻辑问题

AJAX函数只检查`data.success`或`data.redirect`字段，而不检查`data.refresh`字段。

## 2. 实施步骤

### 2.1 修改AJAX错误处理逻辑

修改`ajaxFormSubmit`函数，让它无论响应状态如何，都尝试解析JSON响应，然后根据Success字段来判断是否成功：

```javascript
.then(response => {
    // 无论响应状态如何，都尝试解析JSON响应
    return response.json().catch(() => ({
        success: false,
        message: `HTTP错误 ${response.status}`
    }));
})
```

### 2.2 修改成功判断逻辑

修改成功判断逻辑，检查`data.success`、`data.redirect`或`data.refresh`字段：

```javascript
if (data.success || data.redirect || data.refresh) {
    if (successCallback) successCallback(data);
    // 如果有redirect字段，重定向到指定URL
    if (data.redirect) {
        window.location.href = data.redirect;
    }
} else {
    if (errorCallback) errorCallback(data);
}
```

### 2.3 测试所有表单功能

修改完成后，测试所有表单功能，确保它们都能正确处理AJAX提交：

1. **login.html**：管理员登录表单（包括成功和失败情况）
2. **tasks.html**：领取任务、完成任务表单
3. **shop.html**：兑换物品表单
4. **admin.html**：创建任务模板、删除任务模板、刷新日常任务、验证任务、删除任务、创建物品、更新物品、删除物品、兑换奖励表单

## 3. 预期效果

所有AJAX提交都会正确处理，无论是成功还是失败情况：
- 成功情况：显示成功消息，可能刷新页面或重定向
- 失败情况：显示具体的错误消息，而不是通用的"操作失败"

## 4. 实施顺序

1. 修改`static/js/main.js`中的`ajaxFormSubmit`函数
2. 测试所有表单功能

这个修改使用了简单的方案，通过修改AJAX处理函数的错误处理逻辑，确保了所有AJAX提交都能正确处理，无论后端返回什么状态码。