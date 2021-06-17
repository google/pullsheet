// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"flag"
	"fmt"
	"time"

	"github.com/karrick/tparse"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
)

const dateForm = "2006-01-02"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "pullsheet",
	Long: `pullsheet - Generate spreadsheets based on GitHub contributions

pullsheet generates a CSV (comma separated values) & HTML output about GitHub activity across a series of repositories.`,
	PersistentPreRunE: initCommand,
}

type rootOptions struct {
	repos       []string
	users       []string
	since       string
	until       string
	sinceParsed time.Time
	untilParsed time.Time
	title       string
	tokenPath   string
	branches    []string
	out         string
}

var rootOpts = &rootOptions{}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		klog.Fatal(err)
	}
}

func init() {
	klog.InitFlags(nil)

	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	rootCmd.PersistentFlags().StringSliceVar(
		&rootOpts.repos,
		"repos",
		[]string{},
		"comma-delimited list of repositories. ex: kubernetes/minikube, google/pullsheet",
	)

	rootCmd.PersistentFlags().StringSliceVar(
		&rootOpts.branches,
		"branches",
		[]string{},
		"comma-delimited list of branches ex: master,main,head",
	)
	rootCmd.PersistentFlags().StringSliceVar(
		&rootOpts.users,
		"users",
		[]string{},
		"comma-delimiited list of users",
	)

	rootCmd.PersistentFlags().StringVar(
		&rootOpts.since,
		"since",
		"now-90d",
		"when to query from (date or duration)",
	)

	rootCmd.PersistentFlags().StringVar(
		&rootOpts.until,
		"until",
		"now",
		"when to query till (date or duration)",
	)

	rootCmd.PersistentFlags().StringVar(
		&rootOpts.title,
		"title",
		"",
		"Title to use for output pages",
	)

	rootCmd.PersistentFlags().StringVar(
		&rootOpts.tokenPath,
		"token-path",
		"",
		"GitHub token path",
	)

	rootCmd.PersistentFlags().StringVar(
		&rootOpts.out,
		"out",
		"CSV",
		"Output type - CSV/JSON. Default is CSV",
	)

	// Set up viper flag handling
	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		panic(err)
	}
}

// initRootOpts sets up root options, using env variables to set options if
// they haven't been set by flags
func initRootOpts() error {
	// Set up viper environment variable handling
	viper.SetEnvPrefix("pullsheet")
	envKeys := []string{
		"repos", "branches", "users", "since", "until", "title", "token-path", "out",
	}
	for _, key := range envKeys {
		if err := viper.BindEnv(key); err != nil {
			return err
		}
	}

	if rootOpts.out != "JSON" && rootOpts.out != "CSV" {
		return fmt.Errorf("Invalid out parameter %s. Must be JSON or CSV", rootOpts.out)
	}

	// Set options. viper will prioritize flags over env variables
	rootOpts.repos = viper.GetStringSlice("repos")
	rootOpts.branches = viper.GetStringSlice("branches")
	rootOpts.users = viper.GetStringSlice("users")
	rootOpts.since = viper.GetString("since")
	rootOpts.until = viper.GetString("until")
	rootOpts.title = viper.GetString("title")
	rootOpts.tokenPath = viper.GetString("token-path")
	rootOpts.out = viper.GetString("out")
	return nil
}

func initCommand(*cobra.Command, []string) error {
	if err := initRootOpts(); err != nil {
		return err
	}

	var err error

	t, err := tparse.ParseNow(dateForm, rootOpts.since)
	if err == nil {
		rootOpts.sinceParsed = t
	} else {
		klog.Infof("%q not a duration: %v", rootOpts.since, err)
		rootOpts.sinceParsed, err = time.Parse(dateForm, rootOpts.since)
		if err != nil {
			return errors.Wrap(err, "since time parse")
		}
	}

	rootOpts.untilParsed = time.Now()
	if rootOpts.since != "" {
		t, err := tparse.ParseNow(dateForm, rootOpts.until)
		if err == nil {
			rootOpts.untilParsed = t
		} else {
			klog.Infof("%q not a duration: %v", rootOpts.until, err)
			rootOpts.untilParsed, err = time.Parse(dateForm, rootOpts.until)
			if err != nil {
				return errors.Wrap(err, "until time parse")
			}
		}
	}

	return nil
}
