FROM golang:1.10
COPY . /go/src/github.com/cdiscount/kong-operator
WORKDIR /go/src/github.com/cdiscount/kong-operator/cmd/kong-operator
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s" -o kong-operator

FROM scratch
COPY --from=0 /go/src/github.com/cdiscount/kong-operator/cmd/kong-operator/kong-operator /
LABEL app.language=golang app.name=kong-operator
EXPOSE 8080
ENTRYPOINT ["/kong-operator", "-logtostderr",  "-v=2"]
