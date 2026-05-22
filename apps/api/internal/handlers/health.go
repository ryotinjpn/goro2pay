// Package handlers は HTTP handler を集約する。
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health は GET /health の handler。認証不要。
// API Lambda の生存確認、ロードバランサのヘルスチェック用途。
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
