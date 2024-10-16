package handlers

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/xanzy/go-gitlab"
	"gombot/pkg/configs"
	"log"
)

func newTrue() *bool {
	b := true
	return &b
}
func TestHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

	config := configs.LoadConfig(configs.ConfigPath)

	git, err := gitlab.NewClient(config.GitlabToken)

	if err != nil {
		log.Fatal(err)
	}

	/*	projects, _, _ := git.Projects.ListProjects(&gitlab.ListProjectsOptions{
			Membership: newTrue(), // TODO: read about that:https://stackoverflow.com/questions/28817992/how-to-set-bool-pointer-to-true-in-struct-literal ( gitlab.Ptr() )
		}, nil)

		for _, project := range projects {
			SendMessage(b, ctx, update.Message.Chat.ID, fmt.Sprintf("find projects %d %s %s %s \n", project.ID, project.Name, project.Description, project.WebURL))
		}*/

	/*opt := &gitlab.ListProjectPipelinesOptions{
		Scope:         gitlab.Ptr("branches"),
		Status:        gitlab.Ptr(gitlab.Running),
		Ref:           gitlab.Ptr("master"),
		YamlErrors:    gitlab.Ptr(true),
		Name:          gitlab.Ptr("name"),
		Username:      gitlab.Ptr("username"),
		UpdatedAfter:  gitlab.Ptr(time.Now().Add(-24 * 365 * time.Hour)),
		UpdatedBefore: gitlab.Ptr(time.Now().Add(-7 * 24 * time.Hour)),
		OrderBy:       gitlab.Ptr("status"),
		Sort:          gitlab.Ptr("asc"),
	}*/
	/*pipelines, _, err := git.Pipelines.CreatePipeline(62700716, &gitlab.CreatePipelineOptions{
		Ref: gitlab.Ptr("master"),
	})*/
	pipelines, _, err := git.Pipelines.ListProjectPipelines(62700716, &gitlab.ListProjectPipelinesOptions{})
	if err != nil {
		log.Fatal(err)
	}

	//log.Println(pipelines)
	for _, pipeline := range pipelines {
		log.Printf("Found pipeline: %v", pipeline)
		SendMessage(b, ctx, update.Message.Chat.ID, fmt.Sprintf("find pipeline %s", pipeline))
	}

}
