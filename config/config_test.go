package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FromDefaultPath(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-config")
	if err != nil {
		t.Fatalf("temp dir creation failed %v", err)
	}

	defer os.RemoveAll(tmpDir)

	os.Setenv("RATIFY_CONFIG", tmpDir)

	defer os.Unsetenv("RATIFY_CONFIG")

	fileName := filepath.Join(tmpDir, ConfigFileName)
	content := []byte(`{"stores":  { "version": "1.0.0" }}`)
	err = ioutil.WriteFile(fileName, content, 0644)
	if err != nil {
		t.Fatalf("config file creation failed %v", err)
	}

	config, err := Load("")
	if err != nil {
		t.Fatalf("loading config failed %v", err)
	}

	if config.StoresConfig.Version != "1.0.0" {
		t.Fatalf("mismatch of the loaded config expected version %s actual %s", "1.0.0", config.StoresConfig.Version)
	}

	pluginPath := filepath.Join(tmpDir, PluginsFolder)
	if GetDefaultPluginPath() != pluginPath {
		t.Fatalf("mismatch of the default plugin path expected  %s actual %s", pluginPath, GetDefaultPluginPath())
	}
}

func TestLoad_FromGiventPath(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-config")
	if err != nil {
		t.Fatalf("temp dir creation failed %v", err)
	}

	defer os.RemoveAll(tmpDir)

	fileName := filepath.Join(tmpDir, ConfigFileName)
	content := []byte(`{"stores":  { "version": "1.0.0" }}`)
	err = ioutil.WriteFile(fileName, content, 0644)
	if err != nil {
		t.Fatalf("config file creation failed %v", err)
	}

	config, err := Load(fileName)
	if err != nil {
		t.Fatalf("loading config failed %v", err)
	}

	if config.StoresConfig.Version != "1.0.0" {
		t.Fatalf("mismatch of the loaded config expected version %s actual %s", "1.0.0", config.StoresConfig.Version)
	}

	if GetDefaultPluginPath() == "" {
		t.Fatalf("default plugin path cannot be empty")
	}
}

func TestLoad_NonExistentConfigFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-config")
	if err != nil {
		t.Fatalf("temp dir creation failed %v", err)
	}

	defer os.RemoveAll(tmpDir)

	fileName := filepath.Join(tmpDir, ConfigFileName)
	_, err = Load(fileName)
	if err == nil {
		t.Fatal("loading config expected to fail")
	}
}

func TestLoad_EmptyConfigSucceeds(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-config")
	if err != nil {
		t.Fatalf("temp dir creation failed %v", err)
	}

	defer os.RemoveAll(tmpDir)

	fileName := filepath.Join(tmpDir, ConfigFileName)
	content := []byte("")
	err = ioutil.WriteFile(fileName, content, 0644)
	if err != nil {
		t.Fatalf("config file creation failed %v", err)
	}

	config, err := Load(fileName)
	if err != nil {
		t.Fatalf("loading config failed %v", err)
	}

	if config.StoresConfig.Version != "" {
		t.Fatalf("mismatch of the loaded config expected version %s actual %s", "", config.StoresConfig.Version)
	}
}

func TestLoad_InvalidConfigFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-config")
	if err != nil {
		t.Fatalf("temp dir creation failed %v", err)
	}

	defer os.RemoveAll(tmpDir)

	fileName := filepath.Join(tmpDir, ConfigFileName)
	content := []byte(`"stores":  { "version": "1.0.0" }}`)
	err = ioutil.WriteFile(fileName, content, 0644)
	if err != nil {
		t.Fatalf("config file creation failed %v", err)
	}

	_, err = Load(fileName)
	if err == nil {
		t.Fatalf("loading config is expected to failed")
	}
}
