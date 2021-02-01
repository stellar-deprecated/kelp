#!/bin/sh

function isNode() {
    clear
    if node -v; then
        echo "Node is installed"
        isYarn
    else
        echo "Node is not installed."
    fi
}

function isYarn() {
    if yarn --version; then
        echo "Yarn is installed"
        isGo
    else
        echo "Yarn is not installed."
    fi
}

function isGo() {
    if go version; then
        echo "Golang is installed"
        echo "GOPATH is currently $GOPATH"
    else
        echo "Golang is not installed. Calling install script from git.io/vQhTU"
        if curl --version; then
			# macOS typically has curl installed
			curl -L https://raw.githubusercontent.com/canha/golang-tools-install-script/master/goinstall.sh | bash
        elif wget --version; then
	       	# Linux typically has wget installed
			wget -q -O - https://raw.githubusercontent.com/canha/golang-tools-install-script/master/goinstall.sh | bash
        else
        	echo "curl and wget are not available, install Golang manually youtube.com/watch?v=MbS1wn0B-fk"
        fi
    fi
    isGlide
}

function isGlide() {
	if glide --version; then
        echo "Glide is installed"
    else
        echo "Installing Glide."
        if curl --version; then
        	curl https://glide.sh/get | sh
        elif wget --version; then
        	wget https://glide.sh/get | sh
        else
        	echo "curl and wget are not available, install glide manually https://github.com/Masterminds/glide"
        fi
    fi
    glide install
    isAstilectron
}

function isAstilectron() {
	go get -u github.com/asticode/go-astilectron-bundler/... 
	go install github.com/asticode/go-astilectron-bundler/astilectron-bundler

    cloneAndBuild
}

function cloneAndBuild() {
	# setup Kelp directories in the correct location
	mkdir $GOPATH/src/github.com/stellar/kelp/

	git clone https://github.com/stellar/kelp.git $GOPATH/src/github.com/stellar/kelp

	cd $GOPATH/src/github.com/stellar/kelp

	# Build the binaries using the provided build script (the go install command will produce a faulty binary):
	./scripts/build.sh

	# Confirm one new binary file exists with version information.
	./bin/kelp version

	# run the GUI
	./bin/kelp server 
}

isNode

echo "Remember, PostgreSQL must be running to store data."
echo "Remember, Docker and CCXT must be configured for the expanded set of priceFeeds and orderbooks."
