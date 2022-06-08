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
	"sync"

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

var (
	configHash string
)

// Create a executor from configurationFile and setup config file watcher
func GetExecutorAndWatchForUpdate(configFilePath string) (*ef.Executor, error) {
	cf, err := Load(configFilePath)

	if err != nil {
		return &ef.Executor{}, err
	}

	configHash = cf.FileHash

	stores, verifiers, policyEnforcer, err := createFromConfig(cf)

	if err != nil {
		return &ef.Executor{}, err
	}

	executor := ef.Executor{
		Verifiers:      verifiers,
		ReferrerStores: stores,
		PolicyEnforcer: policyEnforcer,
		Config:         &cf.ExecutorConfig,
		Mu:             sync.RWMutex{},
	}

	err = watchForConfigurationChange(configFilePath, &executor)

	if err != nil {
		return &ef.Executor{}, err
	}

	logrus.Info("configuration successfully loaded.")

	return &executor, nil
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

	logrus.Infof("verifiers address %v ,%v ", verifiers[0].Name(), &verifiers[0])
	logrus.Infof("verifiers successfully created. number of verifiers %d", len(verifiers))

	policyEnforcer, err := pf.CreatePolicyProviderFromConfig(cf.PoliciesConfig)

	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to load policy provider from config")
	}

	logrus.Infof("policies successfully created.")

	return stores, verifiers, policyEnforcer, nil
}

// Setup a watcher on file at configFilePath, reload executor on file change
func watchForConfigurationChange(configFilePath string, executor *ef.Executor) error {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		errors.Wrap(err, "new file watcher on configuration file failed ")
	}

	logrus.Infof("executor pointer %v", executor)

	go func() {

		for {
			select {
			case event, ok := <-watcher.Events:

				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {

					cf, err := Load(configFilePath)

					if err != nil {
						logrus.Warnf("failed to load from config file , err: %v", err)
						continue // we don't want to return, as returning will close the file watcher
					}

					stores, verifiers, policyEnforcer, err := createFromConfig(cf)

					if err != nil {
						logrus.Warnf("failed to store/verifier/policy objects from config, no updates loaded. err: %v", err)
						continue
					}

					if configHash != cf.FileHash {
						executor.ReloadAll(stores, verifiers, policyEnforcer, &cf.ExecutorConfig)
						configHash = cf.FileHash
						logrus.Infof("configuration file has been updated, reloading executor succeeded")
					} else {
						logrus.Infof("no change found in config file, no executor update needed")
					}

				}

			case err, ok := <-watcher.Errors:
				if !ok {
					logrus.Errorf("configuration file watcher returned error : %v, watcher will be closed.", err)
					return
				}
			}
		}
	}()

	err = watcher.Add(configFilePath)
	if err != nil {
		logrus.Error("adding configuration file failed, err: %v", err)
		return err
	}

	logrus.Infof("watcher added on configuration file %v", configFilePath)

	return nil
}
