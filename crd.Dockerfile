# Copyright The Ratify Authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
      
FROM alpine@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c as builder

ARG TARGETOS
ARG TARGETARCH
ARG KUBE_VERSION

RUN echo "Ratify crd building on $TARGETOS, building for $TARGETARCH"

RUN apk add --no-cache curl && \
    curl -LO https://dl.k8s.io/release/v${KUBE_VERSION}/bin/${TARGETOS}/${TARGETARCH}/kubectl && \
    chmod +x kubectl

FROM scratch as build
USER 65532:65532
COPY --chown=65532:65532 * /crds/
COPY --from=builder /kubectl /kubectl
ENTRYPOINT ["/kubectl"]
