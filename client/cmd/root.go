package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/nivista/steady/.gen/protos/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	// Used for flags.
	cfgFile             string
	addr                string
	apiToken, apiSecret string

	conn   *grpc.ClientConn
	client services.SteadyClient

	rootCmd = &cobra.Command{
		Use:   "steady",
		Short: "CLI interface to steady.",
		Long:  `CLI interface to steady. long`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	rootCmd.PersistentFlags().StringVar(&addr, "addr", "127.0.0.1:8080", "address of steady server to connect to.")
	rootCmd.PersistentFlags().StringVar(&apiToken, "apitoken", "", "api key to authenticate requests.")
	rootCmd.PersistentFlags().StringVar(&apiSecret, "apisecret", "", "api key to authenticate requests.")
	viper.BindPFlag("addr", rootCmd.PersistentFlags().Lookup("addr"))

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	client = services.NewSteadyClient(conn)
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return nil //conn.Close()
	}
}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			er(err)
		}

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func basicAuthCtx(ctx context.Context, username, password string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "Authorization", "Basic "+basicAuth(username, password))
}

// See 2 (end of page 4) https://www.ietf.org/rfc/rfc2617.txt
// "To receive authorization, the client sends the userid and password,
// separated by a single colon (":") character, within a base64
// encoded string in the credentials."
// It is not meant to be urlencoded.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
