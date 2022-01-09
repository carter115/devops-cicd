package dao

import (
	"encoding/json"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/pkg/errors"
	"lyyops-cicd/pkg/common"
	"lyyops-cicd/pkg/log"
	"strings"
)

//const DefaultJobPhaseStatus string = "NotReady"

type JobPhaseStatus struct {
	Name    string `json:"name"`
	PodName string `json:"pod_name"`
	Status  string `json:"status"`
}

// 获取每个阶段的状态
func (j *Job) GetPhases() ([]JobPhaseStatus, error) {
	var phases = []JobPhaseStatus{}
	res, err := RedisClient.HGetAll(j.Ctx, JobPhaseKey(j.Id)).Result()
	if err != nil {
		return phases, err
	}
	log.Debugf("fetch data: %+v", res)
	if len(j.PhaseNames) == 0 {
		return phases, errors.New("job.PhaseNames 不能为空")
	}
	for _, k := range j.PhaseNames {
		phase := JobPhaseStatus{}
		if data, ok := res[k]; ok {
			if err := json.Unmarshal([]byte(data), &phase); err != nil {
				log.Warning("phase status unmarshal: %v", err)
				continue
			}
			phases = append(phases, phase)
			log.Debugf("add phase: %v", phase)
		}
	}
	return phases, nil
}

// 保存每个阶段的信息
func (j *Job) phaseSave(wf *wfv1.Workflow) error {
	log.Debugf("phase info: %s, %s", wf.Name, wf.Status.Phase)
	//log.Debugf("workflow: %s", common.ParseJsonStr(wf))

	// 从workflow中提出阶段的信息
	j.Phases = make(map[string]*JobPhaseStatus)
	steps := j.Workflow.Spec.Templates[0].Steps // templates steps
	log.Debugf("steps: %+v", steps)
	for _, step := range steps {
		name := step.Steps[0].Name
		j.Phases[name] = &JobPhaseStatus{
			Name:    name,
			PodName: getPodnameFromWorkflow(&wf.Status.Nodes, name),
			Status:  string(wf.Status.Phase)}
		log.Debugf("phase: %+v", j.Phases[name])
	}
	log.Infof("job.Phases %v", j.Phases)

	// 保存每个阶段的信息
	if len(j.PhaseNames) == 0 {
		return errors.New("job.PhaseNames 不能为空")
	}
	for _, name := range j.PhaseNames {
		if err := RedisClient.HSet(j.Ctx, JobPhaseKey(wf.Name), name, j.Phases[name].String()).Err(); err != nil {
			return errors.Wrapf(err, "保存 %s 失败", name)
		}
	}
	log.Infof("phases: %v save: %+v", j.PhaseNames, j.Phases)
	return RedisClient.Expire(j.Ctx, JobPhaseKey(j.Id), defaultExpired).Err() // 设定过期时间
}

func (j *Job) GetPhaseNames() []string {
	var (
		names []string
		steps = j.Workflow.Spec.Templates[0].Steps
	)
	for _, step := range steps {
		names = append(names, step.Steps[0].Name)
	}
	log.Infof("job.GetPhaseNames %s", names)
	return names
}

// 从workflow status nodes中获取podname
func getPodnameFromWorkflow(nodes *wfv1.Nodes, phaseName string) string {
	var PodName string
	for k, v := range *nodes {
		log.Debugf("node info key[%s]: %s=%s,%s", k, phaseName, v.DisplayName, v.String())
		if phaseName == v.DisplayName {
			PodName = k
			return PodName
		}
	}
	return PodName
}

func (p *JobPhaseStatus) String() string {
	return common.ParseJsonStr(p)
}

func (j *Job) deletePhase() error {
	return RedisClient.Del(j.Ctx, JobPhaseKey(j.Id)).Err()
}

func JobPhaseKey(id string) string {
	return strings.Join([]string{"cicd", "job-phase", id}, sep)
}

func ExtractJobPhaseName(fullname string) string {
	names := strings.Split(fullname, sep)
	return names[2]
}
