package droprights

import (
	"fmt"
	"os"
	"strings"

	"github.com/cosminilie/gitbot/plugins"
	"github.com/go-kit/kit/log"
	gitlab "github.com/xanzy/go-gitlab"
)

const (
	pluginName = "drop_rights"
)

var (
	groupList []*gitlab.Group
)

func init() {
	plugins.RegisterGroupHandler(pluginName, dropRights)
}

//LGTMError is an error struct which implements the error interface
type DropRightsError struct {
	Repo      string
	Group     string
	User      string
	Action    string
	Condition string
	Result    error
}

func (e DropRightsError) Error() string {
	return fmt.Sprintf("DropRightsError:\nRepo:%s,\nGroup:%s,\nUser:%s,\nAction:%s,\nCondition:%s,\nResult:%s,\n", e.Repo, e.Group, e.User, e.Action, e.Condition, e.Result)
}

func dropRights(pc *plugins.PluginClient, ic string) error {
	//logging
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
		logger = log.NewContext(logger).With("plugin", "drop_rights")
	}
	pc.Pmut.Lock()
	defer pc.Pmut.Unlock()

	return handle(logger, pc.GitLabClient, ic)

}

func handle(logger log.Logger, cl *gitlab.Client, mr string) error {
	logger.Log(
		"Repo", mr,
		"Plugin", pluginName,
	)
	upgradeGroupMemberOpts := &gitlab.UpdateGroupMemberOptions{
		AccessLevel: gitlab.AccessLevel(gitlab.ReporterPermissions),
	}

	//get group ID
	listGroupOpts := &gitlab.ListGroupsOptions{
		Search: gitlab.String(""),
	}

	groups, _, err := cl.Groups.ListGroups(listGroupOpts)
	if err != nil {
		myErr := DropRightsError{
			Repo:      "",
			Group:     mr,
			User:      "",
			Action:    "ListGroups",
			Condition: "",
			Result:    err,
		}
		return myErr
	}
	groupList = append(groupList, groups...)

	groupMembers, _, err := cl.Groups.ListGroupMembers(mr)
	if err != nil {
		myErr := DropRightsError{
			Repo:      "",
			Group:     mr,
			User:      "",
			Action:    "ListGroupMembers",
			Condition: "",
			Result:    err,
		}
		return myErr

	}

	for _, m := range groupMembers {
		if m.AccessLevel >= gitlab.DeveloperPermissions {
			if id := groupID(mr); id != 0 {
				logger.Log(
					"Repo", mr,
					"Plugin", pluginName,
					"Action", "UpdateGroupMember",
					"Username", m.Username,
				)
				_, _, err := cl.Groups.UpdateGroupMember(id, m.ID, upgradeGroupMemberOpts)
				if err != nil {
					myErr := DropRightsError{
						Repo:      "",
						Group:     mr,
						User:      m.Username,
						Action:    "UpdateGroupMembership",
						Condition: "",
						Result:    err,
					}
					return myErr
				}

			}

		}
	}

	return nil
}
func groupID(name string) int {
	for _, g := range groupList {
		if strings.EqualFold(g.Name, name) {
			return g.ID
		}
	}
	return 0
}
