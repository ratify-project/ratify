package registry

import "strings"

const (
	regRepoDelimiter = "/"
	defaultRegistry  = "index.docker.io"
)

func GetRegistryRepoString(path string) (string, string) {
	var registry string
	var repo string
	parts := strings.SplitN(path, regRepoDelimiter, 2)
	if len(parts) == 2 && (strings.ContainsRune(parts[0], '.') || strings.ContainsRune(parts[0], ':')) {
		// The first part of the repository is treated as the registry domain
		// iff it contains a '.' or ':' character, otherwise it is all repository
		// and the domain defaults to Docker Hub.
		registry = parts[0]
		repo = parts[1]
	}

	if registry == "" {
		return defaultRegistry, path
	}

	return registry, repo
}
