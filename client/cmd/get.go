package cmd

import (
	"fmt"

	"github.com/nivista/steady/.gen/protos/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {

	getCommand.Flags().StringVar(&id, "id", "", "id of timer.")

	viper.BindPFlag("id", getCommand.Flags().Lookup("id"))

	rootCmd.AddCommand(getCommand)
}

var (
	getCommand = &cobra.Command{
		Use:   "get",
		Short: "Gets a timer.",
		Long:  "Gets a timer.",
		Run: func(cmd *cobra.Command, args []string) {
			req := services.GetTimerRequest{
				Domain:    domain,
				TimerUuid: id,
			}

			res, err := client.GetTimer(cmd.Context(), &req)
			if err != nil {
				fmt.Println("err:", err)
			} else {
				fmt.Println(res)
			}
		},
	}
)
