#!/bin/sh

set -e

rm -rf .git README.md pkg/api/v1/.gitignore
git init

export GIT_COMMITTER_NAME="init-bot"
export GIT_COMMITTER_EMAIL="bot@bluemir.me"

git commit -m "Initial commit" --allow-empty --author="init-bot <bot@bluemir.me>"

read -p "Application REPO? " REPO
echo $REPO
read -p "Application NAME?(default: $(basename $REPO)) " NAME

if [ "$NAME" =  "" ] ; then
	NAME=$(basename $REPO)
fi
echo $NAME

find . -name init.sh -o -name Makefile -prune -o -type f -print | xargs -n 1 sed -i "s#github.com/bluemir/0xC0DE#$REPO#g"
find . -name init.sh -o -name Makefile -prune -o -type f -print | xargs -n 1 sed -i "s#0xC0DE#$NAME#g"

read -p "Do you wish to remove init.sh(Y/n)? " yn
case $yn in
	[Nn]* ) exit;;
	* ) rm init.sh ;;
esac
