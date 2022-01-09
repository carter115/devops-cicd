package common

type StatusCode int

const (
	Success StatusCode = iota
	InvalidParam
	VerifyFailed

	ApplicationNotFound
	ListApplicationFailed
	CreateApplicationFailed
	SaveApplicationFailed
	DeleteApplicationFailed

	CreateJobFailed
	GetJobFailed
	GetJobStatusFailed
	ListJobFailed
	DeleteJobFailed

	GetTemplateFailed
	ListTemplateFailed
	CreateTemplateFailed
	DeleteTemplateFailed

	GetLogsFailed

	GetArgocdApplicationFailed
	GetArgocdApplicationStatusFailed
	CreateArgocdApplicationFailed
	DeleteArgocdApplicationFailed

	UnmarshalWorkflowFailed
	UnmarshalWorkflowTemplateFailed

	Unknown StatusCode = 9999
)

var statusCodeMap = map[StatusCode]string{
	Success:      "",
	InvalidParam: "invalid params",
	VerifyFailed: "verification failed",

	ApplicationNotFound:     "应用Id不存在",
	ListApplicationFailed:   "获取应用列表失败",
	CreateApplicationFailed: "创建应用失败",
	SaveApplicationFailed:   "保存应用失败",
	DeleteApplicationFailed: "删除应用失败",

	CreateJobFailed:    "创建任务失败",
	GetJobFailed:       "获取任务失败",
	GetJobStatusFailed: "获取任务状态失败",
	ListJobFailed:      "获取任务列表失败",
	DeleteJobFailed:    "删除任务失败",

	GetTemplateFailed:    "获取流水线模版失败",
	ListTemplateFailed:   "获取流水线模版列表失败",
	CreateTemplateFailed: "创建流水线模版失败",
	DeleteTemplateFailed: "删除流水线模版失败",

	GetLogsFailed: "获取Pod日志失败",

	GetArgocdApplicationFailed:       "获取Argocd Application失败",
	GetArgocdApplicationStatusFailed: "获取Argocd Application Status失败",
	CreateArgocdApplicationFailed:    "创建Argocd Application失败",
	DeleteArgocdApplicationFailed:    "删除Argocd Application失败",

	UnmarshalWorkflowFailed:         "解析Workflow失败",
	UnmarshalWorkflowTemplateFailed: "解析WorkflowTemplate失败",
	Unknown:                         "unknown",
}

func (c StatusCode) GetMsg() string {
	if s, ok := statusCodeMap[c]; ok {
		return s
	}
	return statusCodeMap[Unknown]
}
