package cmd

import (
	"fmt"

	"github.com/nivista/steady/.gen/protos/common"
	"github.com/nivista/steady/.gen/protos/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {

	createHTTPCommand.Flags().StringVar(&cron, "cron", "@every 5s", "cron schedule for timer.")
	createHTTPCommand.Flags().IntVar(&maxExecutions, "max-executions", 5, "max executions of timer (default: 5, zero means infinite executions")
	createHTTPCommand.Flags().StringVar(&url, "url", "http://example.com", "url endpoint you want to hit (default example.com)")
	createHTTPCommand.Flags().BoolVar(&includeBody, "include-body", false, "whether or not to send the body to elasticsearch.")

	viper.BindPFlag("cron", createHTTPCommand.Flags().Lookup("cron"))
	viper.BindPFlag("max-executions", createHTTPCommand.Flags().Lookup("max-executions"))
	viper.BindPFlag("url", createHTTPCommand.Flags().Lookup("url"))
	viper.BindPFlag("include-body", createHTTPCommand.Flags().Lookup("include-body"))

	rootCmd.AddCommand(createHTTPCommand)
}

var (
	cron          string
	maxExecutions int
	url           string
	includeBody   bool

	createHTTPCommand = &cobra.Command{
		Use:   "create-http",
		Short: "Creates a new http timer.",
		Long:  "Creates a new http timer.",
		Run: func(cmd *cobra.Command, args []string) {
			req := services.CreateTimerRequest{
				Domain: domain,
				Task: &common.Task{
					Task: &common.Task_Http{
						Http: &common.HTTP{
							Url:              url,
							Method:           common.Method_GET,
							SaveResponseBody: includeBody,
						},
					},
				},
				Schedule: &common.Schedule{
					Cron:          cron,
					MaxExecutions: common.Executions(maxExecutions),
				},
			}

			res, err := client.CreateTimer(cmd.Context(), &req)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("ID:", res.TimerUuid)
			}

		},
	}
)
