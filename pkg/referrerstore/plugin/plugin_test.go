package plugin

import (
	"context"
	"strings"
	"testing"

	"github.com/deislabs/ratify/pkg/common"
	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore/config"
)

type TestExecutor struct {
	find    func(plugin string, paths []string) (string, error)
	execute func(ctx context.Context, pluginPath string, cmdArgs []string, stdinData []byte, environ []string) ([]byte, error)
}

func (e *TestExecutor) ExecutePlugin(ctx context.Context, pluginPath string, cmdArgs []string, stdinData []byte, environ []string) ([]byte, error) {
	return e.execute(ctx, pluginPath, cmdArgs, stdinData, environ)
}

func (e *TestExecutor) FindInPaths(plugin string, paths []string) (string, error) {
	return e.find(plugin, paths)
}
func TestPluginMain_GetBlobContent_InvokeExpected(t *testing.T) {
	testPlugin := "test-plugin"
	testExecutor := &TestExecutor{
		find: func(plugin string, paths []string) (string, error) {
			return "testpath", nil
		},
		execute: func(ctx context.Context, pluginPath string, cmdArgs []string, stdinData []byte, environ []string) ([]byte, error) {
			if pluginPath != "testpath" {
				t.Fatalf("mismatch in plugin path expected %s actual %s", "testpath", pluginPath)
			}
			if cmdArgs != nil {
				t.Fatal("cmdArgs is expected to be nil")
			}
			stdData := string(stdinData[:])
			if !strings.Contains(stdData, testPlugin) {
				t.Fatalf("missing config data in stdin expected to have %s actual %s", "test-plugin", stdData)
			}

			commandCheck := false
			versionCheck := false
			subjectCheck := false
			argsCheck := false
			for _, env := range environ {
				if strings.Contains(env, CommandEnvKey) && strings.Contains(env, GetBlobContentCommand) {
					commandCheck = true
				} else if strings.Contains(env, VersionEnvKey) && strings.Contains(env, "1.0.0") {
					versionCheck = true
				} else if strings.Contains(env, SubjectEnvKey) && strings.Contains(env, "localhost") {
					subjectCheck = true
				} else if strings.Contains(env, ArgsEnvKey) && strings.Contains(env, "digest=") {
					argsCheck = true
				}
			}

			if !commandCheck {
				t.Fatalf("missing command env")
			}

			if !versionCheck {
				t.Fatalf("missing version env")
			}

			if !subjectCheck {
				t.Fatalf("missing subject env")
			}

			if !argsCheck {
				t.Fatalf("missing args env")
			}

			return []byte("test data"), nil
		},
	}

	var rawConfig config.StorePluginConfig
	rawConfig = map[string]interface{}{
		testPlugin: StorePlugin{
			name: testPlugin,
		},
	}
	storePlugin := &StorePlugin{
		executor:  testExecutor,
		name:      testPlugin,
		version:   "1.0.0",
		rawConfig: rawConfig,
	}

	subject := common.Reference{
		Original: "localhost",
	}
	result, err := storePlugin.GetBlobContent(context.Background(), subject, "")
	if err != nil {
		t.Fatalf("plugin execution failed %v", err)
	}

	if string(result[:]) != "test data" {
		t.Fatalf("mismatch of result expected %s actual %v", "test data", result)
	}
}

