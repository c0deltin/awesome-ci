package utils

import (
	"fmt"
	"os/exec"
)

func CMD(cmd string, shell bool) string {
	if shell {
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			fmt.Println(err)
		}
		return string(out)
	}

	out, err := exec.Command(cmd).Output()
	if err != nil {
		fmt.Println(err)
	}

	return string(out)
}
