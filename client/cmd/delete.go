package cmd

import (
	"github.com/nivista/steady/.gen/protos/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {

	createHTTPCommand.Flags().StringVar(&id, "id", "", "id of timer.")

	viper.BindPFlag("id", rootCmd.Flags().Lookup("id"))
	rootCmd.AddCommand(deleteCommand)
}

var (
	id string

	deleteCommand = &cobra.Command{
		Use:   "delete",
		Short: "Deletes a timer.",
		Long:  "Deletes a timer.",
		Run: func(cmd *cobra.Command, args []string) {
			req := services.DeleteTimerRequest{
				Domain:    domain,
				TimerUuid: id,
			}

			client.DeleteTimer(cmd.Context(), &req)
		},
	}
)
