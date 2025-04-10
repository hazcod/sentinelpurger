
all: run

run:
	go run ./cmd/... -config=dev.yml

build:
	$$GOPATH/bin/goreleaser build --config=.github/goreleaser.yml --clean --snapshot

clean:
	rm -r dist/ sentinelpurger || true

update:
	go get -u ./...
