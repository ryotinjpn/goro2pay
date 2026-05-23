// Package handlers は HTTP handler を集約する。
package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Health は GET /health の handler。認証不要。
// API Lambda の生存確認、ロードバランサのヘルスチェック用途。
//
// レスポンスに build_sha (環境変数 BUILD_SHA) を含める。
// CodeBuild の post_build で `--environment Variables="{BUILD_SHA=$CODEBUILD_RESOLVED_SOURCE_VERSION}"`
// として Lambda env に注入する想定 (将来対応)。env 未設定のときは "unknown"。
func Health(c *gin.Context) {
	buildSHA := os.Getenv("BUILD_SHA")
	if buildSHA == "" {
		buildSHA = "unknown"
	}
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"build_sha": buildSHA,
	})
}
