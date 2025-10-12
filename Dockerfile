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

# 添加必要的包
RUN apk add --no-cache sqlite-libs tzdata shadow

# 设置中国时区（Asia/Shanghai）
ENV TZ=Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 设置PUID和PGID环境变量，默认为1000
ENV PUID=1000
ENV PGID=1000

# 创建应用程序用户
RUN addgroup -g $PGID minecraft && \
    adduser -D -u $PUID -G minecraft minecraft

# 设置工作目录
WORKDIR /app

# 创建数据目录并设置权限
RUN mkdir -p /app/data && chown -R minecraft:minecraft /app/data

# 从构建环境复制构建好的应用程序
COPY --from=builder /app/minecraft-exchange .

# 只复制必要的静态文件和模板，避免复制过多文件
COPY --chown=minecraft:minecraft static/ static/
COPY --chown=minecraft:minecraft templates/ templates/

# 复制初始数据库文件（如果存在），使用RUN命令和shell语法处理可能不存在的情况
RUN --mount=type=bind,source=.,target=/source \
    if [ -f /source/minecraft_exchange.db ]; then \
        cp -a /source/minecraft_exchange.db* /app/data/ && \
        chown -R minecraft:minecraft /app/data/*.db*; \
    fi

# 创建启动脚本处理PUID/PGID
RUN echo '#!/bin/sh' > /app/entrypoint.sh && \
    echo 'if [ $(id -u) = 0 ]; then' >> /app/entrypoint.sh && \
    echo '    # 调整用户ID和组ID' >> /app/entrypoint.sh && \
    echo '    groupmod -o -g "$PGID" minecraft' >> /app/entrypoint.sh && \
    echo '    usermod -o -u "$PUID" minecraft' >> /app/entrypoint.sh && \
    echo '    # 确保数据目录权限正确' >> /app/entrypoint.sh && \
    echo '    chown -R minecraft:minecraft /app/data' >> /app/entrypoint.sh && \
    echo '    chown -R minecraft:minecraft /app/static' >> /app/entrypoint.sh && \
    echo '    chown -R minecraft:minecraft /app/templates' >> /app/entrypoint.sh && \
    echo '    exec su-exec minecraft:minecraft "$@"' >> /app/entrypoint.sh && \
    echo 'else' >> /app/entrypoint.sh && \
    echo '    exec "$@"' >> /app/entrypoint.sh && \
    echo 'fi' >> /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

# 安装su-exec用于更安全地切换用户
RUN apk add --no-cache su-exec

# 设置环境变量，配置数据库路径为数据目录下
ENV DATABASE_PATH=/app/data/minecraft_exchange.db

# 暴露应用程序端口
EXPOSE 8080

# 声明持久化数据卷，用于保存SQLite数据库文件
VOLUME ["/app/data"]

# 设置入口点
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["./minecraft-exchange"]