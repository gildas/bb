package branch

import (
	"bitbucket.org/gildas_cherruel/bb/cmd/profile"
	"github.com/gildas/go-logger"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all branches",
	Args:  cobra.NoArgs,
	RunE:  listProcess,
}

var listOptions struct {
	Repository string
}

func init() {
	Command.AddCommand(listCmd)

	listCmd.Flags().StringVar(&listOptions.Repository, "repository", "", "Repository to list branches from. Defaults to the current repository")
}

func listProcess(cmd *cobra.Command, args []string) (err error) {
	log := logger.Must(logger.FromContext(cmd.Context())).Child(cmd.Parent().Name(), "list")

	log.Infof("Listing all branches for repository: %s", listOptions.Repository)
	branches, err := GetBranches(log.ToContext(cmd.Context()), cmd)
	if err != nil {
		return err
	}
	if len(branches) == 0 {
		log.Infof("No branch found")
		return
	}
	return profile.Current.Print(cmd.Context(), cmd, Branches(branches))
}
