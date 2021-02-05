#!/bin/sh
# shellcheck disable=SC2016
set -e

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
        echo "Golang is installed at $GOPATH"
    else
        echo "Golang is not installed. Starting Golang install..."

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

        if $ZSH_VERSION; then
            shell_profile="$HOME/.zshrc"
        elif $BASH_VERSION; then
            shell_profile="$HOME/.bashrc"   
        elif $FISH_VERSION; then
            shell="fish"
            if [ -d "$XDG_CONFIG_HOME" ]; then
                shell_profile="$XDG_CONFIG_HOME/fish/config.fish"
            else
                shell_profile="$HOME/.config/fish/config.fish"
            fi
        fi

        echo "shell_profile set to $shell_profile"

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
        mkdir -p "$GOROOT"

        tar -C "$GOROOT" --strip-components=1 -xzf "$TEMP_DIRECTORY/go.tar.gz"

        echo "Configuring shell profile in: $shell_profile"
        touch "$shell_profile"
        if [ "$shell" == "fish" ]; then
            {
                echo '# GoLang'
                echo "set GOROOT '${GOROOT}'"
                echo "set GOPATH '$GOPATH'"
                echo 'set PATH $GOPATH/bin $GOROOT/bin $PATH'
            } >> "$shell_profile"
        else
            {
                echo '# GoLang'
                echo "export GOROOT=${GOROOT}"
                echo 'export PATH=$GOROOT/bin:$PATH'
                echo "export GOPATH=$GOPATH"
                echo 'export PATH=$GOPATH:$PATH'
            } >> "$shell_profile"
        fi

        mkdir -p "${GOPATH}/"{src,pkg,bin}
        echo -e "\nGo $VERSION was installed into $GOROOT.\nMake sure to relogin into your shell or run:"
        echo -e "\n\tsource $shell_profile\n\nto update your environment variables."
        echo "Tip: Opening a new terminal window usually just works. :)"
        rm -f "$TEMP_DIRECTORY/go.tar.gz"
        exit
    fi

    cloneIntoDir
}

# Once we have Go working finish the install processes inside the development directory to avoid errors (Glide)
function cloneIntoDir() {
    if go version; then
        echo "Setting up Kelp folders in the Golang working directory"
        echo $GOPATH
        cd $GOPATH

        # pwd
        sudo mkdir github.com/stellar/kelp

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
