#!/usr/bin/env bash
set -e

VERSION=$1

if [ -z "$VERSION" ]; then
  echo "Usage: ./release.sh v0.x.x"
  exit 1
fi

REPOS=(
  chux-yanzi-core
  chux-yanzi-cli
  chux-yanzi-emitter
  chux-yanzi-library
)

BASE=~/projects/chuxorg

for repo in "${REPOS[@]}"; do
  echo "Tagging $repo with $VERSION"
  cd "$BASE/$repo"
  git tag "$VERSION"
  git push origin "$VERSION"
done

echo "Tagging distribution repo last"
cd "$BASE/yanzi"
git tag "$VERSION"
git push origin "$VERSION"

echo "Release triggered."
