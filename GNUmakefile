MSG ?= chore: provider update


default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

local: 
	make install; cd examples; terraform init; terraform plan; terraform apply --auto-approve; cd -

.PHONY: fmt lint test testacc build install generate

gitpush:
	git add .; git commit -m "$(MSG)"; git push

gittag:
	git tag -a v${VERSION} -m "version ${VERSION}"; git push origin v${VERSION}