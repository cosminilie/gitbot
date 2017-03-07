A gitlab bot used to enforce change control.

[![Build Status](https://travis-ci.org/cosminilie/gitbot.svg?branch=master "Build Status")](https://travis-ci.org/cosminilie/gitbot)

## Overview
Supports resetting permissions for groups. Add's lgtm support in merge requests comments based on a predefined list of approvers. 

Example config (using hcl - https://github.com/hashicorp/hcl) since it more clear to express the repo object in my opinion vs ini style or json/toml,etc):
```
token = ""
git-api-URL = "https://gitlab.company.net/api/v3/"
default-approvers = ["user1","user2","user3"]

repo "monitoring_group/test1" {
plugins = ["lgtm"]
approvers = ["user5"]
}

repo "tools" {
plugins = ["lgtm","drop_rights"]
approvers = ["user6"]
}
```

Assuming you are in the default-approvers or on the approvers list on the group definition using ```/lgtm``` the bot will merge the request. You will also need to create a gitlab user with administrator rights (probably can have more granular rights) and provide the token for that user in the config file.  

Heavily inspired by: https://github.com/kubernetes/test-infra/tree/master/prow

## Not ready for production
The project is missing tests so don't use this in production! It was a side project to understand how the kubernetes bot works and learn the gitlab api. With some work (removing some racing conditions) and a more clear implementation of the group plugins it can get there, but I don't have the time right now. It will also need a way to add tags after a merge request so you can get an artifact out based on the tag (make releases based on tags). Still need to figure out a way on how to do this best, either by adding a new keyword e.g. "/tag 0.1.1" though a comment on the already merged "merge request" or though some other way. This would be simple to add as it's very similar to the lgtm plugin.

## Building locally

You will need to have docker installed and available in your path. To generate RPM's run ```make docker-rpm```. 
Alternatively you can run make help to get the supported list of operations. 

