HUB :=
REPO := etiennecoutaud
IMAGE := kong-operator
TAG := dev

build:
	go build -i github.com/etiennecoutaud/cdiscount/cmd/kong-operator

run: build
	kubectl apply -f manifests/kanary-crd.yml
	./kong-operator -kubeconfig=$(HOME)/.kube/config -v=2 -logtostderr=true

darwin:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s" -o kanary-operator github.com/etiennecoutaud/kanary/cmd/kanary-operator

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s" -o kanary-operator github.com/etiennecoutaud/kanary/cmd/kanary-operator

test: 
	go test  $(shell go list ./... | grep -v fake) -coverprofile=coverage.txt -covermode=atomic

image:
	docker build -t $(REPO)/$(IMAGE):$(TAG) .

dep:
	glide up

gen:
	hack/update-codegen.sh

.PHONY: build run test darwin linux dep gen image
