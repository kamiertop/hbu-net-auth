package cmd

import "github.com/spf13/cobra"

func LoginCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "login",
        Short: "登录校园网",
        Long: "使用账号密码登录校园网",
        RunE: func(cmd *cobra.Command, args []string) error {
            return nil
        },
    }
}
