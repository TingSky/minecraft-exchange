# 使用官方Go镜像作为构建环境
FROM golang:1.22-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制项目源代码
COPY . .

# 构建应用程序
RUN CGO_ENABLED=1 GOOS=linux go build -o minecraft-exchange main.go

# 使用轻量级Alpine镜像作为运行环境
FROM alpine:3.19

# 添加SQLite依赖
RUN apk add --no-cache sqlite-libs

# 设置工作目录
WORKDIR /app

# 从构建环境复制构建好的应用程序
COPY --from=builder /app/minecraft-exchange .

# 复制静态文件和模板
COPY static/ static/
COPY templates/ templates/

# 复制初始数据库文件（如果存在）
COPY --chown=1000:1000 minecraft_exchange.db* ./ 2>/dev/null || true

# 创建非root用户运行应用
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 更改文件所有权
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露应用程序端口
EXPOSE 8080

# 设置环境变量，配置数据库路径
ENV DATABASE_PATH=/app/minecraft_exchange.db

# 设置入口点
CMD ["./minecraft-exchange"]