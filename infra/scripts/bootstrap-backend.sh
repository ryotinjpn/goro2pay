#!/usr/bin/env bash
# Terraform S3 backend (use_lockfile = true) のための tfstate bucket を作成する。
# 1 度だけ実行 (Q-I6 改訂版: DynamoDB 不要)。

set -euo pipefail

REGION="${AWS_REGION:-ap-northeast-1}"
ENV="${ENV:-dev}"
BUCKET="gp-tfstate-${ENV}"

echo "Creating S3 tfstate bucket: ${BUCKET} (region=${REGION})"

# Bucket 作成 (既に存在すれば skip)
if aws s3api head-bucket --bucket "${BUCKET}" 2>/dev/null; then
  echo "Bucket ${BUCKET} already exists, skipping create."
else
  aws s3api create-bucket \
    --bucket "${BUCKET}" \
    --region "${REGION}" \
    --create-bucket-configuration LocationConstraint="${REGION}"
fi

echo "Enabling versioning..."
aws s3api put-bucket-versioning \
  --bucket "${BUCKET}" \
  --versioning-configuration Status=Enabled

echo "Enabling server-side encryption (AES256)..."
aws s3api put-bucket-encryption \
  --bucket "${BUCKET}" \
  --server-side-encryption-configuration \
  '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'

echo "Blocking all public access..."
aws s3api put-public-access-block \
  --bucket "${BUCKET}" \
  --public-access-block-configuration \
  BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true

echo "Done. Bucket ${BUCKET} is ready as Terraform S3 backend (use_lockfile = true)."
