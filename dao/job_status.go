package dao

import (
	"github.com/pkg/errors"
	"lyyops-cicd/dto"
	"lyyops-cicd/pkg/log"
	"strings"
	"time"
)

const DefaultJobStatus string = "Running"

func (j *Job) GetStatusAndPhase(out *dto.GetJobOutput) (err error) {
	var ok bool
	hres, err := RedisClient.HGetAll(j.Ctx, JobStatusKey(j.Id)).Result()
	if err != nil {
		return err
	}
	if j.StartTime, err = time.Parse(time.RFC3339, hres["start_time"]); err != nil {
		return err
	}
	// 如果没有结束时间，则根据当前时间来计算耗时
	j.EndTime, err = time.Parse(time.RFC3339, hres["end_time"])
	j.setCost()

	if j.Status, ok = hres["status"]; !ok {
		j.Status = DefaultJobStatus // DB 查不到，给默认值
	}

	// 处理阶段名字和状态
	fetch_phase_names, ok := hres["phase_names"]
	if !ok {
		return errors.New("key phase_names not found")
	}
	names := strings.Split(fetch_phase_names, ",") // 获取阶段的顺序名字
	j.PhaseNames = names
	phases, err := j.GetPhases()
	if err != nil {
		return err
	}
	log.Debugf("phase name: %v, %v", names, phases)
	out.PhaseList = phases
	return
}

func (j *Job) statusSave() error {
	var err error
	j.setCost() // 保存前先计算时间
	log.Debug(j.Id, j.Status, j.PhaseNames)
	names := strings.Join(j.PhaseNames, ",") // 逗号拼接名字列表
	if err = RedisClient.HMSet(j.Ctx, JobStatusKey(j.Id),
		"start_time", j.StartTime,
		"end_time", j.EndTime,
		"status", j.Status,
		"phase_names", names,
		"cost", j.Cost,
	).Err(); err != nil {
		return err
	}

	err = RedisClient.Expire(j.Ctx, JobStatusKey(j.Id), defaultExpired).Err() // 设定过期时间
	return err
}

func (j *Job) deleteStatus() error {
	return RedisClient.Del(j.Ctx, JobStatusKey(j.Id)).Err()
}

func JobStatusKey(id string) string {
	return strings.Join([]string{"cicd", "job-status", id}, sep)
}

func ExtractJobStatusName(fullname string) string {
	names := strings.Split(fullname, sep)
	return names[2]
}
