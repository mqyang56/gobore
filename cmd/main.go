package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mqyang56/gobore"
)

var clientArgs struct {
	localPort uint16
	localHost string
	to        string
	port      uint16
	secret    string
}

var serverArgs struct {
	minPort uint16
	secret  string
}

func main() {
	gobore.InitLogger()

	var clientCmd = &cobra.Command{
		Use: "client",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := gobore.NewClient(clientArgs.localHost, clientArgs.localPort, clientArgs.to, clientArgs.port, clientArgs.secret)
			if err != nil {
				zap.L().Error("Failed to NewClient", zap.Error(err))
				return
			}
			err = c.Listen()
			if err != nil {
				zap.L().Error("Failed to Listen", zap.Error(err))
				return
			}
		},
	}
	clientCmd.Flags().Uint16Var(&clientArgs.localPort, "local-port", 0, "The local port to expose.")
	clientCmd.Flags().StringVar(&clientArgs.localHost, "local-host", "", "The local host to expose.")
	clientCmd.Flags().StringVar(&clientArgs.to, "to", "", "Address of the remote server to expose local ports to.")
	clientCmd.Flags().Uint16Var(&clientArgs.port, "port", 0, "Optional port on the remote server to select")
	clientCmd.Flags().StringVar(&clientArgs.secret, "secret", "", "Optional secret for authentication")

	var serverCmd = &cobra.Command{
		Use: "server",
		Run: func(cmd *cobra.Command, args []string) {
			err := gobore.NewServer(serverArgs.minPort, serverArgs.secret).Listen()
			if err != nil {
				zap.L().Error("Failed to NewServer", zap.Error(err))
				return
			}
		},
	}
	serverCmd.Flags().Uint16Var(&serverArgs.minPort, "min-port", 1024, "Minimum TCP port number to accept")
	serverCmd.Flags().StringVar(&serverArgs.secret, "secret", "", "Optional secret for authentication")

	var rootCmd = &cobra.Command{}
	rootCmd.AddCommand(clientCmd)
	rootCmd.AddCommand(serverCmd)
	err := rootCmd.Execute()
	if err != nil {
		zap.L().Error("Failed to Execute", zap.Error(err))
		return
	}
}
