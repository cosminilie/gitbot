package gitbot

import (
	"fmt"
	"strings"

	"github.com/cosminilie/gitbot/plugins"
)

func setupGroupHandlers(svc *basicService, reposChan chan plugins.Repo) {
	for r := range reposChan {
		s := r.Name
		s = strings.TrimSuffix(s, "/*")
		s = strings.TrimSuffix(s, "/")

		//Add repos to top level repos list
		svc.Plugins.GroupRepos[s] = r

	}
}

//globalHandlers loads all global handlers for gitlab groups
func groupHandlers(s *basicService) error {
	for _, tpr := range s.Plugins.GroupRepos {

		for _, h := range s.Plugins.GroupHandlers(tpr.Name) {
			s.logger.Log(
				"handler", "groupHandlers",
				"ProjectName", tpr.Name,
				"Plugin", fmt.Sprintf("%T", h),
			)

			pc := &s.Plugins.PluginClient
			if err := h(pc, tpr.Name); err != nil {
				fmt.Println("Error handling groupHandlers ", err)
				return err
			}
		}

	}

	return nil
}
