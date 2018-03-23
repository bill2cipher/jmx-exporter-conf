package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// JMX represents jmx command
type JMX struct {
	url string
}

// NewJMX is used to build new JMX
func NewJMX(url string) *JMX {
	return &JMX{
		url: url,
	}
}

func (j *JMX) execute(cmd, url string) (string, error) {
	var stdOut, stdErr bytes.Buffer
	cmd = fmt.Sprintf("echo %s | java -jar ./jmxterm.jar -l %s -n -v silent", cmd, url)
	process := exec.Command("bash", "-c", cmd)
	process.Stdout = &stdOut
	process.Stderr = &stdErr
	if err := process.Run(); err != nil {
		return "", errors.New(err.Error() + ":" + stdErr.String())
	}
	return stdOut.String(), nil
}

func (j *JMX) domains() ([]string, error) {
	if domains, err := j.execute("domains", j.url); err != nil {
		return nil, err
	} else {
		return strings.Split(domains, "\n"), nil
	}
}

func (j *JMX) beans() ([]string, error) {
	if beans, err := j.execute("beans", j.url); err != nil {
		return nil, err
	} else {
		bs := strings.Split(beans, "\n")
		return bs, nil
	}
}
