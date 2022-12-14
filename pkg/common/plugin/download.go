package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/deislabs/ratify/pkg/ocispecs"
	"github.com/deislabs/ratify/pkg/referrerstore/oras/authprovider"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func DownloadPlugin(name string, source string, pluginBinDir string) error {
	// Initialize a repository
	repository, err := remote.NewRepository(source)
	if err != nil {
		return err
	}

	// TODO: move the ORAS auth code into a separate package that's not referrerstore-specific
	repository.Client = &auth.Client{
		Client: &http.Client{Timeout: 10 * time.Second, Transport: http.DefaultTransport.(*http.Transport).Clone()},
		Header: http.Header{
			"User-Agent": {"ratify"},
		},
		Cache: auth.NewCache(),
		Credential: func(ctx context.Context, registry string) (auth.Credential, error) {
			authProvider, err := authprovider.CreateAuthProviderFromConfig(nil)
			if err != nil {
				return auth.EmptyCredential, err
			}

			authConfig, err := authProvider.Provide(context.TODO(), source)
			if err != nil {
				return auth.EmptyCredential, err
			}

			if authConfig.Username != "" || authConfig.Password != "" || authConfig.IdentityToken != "" {
				return auth.Credential{
					Username:     authConfig.Username,
					Password:     authConfig.Password,
					RefreshToken: authConfig.IdentityToken,
				}, nil
			}
			return auth.EmptyCredential, nil
		},
	}

	// read the reference manifest
	referenceManifestDescriptor, err := repository.Resolve(context.TODO(), source)
	if err != nil {
		return err
	}

	manifestReader, err := repository.Fetch(context.TODO(), referenceManifestDescriptor)
	if err != nil {
		return err
	}
	
	manifestBytes, err := io.ReadAll(manifestReader)
	if err != nil {
		return err
	}

	referenceManifest := ocispecs.ReferenceManifest{}
	if err := json.Unmarshal(manifestBytes, &referenceManifest); err != nil {
		return err
	}

	// Download the contents of the first blob as the named plugin. This matches `oras push registry.example.com/sample-plugin:v1 ./sample`
	// TODO: should this be "first/only blob of media type `application/vnd.ratify.plugin`"?
	blobReference := fmt.Sprintf("%s@%s", source, referenceManifest.Blobs[0].Digest)
	_, blobReader, err := repository.Blobs().FetchReference(context.TODO(), blobReference)
	if err != nil {
		return err
	}
	
	pluginPath := path.Join(pluginBinDir, name)
	pluginFile, err := os.Create(pluginPath)
	if err != nil {
		return err
	}

	defer pluginFile.Close()
	_, err = io.Copy(pluginFile, blobReader)
	if err != nil {
		return err
	}

	// Mark the plugin as executable
	err = os.Chmod(pluginPath, 0700)
	if err != nil {
		return err
	}

	return nil
}
