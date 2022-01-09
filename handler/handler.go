package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"lyyops-cicd/config"
	"lyyops-cicd/controller"
	_ "lyyops-cicd/docs"
	"lyyops-cicd/pkg/common"
	"lyyops-cicd/pkg/log"
)

func InitHandler() *gin.Engine {
	engine := gin.New()
	gin.SetMode(gin.ReleaseMode)

	// 加载中间件
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// -------------------非业务路由-------------------
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	engine.GET("/", HomeHandler)
	engine.GET("/argo/token", ArgoTokenHandler)

	// applicationGroup
	applicationGroup := engine.Group("/application")
	controller.ApplicationControllerGroupRegistry(applicationGroup)

	// templateGroup
	templateGroup := engine.Group("/template")
	controller.TemplateControllerGroupRegistry(templateGroup)

	// jobGroup
	jobGroup := engine.Group("/job")
	controller.JobControllerGroupRegistry(jobGroup)

	// logsGroup
	logsGroup := engine.Group("/logs")
	controller.LogsControllerGroupRegistry(logsGroup)

	// deploymentGroup
	deploymentGroup := engine.Group("/deployment")
	controller.DeploymentControllerGroupRegistry(deploymentGroup)

	return engine
}

func HomeHandler(c *gin.Context) {
	log.Info("home page")
	content := `home page <a href="./swagger/index.html">[swagger]</a> <a href="/argo/token">[argo token]</a>`
	c.Header("Content-Type", "text/html")
	c.String(200, content)
}

// TODO 正式环境应该删除该接口
func ArgoTokenHandler(c *gin.Context) {
	c.JSON(200, common.NewResponse(c, common.Success, config.Config.Argo.Token))
}
