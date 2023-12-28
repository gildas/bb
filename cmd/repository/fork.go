package repository

import (
	"fmt"

	"bitbucket.org/gildas_cherruel/bb/cmd/common"
	"bitbucket.org/gildas_cherruel/bb/cmd/profile"
	"bitbucket.org/gildas_cherruel/bb/cmd/project"
	"bitbucket.org/gildas_cherruel/bb/cmd/workspace"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/spf13/cobra"
)

type RepositoryForkCreator struct {
	Name        string                    `json:"name"                  mapstructure:"name"`
	Description string                    `json:"description,omitempty" mapstructure:"description"`
	Project     *project.ProjectReference `json:"project,omitempty"     mapstructure:"project"`
	IsPrivate   bool                      `json:"is_private"            mapstructure:"is_private"`
	Language    string                    `json:"language,omitempty"    mapstructure:"language"`
	MainBranch  *branch                   `json:"mainbranch,omitempty"  mapstructure:"mainbranch"`
	ForkPolicy  string                    `json:"fork_policy,omitempty" mapstructure:"fork_policy"`
}

var forkCmd = &cobra.Command{
	Use:               "fork [flags] <slug_or_uuid>",
	Short:             "fork a repository by its <slug> or <uuid>.",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: forkValidArgs,
	RunE:              forkProcess,
}

var forkOptions struct {
	Workspace   common.RemoteValueFlag
	Project     common.RemoteValueFlag
	Name        string
	Description string
	Public      bool
	Private     bool
	Language    string
	MainBranch  string
	ForkPolicy  common.EnumFlag
}

func init() {
	Command.AddCommand(forkCmd)

	forkOptions.Workspace = common.RemoteValueFlag{AllowedFunc: workspace.GetWorkspaceSlugs}
	forkOptions.Project = common.RemoteValueFlag{AllowedFunc: project.GetProjectKeys}
	forkOptions.ForkPolicy = common.EnumFlag{Allowed: []string{"allow_forks", "no_public_forks", "no_forks"}, Value: "no_public_forks"}
	forkCmd.Flags().Var(&forkOptions.Workspace, "workspace", "Workspace to fork repositories from")
	forkCmd.Flags().Var(&forkOptions.Project, "project", "Project to fork repositories from")
	forkCmd.Flags().StringVar(&forkOptions.Name, "name", "", "Name of the repository")
	forkCmd.Flags().StringVar(&forkOptions.Description, "description", "", "Description of the repository")
	forkCmd.Flags().BoolVar(&forkOptions.Private, "private", false, "make the repository private")
	forkCmd.Flags().BoolVar(&forkOptions.Public, "public", false, "make the repository public")
	forkCmd.Flags().StringVar(&forkOptions.Language, "language", "", "Language of the repository")
	forkCmd.Flags().StringVar(&forkOptions.MainBranch, "main-branch", "", "Main branch of the repository")
	forkCmd.Flags().Var(&forkOptions.ForkPolicy, "fork-policy", "Fork policy of the repository. Default: no_public_forks")
	forkCmd.MarkFlagsMutuallyExclusive("private", "public")
	_ = forkCmd.MarkFlagRequired("workspace")
	_ = forkCmd.RegisterFlagCompletionFunc("workspace", forkOptions.Workspace.CompletionFunc())
	_ = forkCmd.RegisterFlagCompletionFunc("project", forkOptions.Project.CompletionFunc())
}

func forkValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return GetRepositorySlugs(cmd.Context(), cmd, profile.Current, forkOptions.Workspace.String()), cobra.ShellCompDirectiveNoFileComp
}

func forkProcess(cmd *cobra.Command, args []string) error {
	log := logger.Must(logger.FromContext(cmd.Context())).Child(cmd.Parent().Name(), "fork")

	if profile.Current == nil {
		return errors.ArgumentMissing.With("profile")
	}

	payload := RepositoryForkCreator{
		Name:        forkOptions.Name,
		Description: forkOptions.Description,
		Language:    forkOptions.Language,
		IsPrivate:   forkOptions.Private,
		ForkPolicy:  forkOptions.ForkPolicy.Value,
	}
	if len(forkOptions.MainBranch) > 0 {
		payload.MainBranch = &branch{Type: "branch", Name: forkOptions.MainBranch}
	}
	if len(forkOptions.Project.Value) > 0 {
		payload.Project = project.NewReference(forkOptions.Project.Value)
	}

	log.Infof("Forking repository %s", args[0])
	var forked Repository

	err := profile.Current.Post(
		log.ToContext(cmd.Context()),
		cmd,
		fmt.Sprintf("/repositories/%s/%s/forks", forkOptions.Workspace, args[0]),
		payload,
		&forked,
	)
	if err != nil {
		return err
	}
	return profile.Current.Print(cmd.Context(), forked)
}
