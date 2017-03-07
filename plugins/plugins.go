package plugins

import (
	"strings"
	"sync"

	"github.com/cosminilie/gitbot/gitlabhook"

	"github.com/go-kit/kit/log"
	gitlab "github.com/xanzy/go-gitlab"
)

var (
	allPlugins                = map[string]struct{}{}
	mergeCommentEventHandlers = map[string]MergeCommentEventHandler{}
	groupHandlers             = map[string]GroupHandler{}
)

//Repo struct in loading HCL configuration. Part of the config struct
type Repo struct {
	Name      string   `hcl:",key"`
	Plugins   []string `hcl:"plugins"`
	Approvers []string `hcl:"approvers"`
}

type GroupHandler func(*PluginClient, string) error

func RegisterGroupHandler(name string, fn GroupHandler) {
	allPlugins[name] = struct{}{}
	groupHandlers[name] = fn
}

// PluginClient may be used concurrently, so each entry must be thread-safe.
type PluginClient struct {
	GitLabClient *gitlab.Client
	Repos        map[string]Repo
	Pmut         sync.Mutex
}

//MergeCommentEventHandler func that handle merge request comments
type MergeCommentEventHandler func(*PluginClient, gitlabhook.MergeRequestCommentEvent) error

//RegisterMergeCommentEventHandler registers MergeCommentEventHandler in the global handler register
func RegisterMergeCommentEventHandler(name string, fn MergeCommentEventHandler) {
	allPlugins[name] = struct{}{}
	mergeCommentEventHandlers[name] = fn
}

//NewPluginAgent creates a new plugin agent
func NewPluginAgent(logger log.Logger, gci *gitlab.Client, pluginReposChan chan Repo) *PluginAgent {
	agent := &PluginAgent{}
	agent.PluginClient.GitLabClient = gci
	agent.logger = logger
	agent.PluginClient.Repos = make(map[string]Repo)
	agent.Repos = make(map[string]Repo)
	agent.GroupRepos = make(map[string]Repo)

	go func() {
		agent.mut.Lock()
		defer agent.mut.Unlock()
		for k, v := range expandRepo(pluginReposChan) {
			if _, ok := agent.Repos[k]; !ok {
				agent.Repos[k] = v
				agent.PluginClient.Repos[k] = v
			}
		}
	}()
	//Copy repos to plugin Client
	//	agent.PluginClient.Repos = agent.Repos
	return agent
}

func expandRepo(pluginReposChan chan Repo) map[string]Repo {
	m := make(map[string]Repo)
	for i := range pluginReposChan {
		m[i.Name] = i
	}
	return m

}

//PluginAgent is the main struct which store the information needed to associated plugins with repo and pass PluginClients to each registered handler
type PluginAgent struct {
	PluginClient
	mut        sync.Mutex
	Repos      map[string]Repo
	GroupRepos map[string]Repo
	logger     log.Logger
}

// GlobalHandlers returns a map of plugin names to apply for all repos without waiting for events.
func (pa *PluginAgent) GroupHandlers(repo string) map[string]GroupHandler {
	pa.mut.Lock()
	defer pa.mut.Unlock()

	hs := map[string]GroupHandler{}
	for _, p := range pa.getPlugins(repo) {
		pa.logger.Log(
			"handler", "GlobalHandlers",
			"Plugin", p,
		)

		if h, ok := groupHandlers[p]; ok {
			pa.logger.Log(
				"handler", "GlobalHandlers",
				"Plugin", p,
				"Action", "AddingHandlerforPlugin",
			)
			hs[p] = h
		}
	}

	return hs
}

// StatusEventHandlers returns a map of plugin names to merge request comments handler for the repo.
func (pa *PluginAgent) MergeCommentEventHandlers(repo string) map[string]MergeCommentEventHandler {
	pa.mut.Lock()
	defer pa.mut.Unlock()

	hs := map[string]MergeCommentEventHandler{}
	for _, p := range pa.getPlugins(repo) {
		pa.logger.Log(
			"handler", "MergeCommentEventHandlers",
			"Plugin", p,
		)

		if h, ok := mergeCommentEventHandlers[p]; ok {
			pa.logger.Log(
				"handler", "MergeCommentEventHandlers",
				"Plugin", p,
				"Action", "AddingHandlerforPlugin",
			)
			hs[p] = h
		}
	}

	return hs
}

// getPlugins returns a list of plugins that are enabled on a given (org, repository).
func (pa *PluginAgent) getPlugins(repo string) []string {
	var plugins []string

	plugs := pa.Repos[repo]
	groupPlugs := pa.GroupRepos[repo]
	plugins = append(plugins, plugs.Plugins...)
	plugins = append(plugins, groupPlugs.Plugins...)

	pa.logger.Log(
		"plugins", strings.Join(plugins, ", "),
		"Repo", repo,
	)

	return plugins
}
