package dao

// 处理argocd对象: Application

import (
	"context"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
	appv1 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"lyyops-cicd/config"
	"lyyops-cicd/pkg/log"
)

type Deployment struct {
	Ctx        context.Context
	Id         string   `json:"id"`
	Namespace  string   `json:"namespace"`
	RepoPath   string   `json:"repo_path"`
	ValueFiles []string `json:"value_files"`
}

var ArgoCDClient apiclient.Client

// Argocd Application yaml content
func GetDeploymentYaml(ctx context.Context, id string) (string, error) {
	argoCli, err := newArgocdClient()
	if err != nil {
		return "", errors.Wrap(err, "newArgocdClient")
	}

	clo, appcli, err := argoCli.NewApplicationClient()
	defer clo.Close()
	if err != nil {
		return "nil", errors.Wrap(err, "argoCli.NewApplicationClient")
	}
	log.Debugf("argocd app client: %+v", argoCli)

	inReq := appv1.ApplicationQuery{Name: &id}
	resp, err := appcli.Get(ctx, &inReq)
	if err != nil {
		return "", errors.Wrap(err, "appcli.Get")
	}

	yamlBytes, err := yaml.Marshal(resp)
	//yamlBytes, err := resp.Marshal()
	if err != nil {
		return "", err
	}
	log.Debugf("%s yaml: %s", id, string(yamlBytes))
	return string(yamlBytes), nil
}

// TODO 应用部署的状态和时间
func GetDeploymentStatus(ctx context.Context, id string) (string, error) {
	return "", nil
}

// 创建Argocd Application(yaml content)
func (d *Deployment) CreateYaml(yamlStr string) error {
	var app v1alpha1.Application
	if err := yaml.Unmarshal([]byte(yamlStr), &app); err != nil {
		return err
	}
	log.Debugf("unmarshal yaml object: %+v", app)

	argoCli, err := newArgocdClient()
	if err != nil {
		return errors.Wrap(err, "newArgocdClient")
	}

	clo, appcli, err := argoCli.NewApplicationClient()
	defer clo.Close()
	if err != nil {
		return errors.Wrap(err, "argoCli.NewApplicationClient")
	}
	log.Debugf("argocd app client: %+v", argoCli)

	inReq := appv1.ApplicationCreateRequest{
		Application: app,
	}
	log.Infof("request: %+v", inReq)
	resp, err := appcli.Create(d.Ctx, &inReq)
	if err != nil {
		return errors.Wrap(err, "appcli.Create")
	}
	log.Debugf("argocd application created: %+v", resp)
	return nil
}

// 创建Argocd Application(根据helm文件名: values.yaml)
func (d *Deployment) CreateHelm() error {
	var app = v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "argoproj.io/v1alpha1",
			Kind:       "Application",
		},
		ObjectMeta: metav1.ObjectMeta{Name: d.Id},
		Spec: v1alpha1.ApplicationSpec{
			Project: config.Config.Argocd.Project,
			Source: v1alpha1.ApplicationSource{
				RepoURL: config.Config.Argocd.RepoUrl,
				Path:    d.RepoPath,
				Helm: &v1alpha1.ApplicationSourceHelm{
					ValueFiles: d.ValueFiles,
				},
				TargetRevision: config.Config.Argocd.Revision,
			},
			Destination: v1alpha1.ApplicationDestination{
				Server:    config.Config.Argocd.K8sServer,
				Namespace: d.Namespace,
			},
			SyncPolicy: &v1alpha1.SyncPolicy{
				Automated: &v1alpha1.SyncPolicyAutomated{
					Prune:    true,
					SelfHeal: true,
				},
			},
		},
	}

	argoCli, err := newArgocdClient()
	if err != nil {
		return errors.Wrap(err, "newArgocdClient")
	}

	clo, appcli, err := argoCli.NewApplicationClient()
	defer clo.Close()
	if err != nil {
		return errors.Wrap(err, "argoCli.NewApplicationClient")
	}
	log.Debugf("argocd app client: %+v", argoCli)

	inReq := appv1.ApplicationCreateRequest{
		Application: app,
	}
	log.Infof("request: %+v", inReq)
	resp, err := appcli.Create(d.Ctx, &inReq)
	if err != nil {
		return err
	}
	log.Debugf("argocd application created: %+v", resp)

	return nil
}

// 删除Argocd Application
func (d *Deployment) Delete() error {
	argoCli, err := newArgocdClient()
	if err != nil {
		return err
	}

	clo, appcli, err := argoCli.NewApplicationClient()
	defer clo.Close()
	if err != nil {
		return err
	}
	log.Debugf("argocd app client: %+v", argoCli)

	inReq := appv1.ApplicationDeleteRequest{
		Name: &d.Id,
	}
	log.Infof("request: %+v", inReq)
	resp, err := appcli.Delete(d.Ctx, &inReq)
	if err != nil {
		return errors.Wrap(err, "appcli.Delete")
	}
	log.Debugf("argocd application deleted: %+v", resp)
	return nil
}

func newArgocdClient() (apiclient.Client, error) {
	if ArgoCDClient != nil {
		return ArgoCDClient, nil
	}
	acdClient, err := apiclient.NewClient(&apiclient.ClientOptions{
		ServerAddr:   config.Config.Argocd.Address,
		PlainText:    true,
		Insecure:     true,
		AuthToken:    config.Config.Argocd.Token,
		HttpRetryMax: 5,
	})
	if err != nil {
		return nil, errors.Wrap(err, "apiclient.NewClient")
	}
	log.Infof("logged in successfully: %+v", acdClient)
	ArgoCDClient = acdClient
	return ArgoCDClient, nil
}
