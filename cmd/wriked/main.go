package main

import (
	"log"
	"net/http"
	"os"

	wrikedaemon "github.com/clowre/wrike-token-daemon"
	wrikehttp "github.com/clowre/wrike-token-daemon/internal/wrikedhttp"
	"github.com/spf13/cobra"
)

func main() {

	var (
		clientID     string
		clientSecret string
		daemon       *wrikedaemon.Daemon

		httpPort int

		rootCmd = &cobra.Command{
			Use: "wriked",
			PersistentPreRun: func(cmd *cobra.Command, args []string) {
				daemon = wrikedaemon.NewDaemon(clientID, clientSecret)
				go daemon.StartPolling()
			},
		}
		httpServerCmd = &cobra.Command{
			Use: "http",
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := wrikehttp.StartServer(httpPort, daemon); err != nil && err != http.ErrServerClosed {
					return err
				}

				return nil
			},
		}
	)

	rootFlags := rootCmd.PersistentFlags()
	rootFlags.StringVarP(&clientID, "client-id", "I", "", "specify wrike app's client ID")
	rootFlags.StringVarP(&clientID, "client-secret", "S", "", "specify wrike app's client secret")

	if err := rootCmd.MarkPersistentFlagRequired("client-id"); err != nil {
		log.Fatalf("cannot make client-id a required flag: %v", err)
	}

	if err := rootCmd.MarkPersistentFlagRequired("client-secret"); err != nil {
		log.Fatalf("cannot make client-secret a required flag: %v", err)
	}

	httpFlags := httpServerCmd.PersistentFlags()
	httpFlags.IntVarP(&httpPort, "port", "P", 0, "specify the http server's port")

	rootCmd.AddCommand(httpServerCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(2)
	}
}
