package gitlabhook

//MergeRequestCommentEvent contains information needed to unmarshal the post from a "comment on merge request" gitlab hook
//https://docs.gitlab.com/ce/web_hooks/web_hooks.html#comment-on-merge-request
type MergeRequestCommentEvent struct {
	ObjectKind       string           `json:"object_kind,omitempty"`
	User             User             `json:"user,omitempty"`
	ProjectID        int              `json:"project_id,omitempty"`
	Project          Project          `json:"project,omitempty"`
	ObjectAttributes ObjectAttributes `json:"object_attributes,omitempty"`
	Repository       Repository       `json:"repository,omitempty"`
	MergeRequest     MergeRequest     `json:"merge_request"`
	//	Assignee         User             `json:"asignee"`
}

type User struct {
	Name      string `json:"name,omitempty"`
	Username  string `json:"username,omitempty"`
	AvatarUrl string `json:"avatar_url,omitempty"`
}

type Project struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	WebURL            string `json:"web_url,omitempty"`
	AvatarURL         string `json:"avatar_url,omitempty"`
	GitSSHURL         string `json:"git_ssh_url,omitempty"`
	GitHTTPURL        string `json:"git_http_url,omitempty"`
	Namespace         string `json:"namespace,omitempty"`
	VisibilityLevel   int    `json:"visibility_level,omitempty"`
	PathWithNamespace string `json:"path_with_namespace,omitempty"`
	DefaultBranch     string `json:"default_branch,omitempty"`
	Homepage          string `json:"homepage,omitempty"`
	URL               string `json:"url,omitempty"`
	SSHURL            string `json:"ssh_url,omitempty"`
	HTTPURL           string `json:"http_url,omitempty"`
}

type ObjectAttributes struct {
	ID                   int    `json:"id,omitempty"`
	Note                 string `json:"note"`
	NoteableType         string `json:"noteable_type,omitempty"`
	AuthorID             int    `json:"author_id,omitempty"`
	CreatedAT            string `json:"created_at,omitempty"`
	UpdatedAt            string `json:"updated_at,omitempty"`
	ProjectID            int    `json:"project_id,omitempty"`
	LineCode             string `json:"line_code"`
	CommitID             string `json:"commit_id"`
	NoteableId           int    `json:"noteable_id"`
	System               bool   `json:"system"`
	StDIff               string `json:"st_diff"`
	UpdatedByID          string `json:"updated_by_id"`
	Type                 string `json:"type"`
	Position             string `json:"position"`
	OriginalPosition     string `json:"original_position"`
	ResolvedAt           string `json:"resolved_at"`
	ResolvedById         string `json:"resolved_by_id"`
	DiscussionId         string `json:"discussion_id"`
	OriginalDiscussionID string `json:"original_discussion_id"`
	URL                  string `json:"url"`
}

type Repository struct {
	Name        string `json:"name,omitempty"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
	Homepage    string `json:"homepage,omitempty"`
}

type MergeRequest struct {
	ID              int    `json:"id"`
	TargetBranch    string `json:"target_branch"`
	SourceBranch    string `json:"source_branch"`
	SourceProjectID int    `json:"source_project_id"`
	AuthorID        int    `json:"author_id"`
	AssigneeID      int    `json:"assignee_id"`
	Title           string `json:"title"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	MilestoneID     int    `json:"milestone_id"`
	State           string `json:"state"`
	MergeStatus     string `json:"merge_status"`
	TargetProjectID int    `json:"target_project_id"`
	IID             int    `json:"iid"`
	Description     string `json:"description,omitempty"`
	Position        int    `json:"position"`
	LockedAt        string `json:"locked_at,omitempty"`
	UpdatedByID     int    `json:"updated_by_id"`
	MergeError      string `json:"merge_error"`
	MergeParams     struct {
		ForceRemoveSourceBranch string `json:"force_remove_source_branch"`
	} `json:"merge_params"`
	MergeWhenBuildSucceeds   bool       `json:"merge_when_build_succeeds"`
	MergeUserID              int        `json:"merge_user_id,omitempty"`
	MergeCommitSha           string     `json:"merge_commit_sha,omitempty"`
	DeletedAt                string     `json:"deleted_at,omitempty"`
	InProgressMergeCommitSha string     `json:"in_progress_merge_commit_sha,omitempty"`
	LockVersion              int        `json:"lock_version,omitempty"`
	Source                   Source     `json:"source"`
	Target                   Target     `json:"target"`
	LastCommit               LastCommit `json:"last_commit"`
	WorkInProgress           bool       `json:"work_in_progress"`
}

type Source struct {
	Name              string `json:"name,omitempty"`
	Description       string `json:"description,omitempty"`
	WebURL            string `json:"web_url,omitempty"`
	AvatarURL         string `json:"avatar_url,omitempty"`
	GitSSHURL         string `json:"git_ssh_url,omitempty"`
	GitHTTPURL        string `json:"git_http_url,omitempty"`
	Namespace         string `json:"namespace,omitempty"`
	VisibilityLevel   int    `json:"visibility_level"`
	PathWithNamespace string `json:"path_with_namespace,omitempty"`
	DefaultBranch     string `json:"default_branch,omitempty"`
	Homepage          string `json:"homepage,omitempty"`
	URL               string `json:"url,omitempty"`
	SSHURL            string `json:"ssh_url,omitempty"`
	HTTPURL           string `json:"http_url,omitempty"`
}

type Target struct {
	Name              string `json:"name,omitempty"`
	Description       string `json:"description,omitempty"`
	WebURL            string `json:"web_url,omitempty"`
	AvatarURL         string `json:"avatar_url,omitempty"`
	GitSSHURL         string `json:"git_ssh_url,omitempty"`
	GitHTTPURL        string `json:"git_http_url,omitempty"`
	Namespace         string `json:"namespace,omitempty"`
	VisibilityLevel   int    `json:"visibility_level"`
	PathWithNamespace string `json:"path_with_namespace,omitempty"`
	DefaultBranch     string `json:"default_branch,omitempty"`
	Homepage          string `json:"homepage,omitempty"`
	URL               string `json:"url,omitempty"`
	SSHURL            string `json:"ssh_url,omitempty"`
	HTTPURL           string `json:"http_url,omitempty"`
}

type LastCommit struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	URL       string `json:"url"`
	Author    struct {
		Name  string `json:"name"`
		Email string `json:"Email"`
	} `json:"author,omitempty"`
}
