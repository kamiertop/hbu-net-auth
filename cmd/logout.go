package cmd

import (
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func LogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "登出校园网",
		Long:  "使用学号登出校园网",
		Run: func(cmd *cobra.Command, args []string) {
			_ = runLogout()
		},
	}
}

func runLogout() error {
	if err := client.CheckNetwork(); err != nil {
		color.New(color.BgCyan, color.FgGreen).Println("网络通畅，跳过登录") // nolint: errcheck
		return nil
	}
	username, err := readUsername()
	if err != nil {
		return err
	}

	resp, err := client.Logout(username)
	if err != nil {
		fmt.Println("登出失败:", err)
		return err
	}
	if !resp.Success() {
		fmt.Println("登出失败:", resp.Message())
		return errors.New(resp.Message())
	}
	fmt.Println("登出成功:", resp.Message())
	return nil
}
