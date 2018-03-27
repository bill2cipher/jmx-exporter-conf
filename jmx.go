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
	url   string
	beans map[string][]string
}

// NewJMX is used to build new JMX
func NewJMX(url string) *JMX {
	j := &JMX{
		url:   url,
		beans: make(map[string][]string),
	}
	if err := j.init(); err != nil {
		panic(err.Error())
	}
	return j
}

func (j *JMX) init() error {
	beans, err := j.allBeans()
	if err != nil {
		return err
	}
	for _, b := range beans {
		pair := strings.Split(b, ":")
		if len(pair) != 2 {
			mesg := fmt.Sprintf("bean %s format error", b)
			panic(mesg)
		}
		j.beans[pair[0]] = append(j.beans[pair[0]], pair[1])
	}
	return nil
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

func (j *JMX) Domains() []string {
	keys := make([]string, 0, len(j.beans))
	for k := range j.beans {
		keys = append(keys, k)
	}
	return keys
}

func (j *JMX) Beans(domain string) []string {
	if beans, exist := j.beans[domain]; exist {
		return beans
	}
	return nil
}

func (j *JMX) allBeans() ([]string, error) {
	if beans, err := j.execute("beans", j.url); err != nil {
		return nil, err
	} else {
		bs := strings.Split(beans, "\n")
		var result []string
		for _, b := range bs {
			if b != "" {
				result = append(result, b)
			}
		}
		return result, nil
	}
}
