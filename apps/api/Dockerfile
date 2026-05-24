# API Lambda 用 Dockerfile
# Lambda Web Adapter (LWA) + Go 静的バイナリの multi-stage build (linux/arm64)。
# Build context は apps/api/ 直下を想定 (CodeBuild の buildspec.yml と整合)。

# --- Stage 1: build Go binary ---
# Unit C で AWS SDK v2 (bedrockruntime / dynamodb) を導入。SDK 最新版が
# Go 1.24+ を必須とするため Builder を 1.24 に上げる。
# CodeBuild の未認証 Docker Hub pull が IP 共有でレート制限 (429) に到達しやすい
# ため、ECR Public の Docker Hub ミラー (公式同期) を経由する。タグは Docker Hub
# と同じ semver/tag が使える。
FROM public.ecr.aws/docker/library/golang:1.24-alpine AS builder
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

# LWA 起動時必須の環境変数 (アプリ全体で共通、Dockerfile 側で固定):
#   AWS_LWA_PORT          : LWA → 内部 HTTP server の listen port (Gin と一致)
#   PORT                  : Gin が listen する port (AWS_LWA_PORT と同じ値)
#   READINESS_CHECK_PATH  : LWA が起動完了判定で叩く path
#   AWS_LWA_INVOKE_MODE   : buffered (= デフォルト、Lambda の同期 invoke と等価)。
#                           response streaming を使う場合のみ "response_stream" にする
#
# 一方、Lambda env (実行時注入、Terraform 側で設定) で渡すべきもの:
#   COGNITO_USER_POOL_ID / COGNITO_APP_CLIENT_ID  : 環境ごとに異なる ID
#   LOG_LEVEL                                     : 運用調整したい
#   BUILD_SHA                                     : デプロイごとの commit sha
ENV AWS_LWA_PORT=8080
ENV PORT=8080
ENV READINESS_CHECK_PATH=/health
ENV AWS_LWA_INVOKE_MODE=buffered

# Gin バイナリを Lambda エントリ名 (bootstrap) としてコピー。
# provided.al2023 では LAMBDA_TASK_ROOT (= /var/task) 配下の bootstrap を実行する。
# /var/runtime/ は Lambda runtime 内部ディレクトリで、ユーザバイナリは置かない。
# 参考: https://docs.aws.amazon.com/lambda/latest/dg/runtimes-custom.html
COPY --from=builder /out/bootstrap ${LAMBDA_TASK_ROOT}/bootstrap

CMD ["bootstrap"]
