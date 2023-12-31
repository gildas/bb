package artifact

import (
	"fmt"
	"os"

	"bitbucket.org/gildas_cherruel/bb/cmd/profile"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:               "download [flags] <filename>",
	Aliases:           []string{"get", "fetch"},
	Short:             "download an artifact by its <filename>.",
	ValidArgsFunction: downloadValidArgs,
	Args:              cobra.ExactArgs(1),
	RunE:              getProcess,
}

var downloadOptions struct {
	Repository  string
	Destination string
}

func init() {
	Command.AddCommand(downloadCmd)

	downloadCmd.Flags().StringVar(&downloadOptions.Repository, "repository", "", "Repository to download artifacts from. Defaults to the current repository")
	downloadCmd.Flags().StringVar(&downloadOptions.Destination, "destination", "", "Destination folder to download the artifact to. Defaults to the current folder")
	_ = downloadCmd.MarkFlagDirname("destination")
}

func downloadValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if profile.Current == nil {
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}
	return GetArtifactNames(cmd.Context(), cmd, profile.Current), cobra.ShellCompDirectiveNoFileComp
}

func getProcess(cmd *cobra.Command, args []string) error {
	log := logger.Must(logger.FromContext(cmd.Context())).Child(cmd.Parent().Name(), "download")

	if profile.Current == nil {
		return errors.ArgumentMissing.With("profile")
	}

	log.Infof("Downloading artifact %s to %s", args[0], downloadOptions.Destination)

	err := profile.Current.Download(
		log.ToContext(cmd.Context()),
		cmd,
		fmt.Sprintf("downloads/%s", args[0]),
		downloadOptions.Destination,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to download artifact %s: %s\n", args[0], err)
		os.Exit(1)
	}
	return nil
}
