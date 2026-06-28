package cmd

import "github.com/spf13/cobra"

func LogoutCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "logout",
        Short: "登出校园网",
        Long:  "使用学号登出校园网",
        RunE: func(cmd *cobra.Command, args []string) error {
            return nil
        },
    }
}
