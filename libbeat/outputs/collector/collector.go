package collector

import (
	"fmt"
	"io/ioutil"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"github.com/fsnotify/fsnotify"
)

const selector = "collector"

func init() {
	outputs.RegisterType(selector, makeClient)
}

func makeClient(
	_ outputs.IndexManager,
	_ beat.Info,
	observer outputs.Observer,
	cfg *common.Config,
) (outputs.Group, error) {
	config := defaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return outputs.Fail(err)
	}

	log := logp.NewLogger(selector)

	// If auth type is key and specified accessKey, doesn't start credential watcher.
	// Credential information changes cannot be perception.
	if config.Auth.Type == authTypeKey {
		if config.Auth.Property[authCfgAccessKey] != "" {
			setToken([]byte(config.Auth.Property[authCfgAccessKey]))
		} else {
			// Init credential info
			content, err := ioutil.ReadFile(clusterCredentialFullPath)
			if err != nil {
				return outputs.Fail(fmt.Errorf("read init credential error: %v", err))
			}

			log.Infof("read init content: %v", string(content))
			setToken(content)

			// Start file watcher
			if _, err = NewFileWatcher(log, func(e fsnotify.Event) {
				res, err := ioutil.ReadFile(clusterCredentialFullPath)
				if err != nil {
					log.Errorf("read cluster credential error: %v", err)
					return
				}
				log.Infof("get new credential content: %s", string(res))
				setToken(res)
			}); err != nil {
				log.Errorf("new credential file watcher error: %v", err)
				return outputs.Fail(err)
			}
		}
	}

	hosts, err := outputs.ReadHostList(cfg)
	if err != nil {
		return outputs.Fail(err)
	}

	clients := make([]outputs.NetworkClient, len(hosts))
	for i, host := range hosts {
		var client outputs.NetworkClient
		client, err = newClient(host, config, observer)
		if err != nil {
			return outputs.Fail(err)
		}

		client = outputs.WithBackoff(client, config.Backoff.Init, config.Backoff.Max)
		clients[i] = client
	}

	return outputs.SuccessNet(config.LoadBalance, config.BulkMaxSize, config.MaxRetries, clients)
}
