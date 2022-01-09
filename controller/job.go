package controller

import (
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"lyyops-cicd/dao"
	"lyyops-cicd/dto"
	"lyyops-cicd/pkg/common"
	"lyyops-cicd/pkg/log"
	"strconv"
	"time"
)

type JobController struct{}

func JobControllerGroupRegistry(group *gin.RouterGroup) {
	controller := JobController{}
	group.GET(":id", controller.Get)
	group.GET("/list", controller.List)
	group.GET("/search", controller.Search)
	group.POST("/create", controller.Create)
	group.POST("/delete/:id", controller.Delete)
}

// Get JobController godoc
// @Summary 获取发布任务
// @Tags 发布任务管理
// @Accept json
// @Produce json
// @Param id path string true "任务 ID"
// @Success 200 {string} string ""
// @Router /job/{id} [get]
func (a *JobController) Get(c *gin.Context) {
	var (
		code   = common.Success
		job    = dao.Job{Ctx: c, Id: c.Param("id")}
		output = dto.GetJobOutput{}
		err    error
	)
	if err = job.Validate(); err != nil {
		code = common.InvalidParam
		goto Fail
	}

	if err = job.GetStatusAndPhase(&output); err != nil {
		code = common.GetJobStatusFailed
		err = errors.Wrap(err, "job.GetStatusAndPhase")
		goto Fail
	}
	output.Id = job.Id
	output.Status = job.Status
	output.Cost = job.Cost

	log.Debugf("get job: %+v", output)
	c.JSON(200, common.SuccessResponse(c, output))
	return
Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err.Error()))
}

// List JobController godoc
// @Summary 发布任务列表
// @Tags 发布任务管理
// @Accept json
// @Produce json
// @Param size query int true "数量"
// @Success 200 {string} string ""
// @Router /job/list [get]
func (a *JobController) List(c *gin.Context) {
	var (
		code    common.StatusCode
		jobs    []*dao.Job
		size    int
		outputs dto.ListJobOutput
		out  dto.JobOutput
		err     error
	)
	size, err = strconv.Atoi(c.Query("size"))
	if err != nil || size <= 0 {
		size = 100 // 默认值
	}

	jobs, err = dao.ListJob(c, "", size)
	if err != nil {
		code = common.ListJobFailed
		err = errors.Wrapf(err, code.GetMsg())
		goto Fail
	}

	outputs = make(dto.ListJobOutput, len(jobs))
	// 拼接返回结果
	for k, job := range jobs {
		out = dto.JobOutput{}
		out.Id = job.Id
		out.Start = job.StartTime.Format(time.RFC3339)
		out.End = job.EndTime.Format(time.RFC3339)
		out.Status = job.Status
		outputs[k] = out
	}
	c.JSON(200, common.SuccessResponse(c, outputs))
	return
Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err.Error()))
}

// Search JobController godoc
// @Summary 搜索发布任务
// @Tags 发布任务管理
// @Accept json
// @Produce json
// @Param keyword query string true "关键字"
// @Success 200 {string} string ""
// @Router /job/search [get]
func (a *JobController) Search(c *gin.Context) {
	var (
		code    common.StatusCode
		jobs    []*dao.Job
		outputs dto.ListJobOutput
		out     dto.JobOutput
		err     error
	)

	jobs, err = dao.ListJob(c, c.Query("keyword"), 100)
	if err != nil {
		code = common.ListJobFailed
		err = errors.Wrapf(err, code.GetMsg())
		goto Fail
	}

	outputs = make(dto.ListJobOutput, len(jobs))
	// 拼接返回结果
	for k, job := range jobs {
		out = dto.JobOutput{}
		out.Id = job.Id
		out.Start = job.StartTime.Format(time.RFC3339)
		out.End = job.EndTime.Format(time.RFC3339)
		out.Status = job.Status
		outputs[k] = out
	}
	c.JSON(200, common.SuccessResponse(c, outputs))
	return
Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err.Error()))
}

// Create JobController godoc
// @Summary 创建发布任务
// @Tags 发布任务管理
// @Accept json
// @Produce json
// @Param id query string true "Application ID"
// @Param branch query string true "代码分支"
// @Success 200 {string} string ""
// @Router /job/create [post]
func (j *JobController) Create(c *gin.Context) {
	var (
		code       = common.Success
		id, branch string
		app        *dao.Application
		job        = dao.Job{Ctx: c}
		wf         *wfv1.Workflow
		err        error
	)
	id = c.Query("id")
	branch = c.Query("branch")
	log.Debugf("application: %s, branch: %s", id, branch)

	// 获取applicaion的yaml内容
	if app, err = dao.GetApplication(c, id); err != nil {
		code = common.ApplicationNotFound
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}

	// 解析yaml文件
	if wf, err = dao.UnmarshalWorkflow(app.Content); err != nil {
		code = common.UnmarshalWorkflowFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}

	// 修改workflow的参数branch
	wf.Spec.Templates[0].Steps[0].Steps[0].Arguments.Parameters[1].Value = wfv1.AnyStringPtr(branch)

	//job.Id = wf.GetGenerateName()
	job.Workflow = wf
	//log.Debugf("job workflow: %s", common.ParseJsonStr(job.Workflow))
	job.PhaseNames = job.GetPhaseNames()
	log.Debugf("job.PhaseNames: %v", job.PhaseNames)

	if err = job.Create(); err != nil {
		code = common.CreateJobFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}
	c.JSON(200, common.SuccessResponse(c, dto.CreateJobOutput{Id: job.Id}))
	return
Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err.Error()))
}

// Delete JobController godoc
// @Summary 删除发布任务
// @Tags 发布任务管理
// @Accept json
// @Produce json
// @Param id path string true "Application ID"
// @Success 200 {string} string ""
// @Router /job/delete/{id} [post]
func (a *JobController) Delete(c *gin.Context) {
	var (
		code = common.Success
		job  = dao.Job{Ctx: c, Id: c.Param("id")}
		err  error
	)
	if err := job.Delete(); err != nil {
		code = common.DeleteJobFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}
	c.JSON(200, common.SuccessResponse(c, nil))
	return
Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err))
}
