package comment

import (
	"fmt"

	"bitbucket.org/gildas_cherruel/bb/cmd/common"
	"bitbucket.org/gildas_cherruel/bb/cmd/profile"
	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all issue comments",
	Args:  cobra.NoArgs,
	RunE:  listProcess,
}

var listOptions struct {
	Repository string
	IssueID    common.RemoteValueFlag
}

func init() {
	Command.AddCommand(listCmd)

	listOptions.IssueID = common.RemoteValueFlag{AllowedFunc: GetIssueIDs}
	listCmd.Flags().StringVar(&listOptions.Repository, "repository", "", "Repository to list issue comments from. Defaults to the current repository")
	listCmd.Flags().Var(&listOptions.IssueID, "issue", "Issue to list comments from")
	_ = listCmd.MarkFlagRequired("issue")
	_ = listCmd.RegisterFlagCompletionFunc("issue", listOptions.IssueID.CompletionFunc())
}

func listProcess(cmd *cobra.Command, args []string) (err error) {
	log := logger.Must(logger.FromContext(cmd.Context())).Child(cmd.Parent().Name(), "list")

	if profile.Current == nil {
		return errors.ArgumentMissing.With("profile")
	}

	log.Infof("Listing all comments from repository %s with profile %s", listOptions.Repository, profile.Current)
	comments, err := profile.GetAll[Comment](
		cmd.Context(),
		cmd,
		profile.Current,
		fmt.Sprintf("issues/%s/comments", listOptions.IssueID.Value),
	)
	if err != nil {
		return err
	}
	if len(comments) == 0 {
		log.Infof("No issue found")
		return nil
	}
	return profile.Current.Print(
		cmd.Context(),
		Comments(core.Filter(comments, func(comment Comment) bool {
			return len(comment.Content.Raw) > 0
		})),
	)
}
