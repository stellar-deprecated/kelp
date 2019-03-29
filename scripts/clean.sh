#!/bin/bash

echo "removing files ..."
rm -vrf bin
rm -vrf build
rm -vrf gui/web/build
rm -vf gui/filesystem_vfsdata_release.go
echo "... done"
