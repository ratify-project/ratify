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
	"os"
	"time"

	ef "github.com/deislabs/ratify/pkg/executor/core"
	"github.com/deislabs/ratify/pkg/policyprovider"
	pf "github.com/deislabs/ratify/pkg/policyprovider/factory"
	"github.com/deislabs/ratify/pkg/referrerstore"
	sf "github.com/deislabs/ratify/pkg/referrerstore/factory"
	"github.com/deislabs/ratify/pkg/verifier"
	vf "github.com/deislabs/ratify/pkg/verifier/factory"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type GetExecutor func() *ef.Executor

var (
	configHash string
	executor   ef.Executor
)

// Create a executor from configurationFile and setup config file watcher
func GetExecutorAndWatchForUpdate(configFilePath string) (GetExecutor, error) {

	configFilePath = getConfigurationFile(configFilePath)
	cf, err := Load(configFilePath)

	if err != nil {
		return func() *ef.Executor { return &ef.Executor{} }, err
	}

	configHash = cf.fileHash

	stores, verifiers, policyEnforcer, err := createFromConfig(cf)

	if err != nil {
		return func() *ef.Executor { return &ef.Executor{} }, err
	}

	executor = ef.Executor{
		Verifiers:      verifiers,
		ReferrerStores: stores,
		PolicyEnforcer: policyEnforcer,
		Config:         &cf.ExecutorConfig,
	}

	err = watchForConfigurationChange(configFilePath)

	if err != nil {
		return func() *ef.Executor { return &ef.Executor{} }, err
	}

	logrus.Info("configuration successfully loaded.")

	return func() *ef.Executor { return &executor }, nil
}

// Returns created referer store, verifier, policyprovider objects from config
func createFromConfig(cf Config) ([]referrerstore.ReferrerStore, []verifier.ReferenceVerifier, policyprovider.PolicyProvider, error) {
	stores, err := sf.CreateStoresFromConfig(cf.StoresConfig, GetDefaultPluginPath())

	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to load store from config")
	}
	logrus.Infof("stores successfully created. number of stores %d", len(stores))

	verifiers, err := vf.CreateVerifiersFromConfig(cf.VerifiersConfig, GetDefaultPluginPath())

	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to load verifiers from config")
	}

	logrus.Infof("verifiers successfully created. number of verifiers %d", len(verifiers))

	policyEnforcer, err := pf.CreatePolicyProviderFromConfig(cf.PoliciesConfig)

	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to load policy provider from config")
	}

	logrus.Infof("policies successfully created.")

	return stores, verifiers, policyEnforcer, nil
}

func reloadExecutor(configFilePath string) {

	cf, err := Load(configFilePath)

	if err != nil {
		logrus.Warnf("failed to load from config file , err: %v", err)
		return
	}

	if configHash != cf.fileHash {
		stores, verifiers, policyEnforcer, err := createFromConfig(cf)

		newExecutor := ef.Executor{
			Verifiers:      verifiers,
			ReferrerStores: stores,
			PolicyEnforcer: policyEnforcer,
			Config:         &cf.ExecutorConfig,
		}

		if err != nil {
			logrus.Warnf("failed to store/verifier/policy objects from config, no updates loaded. err: %v", err)
			return
		}

		executor = newExecutor
		configHash = cf.fileHash
		logrus.Infof("configuration file has been updated, reloading executor succeeded")
	} else {
		logrus.Infof("no change found in config file, no executor update needed")
	}
	return
}

// Setup a watcher on file at configFilePath, reload executor on file change
func watchForConfigurationChange(configFilePath string) error {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		errors.Wrap(err, "new file watcher on configuration file failed ")
	}

	err = watcher.Add(configFilePath)

	if err != nil {
		logrus.Errorf("adding configuration file watcher failed, err: %v", err)
		return err
	}

	logrus.Infof("watcher added on configuration file %v", configFilePath)

	// setup for loop to listen for events
	go func() {

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					logrus.Warnf("no longer watching configuration file changes, file watcher event channel closed")
					return
				}

				logrus.Infof("file watcher event detected %v", event)

				// In a cluster scenario, a configMap will recreate the config file
				// after the remove event, the watcher will also be removed
				// since a watcher on a non existent file is not supported, we sleep until the file exist add the watcher back
				if event.Name == configFilePath && event.Op&fsnotify.Remove == fsnotify.Remove {
					logrus.Infof("config file remove event detected")
					sleepTime := 1 * time.Second
					waitTime := 60 //1min

					time.Sleep(sleepTime)
					_, err := os.Stat(configFilePath)

					for err != nil {
						if waitTime < 0 {
							logrus.Warnf("config file not found after waiting for %v sec, os.Stat error %v", waitTime, err)
							return
						}
						logrus.Infof("config file does not exist yet, sleeping again")
						_, err = os.Stat(configFilePath)
						time.Sleep(sleepTime)
						waitTime--
					}
					reloadExecutor(configFilePath)
					err = watcher.Add(configFilePath)

					if err != nil {
						logrus.Errorf("adding configuration file watcher failed, err: %v", err)
						continue
					}

					logrus.Infof("watcher added on configuration directory %v", configFilePath)

				}

				// In a local scenario, the configuration will be updated through a write event
				if event.Name == configFilePath && event.Op&fsnotify.Write == fsnotify.Write {
					reloadExecutor(configFilePath)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					logrus.Errorf("configuration file watcher returned error : %v, watcher will be closed.", err)
					return
				}
			}
		}
	}()

	return nil
}
