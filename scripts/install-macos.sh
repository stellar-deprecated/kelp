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
        echo "Golang is not installed. Calling goinstaller-macos.sh"
        chmod +x ./goinstaller-macos.sh
        ./goinstaller-macos.sh
    fi

    cloneIntoDir
}

# Once we have Go working finish the install processes inside the development directory to avoid errors (Glide)
function cloneIntoDir() {
    if go version; then
        echo "Setting up Kelp folders in the Golang working directory"
        mkdir $GOPATH/src/github.com/stellar/kelp/

        echo "Cloning Kelp into $GOPATH/src/github.com/stellar/kelp"
        git clone https://github.com/stellar/kelp.git $GOPATH/src/github.com/stellar/kelp

        cd $GOPATH/src/github.com/stellar/kelp

        isGlide
    else 
        echo "Golang not installed, try again."
        exit
    fi
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
            exit
        fi
    fi

    isAstilectron
}

function isAstilectron() {
    if go version; then
        go get -u github.com/asticode/go-astilectron-bundler/... 
        go install github.com/asticode/go-astilectron-bundler/astilectron-bundler
    else
        echo "Golang cannot install Astilectron"
        exit
    fi

    buildAndRun
}

function buildAndRun() {
	echo "Building Kelp binaries"
	./scripts/build.sh

    echo "Confirming the Kelp binary exists with version information."
    if ./bin/kelp version; then
        echo "Kelp has built successfully"
    else 
        echo "The Kelp build was not successful"
    fi

	echo "run the GUI"
	./bin/kelp server 
}

isNode

echo "Remember, PostgreSQL must be running to store data."
echo "Remember, Docker and CCXT must be configured for the expanded set of priceFeeds and orderbooks."
