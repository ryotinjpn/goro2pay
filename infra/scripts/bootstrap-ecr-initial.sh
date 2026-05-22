#!/usr/bin/env bash
# CodePipeline が動作する前に Lambda が起動できるよう、
# ECR Repository に :bootstrap タグの初期 image を push する。
# 1 度だけ実行 (それ以降は CodeBuild が image を更新する)。

set -euo pipefail

REGION="${AWS_REGION:-ap-northeast-1}"
ENV="${ENV:-dev}"
REPO_NAME="gp-${ENV}-api-image"
ACCOUNT_ID="$(aws sts get-caller-identity --query Account --output text)"
ECR_URI="${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO_NAME}"

# このスクリプトはリポジトリルートから実行する想定
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

echo "Repo root: ${REPO_ROOT}"
echo "ECR URI:   ${ECR_URI}:bootstrap"

# 1. ECR にログイン
aws ecr get-login-password --region "${REGION}" | \
  docker login --username AWS --password-stdin "${ECR_URI%/*}"

# 2. Build (apps/api/Dockerfile, build context = apps/api/)
docker buildx create --use --name goro2pay-builder 2>/dev/null || docker buildx use goro2pay-builder

docker buildx build \
  --platform linux/arm64 \
  --load \
  -t "${ECR_URI}:bootstrap" \
  -f "${REPO_ROOT}/apps/api/Dockerfile" \
  "${REPO_ROOT}/apps/api/"

# 3. Push
docker push "${ECR_URI}:bootstrap"

echo "Done. Initial image pushed to ${ECR_URI}:bootstrap"
echo "API Lambda is now ready to be created with image_uri = ${ECR_URI}:bootstrap"
echo "Subsequent updates will be handled by CodeBuild via lambda update-function-code."
