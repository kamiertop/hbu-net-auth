package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/kamiertop/hbu-net-auth/srun"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func LoginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "登录校园网",
		Long:  "使用账号密码登录校园网",
		RunE:  runLogin,
	}
}

var client = srun.NewClient(srun.Options{
	PortalURL: "https://gate.hbu.cn/srun_portal_pc?ac_id=1&theme=pro",
	APIBase:   "https://gate.hbu.edu.cn",
	CheckURL:  "https://www.baidu.com",
})

func runLogin(_ *cobra.Command, _ []string) error {
	err := client.CheckNetwork()
	if err == nil {
		color.New(color.BgCyan, color.FgGreen).Println("网络通畅，跳过登录") // nolint: errcheck
		return nil
	}

	username, err := readUsername()
	if err != nil {
		fmt.Fprintln(os.Stderr, "读取学号失败:", err)
		return err
	}

	password, err := readPassword()
	if err != nil {
		fmt.Fprintln(os.Stderr, "读取密码失败:", err)
		return err
	}
	resp, err := client.Login(username, password)
	if err != nil {
		fmt.Fprintln(os.Stderr, "登录失败:", err)
		return err
	}
	if !resp.Success() {
		fmt.Fprintln(os.Stderr, "登录失败:", resp.Message())
		return errors.New(resp.Message())
	}
	fmt.Println("登录成功：", resp.Message())

	if err := client.CheckNetwork(); err != nil {
		fmt.Fprintln(os.Stderr, "网络检查失败:", err)
		return nil
	}

	return nil
}

func readUsername() (string, error) {
	prompt := promptui.Prompt{
		Label: "请输入学号",
		Validate: func(input string) error {
			if len(input) != 11 {
				return errors.New("学号长度不对")
			}
			for _, r := range input {
				if r < '0' || r > '9' {
					return errors.New("学号必须是数字")
				}
			}
			return nil
		},
	}

	return prompt.Run()
}

func readPassword() (string, error) {
	prompt := promptui.Prompt{
		Label: "请输入密码",
		Mask:  '*',
		Validate: func(input string) error {
			if input == "" {
				return errors.New("密码不能为空")
			}
			return nil
		},
	}

	return prompt.Run()
}
