package notaryv2

import (
	"context"
	"os"
	"testing"
)

const (
	pluginDirectory  = "./pluginmanagertest"
	ignoredDirectory = "notation-ignored"
	ignoredFile      = "ignored.file"
	testPluginName   = "mock.plugin"
)

func TestMain(m *testing.M) {
	err := setupTestFiles()
	code := 1
	if err == nil {
		code = m.Run()
	}
	cleanupTestFiles()
	os.Exit(code)
}

func setupTestFiles() error {
	err := os.Mkdir(pluginDirectory, 0700)
	if err != nil {
		return err
	}

	err = os.Mkdir(pluginDirectory+"/"+ignoredDirectory, 0700)
	if err != nil {
		return err
	}

	err = os.WriteFile(pluginDirectory+"/"+notationPluginPrefix+testPluginName, []byte(""), 0700)
	if err != nil {
		return err
	}

	err = os.WriteFile(pluginDirectory+"/"+ignoredFile, []byte(""), 0700)
	if err != nil {
		return err
	}

	return nil
}

func cleanupTestFiles() {
	os.RemoveAll(pluginDirectory)
}

func TestPluginManagerGet(t *testing.T) {
	pluginManager := NewRatifyPluginManager(pluginDirectory)

	_, err := pluginManager.Get(context.TODO(), testPluginName)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPluginManagerList(t *testing.T) {
	pluginManager := NewRatifyPluginManager(pluginDirectory)

	pluginList, err := pluginManager.List(context.TODO())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(pluginList) != 1 {
		t.Errorf("expected to find exactly one plugin, instead found %d", len(pluginList))
		t.Fatalf("plugin list: %+v", pluginList)
	}

	if pluginList[0] != testPluginName {
		t.Errorf("expected to find plugin named %s, instead found %s", testPluginName, pluginList[0])
	}
}
