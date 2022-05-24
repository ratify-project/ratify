/*
Copyright The Ratify Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	ef "github.com/deislabs/ratify/pkg/executor/core"
	"github.com/deislabs/ratify/pkg/policyprovider"
	pf "github.com/deislabs/ratify/pkg/policyprovider/factory"
	"github.com/deislabs/ratify/pkg/referrerstore"
	sf "github.com/deislabs/ratify/pkg/referrerstore/factory"
	"github.com/deislabs/ratify/pkg/verifier"
	vf "github.com/deislabs/ratify/pkg/verifier/factory"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

var (
	configHash string
)

func createFromConfig(cf Config) ([]referrerstore.ReferrerStore, []verifier.ReferenceVerifier, policyprovider.PolicyProvider, error) {
	stores, err := sf.CreateStoresFromConfig(cf.StoresConfig, GetDefaultPluginPath())

	if err != nil {
		return nil, nil, nil, err
	}
	logrus.Infof("stores successfully created. number of stores %d", len(stores))

	verifiers, err := vf.CreateVerifiersFromConfig(cf.VerifiersConfig, GetDefaultPluginPath())

	if err != nil {
		return nil, nil, nil, err
	}

	logrus.Infof("verifiers successfully created. number of verifiers %d", len(verifiers))

	policyEnforcer, err := pf.CreatePolicyProviderFromConfig(cf.PoliciesConfig)

	if err != nil {
		return nil, nil, nil, err
	}

	logrus.Infof("policies successfully created.")

	return stores, verifiers, policyEnforcer, nil
}

func GetExecutorAndWatchForUpdate(configFilePath string) (ef.Executor, error) {
	cf, err := Load(configFilePath)
	configHash = cf.FileHash

	logrus.Infof("configuration loaded %v", configFilePath)

	stores, verifiers, policyEnforcer, err := createFromConfig(cf)

	logrus.Info("configuration successfully loaded.")

	executor := ef.Executor{
		Verifiers:      verifiers,
		ReferrerStores: stores,
		PolicyEnforcer: policyEnforcer,
		Config:         &cf.ExecutorConfig,
	}

	if err != nil {
		return executor, err // todo: wrap
	}
	updateExecutorOnConfigurationChange(configFilePath, &executor) // err handling

	return executor, nil
}

func updateExecutorOnConfigurationChange(configFilePath string, executor *ef.Executor) error {

	// setup file watcher with handler
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Infof("NewWatcher failed: ", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		defer close(done)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					logrus.Infof("Config write event detected! %s %s\n", event.Name, event.Op)
					cf, err := Load(configFilePath)

					stores, verifiers, policyEnforcer, err := createFromConfig(cf)

					if err != nil {
						//return err // todo: wrap
					}

					if configHash != cf.FileHash {
						logrus.Infof("Config change detected!")
						executor.ReloadAll(stores, verifiers, policyEnforcer, &cf.ExecutorConfig)
						configHash = cf.FileHash
					} else {
						logrus.Infof("config file was same ")
					}

				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.Infof("error:", err)
			}
		}
		//close(done)
	}()

	logrus.Infof("added watcher on file %v", configFilePath)
	err = watcher.Add(configFilePath)
	if err != nil {
		logrus.Infof("Add failed:", err)
	}
	<-done

	return nil
}
