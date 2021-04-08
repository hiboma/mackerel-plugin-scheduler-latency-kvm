VERSION  := $(shell git tag | sed 's/v//g' |sort --version-sort | tail -n1)

.PHONY: release_major
release_major: releasedeps
	git semv major --bump

.PHONY: release_minor
release_minor: releasedeps
	git semv minor --bump

.PHONY: release_patch
release_patch: releasedeps
	git semv patch --bump

.PHONY: releasedeps
releasedeps: git-semv

.PHONY: git-semv
git-semv:
	which git-semv > /dev/null || brew tap linyows/git-semv && brew install git-semv

docker_build:
	docker build -t mackerel-plugin-scheduler-latency-kvm:latest -t mackerel-plugin-scheduler-latency-kvm:$(VERSION) .

build:
	go build .

test:
	go test .

release:
	rm -rf dist/*
	goreleaser --rm-dist --skip-validate
