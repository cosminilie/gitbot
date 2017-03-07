package gitbot

import (
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/cosminilie/gitbot/gitlabhook"
	"github.com/cosminilie/gitbot/plugins"

	"github.com/go-kit/kit/log"
	gitlab "github.com/xanzy/go-gitlab"
)

var (
	fullRepo = regexp.MustCompile(`^[a-zA-Z0-9-]+(\/|\/\*)?$`)
	//Right now we identify the boot hook URL by http://ip_addr:9091/hook
	botHook        = regexp.MustCompile(`^http(s)?\:\/\/((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\:9091\/hook$`)
	errUnknownType = errors.New("can't decode gitlab event type")
)

// Service interface
type Service interface {
	GitHook(logger log.Logger, data interface{})
	GetErrors() chan error
}

//RecuringHandlers func
type recuringHandlers func(s *basicService) error

//NewBasicService creates a new basic service. It also performs the necesary steps to setup everything:
//* Expands groups repos to have a complete list of repos. Groups repos are then sent to grouphandlers while individual repos are handled by normal event or time based triggers.
func NewBasicService(logger log.Logger, gcl *gitlab.Client, repos []plugins.Repo, defaultApprovers []string) *basicService {

	var pluginReposChan = make(chan plugins.Repo)
	var groupReposChan = make(chan plugins.Repo)

	logger = log.NewContext(logger).With("Context", "basic_service")

	service := &basicService{
		logger:  logger,
		ErrorCh: make(chan error),
	}

	//Load repos and expand the groups. We also send groups to the groupReposChan while all repos(already completed ones) and the ones we expand from the group are sent to groupReposChan
	go func() {
		err := fanOutRepos(logger, gcl, repos, defaultApprovers, pluginReposChan, groupReposChan)
		if err != nil {
			service.sendError(err)
			return
		}

	}()

	//load plugin agent
	go func() {
		plugin := plugins.NewPluginAgent(logger, gcl, pluginReposChan)
		service.Plugins = plugin

	}()

	//sets up group handlers
	//This is a time intensive operation so we try to run this async and have the service return faster.
	go func() {
		setupGroupHandlers(service, groupReposChan)
	}()

	//start loop to periodic refresh.
	go func() {
		//TO DO: duration should be an global parameter from conf
		service.scheduleHandlersEvery(1*time.Minute, groupHandlers, addRepoEventHook)
	}()
	return service
}

//BasicService implements the Service interface
type basicService struct {
	Plugins *plugins.PluginAgent
	logger  log.Logger
	mut     sync.Mutex
	ErrorCh chan error
}

//Runs an error channel that is used to fan out all the errors from basic service implementation
func (svc *basicService) GetErrors() chan error {
	return svc.ErrorCh
}

//GitHook is called on each git hook
func (svc *basicService) GitHook(logger log.Logger, data interface{}) {
	switch t := data.(type) {
	case gitlabhook.MergeRequestCommentEvent:
		err := svc.handleMergeRequestCommentEvent(logger, t)
		svc.sendError(err)
	default:
		logger.Log(
			"Handler", "GitHook",
			"Error", errUnknownType,
		)
	}
}

//utility functions to trigger periodic handlers.
func (svc *basicService) scheduleHandlersEvery(d time.Duration, rh ...recuringHandlers) {
	//https://golang.org/ref/spec#Passing_arguments_to_..._parameters
	if rh == nil {
		svc.sendError(errors.New("scheduleHandlersEvery, recurringHandlers is nil"))
	}
	for _ = range time.Tick(d) {
		for _, h := range rh {
			svc.logger.Log(
				"Action", "RecurringSchedule",
				"handler", fmt.Sprintf("%T", h),
			)
			if err := h(svc); err != nil {
				svc.sendError(err)
			}
		}

	}
}

//sendError function used to send errors to the error channel
func (svc *basicService) sendError(err error) {
	select {
	case svc.ErrorCh <- err:
	default:
	}
}

//handleMergeRequestCommentEvent is called when new merge comment events happen
func (svc *basicService) handleMergeRequestCommentEvent(logger log.Logger, se gitlabhook.MergeRequestCommentEvent) error {

	for _, h := range svc.Plugins.MergeCommentEventHandlers(se.Project.PathWithNamespace) {
		logger.Log(
			"handler", "handleMergeRequestCommentEvent",
			"ProjectName", se.Project.Name,
			"Plugin", fmt.Sprintf("%T", h),
		)
		pc := &svc.Plugins.PluginClient
		//pc.Repos = s.Plugins.Repos
		if err := h(pc, se); err != nil {
			fmt.Println("Error handling handleMergeRequestCommentEvent.", err)
			return err
		}
	}
	return nil
}
