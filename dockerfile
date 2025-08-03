# 使用官方 Golang 镜像作为构建环境
FROM golang:1.23-alpine AS builder

# 设置 Go 代理，使用国内镜像加速下载
ENV GOPROXY=https://goproxy.cn,direct

# 设置工作目录
WORKDIR /app

# 复制 go mod 和 sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build -o ai-backend .

# 使用小型基础镜像运行应用
FROM alpine:latest

# 安装 ca-certificates 以支持 HTTPS 请求
RUN apk --no-cache add ca-certificates

# 创建工作目录
WORKDIR /root/

# 从构建器复制二进制文件和配置文件
COPY --from=builder /app/ai-backend .

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./ai-backend"]
