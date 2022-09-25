package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const configDir = ".net-manager"

type Config struct {
	Station     Station `yaml:"station"`
	NetDir      string  `yaml:"net-log-directory"`
	HospitalDir string  `yaml:"hospital-log-directory"`
	MailingList string  `yaml:"mailing-list"`
	TimeReport  struct {
		MainMail string `yaml:"main-mail"`
		CcMail   string `yaml:"cc-mail"`
	} `yaml:"time-report"`
}

type Station struct {
	Call      string `yaml:"call"`
	Mail      Mail   `yaml:"mail"`
	Signature string `yaml:"signature"`
}

type Mail struct {
	SmtpHost string `yaml:"smtp-host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Email    string `yaml:"email"`
}

func parseConfig(data []byte) (*Config, error) {
	config := &Config{}
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func openFile(fileName string) (f *os.File, err error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find user home directory: %v\n", err)
		fmt.Fprintf(os.Stderr, "Trying config file in the working directory\n")
		goto workingDir
	}
	f, err = os.Open(filepath.Join(userHomeDir, configDir, fileName))
	if err == nil {
		return f, nil
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read config from home dir: %v\n", err)
		fmt.Fprintf(os.Stderr, "Trying config file in the working directory\n")
	}
workingDir:
	f, err = os.Open(fileName)
	return f, err
}

func readConfig() (config *Config) {
	userHomeDir, err := os.UserHomeDir()
	var data []byte
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find user home directory: %v\n", err)
		fmt.Fprintf(os.Stderr, "Trying config file in the working directory\n")
		goto workingDir
	}
	data, err = ioutil.ReadFile(filepath.Join(userHomeDir, configDir, "net-manager.conf"))
	if err == nil {
		goto parse
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read config from home dir: %v\n", err)
		fmt.Fprintf(os.Stderr, "Trying config file in the working directory\n")
	}
workingDir:
	data, err = ioutil.ReadFile(".net-manager.conf")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read config file:\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		fmt.Fprintf(os.Stderr, "Proceeding without config file.\n")
		return nil
	}
parse:
	config, err = parseConfig(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse config file:\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		fmt.Fprintf(os.Stderr, "Proceeding without config file.\n")
		config = nil
	}
	out, err := yaml.Marshal(config)
	log.Tracef("Parsed config: \n%v\n", string(out))
	return
}
