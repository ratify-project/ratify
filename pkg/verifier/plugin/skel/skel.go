package skel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/deislabs/hora/pkg/common"
	"github.com/deislabs/hora/pkg/common/plugin"
	"github.com/deislabs/hora/pkg/ocispecs"
	"github.com/deislabs/hora/pkg/referrerstore"
	storeConfig "github.com/deislabs/hora/pkg/referrerstore/config"
	"github.com/deislabs/hora/pkg/referrerstore/factory"
	"github.com/deislabs/hora/pkg/utils"
	"github.com/deislabs/hora/pkg/verifier"
	"github.com/deislabs/hora/pkg/verifier/config"
	vp "github.com/deislabs/hora/pkg/verifier/plugin"
	"github.com/deislabs/hora/pkg/verifier/types"
)

type VerifyReference func(args *CmdArgs, subjectReference common.Reference, referenceDescriptor ocispecs.ReferenceDescriptor, referrerStore referrerstore.ReferrerStore) (*verifier.VerifierResult, error)

type CmdArgs struct {
	Version    string
	Subject    string
	subjectRef common.Reference
	StdinData  []byte
}

func PluginMain(name, version string, verifyReference VerifyReference, supportedVersions []string) {
	if e := pluginMainCore(name, version, verifyReference, supportedVersions); e != nil {
		if err := e.Print(); err != nil {
			log.Print("Error writing error JSON to stdout: ", err)
		}
		os.Exit(1)
	}
}

func pluginMainCore(name, version string, verifyReference VerifyReference, supportedVersions []string) *plugin.Error {
	cmd, cmdArgs, err := getCmdArgsFromEnv()
	if err != nil {
		// TODO about string
		return err
	}

	if err = validateVersion(cmdArgs.Version, supportedVersions); err != nil {
		return err
	}

	input, err := validateAndGetConfig(cmdArgs.StdinData)
	if err != nil {
		return err
	}

	// The below is a workaround to be able to use built-in referrer store plugins from within verifier plugins
	storeConfigs := storeConfig.StoresConfig{
		Version:       version,
		PluginBinDirs: nil,
		Stores:        []storeConfig.StorePluginConfig{input.StoreConfig.Store},
	}
	stores, storeErr := factory.CreateStoresFromConfig(storeConfigs, "")
	if storeErr != nil || stores == nil || len(stores) == 0 {
		return plugin.NewError(types.ErrArgsParsingFailure, fmt.Sprintf("create store from input config failed with error #{storeErr}"), "")
	}
	store := stores[0]

	// This is the original implementation for initialization of a referrer store which does not support built-ins
	//store, serr := rp.NewStore(input.StoreConfig.Version, input.StoreConfig.Store, input.StoreConfig.PluginBinDirs)
	//if serr != nil {
	//	return plugin.NewError(types.ErrArgsParsingFailure, fmt.Sprintf("create store from the input config failed with error %v", serr), "")
	//}

	switch cmd {
	case vp.VerifyCommand:
		result, err := verifyReference(cmdArgs, cmdArgs.subjectRef, input.ReferencDesc, store)

		if err != nil {
			return plugin.NewError(types.ErrPluginCmdFailure, fmt.Sprintf("plugin command %s failed", vp.VerifyCommand), err.Error())
		}

		err = types.WriteVerifyResultResult(result, os.Stdout)
		if err != nil {
			return plugin.NewError(types.ErrIOFailure, "failed to write plugin output", err.Error())
		}

		return nil
	default:
		return plugin.NewError(types.ErrUnknownCommand, fmt.Sprintf("unknown %s: %v", vp.CommandEnvKey, cmd), "")
	}
}

func getCmdArgsFromEnv() (string, *CmdArgs, *plugin.Error) {
	argsMissing := make([]string, 0)

	// #1 Command
	var cmd = os.Getenv(vp.CommandEnvKey)
	if cmd == "" {
		argsMissing = append(argsMissing, vp.CommandEnvKey)
	}

	// #2 Version
	var version = os.Getenv(vp.VersionEnvKey)
	if version == "" {
		argsMissing = append(argsMissing, vp.VersionEnvKey)
	}

	// #3 Subject
	var subject = os.Getenv(vp.SubjectEnvKey)
	if subject == "" {
		argsMissing = append(argsMissing, vp.SubjectEnvKey)
	}

	if len(argsMissing) > 0 {
		joined := strings.Join(argsMissing, ",")
		return "", nil, plugin.NewError(types.ErrMissingEnvironmentVariables, fmt.Sprintf("missing env variables [%s]", joined), "")
	}

	// TODO Limit the read length
	stdinData, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return "", nil, plugin.NewError(types.ErrIOFailure, fmt.Sprintf("error reading from stdin: %v", err), "")
	}

	subRef, err := utils.ParseSubjectReference(subject)
	if err != nil {
		return "", nil, plugin.NewError(types.ErrArgsParsingFailure, fmt.Sprintf("cannot parse subject reference %s", subject), err.Error())
	}

	cmdArgs := &CmdArgs{
		Version:    version,
		Subject:    subject,
		StdinData:  stdinData,
		subjectRef: subRef,
	}

	return cmd, cmdArgs, nil
}

func validateVersion(version string, supportedVersions []string) *plugin.Error {
	for _, v := range supportedVersions {
		// TODO check for compatibility using semversion
		if v == version {
			return nil
		}
	}

	return plugin.NewError(types.ErrVersionNotSupported, fmt.Sprintf("plugin doesn't support version %s", version), "")
}

func validateAndGetConfig(jsonBytes []byte) (*config.PluginInputConfig, *plugin.Error) {
	var input config.PluginInputConfig

	if err := json.Unmarshal(jsonBytes, &input); err != nil {
		return nil, plugin.NewError(types.ErrConfigParsingFailure, fmt.Sprintf("error unmarshall verifier config: %v", err), "")
	}
	if input.Config[types.Name] == "" {
		return nil, plugin.NewError(types.ErrInvalidVerifierConfig, "missing verifier name", "")
	}
	return &input, nil
}
