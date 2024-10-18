package repositories

import (
	"github.com/xanzy/go-gitlab"
	"gombot/pkg/configs"
	"log"
	"strings"
)

var git *gitlab.Client

func getGitlabClient() *gitlab.Client {
	config := configs.LoadConfig(configs.ConfigPath)
	if git != nil {
		return git
	}
	var err error
	git, err = gitlab.NewClient(config.GitlabToken)
	if err != nil {
		log.Fatalf("can not connect to database, %v", err)
	}
	return git
}

func GetProjectByApplicationName(applicationName string) *gitlab.Project { //TODO create DTO for data transfers
	git = getGitlabClient()
	projects, _, _ := git.Projects.ListProjects(&gitlab.ListProjectsOptions{
		Membership: gitlab.Ptr(true), // TODO: read about that:https://stackoverflow.com/questions/28817992/how-to-set-bool-pointer-to-true-in-struct-literal ( gitlab.Ptr() )
	}, nil)

	for _, project := range projects {
		if strings.ToLower(project.Name) == strings.ToLower(applicationName) {
			return project
		}
	}
	return nil
}

func CreatePipeline(projectId int, ref string) *gitlab.Pipeline {
	git = getGitlabClient()
	pipeline, _, err := git.Pipelines.CreatePipeline(projectId, &gitlab.CreatePipelineOptions{
		Ref: gitlab.Ptr(ref),
	})
	if err != nil {
		log.Printf("can not create pipeline, %v", err)
	}
	return pipeline
}

func GetPipeline(projectId int, pipelineId int) *gitlab.Pipeline {
	git = getGitlabClient()
	pipeline, _, err := git.Pipelines.GetPipeline(projectId, pipelineId, nil)
	if err != nil {
		log.Printf("can not get pipeline, %v", err)
	}
	return pipeline
}
