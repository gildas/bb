package project

import (
	"fmt"

	"bitbucket.org/gildas_cherruel/bb/cmd/profile"
	"bitbucket.org/gildas_cherruel/bb/cmd/workspace"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-flags"
	"github.com/gildas/go-logger"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all projects",
	Args:  cobra.NoArgs,
	RunE:  listProcess,
}

var listOptions struct {
	Workspace *flags.EnumFlag
}

func init() {
	Command.AddCommand(listCmd)

	listOptions.Workspace = flags.NewEnumFlagWithFunc("", workspace.GetWorkspaceSlugs)
	listCmd.Flags().Var(listOptions.Workspace, "workspace", "Workspace to list projects from")
	_ = listCmd.RegisterFlagCompletionFunc("workspace", listOptions.Workspace.CompletionFunc("workspace"))
}

func listProcess(cmd *cobra.Command, args []string) (err error) {
	log := logger.Must(logger.FromContext(cmd.Context())).Child(cmd.Parent().Name(), "list")

	if profile.Current == nil {
		return errors.ArgumentMissing.With("profile")
	}
	if len(listOptions.Workspace.Value) == 0 {
		listOptions.Workspace.Value = profile.Current.DefaultWorkspace
		if len(listOptions.Workspace.Value) == 0 {
			return errors.ArgumentMissing.With("workspace")
		}
	}

	log.Infof("Listing all projects from workspace %s with profile %s", listOptions.Workspace, profile.Current)
	projects, err := profile.GetAll[Project](
		cmd.Context(),
		cmd,
		profile.Current,
		fmt.Sprintf("/workspaces/%s/projects", listOptions.Workspace),
	)
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		log.Infof("No project found")
		return nil
	}
	return profile.Current.Print(cmd.Context(), cmd, Projects(projects))
}
