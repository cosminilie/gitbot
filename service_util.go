package gitbot

import (
	"errors"
	"net"
	"strings"

	"github.com/cosminilie/gitbot/plugins"
	"github.com/go-kit/kit/log"

	gitlab "github.com/xanzy/go-gitlab"
)

func fanOutRepos(logger log.Logger, cl *gitlab.Client, repos []plugins.Repo, defaultApprovers []string, reposChan chan plugins.Repo, groupReposChan chan plugins.Repo) error {
	defer func() {
		close(reposChan)
		close(groupReposChan)
	}()
	for _, r := range repos {

		if fullRepo.MatchString(r.Name) {

			//cleanup namespace
			s := r.Name
			s = strings.TrimSuffix(s, "/*")
			s = strings.TrimSuffix(s, "/")
			//found a group, sent to global handlers

			logger.Log(
				"FanOutRepos", "groupRepos",
				"Repo", s,
			)
			//send top level repos to groupReposChan
			r.Name = s
			groupReposChan <- r

			//get members
			pr, _, err := cl.Groups.ListGroupProjects(s)
			if err != nil {
				return err
			}
			if len(pr) > 0 {
				for _, p := range pr {
					//create new repo struct
					rep := plugins.Repo{
						Name:      strings.Replace(p.NameWithNamespace, " ", "", -1),
						Plugins:   r.Plugins,
						Approvers: append(r.Approvers, defaultApprovers...),
					}
					logger.Log(
						"Handler", "fan_out_repos",
						"ProjectName", rep.Name,
						"Aprovers", strings.Join(rep.Approvers, " "),
						"Plugin", strings.Join(rep.Plugins, " "),
					)

					//append to existing list
					reposChan <- rep

				}
			}
			continue
		}
		//don't need to expand
		r.Approvers = append(r.Approvers, defaultApprovers...)
		reposChan <- r
	}
	return nil
}

//taken from util/helper.go
func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
