.PHONY: build clean

# Lambda 用 bootstrap バイナリ生成 (provided.al2023 ランタイム + arm64)。
# terraform apply 前にローカルで本ターゲットを実行し、bootstrap バイナリを
# apps/api/cmd/scheduler/ 配下に生成しておく必要がある (Q-I2=A、Q-I3=A)。
#
# go.mod は apps/api と共有 (internal package 利用のため)。
build:
	cd ../.. && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o cmd/scheduler/bootstrap ./cmd/scheduler

clean:
	rm -f bootstrap
