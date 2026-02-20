BINARY_NAME=things3-api
INSTALL_PATH=/usr/local/bin/$(BINARY_NAME)
PLIST_NAME=com.things3api.plist
PLIST_SRC=launchd/$(PLIST_NAME)
PLIST_DEST=$(HOME)/Library/LaunchAgents/$(PLIST_NAME)

.PHONY: build run install uninstall restart logs clean token

build:
	CGO_ENABLED=0 go build -o $(BINARY_NAME) .

run:
	go run .

install: build
	cp $(BINARY_NAME) $(INSTALL_PATH)
	mkdir -p $(HOME)/Library/LaunchAgents
	cp $(PLIST_SRC) $(PLIST_DEST)
	launchctl load $(PLIST_DEST)

uninstall:
	-launchctl unload $(PLIST_DEST)
	rm -f $(PLIST_DEST)
	rm -f $(INSTALL_PATH)

restart:
	-launchctl unload $(PLIST_DEST)
	launchctl load $(PLIST_DEST)

logs:
	tail -f /tmp/things3-api.log

clean:
	rm -f $(BINARY_NAME)

token:
	@openssl rand -hex 32
