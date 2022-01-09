package dao

import (
	"context"
	"encoding/json"
	"github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflow"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/util"
	"github.com/argoproj/argo-workflows/v3/workflow/common"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"lyyops-cicd/config"
	common2 "lyyops-cicd/pkg/common"
	"lyyops-cicd/pkg/log"
	"sort"
	"strings"
	"time"
)

const DefaultCostDuration = time.Second

type Job struct {
	Ctx        context.Context            `json:"-"`
	Id         string                     `json:"id"`
	StartTime  time.Time                  `json:"start_time"`
	EndTime    time.Time                  `json:"end_time"`
	Cost       string                     `json:"cost"`
	Status     string                     `json:"status"`
	Workflow   *wfv1.Workflow             `json:"-"`
	PhaseNames []string                   `json:"phase_names"`
	Phases     map[string]*JobPhaseStatus `json:"phases"`
}

func (j Job) Validate() error {
	return validation.ValidateStruct(&j,
		validation.Field(&j.Ctx, validation.NotNil),
		validation.Field(&j.Id, validation.Required, validation.Length(3, 100)),
	)
}

// 创建任务
func (j *Job) Create() error {
	log.Debugf("job: %+v", j)
	j.StartTime = time.Now()
	j.Status = DefaultJobStatus

	ctx, cli, err := NewArgoClient()
	if err != nil {
		return err
	}
	svcCli := cli.NewWorkflowServiceClient()

	j.Workflow.Spec.ServiceAccountName = config.Config.Argo.ServiceAccount
	inReq := workflow.WorkflowCreateRequest{
		Namespace: getNamespace(ctx),
		Workflow:  j.Workflow,
		//CreateOptions: &metav1.CreateOptions{DryRun: []string{"All"}},
	}

	// 启动任务
	created, err := svcCli.CreateWorkflow(ctx, &inReq)
	if err != nil {
		return err
	}

	j.Id = created.GetName() // workflow name
	log.Debugf("namespace: %s, name: %s, serviceaccount: %s", created.GetNamespace(), j.Id, j.Workflow.Spec.ServiceAccountName)

	// 创建任务
	if err := j.statusSave(); err != nil {
		return err
	}

	go waitWatchOrLog(ctx, svcCli, inReq.Namespace, created.Name, j, true, true)

	return nil
}

// 任务耗时
func (j *Job) setCost() {
	dur := j.EndTime.Sub(j.StartTime)
	if dur < DefaultCostDuration {
		j.EndTime = time.Now() // 如果没有结束，则设为当前时间
	}
	j.Cost = j.EndTime.Sub(j.StartTime).String()
}

func waitWatchOrLog(ctx context.Context, serviceClient workflow.WorkflowServiceClient, namespace string, workflowName string, job *Job, ignoreNotFound, saveLog bool) {
	defer func() {
		job.EndTime = time.Now() // 更新job结束时间
		log.Debugf("start save job status: %+v", job)
		if err := job.statusSave(); err != nil {
			log.Errorf("save job status: %+v", job)
		} // 结束后更新job状态
		ctx.Done()
	}()

	go logWorkflow(ctx, serviceClient, job, namespace, workflowName, "", &corev1.PodLogOptions{
		Container: common.MainContainerName,
		Follow:    true,
		Previous:  false,
	})

	req := &workflow.WatchWorkflowsRequest{
		Namespace: namespace,
		ListOptions: &metav1.ListOptions{
			FieldSelector: util.GenerateFieldSelectorFromWorkflowName(workflowName),
			//ResourceVersion: "0",
		},
	}
	stream, err := serviceClient.WatchWorkflows(ctx, req)
	log.Debugf("starting watch workflow[%s]......", job.Id)
	if err != nil {
		if status.Code(err) == codes.NotFound && ignoreNotFound {
			log.Errorf("workflow[%s] not found", job.Id)
			return
		}
		log.Errorf("workflow[%s] error: %+v", job.Id, err)
		return
	}

	log.Debugf("workflow[%s] start watch stream......", job.Id)
	for {
		event, err := stream.Recv()
		log.Debugf("recv job event: %s", event.Object.Name)
		if err == io.EOF {
			log.Debugf("Re-establishing workflow[%s] watch", job.Id)
			stream, err = serviceClient.WatchWorkflows(ctx, req)

			if err != nil {
				log.Errorf("workflow[%s] error: %+v", job.Id, err)
				return
			}
			continue
		}
		if err != nil {
			log.Errorf("workflow[%s] error: %+v", job.Id, err)
			return
		}
		if event == nil {
			continue
		}

		eventWf := event.Object
		log.Debugf("event workflow: %s", common2.ParseJsonStr(eventWf))

		if err := job.phaseSave(eventWf); err != nil {
			log.Error(err)
			continue
		}

		job.Status = string(eventWf.Status.Phase) // 更新job状态
		log.Infof("job status phase save: %s", common2.ParseJsonStr(job))

		// 完成后退出
		if !eventWf.Status.FinishedAt.IsZero() {
			log.Debugf("workflow[%s] successful...", job.Id)
			return
		}
	}
}

func (j *Job) Delete() error {
	if err := j.deleteStatus(); err != nil {
		return errors.Wrap(err, "job.deleteStatus")
	}
	if err := j.deletePhase(); err != nil {
		return errors.Wrap(err, "job.deletePhase")
	}
	return nil
}

// 发布任务列表
func ListJob(ctx context.Context, keyword string, size int) ([]*Job, error) {
	var jobs Jobs

	jobNames, err := RedisClient.Keys(ctx, JobStatusKey("*")).Result()
	if err != nil {
		err = errors.Wrap(err, "ListJob")
		return jobs, err
	}
	if len(jobNames) == 0 {
		return jobs, nil
	}

	for _, name := range jobNames {
		objMap, err := RedisClient.HGetAll(ctx, name).Result()
		if err != nil {
			continue
		}

		// 过滤搜索关键字
		if keyword != "" {
			if !strings.Contains(strings.ToLower(name), strings.ToLower(keyword)) {
				continue
			}
		}

		log.Debugf("fetch job %s map: %+v", name, objMap)
		obj := Job{Id: ExtractJobStatusName(name)}
		bs, _ := json.Marshal(objMap)
		json.Unmarshal(bs, &obj)
		jobs = append(jobs, &obj)
	}

	// 排序
	sort.Sort(jobs)
	if size > len(jobs) {
		size = len(jobs)
	}
	log.Infof("jobs list: %+v", jobs[:size])
	return jobs[:size], nil
}

type Jobs []*Job

// 发布任务列表排序
func (jobs Jobs) Len() int {
	return len(jobs)
}

func (jobs Jobs) Less(i, j int) bool {
	return jobs[i].StartTime.After(jobs[j].StartTime) // 按开始时间倒序
}

func (jobs Jobs) Swap(i, j int) {
	jobs[i], jobs[j] = jobs[j], jobs[i]
}
