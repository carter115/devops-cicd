package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"lyyops-cicd/dao"
	"lyyops-cicd/pkg/common"
	"lyyops-cicd/pkg/log"
	"regexp"
	"strings"
)

// 用于手动操作argocd的对象

type DeploymentController struct{}

func DeploymentControllerGroupRegistry(group *gin.RouterGroup) {
	controller := DeploymentController{}
	group.GET(":id", controller.Get)
	group.GET("/status/:id", controller.GetStatus)
	//group.GET("/list", controller.List)
	group.POST("/create", controller.Create)
	//group.POST("/create_yaml", controller.CreateYaml)
	group.POST("/delete/:id", controller.Delete)
}

// Get DeploymentController godoc
// @Summary 获取Deployment Yaml
// @Tags 自动部署CD管理
// @Accept json
// @Produce json
// @Param id path string true "Deployment ID"
// @Success 200 {string} string ""
// @Router /deployment/{id} [get]
func (a *DeploymentController) Get(c *gin.Context) {
	var (
		code      common.StatusCode
		yamlValue string
		err       error
	)
	yamlValue, err = dao.GetDeploymentYaml(c, c.Param("id"))
	if err != nil {
		code = common.GetArgocdApplicationFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}
	c.JSON(200, common.SuccessResponse(c, yamlValue))
	return
Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err))
}

// GetStatus DeploymentController godoc
// @Summary 获取Deployment状态
// @Tags 自动部署CD管理
// @Accept json
// @Produce json
// @Param id path string true "Deployment ID"
// @Success 200 {string} string ""
// @Router /deployment/status/{id} [get]
func (a *DeploymentController) GetStatus(c *gin.Context) {
	var (
		code   common.StatusCode
		status string
		err    error
	)
	status, err = dao.GetDeploymentStatus(c, c.Param("id"))
	if err != nil {
		code = common.GetArgocdApplicationStatusFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}

	c.JSON(200, common.SuccessResponse(c, status))
	return
Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err))
}

// Create DeploymentController godoc
// @Summary 创建Deployment
// @Tags 自动部署CD管理
// @Accept json
// @Produce json
// @Param id query string true "argocd application name"
// @Param repo_path query string true "git repo path"
// @Param namespace query string true "app namespace"
// @Param value_files query string true "argocd: helm values files(ex: a.yaml,b.yaml)"
// @Success 200 {string} string ""
// @Router /deployment/create [post]
func (d *DeploymentController) Create(c *gin.Context) {
	var (
		code   = common.Success
		deploy = dao.Deployment{Ctx: c}
		err    error
	)

	deploy.Id = c.Query("id")
	deploy.RepoPath = c.Query("repo_path")
	deploy.Namespace = c.Query("namespace")
	deploy.ValueFiles = splitValueFilename(c.Query("value_files"))

	log.Debugf("deployment: %+v", deploy)

	if err = deploy.CreateHelm(); err != nil {
		code = common.CreateArgocdApplicationFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}
	c.JSON(200, common.SuccessResponse(c, ""))
	return
Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err))
}

// CreateYaml DeploymentController godoc
// @Summary 创建DeploymentFromYaml
// @Tags 自动部署CD管理
// @Accept json
// @Produce json
// @Param content body string true "内容"
//@Success 200 {string} string ""
// @Router /deployment/create_yaml [post]
func (d *DeploymentController) CreateYaml(c *gin.Context) {
	c.JSON(200, common.SuccessResponse(c, "TODO"))
}

// Delete DeploymentController godoc
// @Summary 删除Deployment
// @Tags 自动部署CD管理
// @Accept json
// @Produce json
// @Param id path string true "Deployment ID"
// @Success 200 {string} string ""
// @Router /deployment/delete/{id} [post]
func (d *DeploymentController) Delete(c *gin.Context) {
	var (
		code   common.StatusCode
		deploy = dao.Deployment{Ctx: c}
		err    error
	)
	deploy.Id = c.Param("id")
	if err = deploy.Delete(); err != nil {
		code = common.DeleteArgocdApplicationFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}
	c.JSON(200, common.SuccessResponse(c, ""))
	return
Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err))
}

func splitValueFilename(str string) []string {
	var (
		files = []string{}
		sep   = ","
	)
	reg := regexp.MustCompile("\\s+")
	notEmpty := reg.ReplaceAllString(str, sep)
	for _, v := range strings.Split(notEmpty, sep) {
		if v != "" {
			files = append(files, v)
		}
	}
	return files
}
