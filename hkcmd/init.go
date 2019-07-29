package hkcmd

/*
	Init the database
 */

import (
	"fmt"
	"github.com/spf13/cobra"
	"hkcomm"
)

var initCmd = &cobra.Command{
	Use: "init",
	Short: "Init database",
	Long: "Init database which HKComm need.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := hkcomm.InitDatabase(); err != nil {
			fmt.Println(err)
		}
	},
}

func init()  {
	rootCmd.AddCommand(initCmd)
}