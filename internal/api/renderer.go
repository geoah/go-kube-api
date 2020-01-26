package api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

// renderer chooses the correct renderer based on the
// contet-type header
func renderer(c *gin.Context, data interface{}) render.Render {
	contentType := strings.ToLower(c.ContentType())
	if strings.Contains(contentType, "yaml") {
		return &render.YAML{
			Data: data,
		}
	}

	return &render.JSON{
		Data: data,
	}
}
