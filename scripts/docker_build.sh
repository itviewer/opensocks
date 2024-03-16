#!bin/bash
NAME="opensocks"
ARCH=`uname -m`
if [ $ARCH == 'x86_64' ]
then
TAG="latest"
elif [ $ARCH == 'aarch64' ]
then
TAG="arm64"
else
TAG=$ARCH
fi
echo "build $NAME:$TAG"
git pull
docker build . -t itviewer/$NAME:$TAG
docker image push itviewer/$NAME:$TAG

echo "DONE!!!"
