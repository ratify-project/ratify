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

package context

import (
	"context"
	"fmt"
)

type contextKey string

const contextKeyNamespace = contextKey("namespace")

// SetContextWithNamespace embeds namespace to the context.
func SetContextWithNamespace(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, contextKeyNamespace, namespace)
}

// CreateCacheKey creates a new cache key prefixed with embedded namespace.
func CreateCacheKey(ctx context.Context, key string) string {
	namespace := ctx.Value(contextKeyNamespace)
	if namespace == nil {
		return key
	}

	namespaceStr := namespace.(string)
	if namespaceStr == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", namespaceStr, key)
}
