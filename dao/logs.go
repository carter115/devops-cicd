package dao

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflow"
	"io"
	corev1 "k8s.io/api/core/v1"
	"lyyops-cicd/pkg/log"
	"strings"
	"time"
)

const defaultExpired = time.Hour * 24 * 100 // 100天

// 获取pod日志
func GetLog(ctx context.Context, podName string) ([]string, error) {
	log.Debugf("log key: %s", LogsKey(podName))
	str, err := RedisClient.LRange(ctx, LogsKey(podName), 0, -1).Result()
	return str, err
}

// 创建pod日志
func LogsSave(ctx context.Context, podName, line string) error {
	if err := RedisClient.RPush(ctx, LogsKey(podName), line).Err(); err != nil {
		return err
	}
	return RedisClient.Expire(ctx, LogsKey(podName), defaultExpired).Err()
}

func logWorkflow(ctx context.Context, serviceClient workflow.WorkflowServiceClient, job *Job, namespace, workflowName, podName string, logOptions *corev1.PodLogOptions) {
	// logs
	stream, err := serviceClient.WorkflowLogs(ctx, &workflow.WorkflowLogRequest{
		Name:       workflowName,
		Namespace:  namespace,
		PodName:    podName,
		LogOptions: logOptions,
	})
	if err != nil {
		log.Errorf("log workflow[%s]: %+v", job.Id, err)
		return
	}

	// loop on log lines
	log.Debugf("workflow[%s] start recv log stream......", job.Id)
	for {
		event, err := stream.Recv()

		if err == io.EOF {
			return
		}
		if err != nil {
			log.Errorf("log workflow[%s] stream recv error: %+v", job.Id, err)
			return
		}
		// add now time
		text := fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), event.Content)
		if err := LogsSave(ctx, event.PodName, text); err != nil {
			log.Errorf("logs workflow[%s] save: %+v", job.Id, err)
			continue
		}
	}
}

func LogsKey(id string) string {
	return strings.Join([]string{"cicd", "logs", id}, sep)
}

func ExtractLogsName(fullname string) string {
	names := strings.Split(fullname, sep)
	return names[2]
}
