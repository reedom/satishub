package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s Server) readConfig(ctx *gin.Context) {
	http.ServeFile(ctx.Writer, ctx.Request, s.service.ConfigPath())
}
