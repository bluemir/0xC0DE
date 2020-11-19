#!/bin/sh

rm -rf .git readme.md
git init
git commit -m "Initial commit" --allow-empty --author="init-bot <bot@bluemir.me>"
rm init.sh
