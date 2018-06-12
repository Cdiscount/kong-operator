HUB :=
REPO := etiennecoutaud
IMAGE := kong-operator
TAG := dev

build:
	go build github.com/cdiscount/kong-operator/cmd/kong-operator

run: build
	kubectl apply -f manifests/crd.yml
	./kong-operator -kubeconfig=$(HOME)/.kube/config -v=2 -logtostderr=true

darwin:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s" -o kong-operator github.com/cdiscount/cmd/kong-operator

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s" -o kong-operator github.com/cdiscount/cmd/kong-operator

test: 
	go test  $(shell go list ./... | grep -v fake) -coverprofile=coverage.txt -covermode=atomic

image:
	docker build -t $(REPO)/$(IMAGE):$(TAG) .

dep:
	glide up

gen:
	hack/update-codegen.sh

.PHONY: image build run test darwin linux dep gen 
