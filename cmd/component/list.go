package component

import (
	"bitbucket.org/gildas_cherruel/bb/cmd/profile"
	"github.com/gildas/go-logger"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all components",
	Args:  cobra.NoArgs,
	RunE:  listProcess,
}

var listOptions struct {
	Repository string
}

func init() {
	Command.AddCommand(listCmd)

	listCmd.Flags().StringVar(&listOptions.Repository, "repository", "", "Repository to list components from. Defaults to the current repository")
}

func listProcess(cmd *cobra.Command, args []string) (err error) {
	log := logger.Must(logger.FromContext(cmd.Context())).Child(cmd.Parent().Name(), "list")

	log.Infof("Listing all issues from repository %s", listOptions.Repository)
	components, err := profile.GetAll[Component](cmd.Context(), cmd, "components")
	if err != nil {
		return err
	}
	if len(components) == 0 {
		log.Infof("No component found")
		return nil
	}
	return profile.Current.Print(cmd.Context(), cmd, Components(components))
}
