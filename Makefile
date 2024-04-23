# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

ifeq ($(shell uname),Darwin)
SED_COMMAND=sed -i '' -n '/rules/,$$p'
else
SED_COMMAND=sed -i -n '/rules/,$$p'
endif

# generate: generate-controller generate-groups rbacs manifests fmt
generate: generate-controller rbacs manifests fmt

#generate helm documentation
docs: helm-docs
	$(HELM_DOCS) -t deployments/node/README.gotmpl deployments/node

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	rm -f deployments/node/crds/*
	$(CONTROLLER_GEN) paths="./apis/..." crd:generateEmbeddedObjectMeta=true output:crd:artifacts:config=deployments/node/crds

#Generate RBAC for each controller
rbacs: controller-gen
	rm -f deployments/node/files/*

	$(CONTROLLER_GEN) paths="./pkg/local-resource-manager" rbac:roleName=node-local-resource-manager output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/node/files/node-local-resource-manager-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' && $(SED_COMMAND) deployments/node/files/node-local-resource-manager-ClusterRole.yaml
	$(CONTROLLER_GEN) paths="./pkg/rear-manager/" rbac:roleName=node-rear-manager output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/node/files/node-rear-manager-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' && $(SED_COMMAND) deployments/node/files/node-rear-manager-ClusterRole.yaml
	$(CONTROLLER_GEN) paths="./pkg/rear-controller/..." rbac:roleName=node-rear-controller output:rbac:stdout | awk -v RS="---\n" 'NR>1{f="./deployments/node/files/node-rear-controller-" $$4 ".yaml";printf "%s",$$0 > f; close(f)}' &&  $(SED_COMMAND) deployments/node/files/node-rear-controller-ClusterRole.yaml

# Install gci if not available
gci:
ifeq (, $(shell which gci))
	@go install github.com/daixiang0/gci@v0.11.2
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
	find . -type f -name '*.go' -a ! -name '*zz_generated*' -exec $(GCI) write -s standard -s default -s "prefix(github.com/fluidos-project/node)" {} \;
	find . -type f -name '*.go' -exec $(ADDLICENSE) -l apache -c "FLUIDOS Project" -y "2022-$(shell date +%Y)" {} \;

# Install golangci-lint if not available
golangci-lint:
ifeq (, $(shell which golangci-lint))
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2
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

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0
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