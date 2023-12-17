package comment

import (
	"context"
	"fmt"
	"os"

	"bitbucket.org/gildas_cherruel/bb/cmd/common"
	"bitbucket.org/gildas_cherruel/bb/cmd/profile"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:               "get",
	Aliases:           []string{"show", "info", "display"},
	Short:             "get an issue comment",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getValidArgs,
	RunE:              getProcess,
}

var getOptions struct {
	IssueID    common.RemoteValueFlag
	Repository string
}

func init() {
	Command.AddCommand(getCmd)

	getOptions.IssueID = common.RemoteValueFlag{AllowedFunc: func(ctx context.Context) []string {
		// maybe we should store the Command in the flag and use it here
		return GetIssueIDs(ctx, profile.Current, getOptions.Repository) // not sure the profile is set at this point
	}}
	getCmd.Flags().StringVar(&getOptions.Repository, "repository", "", "Repository to get an issue comment from. Defaults to the current repository")
	getCmd.Flags().Var(&getOptions.IssueID, "issue", "Issue to get comments from. Defaults to the current issue")
	_ = getCmd.MarkFlagRequired("issue")
	_ = getCmd.RegisterFlagCompletionFunc("issue", getOptions.IssueID.CompletionFunc())
}

func getValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if profile.Current == nil {
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}
	return GetIssueCommentIDs(cmd.Context(), profile.Current, getOptions.Repository, getOptions.IssueID.Value), cobra.ShellCompDirectiveNoFileComp
}

func getProcess(cmd *cobra.Command, args []string) (err error) {
	log := logger.Must(logger.FromContext(cmd.Context())).Child(cmd.Parent().Name(), "get")

	if profile.Current == nil {
		return errors.ArgumentMissing.With("profile")
	}

	log.Infof("Displaying issue %s", args[0])
	var comment Comment

	err = profile.Current.Get(
		log.ToContext(cmd.Context()),
		getOptions.Repository,
		fmt.Sprintf("issues/%d/comment/%s", getOptions.IssueID, args[0]),
		&comment,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get issue comment %s: %s\n", args[0], err)
		os.Exit(1)
	}
	return profile.Current.Print(cmd.Context(), comment)
}
