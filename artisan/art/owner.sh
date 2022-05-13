#!/bin/zsh
# changes the ownership to root before tarring
sudo chown root bin/darwin/amd64/art
sudo chown root bin/darwin/arm64/art
sudo chown root bin/linux/amd64/art
# tarring binaries
tar -zcvf art_linux_amd64.tar.gz -C bin/linux/amd64 .
tar -zcvf art_darwin_amd64.tar.gz -C bin/darwin/amd64 .
tar -zcvf art_darwin_arm64.tar.gz -C bin/darwin/arm64 .
# removing binaries with root owner
rm -r bin