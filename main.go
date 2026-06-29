package main

import (
	"fmt"
	"os"

	"github.com/kamiertop/hbu-net-auth/cmd"
	"github.com/spf13/cobra"
)

func main() {
	root := cobra.Command{
		Use:           "hbu-net-auth",
		Short:         "河北大学校园网上网认证CLI工具",
		Long:          "河北大学校园网上网认证CLI工具，支持登录、登出、查询状态等功能",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return nil
		},
	}
	root.AddCommand(cmd.LoginCommand())

	root.AddCommand(cmd.LogoutCommand())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
