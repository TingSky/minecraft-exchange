# 第一阶段：使用Go镜像进行构建
FROM golang:1.22-alpine AS builder

# 在Alpine镜像中安装C编译器和必要的构建工具以支持CGO
RUN apk add --no-cache gcc musl-dev

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制项目源代码
COPY . .

# 优化构建参数：禁用调试信息以减小二进制文件大小
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o minecraft-exchange main.go

# 第二阶段：使用scratch作为基础镜像，这是最小的可能镜像
FROM alpine:3.19

# 使用--virtual创建一个虚拟包组，方便后续清理
RUN apk add --no-cache --virtual .build-deps sqlite-libs

# 设置工作目录
WORKDIR /app

# 从构建环境复制构建好的应用程序
COPY --from=builder /app/minecraft-exchange .

# 只复制必要的静态文件和模板，避免复制过多文件
COPY --chown=65534:65534 static/ static/
COPY --chown=65534:65534 templates/ templates/

# 复制初始数据库文件（如果存在），使用RUN命令和shell语法处理可能不存在的情况
    RUN --mount=type=bind,source=.,target=/source \
        if [ -f /source/minecraft_exchange.db ]; then \
            cp -a /source/minecraft_exchange.db* ./ && \
            chown -R 65534:65534 *.db*; \
        fi

# 切换到nobody用户（已存在的非root用户），避免创建新用户
USER nobody

# 暴露应用程序端口
EXPOSE 8080

# 设置环境变量，配置数据库路径
ENV DATABASE_PATH=/app/minecraft_exchange.db

# 设置入口点
CMD ["./minecraft-exchange"]