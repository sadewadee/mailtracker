# Eximmon Makefile

.PHONY: build build-linux install clean

build:
	go build -o bin/eximmon main.go config.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/eximmon main.go config.go

install: build-linux
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "Error: Run as root (sudo make install)"; \
		exit 1; \
	fi
	@./install.sh

clean:
	rm -f bin/eximmon
	rm -f .eximmon.conf
