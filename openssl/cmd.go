package openssl

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func GetCertsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "share", "amfora"), nil
}

func GetPageDir(dir string, url string) (string, error) {
	sp := strings.Split(url, "//")
	if len(sp) <= 1 || sp[1] == ""{
		return "", errors.New("not a proper url")
	}
	pageDir := filepath.Join(dir, sp[1])

	return pageDir, nil
}

func CallOpenSSL(pageName string, userName string, expireDays int) error {
	if expireDays == 0 {
		expireDays = 1825
	}
	daysStr := strconv.Itoa(expireDays)

	dir, err := GetCertsDir()
	if err != nil {
		return err
	}

	pageDir, err := GetPageDir(dir, pageName)
	if err != nil {
		return err
	}

	// Based on:
	// openssl req -new -subj "/CN=username" -x509 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -days 1825 -nodes -out cert.pem -keyout key.pem
	cmd := exec.Command(
		"openssl",
		"req",
		"-new",
		"-subj", fmt.Sprint("/CN=", userName),
		"-x509",
		"-newkey", "ec",
		"-pkeyopt", "ec_paramgen_curve:prime256v1",
		"-days", daysStr,
		"-nodes",
		"-out", filepath.Join(pageDir, "cert.pem"),
		"-keyout", filepath.Join(pageDir, "key.pem"),
	)

	stdout, stderr := cmd.CombinedOutput()

	if stderr != nil {
		return errors.New(fmt.Sprint(string(stdout), ": ", stderr.Error()))
	}

	return nil
}
