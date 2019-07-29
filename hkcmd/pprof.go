/**
* @Author: HaoKunT
* @Date: 2019/7/24 6:43
* @File: pprof.go
*/
package hkcmd

import (
	"github.com/spf13/cobra"
	"log"
	"net/http"
	_ "net/http/pprof"
)

var pprofCmd = &cobra.Command{
	Use: "pprof",
	Short: "Open pprof web tool",
	Long: "Open pprof web tool on :10000",
	Run: func(cmd *cobra.Command, args []string) {
		go func() {
			log.Println(http.ListenAndServe("localhost:10000", nil))
		}()
		rootCmd.Run(cmd, args)
	},
}

func init()  {
	rootCmd.AddCommand(pprofCmd)
}
