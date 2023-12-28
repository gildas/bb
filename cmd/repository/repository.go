package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"bitbucket.org/gildas_cherruel/bb/cmd/common"
	"bitbucket.org/gildas_cherruel/bb/cmd/profile"
	"bitbucket.org/gildas_cherruel/bb/cmd/project"
	"bitbucket.org/gildas_cherruel/bb/cmd/user"
	"bitbucket.org/gildas_cherruel/bb/cmd/workspace"
	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/spf13/cobra"
)

type Repository struct {
	Type                 string              `json:"type"               mapstructure:"type"`
	ID                   common.UUID         `json:"uuid"               mapstructure:"uuid"`
	Name                 string              `json:"name"               mapstructure:"name"`
	FullName             string              `json:"full_name"          mapstructure:"full_name"`
	Slug                 string              `json:"slug"               mapstructure:"slug"`
	Owner                user.Account        `json:"owner"              mapstructure:"owner"`
	Workspace            workspace.Workspace `json:"workspace"          mapstructure:"workspace"`
	Project              project.Project     `json:"project"            mapstructure:"project"`
	HasIssues            bool                `json:"has_issues"         mapstructure:"has_issues"`
	HasWiki              bool                `json:"has_wiki"           mapstructure:"has_wiki"`
	IsPrivate            bool                `json:"is_private"         mapstructure:"is_private"`
	ForkPolicy           string              `json:"fork_policy"        mapstructure:"fork_policy"`
	Size                 int64               `json:"size"               mapstructure:"size"`
	Language             string              `json:"language,omitempty" mapstructure:"language"`
	MainBranch           string              `json:"-"                  mapstructure:"-"`
	DefaultMergeStrategy string              `json:"-"                  mapstructure:"-"`
	BranchingModel       string              `json:"-"                  mapstructure:"-"`
	Parent               *Repository         `json:"parent"             mapstructure:"parent"`
	Links                common.Links        `json:"links"              mapstructure:"links"`
	CreatedOn            time.Time           `json:"created_on"         mapstructure:"created_on"`
	UpdatedOn            time.Time           `json:"updated_on"         mapstructure:"updated_on"`
}

/*
type repositorySettings struct {
	DefaultMergeStrategy bool `json:"default_merge_strategy" mapstructure:"default_merge_strategy"`
	BranchingModel       bool `json:"branching_model"        mapstructure:"branching_model"`
}
*/

type branch struct {
	Type string `json:"type" mapstructure:"type"`
	Name string `json:"name" mapstructure:"name"`
}

// Command represents this folder's command
var Command = &cobra.Command{
	Use:     "repo",
	Aliases: []string{"repository"},
	Short:   "Manage repositories",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Workspace requires a subcommand:")
		for _, command := range cmd.Commands() {
			fmt.Println(command.Name())
		}
	},
}

// GetHeader gets the header for a table
//
// implements common.Tableable
func (repository Repository) GetHeader(short bool) []string {
	return []string{"ID", "Name", "Full Name"}
}

// GetRow gets the row for a table
//
// implements common.Tableable
func (repository Repository) GetRow(headers []string) []string {
	return []string{
		repository.ID.String(),
		repository.Name,
		repository.FullName,
	}
}

// String returns the string representation of the repository
//
// implements fmt.Stringer
func (repository Repository) String() string {
	return repository.FullName
}

// GetRepositorySlugs gets the slugs of all repositories
func GetRepositorySlugs(context context.Context, cmd *cobra.Command, currentProfile *profile.Profile, workspace string) (slugs []string) {
	log := logger.Must(logger.FromContext(context)).Child("repository", "slugs")

	repositories, err := profile.GetAll[Repository](context, cmd, currentProfile, fmt.Sprintf("/repositories/%s", workspace))
	if err != nil {
		log.Errorf("Failed to get repositories", err)
		return
	}
	return core.Map(repositories, func(repository Repository) string {
		return repository.Slug
	})
}

// MarshalJSON implements the json.Marshaler interface.
//
// Implements json.Marshaler
func (repository Repository) MarshalJSON() (data []byte, err error) {
	type surrogate Repository
	var owner *user.Account
	var wspace *workspace.Workspace
	var proj *project.Project
	var br *branch
	var createdOn string
	var updatedOn string

	if !repository.Owner.ID.IsNil() {
		owner = &repository.Owner
	}
	if !repository.Workspace.ID.IsNil() {
		wspace = &repository.Workspace
	}
	if !repository.Project.ID.IsNil() {
		proj = &repository.Project
	}
	if len(repository.MainBranch) > 0 {
		br = &branch{Type: "branch", Name: repository.MainBranch}
	}
	if !repository.CreatedOn.IsZero() {
		createdOn = repository.CreatedOn.Format("2006-01-02T15:04:05.999999999-07:00")
	}
	if !repository.UpdatedOn.IsZero() {
		updatedOn = repository.UpdatedOn.Format("2006-01-02T15:04:05.999999999-07:00")
	}

	data, err = json.Marshal(struct {
		surrogate
		Owner      *user.Account        `json:"owner,omitempty"`
		Workspace  *workspace.Workspace `json:"workspace,omitempty"`
		Project    *project.Project     `json:"project,omitempty"`
		MainBranch *branch              `json:"mainbranch,omitempty"`
		CreatedOn  string               `json:"created_on,omitempty"`
		UpdatedOn  string               `json:"updated_on,omitempty"`
	}{
		surrogate:  surrogate(repository),
		Owner:      owner,
		Workspace:  wspace,
		Project:    proj,
		MainBranch: br,
		CreatedOn:  createdOn,
		UpdatedOn:  updatedOn,
	})
	return data, errors.JSONMarshalError.Wrap(err)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// Implements json.Unmarshaler
func (repository *Repository) UnmarshalJSON(data []byte) (err error) {
	type surrogate Repository
	var inner struct {
		surrogate
		MainBranch branch `json:"mainbranch"`
	}
	if err = json.Unmarshal(data, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*repository = Repository(inner.surrogate)
	repository.MainBranch = inner.MainBranch.Name
	return nil
}
