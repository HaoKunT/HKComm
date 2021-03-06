package hkcmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"hkcomm"
)

/*
`This command used to create a superuser in the database`
 */

 var createSuperUser = &cobra.Command{
 	Use: "createSuperUser",
 	Short: "This command is used to create a super user",
 	Long: "This command is used to create a super user, it happened after you init your database",
 	Run: func(cmd *cobra.Command, args []string) {
		if err := hkcomm.CreateSuperUser(); err != nil {
			fmt.Println(err)
		}
	},
 }

func init()  {
	rootCmd.AddCommand(createSuperUser)
}