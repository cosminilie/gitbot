package gitbot

import (
	"fmt"
	"strings"

	gitlab "github.com/xanzy/go-gitlab"
)

//addRepoEventHook adds an event hook or updates an existing event hook to point at this service.
func addRepoEventHook(s *basicService) error {

	//get server Ip
	ip, err := externalIP()
	if err != nil {
		fmt.Println(err)
	}
	//init hook options
	hookOpts := &gitlab.AddProjectHookOptions{
		URL:                   gitlab.String(fmt.Sprintf("http://%s:9091/hook", ip)),
		PushEvents:            gitlab.Bool(true),
		IssuesEvents:          gitlab.Bool(false),
		MergeRequestsEvents:   gitlab.Bool(true),
		TagPushEvents:         gitlab.Bool(false),
		NoteEvents:            gitlab.Bool(true),
		BuildEvents:           gitlab.Bool(false),
		PipelineEvents:        gitlab.Bool(false),
		WikiPageEvents:        gitlab.Bool(false),
		EnableSSLVerification: gitlab.Bool(false),
	}
	//init gitlab List HookOpts
	listOptions := &gitlab.ListProjectHooksOptions{}

	//Loop though each repo
	for _, r := range s.Plugins.Repos {

		//Get project details
		proj, _, err := s.Plugins.GitLabClient.Projects.GetProject(r.Name)
		if err != nil {
			return fmt.Errorf("AddRepoEventHook Error: Failed to get Project %s. Returned error: %s", r.Name, err)
		}

		//Get project hooks
		hooks, _, err := s.Plugins.GitLabClient.Projects.ListProjectHooks(proj.ID, listOptions)
		if err != nil {
			return fmt.Errorf("AddRepoEventHook Error: Failed to list hooks for Project %s. Returned error: %s", proj.NameWithNamespace, err)
		}

		//mark projects that don't have hooks
		var repoHook = false
		for _, h := range hooks {
			if strings.Contains(h.URL, ip) {
				s.logger.Log(
					"Func", "addRepoEventHook",
					"Action", "HookMatch",
					"Project", proj.NameWithNamespace,
					"Hook", h.URL,
				)
				repoHook = true
				break
			}
			//check to see if it was created by us, and if it was, delete it to clean up (in case our IP address changed)
			if botHook.MatchString(h.URL) {
				s.logger.Log(
					"Func", "addRepoEventHook",
					"Action", "FoundOldHook",
					"Project", proj.NameWithNamespace,
					"Hook", h.URL,
				)
				//delete hook
				_, err = s.Plugins.GitLabClient.Projects.DeleteProjectHook(proj.ID, h.ID)
				if err != nil {
					return fmt.Errorf("error Deleting hook:%s for project :%s. Returned errror is:%s", h.URL, proj.NameWithNamespace, err)
				}

			}
		}

		if !repoHook {
			s.logger.Log(
				"Func", "addRepoEventHook",
				"Action", "CreatingHook",
				"Project", proj.NameWithNamespace,
				"Hook", *hookOpts.URL,
			)
			//create hook
			_, _, err := s.Plugins.GitLabClient.Projects.AddProjectHook(proj.ID, hookOpts)
			if err != nil {
				return fmt.Errorf("error Creating hook:%s for project :%s. Returned errror is:%s", *hookOpts.URL, proj.NameWithNamespace, err)

			}

		}
	}

	return nil
}
