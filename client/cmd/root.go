package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/nivista/steady/.gen/protos/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var (
	// Used for flags.
	cfgFile string
	addr    string
	domain  string

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

	rootCmd.PersistentFlags().StringVar(&addr, "addr", "localhost:8080", "address of steady server to connect to.")

	viper.BindPFlag("addr", rootCmd.PersistentFlags().Lookup("addr"))

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		var err error
		conn, err = grpc.Dial("localhost:8080", grpc.WithInsecure())
		if err != nil {
			return err
		}
		client = services.NewSteadyClient(conn)
		return nil
	}
	var err error
	conn, err = grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		//return err
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
