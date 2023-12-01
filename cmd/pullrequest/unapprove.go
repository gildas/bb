package pullrequest

import (
	"context"
	"fmt"
	"os"

	"bitbucket.org/gildas_cherruel/bb/cmd/profile"
	"github.com/gildas/go-errors"
	"github.com/spf13/cobra"
)

var unapproveCmd = &cobra.Command{
	Use:               "unapprove",
	Short:             "unapprove a pullrequest",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: unapproveValidArgs,
	RunE:              unapproveProcess,
}

var unapproveOptions struct {
	Repository string
}

func init() {
	Command.AddCommand(unapproveCmd)

	unapproveCmd.Flags().StringVar(&unapproveOptions.Repository, "repository", "", "Repository to unapprove pullrequest from. Defaults to the current repository")
}

func unapproveValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var log = Log.Child(nil, "validargs")

	if profile.Current == nil {
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}

	log.Infof("Getting pullrequests for repository %s", approveOptions.Repository)
	var pullrequests struct {
		Values   []PullRequest `json:"values"`
		PageSize int           `json:"pagelen"`
		Size     int           `json:"size"`
		Page     int           `json:"page"`
	}

	err := profile.Current.Get(
		log.ToContext(context.Background()),
		unapproveOptions.Repository,
		"pullrequests?state=OPEN",
		&pullrequests,
	)
	if err != nil {
		log.Errorf("Failed to get pullrequests for repository %s", unapproveOptions.Repository, err)
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}

	var result []string
	for _, pullrequest := range pullrequests.Values {
		result = append(result, fmt.Sprintf("%d", pullrequest.ID))
	}
	return result, cobra.ShellCompDirectiveNoFileComp
}

func unapproveProcess(cmd *cobra.Command, args []string) (err error) {
	var log = Log.Child(nil, "unapprove")

	if profile.Current == nil {
		return errors.ArgumentMissing.With("profile")
	}

	log.Infof("Unapproving pullrequest %s", args[0])
	err = profile.Current.Delete(
		log.ToContext(context.Background()),
		unapproveOptions.Repository,
		fmt.Sprintf("pullrequests/%s/approve", args[0]),
		nil,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to unapprove pullrequest %s: %s\n", args[0], err)
		return nil
	}
	return
}
