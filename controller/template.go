package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"lyyops-cicd/dao"
	"lyyops-cicd/pkg/common"
	"lyyops-cicd/pkg/log"
)

type TemplateController struct{}

func TemplateControllerGroupRegistry(group *gin.RouterGroup) {
	controller := TemplateController{}
	group.GET(":id", controller.Get)
	group.GET("/list", controller.List)
	group.GET("/search", controller.Search)
	group.POST("/create", controller.Create)
	group.POST("/delete/:id", controller.Delete)
}

// Get TemplateController godoc
// @Summary 获取WorkflowTemplate
// @Tags 工作流模板
// @Accept json
// @Produce json
// @Param id path string true "WorkflowTemplate ID"
// @Success 200 {string} string ""
// @Router /template/{id} [get]
func (t *TemplateController) Get(c *gin.Context) {
	var (
		code = common.Success
		tmpl *dao.Template
		err  error
	)
	tmpl, err = dao.GetTemplate(c, c.Param("id"))
	if err != nil {
		code = common.GetTemplateFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}

	c.JSON(200, common.SuccessResponse(c, string(tmpl.Content)))
	return
Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err.Error()))
}

// List TemplateController godoc
// @Summary 获取WorkflowTemplate列表
// @Tags 工作流模板
// @Accept json
// @Produce json
// @Success 200 {string} string ""
// @Router /template/list [get]
func (t *TemplateController) List(c *gin.Context) {
	var (
		code      = common.Success
		tmplNames []string
		err       error
	)
	tmplNames, err = dao.ListTemplate(c, "")
	if err != nil {
		code = common.ListTemplateFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}
	c.JSON(200, common.SuccessResponse(c, tmplNames))
	return
Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err.Error()))
}

// Search TemplateController godoc
// @Summary 搜索WorkflowTemplate
// @Tags 工作流模板
// @Accept json
// @Produce json
// @Param keyword query string true "关键字"
// @Success 200 {string} string ""
// @Router /template/search [get]
func (t *TemplateController) Search(c *gin.Context) {
	var (
		code      = common.Success
		tmplNames []string
		err       error
	)
	tmplNames, err = dao.ListTemplate(c, c.Query("keyword"))
	if err != nil {
		code = common.ListTemplateFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}
	c.JSON(200, common.SuccessResponse(c, tmplNames))
	return
Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err.Error()))
}

// Create TemplateController godoc
// @Summary 创建WorkflowTemplate
// @Tags 工作流模板
// @Accept json
// @Produce json
// @Param context body string true "内容"
// @Success 200 {string} string ""
// @Router /template/create [post]
func (t *TemplateController) Create(c *gin.Context) {
	var (
		code = common.Success
		tmpl = dao.Template{Ctx: c}
		err  error
	)
	tmpl.Content, err = c.GetRawData()
	if err != nil || len(tmpl.Content) == 0 {
		code = common.InvalidParam
		err = errors.Wrap(err, "template content 不能为空")
		goto Fail
	}

	if err = tmpl.CreateWorkflowTemplate(); err != nil {
		code = common.CreateTemplateFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}
	c.JSON(200, common.SuccessResponse(c, tmpl.Id))
	return

Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err.Error()))
}

// Delete TemplateController godoc
// @Summary 删除WorkflowTemplate
// @Tags 工作流模板
// @Accept json
// @Produce json
// @Param id path string true "WorkflowTemplate ID"
// @Success 200 {string} string ""
// @Router /template/delete/{id} [post]
func (t *TemplateController) Delete(c *gin.Context) {
	var (
		code = common.Success
		tmpl = dao.Template{Ctx: c}
		err  error
	)
	tmpl.Id = c.Param("id")
	if tmpl.Id == "" {
		code = common.InvalidParam
		err = errors.Wrap(err, "template Id 不能为空")
		goto Fail
	}
	log.Debugf("template: %+v", tmpl)
	if err = tmpl.DeleteWorkflowTemplate(); err != nil {
		code = common.DeleteTemplateFailed
		err = errors.Wrap(err, "DeleteWorkflowTemplate")
		goto Fail
	}
	c.JSON(200, common.SuccessResponse(c, ""))
	return
Fail:
	log.Error(err.Error())
	c.JSON(400, common.NewResponse(c, code, err.Error()))
}
