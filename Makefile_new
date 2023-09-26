# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# generate: generate-controller generate-groups rbacs manifests fmt
generate: generate-controller rbacs manifests fmt

#generate helm documentation
docs: helm-docs
	$(HELM_DOCS) -t deployments/fluidos/README.gotmpl deployments/fluidos

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	rm -f deployments/fluidos/crds/*
	$(CONTROLLER_GEN) paths="./apis/..." crd:generateEmbeddedObjectMeta=true output:crd:artifacts:config=deployments/fluidos/crds

#Generate RBAC for each controller
rbacs: controller-gen
	rm -f deployments/fluidos/files/*
#	$(CONTROLLER_GEN) paths="./pkg/liqo-controller-manager/..." rbac:roleName=liqo-controller-manager output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-controller-manager-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-controller-manager-ClusterRole.yaml deployments/liqo/files/liqo-controller-manager-Role.yaml
	$(CONTROLLER_GEN) paths="./pkg/local-resource-manager" rbac:roleName=fluidos-local-resource-manager output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/fluidos/files/fluidos-local-resource-manager-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/fluidos/files/fluidos-local-resource-manager-ClusterRole.yaml
	$(CONTROLLER_GEN) paths="./pkg/rear-manager/" rbac:roleName=fluidos-rear-manager output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/fluidos/files/fluidos-rear-manager-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/fluidos/files/fluidos-rear-manager-ClusterRole.yaml
	$(CONTROLLER_GEN) paths="./pkg/rear-controller/..." rbac:roleName=fluidos-rear-controller output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/fluidos/files/fluidos-rear-controller-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/fluidos/files/fluidos-rear-controller-ClusterRole.yaml
#	$(CONTROLLER_GEN) paths="./internal/liqonet/route-operator" rbac:roleName=liqo-route output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-route-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-route-ClusterRole.yaml deployments/liqo/files/liqo-route-Role.yaml
#	$(CONTROLLER_GEN) paths="./internal/liqonet/tunnel-operator" rbac:roleName=liqo-gateway output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-gateway-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-gateway-ClusterRole.yaml deployments/liqo/files/liqo-gateway-Role.yaml
#	$(CONTROLLER_GEN) paths="./internal/liqonet/network-manager/..." rbac:roleName=liqo-network-manager output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-network-manager-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-network-manager-ClusterRole.yaml deployments/liqo/files/liqo-network-manager-Role.yaml
#	$(CONTROLLER_GEN) paths="./internal/crdReplicator" rbac:roleName=liqo-crd-replicator output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-crd-replicator-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-crd-replicator-ClusterRole.yaml deployments/liqo/files/liqo-crd-replicator-Role.yaml
#	$(CONTROLLER_GEN) paths="./pkg/discoverymanager" rbac:roleName=liqo-discovery output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-discovery-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-discovery-ClusterRole.yaml deployments/liqo/files/liqo-discovery-Role.yaml
#	$(CONTROLLER_GEN) paths="./internal/auth-service" rbac:roleName=liqo-auth-service output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-auth-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-auth-ClusterRole.yaml deployments/liqo/files/liqo-auth-Role.yaml
#	$(CONTROLLER_GEN) paths="./pkg/peering-roles/basic" rbac:roleName=liqo-remote-peering-basic output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-remote-peering-basic-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-remote-peering-basic-ClusterRole.yaml
#	$(CONTROLLER_GEN) paths="./pkg/peering-roles/incoming" rbac:roleName=liqo-remote-peering-incoming output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-remote-peering-incoming-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-remote-peering-incoming-ClusterRole.yaml
#	$(CONTROLLER_GEN) paths="./pkg/peering-roles/outgoing" rbac:roleName=liqo-remote-peering-outgoing output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-remote-peering-outgoing-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-remote-peering-outgoing-ClusterRole.yaml
#	$(CONTROLLER_GEN) paths="./pkg/liqo-controller-manager/..." rbac:roleName=liqo-controller-manager output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-controller-manager-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-controller-manager-ClusterRole.yaml deployments/liqo/files/liqo-controller-manager-Role.yaml
#	$(CONTROLLER_GEN) paths="./pkg/virtualKubelet/roles/local" rbac:roleName=liqo-virtual-kubelet-local output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-virtual-kubelet-local-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-virtual-kubelet-local-ClusterRole.yaml
#	$(CONTROLLER_GEN) paths="./pkg/virtualKubelet/roles/remote" rbac:roleName=liqo-virtual-kubelet-remote output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-virtual-kubelet-remote-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-virtual-kubelet-remote-ClusterRole.yaml
#	$(CONTROLLER_GEN) paths="./cmd/uninstaller" rbac:roleName=liqo-pre-delete output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-pre-delete-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-pre-delete-ClusterRole.yaml
#	$(CONTROLLER_GEN) paths="./cmd/metric-agent" rbac:roleName=liqo-metric-agent output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-metric-agent-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-metric-agent-ClusterRole.yaml
#	$(CONTROLLER_GEN) paths="./cmd/telemetry" rbac:roleName=liqo-telemetry output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/liqo/files/liqo-telemetry-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  sed -i -n '/rules/,$$p' deployments/liqo/files/liqo-telemetry-ClusterRole.yaml

# Install gci if not available
gci:
ifeq (, $(shell which gci))
	@go install github.com/daixiang0/gci@v0.11.0
GCI=$(GOBIN)/gci
else
GCI=$(shell which gci)
endif

# Install addlicense if not available
addlicense:
ifeq (, $(shell which addlicense))
	@go install github.com/google/addlicense@v1.0.0
ADDLICENSE=$(GOBIN)/addlicense
else
ADDLICENSE=$(shell which addlicense)
endif

# Run go fmt against code
fmt: gci addlicense
	go mod tidy
	go fmt ./...
	find . -type f -name '*.go' -a ! -name '*zz_generated*' -exec $(GCI) write -s standard -s default -s "prefix(github.com/fluidos-project/WP3_Node)" {} \;
	find . -type f -name '*.go' -exec $(ADDLICENSE) -l apache -c "FLUIDOS Project" -y "2022-$(shell date +%Y)" {} \;

# Install golangci-lint if not available
golangci-lint:
ifeq (, $(shell which golangci-lint))
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3
GOLANGCILINT=$(GOBIN)/golangci-lint
else
GOLANGCILINT=$(shell which golangci-lint)
endif

markdownlint:
ifeq (, $(shell which markdownlint))
	@echo "markdownlint is not installed. Please install it: https://github.com/igorshubovych/markdownlint-cli#installation"
	@exit 1
else
MARKDOWNLINT=$(shell which markdownlint)
endif

md-lint: markdownlint
	@find . -type f -name '*.md' -a -not -path "./.github/*" \
		-not -path "./docs/_legacy/*" \
		-not -path "./deployments/*" \
		-not -path "./hack/code-generator/*" \
		-exec $(MARKDOWNLINT) {} +

lint: golangci-lint
	$(GOLANGCILINT) run --new

generate-controller: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./apis/..."

generate-groups:
	if [ ! -d  "hack/code-generator" ]; then \
		git clone --depth 1 -b v0.25.0 https://github.com/kubernetes/code-generator.git hack/code-generator; \
	fi
	rm -rf pkg/client
	hack/code-generator/generate-groups.sh client,lister,informer \
		github.com/liqotech/liqo/pkg/client github.com/liqotech/liqo/apis \
		"virtualkubelet:v1alpha1" \
		--output-base ./ \
		-h hack/boilerplate.go.txt && \
	mv github.com/liqotech/liqo/pkg/client pkg/ && \
	rm -rf github.com

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

helm-docs:
ifeq (, $(shell which helm-docs))
	@{ \
	set -e ;\
	HELM_DOCS_TMP_DIR=$$(mktemp -d) ;\
	cd $$HELM_DOCS_TMP_DIR ;\
	version=1.11.0 ;\
    arch=x86_64 ;\
    echo  $$HELM_DOCS_PATH ;\
    echo https://github.com/norwoodj/helm-docs/releases/download/v$${version}/helm-docs_$${version}_linux_$${arch}.tar.gz ;\
    curl -LO https://github.com/norwoodj/helm-docs/releases/download/v$${version}/helm-docs_$${version}_linux_$${arch}.tar.gz ;\
    tar -zxvf helm-docs_$${version}_linux_$${arch}.tar.gz ;\
    mv helm-docs $(GOBIN)/helm-docs ;\
	rm -rf $$HELM_DOCS_TMP_DIR ;\
	}
HELM_DOCS=$(GOBIN)/helm-docs
else
HELM_DOCS=$(shell which helm-docs)
endif