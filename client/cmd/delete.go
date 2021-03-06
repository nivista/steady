package cmd

import (
	"fmt"

	"github.com/nivista/steady/.gen/protos/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {

	deleteCommand.Flags().StringVar(&id, "id", "", "id of timer.")

	viper.BindPFlag("id", deleteCommand.Flags().Lookup("id"))

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
				TimerUuid: id,
			}

			_, err := client.DeleteTimer(cmd.Context(), &req)
			if err != nil {
				fmt.Println("err:", err)
			} else {
				fmt.Println("OK")
			}
		},
	}
)
