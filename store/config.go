package store

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/hashicorp/go-hclog"

	"github.com/spf13/viper"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	accountTokenParam   = "ACCOUNT_TOKEN"
	apiTokenParam       = "API_TOKEN"
	regionParam         = "REGION"
	customListenerParam = "CUSTOM_LISTENER_URL"
	customAPIParam      = "CUSTOM_API"
	usRegionCode        = "us"
)

// PandoraConfig struct for pandora span store
type PandoraConfig struct {
	AccountToken      string `yaml:"accountToken"`
	Region            string `yaml:"region"`
	APIToken          string `yaml:"apiToken"`
	CustomListenerURL string `yaml:"customListenerUrl"`
	CustomAPIURL      string `yaml:"customAPIUrl"`
	SourceType        string `yaml:"sourceType"`
	Repo              string `yaml:"repo"`
}

// validate pandora config, return error if invalid
func (config *PandoraConfig) validate(logger hclog.Logger) error {
	if config.AccountToken == "" && config.APIToken == "" {
		return errors.New("At least one of pandora account token or api-token has to be valid")
	}
	if config.APIToken == "" {
		logger.Warn("No api token found, can't create span reader")
	}
	if config.AccountToken == "" {
		logger.Warn("No account token found, spans will not be saved")
	}
	return nil
}

//ParseConfig receives a config file path, parse it and returns pandora span store config
func ParseConfig(filePath string, logger hclog.Logger) (*PandoraConfig, error) {
	var pandoraConfig *PandoraConfig
	if filePath != "" {
		pandoraConfig = &PandoraConfig{}
		yamlFile, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(yamlFile, &pandoraConfig)
	} else {
		v := viper.New()
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.SetDefault(regionParam, "")
		v.SetDefault(customAPIParam, "")
		v.SetDefault(customListenerParam, "")
		v.AutomaticEnv()

		pandoraConfig = &PandoraConfig{
			Region:            v.GetString(regionParam),
			AccountToken:      v.GetString(accountTokenParam),
			APIToken:          v.GetString(apiTokenParam),
			CustomAPIURL:      v.GetString(customAPIParam),
			CustomListenerURL: v.GetString(customListenerParam),
		}
	}

	if err := pandoraConfig.validate(logger); err != nil {
		return nil, err
	}
	return pandoraConfig, nil
}

// ListenerURL returns the constructed listener URL to write spans to
func (config *PandoraConfig) ListenerURL() string {
	if config.CustomListenerURL != "" {
		return config.CustomListenerURL
	}
	return fmt.Sprintf("https://listener%s.pandora:8071", config.regionCode())
}

// APIURL returns the constructed API URL to read spans from
func (config *PandoraConfig) APIURL() string {
	if config.CustomAPIURL != "" {
		return config.CustomAPIURL
	}
	return fmt.Sprintf("https://api%s.pandora/v1/elasticsearch/_msearch", config.regionCode())
}

func (config *PandoraConfig) regionCode() string {
	regionCode := ""
	if config.Region != "" && config.Region != usRegionCode {
		regionCode = fmt.Sprintf("-%s", config.Region)
	}
	return regionCode
}

func (config *PandoraConfig) String() string {
	desc := fmt.Sprintf("account token: %v \n api token: %v \n listener url: %v \n api url: %s", censorString(config.AccountToken, 4), censorString(config.APIToken, 9), config.ListenerURL(), config.APIURL())
	return desc
}

func censorString(word string, n int) string {
	if len(word) > 2*n {
		return word[:n] + strings.Repeat("*", len(word)-(n*2)) + word[len(word)-n:]
	}
	return strings.Repeat("*", len(word))
}
