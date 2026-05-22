# API Lambda 用 Dockerfile
# Lambda Web Adapter (LWA) + Go 静的バイナリの multi-stage build (linux/arm64)。
# Build context は apps/api/ 直下を想定 (CodeBuild の buildspec.yml と整合)。

# --- Stage 1: build Go binary ---
FROM golang:1.22-alpine AS builder
WORKDIR /src

# モジュールキャッシュ用に先にコピー
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# arm64 / Linux / 静的バイナリ
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
    go build -trimpath -ldflags="-s -w" -o /out/bootstrap .

# --- Stage 2: runtime ---
# AWS Lambda Web Adapter (LWA) v1.0.0 GA を /opt/extensions に同梱。
# 公式 README 推奨タグ: public.ecr.aws/awsguru/aws-lambda-adapter:1.0.0
# (https://github.com/awslabs/aws-lambda-web-adapter/blob/main/README.md)
FROM public.ecr.aws/lambda/provided:al2023-arm64

# LWA を /opt/extensions/lambda-adapter に配置
COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:1.0.0 /lambda-adapter /opt/extensions/lambda-adapter

# Gin の listen ポートを LWA のデフォルトに合わせる
ENV AWS_LWA_PORT=8080
ENV PORT=8080
ENV READINESS_CHECK_PATH=/health

# Gin バイナリを Lambda エントリ名 (bootstrap) としてコピー。
# provided.al2023 では LAMBDA_TASK_ROOT (= /var/task) 配下の bootstrap を実行する。
# /var/runtime/ は Lambda runtime 内部ディレクトリで、ユーザバイナリは置かない。
# 参考: https://docs.aws.amazon.com/lambda/latest/dg/runtimes-custom.html
COPY --from=builder /out/bootstrap ${LAMBDA_TASK_ROOT}/bootstrap

CMD ["bootstrap"]
