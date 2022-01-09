package dao

import (
	"context"
	"github.com/pkg/errors"
	"lyyops-cicd/pkg/log"
	"strings"
)

type Application struct {
	Ctx     context.Context
	Id      string `json:"id"`
	Content []byte `json:"content"`
}

func GetApplication(ctx context.Context, id string) (*Application, error) {
	content, err := RedisClient.Get(ctx, ApplicationKey(id)).Bytes()
	return &Application{Id: id, Content: content}, err
}

func ListApplication(ctx context.Context) ([]string, error) {
	res, err := RedisClient.Keys(ctx, ApplicationKey("*")).Result()
	if err != nil {
		err = errors.Wrap(err, "RedisClient.Keys")
		return nil, err
	}
	log.Debug(res)
	var apps = []string{}
	for _, val := range res {
		apps = append(apps, ExtractTemplateName(val))
	}
	return apps, nil
}

func (a *Application) Save() error {
	return RedisClient.Set(a.Ctx, ApplicationKey(a.Id), a.Content, -1).Err()
}

func (a *Application) Delete() error {
	return RedisClient.Del(a.Ctx, ApplicationKey(a.Id)).Err()
}

func ApplicationKey(id string) string {
	return strings.Join([]string{"cicd", "application", id}, sep)
}

func ExtractApplicationName(fullname string) string {
	names := strings.Split(fullname, sep)
	return names[2]
}
