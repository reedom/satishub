package api

import (
	"github.com/gin-gonic/gin"
	"github.com/reedom/satishub/pkg/satis"
)

func (s Server) handleGitlab(ctx *gin.Context) {
	var req struct {
		Repository struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		}
	}
	err := ctx.BindJSON(&req)
	if err != nil {
		s.log.Println("GitLab WebHook content is broken?")
		ctx.JSON(200, "OK")
		return
	}

	if req.Repository.URL == "" {
		if s.debug {
			s.log.Println("repository URL not found in request payload")
		}
		ctx.JSON(200, "OK")
		return
	}

	pkg := satis.PackageInfo{
		Name:    ctx.Query("name"),
		Version: ctx.Query("version"),
		URL:     req.Repository.URL,
		Type:    "vcs",
	}

	go func() {
		if s.debug {
			s.log.Printf("process repository %v(%v)", pkg.Name, pkg.URL)
		}
		<-s.service.UpdatePackage(pkg)
	}()

	ctx.JSON(200, "OK")
}
