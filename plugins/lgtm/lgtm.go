package lgtm

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/go-kit/kit/log"

	"github.com/cosminilie/gitbot/gitlabhook"
	"github.com/cosminilie/gitbot/plugins"
	gitlab "github.com/xanzy/go-gitlab"
)

const (
	pluginName                           = "lgtm"
	ActionStrCreateMergeRequestNote      = "CreateNote"
	ActionStrCreateMergeRequest          = "CreateMergeRequest"
	ConditionStrCantBeMerged             = "MergeRequest.MergeStatus=cannot_be_merged"
	ConditionStrWorkInProgress           = "MergeRequest.WorkInProgress=true"
	ConditionStrState                    = "MergeRequest.State=closed"
	ConditionStrAuthorAndWantLGTM        = "isAuthor&&wantLGTM"
	ConditionsStrIsNotAproverAndWantLGTM = "!isApprover&&wantLGTM"
	ConditionsStrAllOK                   = "AllOK"
)

var (
	lgtmRe = regexp.MustCompile(`(?mi)^\/lgtm\r?$`)
)

func init() {
	plugins.RegisterMergeCommentEventHandler(pluginName, handleMergeRequestCommentHandler)
}

//LGTMError is an error struct which implements the error interface
type LGTMError struct {
	Repo      string
	Group     string
	User      string
	Action    string
	Condition string
	Result    error
}

func (e LGTMError) Error() string {
	return fmt.Sprintf("LGTMError:\nRepo:%s,\nGroup:%s,\nUser:%s,\nAction:%s,\nCondition:%s,\nResult:%s,\n", e.Repo, e.Group, e.User, e.Action, e.Condition, e.Result)
}

func handleMergeRequestCommentHandler(pc *plugins.PluginClient, ic gitlabhook.MergeRequestCommentEvent) error {
	//logging
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
		logger = log.NewContext(logger).With("plugin", "lgtm")

	}
	pc.Pmut.Lock()
	defer pc.Pmut.Unlock()
	if p, ok := pc.Repos[ic.Project.PathWithNamespace]; ok {
		logger.Log(
			"Func", "handleMergeRequestCommentHandler",
			"Approvers", strings.Join(p.Approvers, " "),
			"Name", p.Name,
			"Plugins", strings.Join(p.Plugins, " "),
		)

		return handle(logger, pc.GitLabClient, ic, p.Approvers)
	}

	return fmt.Errorf("Could not find any plugin for this repo: %v\n", pc.Repos)
}

