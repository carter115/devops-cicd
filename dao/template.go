package dao

// 处理argo对象: workflow, workflow template

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apiclient/workflowtemplate"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/workflow/common"
	"github.com/pkg/errors"
	"io/ioutil"
	"lyyops-cicd/pkg/log"
	"strings"
)

type Template struct {
	Ctx     context.Context
	Id      string `json:"id"`
	Content []byte `json:"content"`
}

func GetTemplate(ctx context.Context, id string) (*Template, error) {
	bs, err := RedisClient.Get(ctx, TemplateKey(id)).Bytes()
	if err != nil {
		return nil, err
	}
	return &Template{Id: id, Content: bs}, nil
}

func ListTemplate(ctx context.Context, keyword string) ([]string, error) {
	res, err := RedisClient.Keys(ctx, TemplateKey("*")).Result()
	if err != nil {
		return nil, err
	}
	log.Debug("redis keys:",res)
	var tmpls []string
	for _, val := range res {
		name := ExtractTemplateName(val)
		// 过滤搜索关键字
		if keyword != "" {
			if !strings.Contains(strings.ToLower(name), strings.ToLower(keyword)) {
				continue
			}
		}
		tmpls = append(tmpls, name)
	}
	return tmpls, nil
}

func (t *Template) Save() error {
	return RedisClient.Set(t.Ctx, TemplateKey(t.Id), t.Content, -1).Err()
}

func (t *Template) Delete() error {
	return RedisClient.Del(t.Ctx, TemplateKey(t.Id)).Err()
}

func ListWorkflowTemplate() error {
	ctx, cli, err := NewArgoClient()
	if err != nil {
		return err
	}
	svcCli, err := cli.NewWorkflowTemplateServiceClient()
	if err != nil {
		return err
	}
	ns := ctx.Value("namespace").(string)
	fmt.Println("namespace:", ns)
	inReq := workflowtemplate.WorkflowTemplateListRequest{
		Namespace: ns,
	}
	resp, err := svcCli.ListWorkflowTemplates(ctx, &inReq)
	if err != nil {
		return err
	}
	fmt.Println(resp)
	return nil

}

func (t *Template) CreateWorkflowTemplate() error {
	ctx, cli, err := NewArgoClient()
	if err != nil {
		return err
	}
	svcCli, err := cli.NewWorkflowTemplateServiceClient()
	if err != nil {
		return err
	}

	workflowTemplate, err := UnmarshalWorkflowTemplates(t.Content)
	if err != nil {
		return err
	}

	t.Id = workflowTemplate.Name
	inReq := workflowtemplate.WorkflowTemplateCreateRequest{
		Namespace: getNamespace(ctx),
		Template:  workflowTemplate,
	}
	resp, err := svcCli.CreateWorkflowTemplate(ctx, &inReq)
	if err != nil {
		return err
	}
	log.Infof("name: %s, namespace: %s", resp.Name, resp.Namespace)
	return t.Save()
}

func (t *Template) DeleteWorkflowTemplate() error {
	ctx, cli, err := NewArgoClient()
	if err != nil {
		return err
	}
	svcCli, err := cli.NewWorkflowTemplateServiceClient()
	if err != nil {
		return err
	}

	inReq := workflowtemplate.WorkflowTemplateDeleteRequest{
		Namespace: getNamespace(ctx),
		Name:      t.Id,
	}
	resp, err := svcCli.DeleteWorkflowTemplate(ctx, &inReq)
	if err != nil {
		return err
	}
	log.Debugf("response: %s", resp.String())
	log.Infof("delete workflow template: %s", t.Id)
	return t.Delete()
}

func UnmarshalWorkflowTemplates(wfBytes []byte) (*wfv1.WorkflowTemplate, error) {
	yamlWfs, err := common.SplitWorkflowTemplateYAMLFile(wfBytes, false)
	if err == nil && len(yamlWfs) > 0 {
		return &yamlWfs[0], nil // 只返回第1个WorkflowTemplate
	}
	return nil, errors.New(fmt.Sprintf("Failed to parse workflow template: %v", err))
}

func UnmarshalWorkflow(wfBytes []byte) (*wfv1.Workflow, error) {
	yamlWfs, err := common.SplitWorkflowYAMLFile(wfBytes, false)
	if err == nil && len(yamlWfs) > 0 {
		return &yamlWfs[0], nil // 只返回第1个Workflow
	}
	return nil, errors.New(fmt.Sprintf("Failed to parse workflow: %v", err))
}

func LoadLocalTemplates() {
	//files := []string{ "./templates/app_abc.yaml"}
	files := []string{"./templates/checkout.yaml", "./templates/build_golang.yaml", "./templates/app_abc.yaml"}
	for _, f := range files {
		yamlBytes, err := ioutil.ReadFile(f)
		if err != nil {
			return
		}
		log.Info("load yaml file:", string(yamlBytes))
		//CreateWorkflowTemplate(yamlBytes)
	}
}

func TemplateKey(id string) string {
	return strings.Join([]string{"cicd", "template", id}, sep)
}

func ExtractTemplateName(fullname string) string {
	names := strings.Split(fullname, sep)
	return names[2]
}
