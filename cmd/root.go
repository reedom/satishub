// Copyright © 2018 HANAI Tohru <tohru@reedom.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	version  = "1.0.0"
	revision = ""
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "satishub",
	Short: "SatisHub",
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("version").Changed {
			fmt.Println("satishub  version:", version)
			if !strings.HasSuffix(version, revision) {
				fmt.Println("         revision:", revision)
			}
			return
		}
		cmd.Usage()
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	var appFlags = []struct {
		flagKey string
		envKey  string
		defVal  interface{}
		usage   string
	}{
		{"addr", "SATIS_HTTP_ADDR", ":80", "HTTP service server listen address"},
		{"tlsaddr", "SATIS_TLS_ADDR", ":443", "TLS(HTTPS) service server listen address"},
		{"no-http", "SATIS_NO_HTTP", false, "do not setup HTTP server"},
		{"no-tls", "SATIS_NO_TLS", false, "do not setup HTTPS server"},
		{"debug", "SATIS_DEBUG", false, "output verbose messages for debugging"},
	}
	for _, f := range appFlags {
		switch f.defVal.(type) {
		case string:
			RootCmd.PersistentFlags().String(f.flagKey, f.defVal.(string), f.usage)
		case bool:
			RootCmd.PersistentFlags().Bool(f.flagKey, f.defVal.(bool), f.usage)
		default:
			panic(fmt.Sprintf("Unhandled type: %v", f.defVal))
		}
		viper.BindPFlag(f.flagKey, RootCmd.Flag(f.flagKey))
		if f.envKey != "" {
			viper.BindEnv(f.flagKey, f.envKey)
		}
	}
	RootCmd.Flags().Bool("version", false, "show version")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName(".satishub") // name of config file (without extension)
	viper.AddConfigPath("$HOME")     // adding home directory as first search path
	if cfgFile != "" {               // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
