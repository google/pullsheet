package cmd

import (
	"context"
	"fmt"
	"github.com/google/pullsheet/pkg/client"
	"github.com/google/pullsheet/pkg/server"
	"github.com/google/pullsheet/pkg/server/job"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

// serverCmd represents the subcommand for `pullsheet server`
var serverCmd = &cobra.Command{
	Use:           "server",
	Short:         "Serve leaderboard data with web UI",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer(rootOpts)
	},
}

type serverOptions struct {
	port int
}
var serverOpts = &serverOptions{}

func init() {
	serverCmd.Flags().IntVar(
		&serverOpts.port,
		"port",
		8080,
		"Port for server to listen on")

	rootCmd.AddCommand(serverCmd)
}

func runServer(rootOpts *rootOptions) error {
	ctx := context.Background()
	c, err := client.New(ctx, rootOpts.tokenPath)
	if err != nil {
		return err
	}

	// setup initial job
	j := job.New(
		job.Opts{
		Repos: rootOpts.repos,
		Users: rootOpts.users,
		Since: rootOpts.sinceParsed,
		Until: rootOpts.untilParsed,
		Title: rootOpts.title,
	})

	s := server.New(ctx, c, j)
	http.HandleFunc("/", s.Root())

	listenAddr := fmt.Sprintf(":%s", os.Getenv("PORT"))
	if listenAddr == ":" {
		listenAddr = fmt.Sprintf(":%d", serverOpts.port)
	}
	return http.ListenAndServe(listenAddr, nil)
}