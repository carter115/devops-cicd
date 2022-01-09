package controller

import (
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"lyyops-cicd/dao"
	"lyyops-cicd/pkg/common"
	"lyyops-cicd/pkg/log"
	"strconv"
	"strings"
)

type ApplicationController struct{}

func ApplicationControllerGroupRegistry(group *gin.RouterGroup) {
	controller := ApplicationController{}
	group.GET(":id", controller.Get)
	group.GET("/list", controller.List)
	group.POST("/create", controller.Create)
	group.POST("/delete/:id", controller.Delete)
}

// Get ApplicationController godoc
// @Summary 获取Application
// @Tags 应用管理
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Success 200 {string} string ""
// @Router /application/{id} [get]
func (a *ApplicationController) Get(c *gin.Context) {
	var (
		code common.StatusCode
		app  *dao.Application
		err  error
	)
	// 获取applicaion的yaml内容
	if app, err = dao.GetApplication(c, c.Param("id")); err != nil {
		code = common.ApplicationNotFound
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}
	c.JSON(200, common.SuccessResponse(c, string(app.Content)))
	return
Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err))
}

// List ApplicationController godoc
// @Summary 获取Application列表
// @Tags 应用管理
// @Accept json
// @Produce json
// @Success 200 {string} string ""
// @Router /application/list [get]
func (a *ApplicationController) List(c *gin.Context) {
	var (
		code common.StatusCode
		apps []string
		err  error
	)
	apps, err = dao.ListApplication(c)
	if err != nil {
		code = common.ListApplicationFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}
	c.JSON(200, common.SuccessResponse(c, apps))
	return
Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err))
}

// Create ApplicationController godoc
// @Summary 创建Application
// @Tags 应用管理
// @Accept json
// @Produce json
// @Param id query string true "application name"
// @Param content body string true "yaml内容"
// @Param create_argocd query boolean true "是否创建ArgoCD对象" Enums(true,false)
// @Param repo_path query string false "argocd: git repo path"
// @Param namespace query string false "argocd: namespace"
// @Param value_files query string false "argocd: helm values files(ex: a.yaml,b.yaml)"
// @Success 200 {string} string ""
// @Router /application/create [post]
func (a *ApplicationController) Create(c *gin.Context) {
	var (
		code          = common.Success
		app           = dao.Application{}
		deploy        = dao.Deployment{}
		create_argocd bool
		wf            *wfv1.Workflow
		err           error
	)

	create_argocd, err = strconv.ParseBool(c.Query("create_argocd"))
	if err != nil {
		code = common.InvalidParam
		err = errors.Wrap(err, "create_argocd字段非法")
		goto Fail
	}

	app.Ctx = c
	app.Id = c.Query("id")
	if app.Content, err = c.GetRawData(); err != nil {
		code = common.InvalidParam
		err = errors.Wrap(err, "GetRawData")
		goto Fail
	}
	log.Debugf("creating application id: %s", app.Id)

	// 解析yaml文件
	if wf, err = dao.UnmarshalWorkflow(app.Content); err != nil {
		code = common.CreateApplicationFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}
	log.Debugf("workflow object: %+v", wf)

	if app.Id == "" || app.Id != strings.TrimRight(wf.GenerateName, "-") {
		code = common.InvalidParam
		err = errors.Wrapf(err, "application id %s, workflow GenerateName %s", app.Id, wf.GenerateName)
		goto Fail
	}

	// 生成Argocd application对象
	if create_argocd {
		deploy.Ctx = c
		deploy.Id = app.Id
		deploy.Namespace = c.Query("namespace")
		deploy.RepoPath = c.Query("repo_path")
		deploy.ValueFiles = splitValueFilename(c.Query("value_files"))
		log.Debugf("deployment: %+v", deploy)
		if deploy.Namespace == "" || deploy.RepoPath == "" || len(deploy.ValueFiles) == 0 {
			code = common.CreateArgocdApplicationFailed
			err = errors.Wrap(err, "argocd 参数不能为空")
			goto Fail
		}

		if err = deploy.CreateHelm(); err != nil {
			code = common.CreateArgocdApplicationFailed
			err = errors.Wrap(err, code.GetMsg())
			goto Fail
		}
		log.Infof("deployment created: %s", deploy.Id)
	}

	// 把yaml内容保存到DB
	if err = app.Save(); err != nil {
		code = common.SaveApplicationFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}

	log.Info("application created: %s", app.Id)
	c.JSON(200, common.SuccessResponse(c, app.Id))
	return

Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err))
}

// Delete ApplicationController godoc
// @Summary 删除Application
// @Tags 应用管理
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Param delete_argocd query boolean true "是否删除ArgoCD对象" Enums(true,false)
// @Success 200 {string} string ""
// @Router /application/delete/{id} [post]
func (a *ApplicationController) Delete(c *gin.Context) {
	var (
		code          = common.Success
		app           = dao.Application{Ctx: c}
		deploy        = dao.Deployment{Ctx: c}
		delete_argocd bool
		err           error
	)

	delete_argocd, err = strconv.ParseBool(c.Query("delete_argocd"))
	if err != nil {
		code = common.InvalidParam
		err = errors.Wrap(err, "delete_argocd字段非法")
		goto Fail
	}

	app.Id = c.Param("id")
	deploy.Id = app.Id
	if app.Id == "" {
		code = common.InvalidParam
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}
	log.Debugf("application: %+v", app)

	if err = app.Delete(); err != nil {
		code = common.DeleteApplicationFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}

	if delete_argocd {
		if err = deploy.Delete(); err != nil {
			code = common.DeleteArgocdApplicationFailed
			err = errors.Wrap(err, code.GetMsg())
			goto Fail
		}
	}
	log.Infof("delete application successful")
	c.JSON(200, common.SuccessResponse(c, app.Id))
	return

Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err))
}
