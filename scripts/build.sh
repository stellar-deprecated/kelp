#!/bin/bash

function usage() {
    echo "Usage: $0 [flags]"
    echo ""
    echo "Flags:"
    echo "    -d, --deploy        prepare tar archives in build/, only works on a tagged commit in the format v1.0.0 or v1.0.0-rc1"
    echo "    -h, --help          show this help info"
    exit 1
}

if [[ ($# -gt 1 || `pwd | rev | cut -d'/' -f1 | rev` != "kelp") ]]
then
    echo "need to invoke from the root 'kelp' directory"
    exit 1
fi

if [[ ($# -eq 1 && ("$1" == "-d" || "$1" == "--deploy")) ]]; then
    MODE=deploy
elif [[ ($# -eq 1 && ("$1" == "-h" || "$1" == "--help")) ]]; then
    usage
elif [[ $# -eq 1 ]]; then
    usage
else
    MODE=build
fi

# version is git tag if it's available, otherwise git hash
VERSION=$(git describe --always --abbrev=8 --dirty --tags)
GIT_HASH=$(git describe --always --abbrev=50 --dirty --long)
DATE=$(date -u +%"Y%m%dT%H%M%SZ")
LDFLAGS="-X github.com/lightyeario/kelp/cmd.version=$VERSION -X github.com/lightyeario/kelp/cmd.gitHash=$GIT_HASH -X github.com/lightyeario/kelp/cmd.buildDate=$DATE"

echo "version: $VERSION"
echo "git hash: $GIT_HASH"
echo "build date: $DATE"
echo ""

if [[ $MODE == "build" ]]
then
    # explicit check for windows
    EXTENSION=""
    if [[ `go env GOOS` == "windows" ]]
    then
        EXTENSION=".exe"
    fi

    # generate outfile
    OUTFILE=bin/kelp$EXTENSION
    mkdir -p bin

    echo -n "compiling ... "
    go build -ldflags "$LDFLAGS" -o $OUTFILE
    BUILD_RESULT=$?
    if [[ $BUILD_RESULT -ne 0 ]]
    then
        echo ""
        echo "build failed with error code $BUILD_RESULT"
        exit $BUILD_RESULT
    fi
    echo "successful: $OUTFILE"
    echo ""
    echo "BUILD SUCCESSFUL"
    exit 0
fi
# else, we are in deploy mode

if ! [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-rc[1-9]+)?$ ]]
then
    echo "error: the git commit needs to be tagged with a valid version to prepare archives, see $0 -h for more information"
    exit 1
fi

ARCHIVE_DIR=build/$DATE
ARCHIVE_FOLDER_NAME=kelp-$VERSION
ARCHIVE_DIR_SOURCE=$ARCHIVE_DIR/$ARCHIVE_FOLDER_NAME
mkdir -p $ARCHIVE_DIR_SOURCE
OUTFILE=$ARCHIVE_DIR_SOURCE/kelp
cp examples/configs/trader/* $ARCHIVE_DIR_SOURCE/

PLATFORM_ARGS=("darwin amd64" "linux amd64" "windows amd64" "linux arm64" "linux arm 5" "linux arm 6" "linux arm 7")
for args in "${PLATFORM_ARGS[@]}"
do
    # extract vars
    GOOS=`echo $args | cut -d' ' -f1 | tr -d ' '`
    GOARCH=`echo $args | cut -d' ' -f2 | tr -d ' '`
    GOARM=`echo $args | cut -d' ' -f3 | tr -d ' '`
    echo -n "compiling for (GOOS=$GOOS, GOARCH=$GOARCH, GOARM=$GOARM) ... "

    # explicit check for windows
    BINARY="$OUTFILE"
    if [[ "$GOOS" == "windows" ]]
    then
        BINARY="$OUTFILE.exe"
    fi

    # compile
    env GOOS=$GOOS GOARCH=$GOARCH GOARM=$GOARM go build -ldflags "$LDFLAGS" -o $BINARY
    BUILD_RESULT=$?
    if [[ $BUILD_RESULT -ne 0 ]]
    then
        echo ""
        echo "build failed with error code $BUILD_RESULT"
        exit $BUILD_RESULT
    fi
    echo "successful"

    # archive
    ARCHIVE_FILENAME=kelp-$VERSION-$GOOS-$GOARCH$GOARM.tar
    cd $ARCHIVE_DIR
    echo -n "archiving binary file ... "
    tar cf ${ARCHIVE_FILENAME} $ARCHIVE_FOLDER_NAME
    TAR_RESULT=$?
    cd ../../
    if [[ $TAR_RESULT -ne 0 ]]
    then
        echo ""
        echo "archiving failed with error code $TAR_RESULT"
        exit $TAR_RESULT
    fi
    echo "successful: ${ARCHIVE_DIR}/${ARCHIVE_FILENAME}"

    echo -n "cleaning up binary: $BINARY ... "
    rm $BINARY
    echo "successful"
    echo ""
done

echo -n "cleaning up kelp folder: $ARCHIVE_DIR_SOURCE/ ... "
cd $ARCHIVE_DIR
rm $ARCHIVE_FOLDER_NAME/*
rmdir $ARCHIVE_FOLDER_NAME
cd ../../
echo "done"

echo ""
echo "BUILD SUCCESSFUL"
