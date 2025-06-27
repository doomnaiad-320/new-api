FROM node:18-alpine AS builder

# 设置工作目录
WORKDIR /build

# 复制package.json并安装依赖
COPY web/package.json ./
RUN npm install --legacy-peer-deps

# 复制源代码和版本文件
COPY ./web .
COPY ./VERSION .

# 设置环境变量并构建
ENV DISABLE_ESLINT_PLUGIN=true
ENV NODE_OPTIONS="--max-old-space-size=4096"
RUN VITE_REACT_APP_VERSION=$(cat VERSION) npm run build

FROM golang:alpine AS builder2

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build

ADD go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=builder /build/dist ./web/dist
RUN go build -ldflags "-s -w -X 'one-api/common.Version=$(cat VERSION)'" -o one-api

FROM alpine

RUN apk update \
    && apk upgrade \
    && apk add --no-cache ca-certificates tzdata ffmpeg \
    && update-ca-certificates

COPY --from=builder2 /build/one-api /
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/one-api"]
