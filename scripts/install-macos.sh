#!/bin/sh
# shellcheck disable=SC2016
set -e

# Set this machine's dependencies status(es) with the OK flag
OK=true;

# Dependency test functions begin here

function isNode() {
    if node -v; then
        OK=true;
    else
        OK=false;
    fi
}

function isYarn() {
    if yarn --version; then
        OK=true;    
    else
        OK=false;
    fi
}

function isGit() {
    if git --version; then
        OK=true;
    else
        OK=false;
    fi
}

function isGo() {
    if go version; then
        OK=true;
    else
        OK=false;
    fi
}

function isGlide() {
    if glide --version; then
        OK=true;
    else
        OK=false;
    fi
}

# Installation functions begin here

function installGo() {
    echo 'Installing Golang on your machine.'
    # Go install script based on https://github.com/canha/golang-tools-install-script
    VERSION="1.15.7"

    [ -z "$GOROOT" ] && GOROOT="/usr/local/go"
    [ -z "$GOPATH" ] && GOPATH="$HOME/go"

    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case $OS in
        "Linux")
            case $ARCH in
            "x86_64")
                ARCH=amd64\
                ;;
            "aarch64")
                ARCH=arm64
                ;;
            "armv6" | "armv7l")
                ARCH=armv6l
                ;;
            "armv8")
                ARCH=arm64
                ;;
            .*386.*)
                ARCH=386
                ;;
            esac
            PLATFORM="linux-$ARCH"
        ;;
        "Darwin")
            PLATFORM="darwin-amd64"
        ;;
    esac

    if [ -z "$PLATFORM" ]; then
        echo "Your operating system is not supported by the script."
        exit 1
    fi

    if [ -n "$($SHELL -c 'echo $BASH_VERSION')" ]; then
        shell_profile="$HOME/.bashrc"
    else
        echo "Kelp requires bash"
    fi

    if [ -d "$GOROOT" ]; then
        echo "The Go install directory ($GOROOT) already exists. Exiting."
        exit 1
    fi

    PACKAGE_NAME="go$VERSION.$PLATFORM.tar.gz"
    TEMP_DIRECTORY=$(mktemp -d)

    echo "Downloading $PACKAGE_NAME ..."
    if hash wget 2>/dev/null; then
        wget https://storage.googleapis.com/golang/$PACKAGE_NAME -O "$TEMP_DIRECTORY/go.tar.gz"
    else
        curl -o "$TEMP_DIRECTORY/go.tar.gz" https://storage.googleapis.com/golang/$PACKAGE_NAME
    fi

    if [ $? -ne 0 ]; then
        echo "Download failed! Exiting."
        exit 1
    fi

    echo "Extracting File..."

    sudo chown -R $USER: $HOME # https://github.com/golang/go/issues/27187
    mkdir -p "$GOROOT"

    tar -C "$GOROOT" --strip-components=1 -xzf "$TEMP_DIRECTORY/go.tar.gz"

    GOROOT=${GOROOT}
    PATH=$GOROOT/bin:$PATH
    GOPATH=$GOPATH
    PATH=$GOPATH:$PATH

    mkdir -p "${GOPATH}/"{src,pkg,bin}
    echo -e "\nGo $VERSION was installed into $GOROOT.\nMake sure to relogin into your shell or run:"
    echo -e "\n\ restart script to update your environment variables."
    echo "Tip: Opening a new terminal window usually just works. :)"
    rm -f "$TEMP_DIRECTORY/go.tar.gz"
    exit
}

# Once we have Golang; finish the install processes inside the development directory to avoid errors (Glide)
function cloneIntoDir() {
    echo "Setting up Kelp folders in the Golang working directory"
    echo $GOPATH
    cd $GOPATH

    # pwd
    sudo mkdir github.com/stellar/kelp

    echo "Cloning Kelp into $GOPATH/src/github.com/stellar/kelp"
    git clone https://github.com/stellar/kelp.git $GOPATH/src/github.com/stellar/kelp

    cd $GOPATH/src/github.com/stellar/kelp
}

# After Golang install Glide
function installGlide() {
    echo "Installing Glide."
    if curl --version; then
        curl https://glide.sh/get | sh
    elif wget --version; then
        wget https://glide.sh/get | sh
    else
        echo "curl and wget are not available, install glide manually https://github.com/Masterminds/glide"
        exit
    fi
}

# After Glide install Astilectron
function installAstilectron() {
    go get -u github.com/asticode/go-astilectron-bundler/... 
    go install github.com/asticode/go-astilectron-bundler/astilectron-bundler
}

# Finally install build Kelp and run it
function buildKelpAndRun() {
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
if [ $OK ]; then
    echo "Node is installed on your machine."
else
    echo "Node is not installed on your machine!"
    exit 1
fi

isYarn
if [ $OK ]; then
    echo "Yarn is installed on your machine."
else
    echo "Yarn is not installed on your machine!"
    exit 1
fi

isGit
if [ $OK ]; then
    echo "Git is installed on your machine."
else
    echo "Git is not installed on your machine!"
    exit 1
fi

isGo
echo $OK
if $OK; then
    echo "true for some reason"
    cloneIntoDir
    isGlide
    installAstilectron
    buildAndRun
    echo "Remember, PostgreSQL must be running to store data."
    echo "Remember, Docker and CCXT must be configured for the expanded set of priceFeeds and orderbooks."
else
    installGo
fi
