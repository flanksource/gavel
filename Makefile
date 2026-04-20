# Makefile stub - forwards to Taskfile
# Install task: https://taskfile.dev/installation/

.PHONY: build lint test install restart fmt tidy clean all

build:
	@task build

lint:
	@task lint

test:
	@task test

install:
	@task install

restart:
	@task restart

fmt:
	@task fmt

tidy:
	@task mod

clean:
	@task clean

all:
	@task ci
