package cmd

/*
	This is HKComm Server
 */

import (
	"fmt"
	"github.com/kataras/iris"
	"github.com/spf13/cobra"
	"./HKComm"
)

func Execute()  {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

var rootCmd = &cobra.Command{
	Use: "HKComm-server",
	Short: "HKComm server is an IM server",
	Long: "HKComm server provide the cross Internet communication.",

	Run: func(cmd *cobra.Command, args []string) {
		// close
		iris.RegisterOnInterrupt(func() {
			hkcomm.SafeExit()
		})
		defer hkcomm.SafeExit()
		hkcomm.Server()
	},
}