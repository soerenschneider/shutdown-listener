package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
)

const (
	AppName             = "shutdown-listener"
	defaultMetricConfig = ":9194"
)

var (
	defaultCmd = []string{
		"sudo",
		"systemctl",
		"poweroff",
	}

	// This regex is not a very strict check, we don't validate hostname or ip (v4, v6) addresses...
	mqttHostRegex = regexp.MustCompile(`^\w{3,}://.{3,}:\d{2,5}$`)

	// We don't care that technically it's allowed to start with a slash
	mqttTopicRegex = regexp.MustCompile("^([\\w%]+)(/[\\w%]+)*$")
)

type Config struct {
	MetricConfig string   `json:"metrics_addr,omitempty"`
	Command      []string `json:"cmd"`
	MqttConfig
}

type MqttConfig struct {
	Host     string `json:"mqtt_host,omitempty"`
	Topic    string `json:"mqtt_topic,omitempty"`
	User     string `json:"mqtt_user,omitempty"`
	Password string `json:"mqtt_password,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		MetricConfig: defaultMetricConfig,
		Command:      defaultCmd,
	}
}

func ReadJsonConfig(filePath string) (*Config, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read config from file: %v", err)
	}

	ret := DefaultConfig()
	err = json.Unmarshal(fileContent, &ret)
	return &ret, err
}

func (conf *Config) Validate() error {
	if err := matchTopic(conf.Topic); err != nil {
		return errors.New("invalid mqtt topic provided")
	}

	if err := matchHost(conf.MqttConfig.Host); err != nil {
		return err
	}

	return nil
}

func (conf *Config) Print() {
	log.Println("-----------------")
	log.Println("Configuration:")
	log.Printf("MetricConfig=%s", conf.MetricConfig)
	log.Printf("Command=%d", conf.Command)
	log.Printf("Host=%s", conf.Host)
	log.Printf("Topic=%s", conf.Topic)
	log.Println("-----------------")
}

func matchTopic(topic string) error {
	if !mqttTopicRegex.MatchString(topic) {
		return fmt.Errorf("invalid topic format used")
	}
	return nil
}

func matchHost(host string) error {
	if !mqttHostRegex.Match([]byte(host)) {
		return fmt.Errorf("invalid host format used")
	}
	return nil
}
