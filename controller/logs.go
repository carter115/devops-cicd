package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"lyyops-cicd/dao"
	"lyyops-cicd/pkg/common"
	"lyyops-cicd/pkg/log"
)

type LogsController struct{}

func LogsControllerGroupRegistry(group *gin.RouterGroup) {
	controller := LogsController{}
	group.GET(":id", controller.Get)
}

// Get LogsController godoc
// @Summary 获取pod日志
// @Tags 日志管理
// @Accept json
// @Produce json
// @Param id path string true "pod name"
// @Success 200 {string} string ""
// @Router /logs/{id} [get]
func (l *LogsController) Get(c *gin.Context) {
	var (
		code = common.Success
		res  []string
		err  error
	)
	res, err = dao.GetLog(c, c.Param("id"))
	if err != nil {
		code = common.GetLogsFailed
		err = errors.Wrap(err, code.GetMsg())
		goto Fail
	}
	c.JSON(200, common.SuccessResponse(c, res))
	return
Fail:
	log.Error(err)
	c.JSON(400, common.NewResponse(c, code, err.Error()))
}
