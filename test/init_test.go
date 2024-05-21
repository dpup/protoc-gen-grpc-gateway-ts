package test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var projectRoot = ""

func init() {
	wd, _ := os.Getwd()
	for !strings.HasSuffix(wd, "protoc-gen-grpc-gateway-ts") {
		wd = filepath.Dir(wd)
	}
	projectRoot = wd
}

func runTsc() cmdResult {
	cmd := exec.Command("npx", "tsc", "--project", ".", "--noEmit")
	cmd.Dir = projectRoot + "/test/testdata/"

	cmdOutput := new(bytes.Buffer)
	cmdError := new(bytes.Buffer)
	cmd.Stderr = cmdOutput
	cmd.Stdout = cmdError

	err := cmd.Run()

	return cmdResult{
		stdout:   cmdOutput.String(),
		stderr:   cmdError.String(),
		err:      err,
		exitCode: cmd.ProcessState.ExitCode(),
	}
}

type cmdResult struct {
	stdout   string
	stderr   string
	err      error
	exitCode int
}

func createTestFile(fname, content string) {
	f, err := os.Create(projectRoot + "/test/testdata/" + fname)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.WriteString(content); err != nil {
		panic(err)
	}
}

func removeTestFile(fname string) {
	if err := os.Remove(projectRoot + "/test/testdata/" + fname); err != nil {
		panic(err)
	}
}
