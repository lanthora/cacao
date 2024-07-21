package frontend

import (
	"embed"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
)

//go:embed dist/*
var staticFS embed.FS

var contentType = map[string]string{
	".html": "text/html; charset=UTF-8",
	".css":  "text/css; charset=UTF-8",
	".js":   "text/javascript; charset=UTF-8",
	".ico":  "image/x-icon",
	".svg":  "image/svg+xml",
}

func Static(c *gin.Context) {
	filePath := "dist" + c.Request.URL.String()
	if data, err := staticFS.ReadFile(filePath); err == nil {
		c.Header("Cache-Control", "public, max-age=604800")
		c.Data(http.StatusOK, contentType[path.Ext(filePath)], data)
		return
	}

	indexPath := "dist/index.html"
	if data, err := staticFS.ReadFile(indexPath); err == nil {
		c.Data(http.StatusOK, contentType[path.Ext(indexPath)], data)
		return
	}

	c.Status(http.StatusNotFound)
}
