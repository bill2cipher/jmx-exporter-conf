package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"gopkg.in/yaml.v2"
)

type rule struct {
	Pattern string            `yaml:"pattern"`
	Name    string            `yaml:"name"`
	Bean    *ViewBean         `yaml:"-"`
	Labels  map[string]string `yaml:"labels"`
}

type Conf struct {
	HostPort                  string  `yaml:"hostPort"`
	StartDelaySeconds         int     `yaml:"startDelaySeconds"`
	Ssl                       bool    `yaml:"ssl"`
	LowercaseOutputName       bool    `yaml:"lowercaseOutputName"`
	LowercaseOutputLabelNames bool    `yaml:"lowercaseOutputLabelNames"`
	Rules                     []*rule `yaml:"rules"`
}

func NewConf(url string) *Conf {
	c := new(Conf)
	c.HostPort = url
	c.StartDelaySeconds = 30
	c.Ssl = false
	c.LowercaseOutputLabelNames = false
	c.LowercaseOutputName = false
	return c
}

func (c *Conf) addRule(bean *ViewBean) {
	if c.removeIfExist(bean) {
		return
	}
	labels := make(map[string]string)
	for _, l := range bean.labels {
		if !l.used {
			continue
		}
		labels[l.name] = fmt.Sprintf("$%d", l.index)
	}
	c.Rules = append(c.Rules, &rule{Pattern: c.parsePattern(bean), Name: bean.domain, Bean: bean, Labels: labels})
}

func (c *Conf) removeIfExist(bean *ViewBean) bool {
	var result []*rule
	exist := false
	for _, v := range c.Rules {
		if v.Bean != bean {
			result = append(result, v)
		} else {
			exist = true
		}
	}
	c.Rules = result
	return exist
}

func (c *Conf) parsePattern(bean *ViewBean) string {
	var result bytes.Buffer
	var pairs []string
	result.WriteString(bean.domain)
	result.WriteByte('<')
	for _, l := range bean.labels {
		pairs = append(pairs, fmt.Sprintf("%s=\"([\\w-]*)\"", l.name))
	}
	result.WriteString(strings.Join(pairs, ","))
	result.WriteString("><>([^:]*):(.*)")
	return result.String()
}

func (c *Conf) dump() (string, error) {
	if d, err := yaml.Marshal(c); err != nil {
		return "", err
	} else {
		return string(d), nil
	}
}

func (c *Conf) save() error {
	content, err := c.dump()
	if err != nil {
		return err
	} else if err := clipboard.WriteAll(content); err == nil {
		return nil
	} else if file, err := os.OpenFile("./conf.yaml", os.O_WRONLY|os.O_CREATE, 0666); err != nil {
		return err
	} else if _, err := io.WriteString(file, content); err != nil {
		return err
	} else {
		return errors.New("save to clipboard failed, save into file conf.yaml instead")
	}
}
