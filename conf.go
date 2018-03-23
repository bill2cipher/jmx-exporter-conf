package main

import "gopkg.in/yaml.v2"

type Conf struct {
	HostPort                  string `yaml:"hostPort"`
	StartDelaySeconds         int    `yaml:"startDelaySeconds"`
	Ssl                       bool   `yaml:"ssl"`
	LowercaseOutputName       bool   `yaml:"lowercaseOutputName"`
	LowercaseOutputLabelNames bool   `yaml:"lowercaseOutputLabelNames"`
	Rules                     []struct {
		Pattern string `yaml:"pattern"`
		name    string `yaml:"name"`
	} `yaml:"rules"`
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

func (c *Conf) dump() (string, error) {
	if d, err := yaml.Marshal(c); err != nil {
		return "", err
	} else {
		return string(d), nil
	}
}
