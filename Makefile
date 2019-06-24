RDL ?= $(GOPATH)/bin/rdl
DOCKER_IMAGE_NAME ?= registry.gitlab.com/cty3000/superman-detector

all: go/bin/supermandetectord

go/bin/supermandetectord: go/src/supermandetector supermandetectord-external-libs test
	go build

supermandetectord-external-libs:
	GO111MODULE=on go mod vendor

go/src/supermandetector: $(RDL)
	rm -rf supermandetector
	mkdir -p supermandetector
	$(RDL) -ps generate -t -o supermandetector go-model rdl/SupermanDetector.rdl
	$(RDL) -ps generate -t -o supermandetector go-server rdl/SupermanDetector.rdl
	$(RDL) -ps generate -t -o supermandetector go-client rdl/SupermanDetector.rdl
	echo "package supermandetector" > supermandetector/doc.go

$(RDL):
	go get github.com/ardielle/ardielle-tools/...

test: go/bin/supermandetectord
	go test -v ./...

ifeq  ($(shell uname),Darwin)
coverage:
	go test -v -race -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	rm -f coverage.out
	open coverage.html
else
coverage:
	go test -v -race -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	rm -f coverage.out
endif

lint:
	gometalinter --enable-all . | rg -v comment

contributors:
	git log --format='%aN <%aE>' | sort -fu > CONTRIBUTORS

bench: clean supermandetectord-external-libs
	go test -count=5 -run=NONE -bench . -benchmem

docker-build: test
	docker build -t $(DOCKER_IMAGE_NAME) .

docker-publish: docker-build
	docker push $(DOCKER_IMAGE_NAME)

docker: docker-publish

clean::
	rm -rf superman-detector supermandetector $$GOPATH/bin/rdl vendor go.mod go.sum