func handle(logger log.Logger, gc *gitlab.Client, ic gitlabhook.MergeRequestCommentEvent, approversList []string) error {

	//hadle
	logger.Log(
		"Func", "handle",
		"Repo", ic.Project.Name,
		"Group", ic.Project.Namespace,
		"Reff", ic.Project.GitHTTPURL,
	)
	if ic.User.Username == "lgtm-bot" {
		logger.Log(
			"Func", "handle",
			"Repo", ic.Project.Name,
			"Group", ic.Project.Namespace,
			"User", ic.User.Username,
			"Action", "Comment author is lgtm-bot, skipping",
		)
		return nil
	}
	// Only consider open PRs.
	if ic.MergeRequest.WorkInProgress {

		//Create comment
		response := plugins.FormatResponse(ic, "LGTM plugin -> Can't merge as request is still work in progress.")
		gitComment := gitlab.CreateMergeRequestNoteOptions{
			Body: &response,
		}
		_, _, err := gc.Notes.CreateMergeRequestNote(ic.ProjectID, ic.MergeRequest.ID, &gitComment)
		if err != nil {
			myErr := LGTMError{
				Repo:      ic.Project.Name,
				Group:     ic.Project.Namespace,
				User:      ic.User.Username,
				Action:    ActionStrCreateMergeRequestNote,
				Condition: ConditionStrWorkInProgress,
				Result:    err,
			}
			return myErr
		}
		return nil

	}
	if ic.MergeRequest.State == "closed" {
		//Create comment
		response := plugins.FormatResponse(ic, "LGTM plugin -> Can't merge as request is Closed")
		gitComment := gitlab.CreateMergeRequestNoteOptions{
			Body: &response,
		}
		_, _, err := gc.Notes.CreateMergeRequestNote(ic.ProjectID, ic.MergeRequest.ID, &gitComment)
		if err != nil {
			myErr := LGTMError{
				Repo:      ic.Project.Name,
				Group:     ic.Project.Namespace,
				User:      ic.User.Username,
				Action:    ActionStrCreateMergeRequestNote,
				Condition: ConditionStrState,
				Result:    err,
			}
			return myErr
		}
		return nil

	}

	if ic.MergeRequest.MergeStatus == "cannot_be_merged" {
		//Create comment
		response := plugins.FormatResponse(ic, "LGTM plugin -> Can't merge as request has some problems. Status is `cannot_be_merged`")
		gitComment := gitlab.CreateMergeRequestNoteOptions{
			Body: &response,
		}
		_, _, err := gc.Notes.CreateMergeRequestNote(ic.ProjectID, ic.MergeRequest.ID, &gitComment)
		if err != nil {
			myErr := LGTMError{
				Repo:      ic.Project.Name,
				Group:     ic.Project.Namespace,
				User:      ic.User.Username,
				Action:    ActionStrCreateMergeRequestNote,
				Condition: ConditionStrCantBeMerged,
				Result:    err,
			}
			return myErr
		}

		return nil

	}

	// If we create an "/lgtm" comment, add lgtm if necameessary.
	wantLGTM := false
	if lgtmRe.MatchString(ic.ObjectAttributes.Note) {
		wantLGTM = true
	} else {
		logger.Log(
			"Func", "handle",
			"Repo", ic.Project.Name,
			"Group", ic.Project.Namespace,
			"User", ic.User.Username,
			"Action", "Create Merge comment",
			"Condition", "WantLGTM=false",
			"Result", "N/A",
		)
		return nil
	}

	var isApprover bool
	//Check if the person which submited the comment is the aproval list of people
	if userInApproverList(ic.User.Username, approversList) {
		isApprover = true
	}

	var isAuthor bool
	//Check if the same person which submited the merge request also submited the comment
	if ic.MergeRequest.AuthorID == ic.ObjectAttributes.AuthorID {
		isAuthor = true
	}

	if isAuthor && wantLGTM {

		//create comment
		response := plugins.FormatResponse(ic, "LGTM plugin -> You can't LGTM your own Merge Request")
		gitComment := gitlab.CreateMergeRequestNoteOptions{
			Body: &response,
		}
		_, _, err := gc.Notes.CreateMergeRequestNote(ic.ProjectID, ic.MergeRequest.ID, &gitComment)
		if err != nil {
			myErr := LGTMError{
				Repo:      ic.Project.Name,
				Group:     ic.Project.Namespace,
				User:      ic.User.Username,
				Action:    ActionStrCreateMergeRequestNote,
				Condition: ConditionStrAuthorAndWantLGTM,
				Result:    err,
			}

			return myErr
		}
		return nil

	} else if !isApprover && wantLGTM {

		response := plugins.FormatResponse(ic, "LGTM plugin -> You can't LGTM a PR unless you are an in the list of Approvers")
		gitComment := gitlab.CreateMergeRequestNoteOptions{
			Body: &response,
		}
		_, _, err := gc.Notes.CreateMergeRequestNote(ic.ProjectID, ic.MergeRequest.ID, &gitComment)
		if err != nil {
			myErr := LGTMError{
				Repo:      ic.Project.Name,
				Group:     ic.Project.Namespace,
				User:      ic.User.Username,
				Action:    ActionStrCreateMergeRequestNote,
				Condition: ConditionsStrIsNotAproverAndWantLGTM,
				Result:    err,
			}

			return myErr
		}
		return nil

	} else {
		//Add merge request comment
		response := plugins.FormatResponse(ic, "LGTM plugin -> All OK. Merging ...")
		gitComment := gitlab.CreateMergeRequestNoteOptions{
			Body: &response,
		}
		_, _, err := gc.Notes.CreateMergeRequestNote(ic.ProjectID, ic.MergeRequest.ID, &gitComment)
		if err != nil {
			myErr := LGTMError{
				Repo:      ic.Project.Name,
				Group:     ic.Project.Namespace,
				User:      ic.User.Username,
				Action:    ActionStrCreateMergeRequestNote,
				Condition: ConditionsStrAllOK,
				Result:    err}

			return myErr
		}
		//merge the request

		//Create merge request ops
		mergeRequestOpts := &gitlab.AcceptMergeRequestOptions{
			MergeCommitMessage:       gitlab.String(fmt.Sprintf("LGTM Plugin Merged Request based on Aproval from %s\n", ic.User.Username)),
			ShouldRemoveSourceBranch: gitlab.Bool(true),
			MergeWhenBuildSucceeds:   gitlab.Bool(true),
		}

		//Call accept merge request
		_, _, err = gc.MergeRequests.AcceptMergeRequest(ic.ProjectID, ic.MergeRequest.ID, mergeRequestOpts)
		if err != nil {
			myErr := LGTMError{
				Repo:      ic.Project.Name,
				Group:     ic.Project.Namespace,
				User:      ic.User.Username,
				Action:    ActionStrCreateMergeRequest,
				Condition: ConditionsStrAllOK,
				Result:    err,
			}

			return myErr
		}

	}

	logger.Log(
		"Func", "handle",
		"Repo", ic.Project.Name,
		"Group", ic.Project.Namespace,
		"User", ic.User.Username,
		"Action", "N/A",
		"Condition", "LGTM is ignored",
		"Result", "None",
	)

	return nil
}

func userInApproverList(a string, list []string) bool {
	for _, b := range list {
		if strings.EqualFold(b, a) {
			return true
		}
	}
	return false
}
