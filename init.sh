#!/bin/sh

rm -rf .git readme.md
git init
git commit -m "Initial commit" --allow-empty --author="init-bot <bot@bluemir.me>"

read -p "Application Name? " NAME

find . -type f | xargs -n 1 sed -i "s/0xC0DE/$NAME/g"

read -p "Do you wish to remove init.sh(Y/n)?" yn
case $yn in
	[Nn]* ) exit;;
	* ) rm init.sh ;;
esac
