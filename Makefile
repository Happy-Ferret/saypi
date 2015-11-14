COW_PATH = say/internal/cows/
COW_FILES = $(wildcard $(COW_PATH)*.cow)
COW_BUILD = $(COW_PATH)cows.go

.PHONY: test cows clean

test: cows
	go test ./...

cows: $(COW_BUILD)

clean:
	rm $(COW_BUILD)

$(COW_BUILD): $(COW_FILES)
	go-bindata -o $@ -nomemcopy -prefix $(COW_PATH) $(COW_PATH)
	sed -i'' 's/package main/package cows/' $@
