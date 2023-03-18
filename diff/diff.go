package diff

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
)

// Diff returns the difference between two strings, using the diff command line tool.
func Diff(a, b string) (string, error) {
	afilePath, err := writeTempFile(a)
	if err != nil {
		return "", err
	}

	bFilePath, err := writeTempFile(b)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("diff", "--color=always", "-u", afilePath, bFilePath)
	out, _ := cmd.Output()
	return string(out), nil
}

func writeTempFile(content string) (string, error) {
	name := randString()
	err := os.WriteFile(os.TempDir()+"/"+name, []byte(content), 0644)
	if err != nil {
		return "", err
	}
	return os.TempDir() + "/" + name, nil

}

func randString() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
