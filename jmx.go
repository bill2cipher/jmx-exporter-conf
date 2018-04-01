package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// JMXLabel represents labels
type JMXLabel struct {
	Name, Value string
	Index       int
}

// JMXBean represents a bean
type JMXBean struct {
	Domain, Name     string
	Labels           []*JMXLabel
	ValueName, Value string
}

// JMX represents jmx command
type JMX struct {
	url   string
	beans map[string]map[string]*JMXBean
}

// NewJMX is used to build new JMX
func NewJMX(url string) *JMX {
	j := &JMX{
		url:   url,
		beans: make(map[string]map[string]*JMXBean),
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

	reg := regexp.MustCompile(`([^<>]+)<([^<>]+)><([^<>]*)>([^:]+)(.+)`)
	for _, b := range beans {
		parts := reg.FindStringSubmatch(b)
		if len(parts) != 6 {
			mesg := fmt.Sprintf("bean %s format error %d:%v", b, len(parts), parts)
			panic(mesg)
		}
		parts = parts[1:]
		jb := &JMXBean{
			Domain:    parts[0],
			Name:      parts[1],
			ValueName: parts[3],
			Value:     parts[4],
			Labels:    j.buildLabels(parts[1]),
		}
		if _, exist := j.beans[parts[0]]; !exist {
			j.beans[parts[0]] = make(map[string]*JMXBean)
		}
		j.beans[parts[0]][parts[1]] = jb
	}
	return nil
}

func (j *JMX) buildLabels(name string) []*JMXLabel {
	parts := strings.Split(name, ",")
	var labels []*JMXLabel
	for i, p := range parts {
		values := strings.SplitN(p, "=", 2)
		labels = append(labels, &JMXLabel{Name: strings.TrimSpace(values[0]), Value: values[1], Index: i + 1})
	}
	return labels
}

func (j *JMX) execute(url string) (string, error) {
	var stdOut, stdErr bytes.Buffer
	cmd := fmt.Sprintf("java -jar ./jmx_dump.jar %s", url)
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

func (j *JMX) Beans(domain string) []*JMXBean {
	var result []*JMXBean
	if beans, exist := j.beans[domain]; exist {
		for _, b := range beans {
			result = append(result, b)
		}
	}
	return result
}

func (j *JMX) allBeans() ([]string, error) {
	if beans, err := j.execute(j.url); err != nil {
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
