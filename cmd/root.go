package cmd

/*
Copyright © 2022 dariuszSki dsliwinski@aol.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	conn              *grpc.ClientConn
	cfgFile           = flag.String("cIdentity", "", "Ziti Client Identity file")
	sIdentity         = flag.String("sIdentity", "", "Optional Ziti Server Identity if you require a specific destination")
	service           = flag.String("service", "", "Ziti Service")
	addressByIdentity = flag.Bool("addressByIdentity", false, "Enable addressable identity")
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "grpc-echo",
		Short: "grpc-echo: echo message app",
		Long: `It is an application that a client can test a network using a message and get a response back from a server;
i.e. client ==message==>server==message==>client.`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(cfgFile, "config", "", "identity config file (default is $HOME/.netfoundry/identities/.identity.json)")
	rootCmd.PersistentFlags().StringVar(service, "service", "", "Ziti Service")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if *cfgFile != "" {
		// client/server identity file.
		viper.SetConfigFile(*cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".grpc-echo" (without extension).
		viper.AddConfigPath(home + "/.netfoundry/identities")
		viper.SetConfigType("json")
		viper.SetConfigName(".identity")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		_, err := fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		if err != nil {
			return
		}
	}
}
