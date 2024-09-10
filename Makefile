# See https://makefiletutorial.com for in depth examples

CURRENT_VERSION := $(shell cat VERSION)
NEXT_MAJOR_VERSION := $(shell sh scripts/semverinc.sh $(CURRENT_VERSION) 0)
NEXT_MINOR_VERSION := $(shell sh scripts/semverinc.sh $(CURRENT_VERSION) 1)
NEXT_PATCH_VERSION := $(shell sh scripts/semverinc.sh $(CURRENT_VERSION) 2)

build-timezone-data:
	mkdir -p bin
	go build -o bin/tz cmd/timezone/main.go
	./bin/tz -build -db data/timezone.data

.PHONY: test
test:
	go test ./...

test-verbose:
	go test -v

code-coverage:
	go test -cover -coverprofile coverage.out ./...
	gocovsh

test-versioning:
	@echo Sematic versioning test...
	@echo ___Current version: v$(CURRENT_VERSION)
	@echo Next patch version: v$(NEXT_PATCH_VERSION)
	@echo Next minor version: v$(NEXT_MINOR_VERSION)
	@echo Next major version: v$(NEXT_MAJOR_VERSION)

update: update_internal lint test-verbose

update_internal:
	go get -u ./...
	go mod tidy

lint:
	golangci-lint run

create-next-major-version:
	@echo    Current Version: v$(CURRENT_VERSION)
	@echo Next Major Version: v$(NEXT_MAJOR_VERSION)
	@echo $(NEXT_MAJOR_VERSION) > VERSION
	@git add .
	@git commit -m "ARelease new major version: (v$(NEXT_MAJOR_VERSION))"
	@git tag v$(NEXT_MAJOR_VERSION)
	@echo In order to update tags run: git push --tags

create-next-minor-version:
	@echo    Current Version: v$(CURRENT_VERSION)
	@echo Next Minor Version: v$(NEXT_MINOR_VERSION)
	@echo $(NEXT_MINOR_VERSION) > VERSION
	@git add .
	@git commit -m "Release new minor version: (v$(NEXT_MINOR_VERSION))"
	@git tag v$(NEXT_MINOR_VERSION)
	@echo In order to update tags run: git push --tags

create-next-patch-version:
	@echo    Current Version: v$(CURRENT_VERSION)
	@echo Next Patch Version: v$(NEXT_PATCH_VERSION)
	@echo $(NEXT_PATCH_VERSION) > VERSION
	@git add .
	@git commit -m "Release new patch version: (v$(NEXT_PATCH_VERSION))"
	@git tag v$(NEXT_PATCH_VERSION)
	@echo In order to update tags run: git push --tags

#delete-all-tags:
#	git tag -l | xargs git tag -d