func TestPluginMain_GetReferenceManifest_InvokeExpected(t *testing.T) {
	testPlugin := "test-plugin"
	testExecutor := &TestExecutor{
		find: func(plugin string, paths []string) (string, error) {
			return "testpath", nil
		},
		execute: func(ctx context.Context, pluginPath string, cmdArgs []string, stdinData []byte, environ []string) ([]byte, error) {
			if pluginPath != "testpath" {
				t.Fatalf("mismatch in plugin path expected %s actual %s", "testpath", pluginPath)
			}
			if cmdArgs != nil {
				t.Fatal("cmdArgs is expected to be nil")
			}
			stdData := string(stdinData[:])
			if !strings.Contains(stdData, testPlugin) {
				t.Fatalf("missing config data in stdin expected to have %s actual %s", "test-plugin", stdData)
			}

			commandCheck := false
			versionCheck := false
			subjectCheck := false
			argsCheck := false
			for _, env := range environ {
				if strings.Contains(env, CommandEnvKey) && strings.Contains(env, GetRefManifestCommand) {
					commandCheck = true
				} else if strings.Contains(env, VersionEnvKey) && strings.Contains(env, "1.0.0") {
					versionCheck = true
				} else if strings.Contains(env, SubjectEnvKey) && strings.Contains(env, "localhost") {
					subjectCheck = true
				} else if strings.Contains(env, ArgsEnvKey) && strings.Contains(env, "digest=") {
					argsCheck = true
				}
			}

			if !commandCheck {
				t.Fatalf("missing command env")
			}

			if !versionCheck {
				t.Fatalf("missing version env")
			}

			if !subjectCheck {
				t.Fatalf("missing subject env")
			}

			if !argsCheck {
				t.Fatalf("missing args env")
			}

			manifestData := ` {"artifactType":"test-type"}`
			return []byte(manifestData), nil
		},
	}

	var rawConfig config.StorePluginConfig
	rawConfig = map[string]interface{}{
		testPlugin: StorePlugin{
			name: testPlugin,
		},
	}
	storePlugin := &StorePlugin{
		executor:  testExecutor,
		name:      testPlugin,
		version:   "1.0.0",
		rawConfig: rawConfig,
	}

	subject := common.Reference{
		Original: "localhost",
	}
	ref := ocispecs.ReferenceDescriptor{
		ArtifactType: "test-type",
	}
	result, err := storePlugin.GetReferenceManifest(context.Background(), subject, ref)
	if err != nil {
		t.Fatalf("plugin execution failed %v", err)
	}

	if result.ArtifactType != "test-type" {
		t.Fatalf("mismatch of result expected %s actual %v", "test-type", result)
	}
}

func TestPluginMain_ListReferrers_InvokeExpected(t *testing.T) {
	testPlugin := "test-plugin"
	testExecutor := &TestExecutor{
		find: func(plugin string, paths []string) (string, error) {
			return "testpath", nil
		},
		execute: func(ctx context.Context, pluginPath string, cmdArgs []string, stdinData []byte, environ []string) ([]byte, error) {
			if pluginPath != "testpath" {
				t.Fatalf("mismatch in plugin path expected %s actual %s", "testpath", pluginPath)
			}
			if cmdArgs != nil {
				t.Fatal("cmdArgs is expected to be nil")
			}
			stdData := string(stdinData[:])
			if !strings.Contains(stdData, testPlugin) {
				t.Fatalf("missing config data in stdin expected to have %s actual %s", "test-plugin", stdData)
			}

			commandCheck := false
			versionCheck := false
			subjectCheck := false
			argsCheck := false
			for _, env := range environ {
				if strings.Contains(env, CommandEnvKey) && strings.Contains(env, ListReferrersCommand) {
					commandCheck = true
				} else if strings.Contains(env, VersionEnvKey) && strings.Contains(env, "1.0.0") {
					versionCheck = true
				} else if strings.Contains(env, SubjectEnvKey) && strings.Contains(env, "localhost") {
					subjectCheck = true
				} else if strings.Contains(env, ArgsEnvKey) && strings.Contains(env, "nextToken=") {
					argsCheck = true
				}
			}

			if !commandCheck {
				t.Fatalf("missing command env")
			}

			if !versionCheck {
				t.Fatalf("missing version env")
			}

			if !subjectCheck {
				t.Fatalf("missing subject env")
			}

			if !argsCheck {
				t.Fatalf("missing args env")
			}

			listReferrers := ` {"nextToken":"test-token"}`
			return []byte(listReferrers), nil
		},
	}

	var rawConfig config.StorePluginConfig
	rawConfig = map[string]interface{}{
		testPlugin: StorePlugin{
			name: testPlugin,
		},
	}
	storePlugin := &StorePlugin{
		executor:  testExecutor,
		name:      testPlugin,
		version:   "1.0.0",
		rawConfig: rawConfig,
	}

	subject := common.Reference{
		Original: "localhost",
	}
	result, err := storePlugin.ListReferrers(context.Background(), subject, nil, "")
	if err != nil {
		t.Fatalf("plugin execution failed %v", err)
	}

	if result.NextToken != "test-token" {
		t.Fatalf("mismatch of result expected %s actual %v", "test-token", result)
	}
}
