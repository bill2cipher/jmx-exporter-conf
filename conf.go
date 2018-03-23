package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"gopkg.in/yaml.v2"
)

type rule struct {
	Pattern string `yaml:"pattern"`
	Name    string `yaml:"name"`
	content string `ymal:"-"`
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

func (c *Conf) addRule(r string) {
	if c.removeIfExist(r) {
		return
	}
	c.Rules = append(c.Rules, &rule{Pattern: c.parsePattern(r), Name: r, content: r})
}

func (c *Conf) removeIfExist(r string) bool {
	var result []*rule
	exist := false
	for _, v := range c.Rules {
		if v.content != r {
			result = append(result, v)
		} else {
			exist = true
		}
	}
	c.Rules = result
	return exist
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
	} else {
		return clipboard.WriteAll(content)
	}
}

func (c *Conf) parsePattern(r string) string {
	var result bytes.Buffer
	var pairs []string
	rule := strings.Split(r, ":")
	result.WriteString(rule[0])
	result.WriteByte('<')
	values := strings.Split(rule[1], ",")
	for _, v := range values {
		pairs = append(pairs, fmt.Sprintf("%s=\"(.*)\"", strings.Split(v, "=")[0]))
	}
	result.WriteString(strings.Join(pairs, ","))
	result.WriteString("><>([^:]*):\\s(.*)")
	return result.String()
}
