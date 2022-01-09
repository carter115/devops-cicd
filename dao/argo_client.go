package dao

import (
	"context"
	"github.com/argoproj/argo-workflows/v3/cmd/argo/commands/client"
	"github.com/argoproj/argo-workflows/v3/pkg/apiclient"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	"lyyops-cicd/config"
)

// 默认的argo client配置
func NewArgoClient() (context.Context, apiclient.Client, error) {
	opts := apiclient.Opts{
		ArgoServerOpts: apiclient.ArgoServerOpts{
			URL:                config.Config.Argo.Address,
			Secure:             config.Config.Argo.Secret,
			InsecureSkipVerify: config.Config.Argo.InsecureSkipVerify,
			HTTP1:              config.Config.Argo.Http1,
		},
		AuthSupplier: getAuth(),
		ClientConfigSupplier: func() clientcmd.ClientConfig {
			return client.GetConfig()
		},
	}
	ctx, cli, err := apiclient.NewClientFromOpts(opts)
	if err != nil {
		err = errors.Wrap(err, "apiclient.NewClientFromOpts")
		return ctx, cli, err
	}
	setNamespace(&ctx)
	return ctx, cli, err
}

func setNamespace(ctx *context.Context) {
	ns := config.Config.Argo.Namespace
	if ns == "" {
		ns = "default"
	}
	*ctx = context.WithValue(*ctx, "namespace", ns)
}

func getNamespace(ctx context.Context) string {
	ns := ctx.Value("namespace").(string)
	if ns == "" {
		ns = "default"
	}
	return ns
}

func getAuth() func() string {
	return func() string {
		return config.Config.Argo.Token
	}
}
