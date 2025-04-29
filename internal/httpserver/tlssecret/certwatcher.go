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

package tlssecret

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

type TLSSecretWatcher struct {
	sync.RWMutex
	watcher             *fsnotify.Watcher
	ratifyServerTLSCert *tls.Certificate
	clientCAs           *x509.CertPool

	gatekeeperCACertPath    string
	ratifyServerTLSCertPath string
	ratifyServerTLSKeyPath  string
}

func NewTLSSecretWatcher(gatekeeperCACertPath, ratifyServerTLSCertPath, ratifyServerTLSKeyPath string) (*TLSSecretWatcher, error) {
	if ratifyServerTLSCertPath == "" || ratifyServerTLSKeyPath == "" {
		return nil, fmt.Errorf("ratify server TLS cert and key paths must be set")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}
	
	tlsWatcher := &TLSSecretWatcher{
		watcher:                 watcher,
		gatekeeperCACertPath:    gatekeeperCACertPath,
		ratifyServerTLSCertPath: ratifyServerTLSCertPath,
		ratifyServerTLSKeyPath:  ratifyServerTLSKeyPath,
	}

	if err = tlsWatcher.loadCerts(); err != nil {
		return nil, fmt.Errorf("failed to initialize TLS certs: %w", err)
	}

	return tlsWatcher, nil
}

func (w *TLSSecretWatcher) Start() error {
	files := []string{w.ratifyServerTLSCertPath, w.ratifyServerTLSKeyPath}
	if w.gatekeeperCACertPath != "" {
		files = append(files, w.gatekeeperCACertPath)
	}
	for _, file := range files {
		if err := w.watcher.Add(file); err != nil {
			return fmt.Errorf("failed to watch file %s: %w", file, err)
		}
	}

	go w.watch()
	return nil
}

func (w *TLSSecretWatcher) loadCerts() error {
	if w.gatekeeperCACertPath != "" {
		caCert, err := os.ReadFile(w.gatekeeperCACertPath)
		if err != nil {
			return err
		}

		clientCAs := x509.NewCertPool()
		clientCAs.AppendCertsFromPEM(caCert)
		w.Lock()
		w.clientCAs = clientCAs
		w.Unlock()
	}

	ratifyServerTLSCert, err := tls.LoadX509KeyPair(w.ratifyServerTLSCertPath, w.ratifyServerTLSKeyPath)
	if err != nil {
		return err
	}
	w.Lock()
	w.ratifyServerTLSCert = &ratifyServerTLSCert
	w.Unlock()
	return nil
}

// watcher monitors the CA cert file and reloads it on change
func (w *TLSSecretWatcher) watch() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			// If the watcher is closed, exit the loop.
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) != 0 {
				if event.Op&fsnotify.Remove != 0 {
					if err := w.watcher.Add(event.Name); err != nil {
						logrus.Errorf("error re-watching file: %v", err)
					}
				}
				if err := w.loadCerts(); err != nil {
					logrus.Errorf("failed to reload CA certs: %v", err)
				}
			}
		case err, ok := <-w.watcher.Errors:
			// If the watcher is closed, exit the loop.
			if !ok {
				return
			}
			logrus.Errorf("error watching file: %v", err)
		}
	}
}

func (w *TLSSecretWatcher) Stop() {
	if err := w.watcher.Close(); err != nil {
		logrus.Errorf("error closing watcher: %v", err)
	}
}

func (w *TLSSecretWatcher) GetConfigForClient(*tls.ClientHelloInfo) (*tls.Config, error) {
	w.RLock()
	defer w.RUnlock()

	config := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		Certificates:       []tls.Certificate{*w.ratifyServerTLSCert},
		GetConfigForClient: w.GetConfigForClient,
	}

	if w.clientCAs != nil {
		config.ClientCAs = w.clientCAs
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return config, nil
}
