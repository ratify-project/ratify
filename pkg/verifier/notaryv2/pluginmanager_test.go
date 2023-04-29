package notaryv2

import (
	"context"
	"testing"
)

const (
	expectedPluginName = "mock.plugin"
)

func TestPluginManagerGet(t *testing.T) {
	pluginManager := NewRatifyPluginManager("./")

	_, err := pluginManager.Get(context.TODO(), expectedPluginName)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPluginManagerList(t *testing.T) {
	pluginManager := NewRatifyPluginManager("./")

	pluginList, err := pluginManager.List(context.TODO())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(pluginList) != 1 {
		t.Errorf("expected to find exactly one plugin, instead found %d", len(pluginList))
		t.Fatalf("plugin list: %+v", pluginList)
	}

	if pluginList[0] != expectedPluginName {
		t.Errorf("expected to find plugin named %s, instead found %s", expectedPluginName, pluginList[0])
	}
}
