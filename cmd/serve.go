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
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/reedom/satishub/api"
	"github.com/reedom/satishub/pkg/satis"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start satis service server",
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.New(os.Stdout, "service ", log.Ldate|log.Ltime)
		debug := viper.GetBool("debug")

		satisParam := satis.ServiceParam{
			SatisPath:   viper.GetString("satis"),
			ConfigPath:  viper.GetString("config"),
			RepoPath:    viper.GetString("repo"),
			Debug:       debug,
			Timeout:     time.Second * time.Duration(viper.GetInt("timeout")),
			SNSTopicARN: viper.GetString("sns-topic-arn"),
		}

		useHTTP := !viper.GetBool("no-http")
		useTLS := !viper.GetBool("no-tls")
		addr := viper.GetString("addr")
		tlsAddr := viper.GetString("tlsaddr")
		tlsCert := viper.GetString("tlscert")
		tlsKey := viper.GetString("tlskey")

		table := tablewriter.NewWriter(os.Stdout)
		table.Append([]string{"satis executable path", satisParam.SatisPath})
		table.Append([]string{"satis config file path", satisParam.ConfigPath})
		table.Append([]string{"output directory path", satisParam.RepoPath})
		if useHTTP {
			table.Append([]string{"HTTP listen address", addr})
		} else {
			table.Append([]string{"HTTP", "false"})
		}
		if useTLS {
			table.Append([]string{"HTTPS listen address", tlsAddr})
			table.Append([]string{"TLS certificate file path", tlsCert})
			table.Append([]string{"TLS secret key file path", tlsKey})
		} else {
			table.Append([]string{"HTTPS", "false"})
		}
		table.Render()

		if !useTLS && !useHTTP {
			log.Println("nothing to do")
			return
		}

		ctx, cancel := context.WithCancel(context.Background())

		service := satis.NewService(satisParam)
		go func() {
			stream := service.Run(ctx)
			for {
				select {
				case <-ctx.Done():
					return
				case result := <-stream:
					if result.Error != nil {
						logger.Printf("ERROR: %v", result.Error.Error())
						continue
					}
					if debug {
						fmt.Println("cmd done", result)
					}
				}
			}
		}()

		server := api.NewServer(service, logger, satisParam.Debug)

		errTLS := make(chan error)
		if useTLS {
			logger.Printf("start TLS server on %v", tlsAddr)
			go func() {
				errTLS <- server.ServeTLS(ctx, tlsAddr, tlsCert, tlsKey)
				close(errTLS)
			}()
		}

		errHTTP := make(chan error)
		if useHTTP {
			logger.Printf("start HTTP server on %v", addr)
			go func() {
				errHTTP <- server.Serve(ctx, addr)
				close(errHTTP)
			}()
		}

		interrupt := make(chan os.Signal)
		signal.Notify(interrupt, os.Interrupt)
		select {
		case err := <-errTLS:
			log.Println(err.Error())
			cancel()
		case err := <-errHTTP:
			log.Println(err.Error())
			cancel()
		case <-interrupt:
			cancel()
		}

		// wait for graceful shutdowns
		if !useHTTP {
			close(errHTTP)
		}
		if !useTLS {
			close(errTLS)
		}
		<-errHTTP
		<-errTLS
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)

	var appFlags = []struct {
		flagKey string
		envKey  string
		defVal  interface{}
		usage   string
	}{
		{"satis", "SATIS_EXEC_PATH", "satis", "satis executable path"},
		{"config", "SATIS_CONFIG_PATH", "satis.json", "satis config file path"},
		{"repo", "SATIS_REPO_PATH", "repo", "satis output directory path"},
		{"timeout", "SATIS_TIMEOUT", int(60 * 20), "satis build process timeout in seconds"},
		{"tlscert", "SATIS_TLS_CERT_PATH", "satis.crt", "TLS certificate file path"},
		{"tlskey", "SATIS_TLS_SECRET_KEY_PATH", "satis.key", "TLS secret key file path"},
		{"sns-topic-arn", "SATIS_SNS_TOPIC_ARN", "", "AWS Simple Notification Service ARN"},
	}
	for _, f := range appFlags {
		switch f.defVal.(type) {
		case string:
			serveCmd.Flags().String(f.flagKey, f.defVal.(string), f.usage)
		case int:
			serveCmd.Flags().Int(f.flagKey, f.defVal.(int), f.usage)
		case bool:
			serveCmd.Flags().Bool(f.flagKey, f.defVal.(bool), f.usage)
		default:
			panic(fmt.Sprintf("Unhandled type: %v", f.defVal))
		}
		viper.BindPFlag(f.flagKey, serveCmd.Flag(f.flagKey))
		if f.envKey != "" {
			viper.BindEnv(f.flagKey, f.envKey)
		}
	}
}
