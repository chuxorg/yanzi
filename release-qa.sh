#!/usr/bin/env bash
set -euo pipefail

VERSION=${1:-}

if [ -z "$VERSION" ]; then
  echo "Usage: ./release-qa.sh vX.Y.Z-qa"
  exit 1
fi

if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-qa$ ]]; then
  echo "Invalid version: $VERSION"
  echo "Expected format: vX.Y.Z-qa"
  exit 1
fi

REPOS=(
  chux-yanzi-core
  chux-yanzi-cli
  chux-yanzi-emitter
  chux-yanzi-library
)

BASE=~/projects/chuxorg

ensure_dev_branch() {
  local repo="$1"
  local branch

  branch="$(git rev-parse --abbrev-ref HEAD)"
  if [ "$branch" != "development" ]; then
    echo "Switching $repo to development branch"
    git fetch origin development
    git checkout development
  fi

  git pull --ff-only origin development
}

tag_if_missing() {
  local repo="$1"
  local version="$2"

  if git rev-parse -q --verify "refs/tags/$version" >/dev/null; then
    echo "Tag $version already exists locally in $repo. Skipping."
    return 0
  fi

  if git ls-remote --tags origin "$version" | grep -q "$version"; then
    echo "Tag $version already exists on origin for $repo. Skipping."
    return 0
  fi

  git tag "$version"
  git push origin "$version"
}

for repo in "${REPOS[@]}"; do
  echo "Tagging $repo with $VERSION from development"
  cd "$BASE/$repo"
  ensure_dev_branch "$repo"
  tag_if_missing "$repo" "$VERSION"
done

echo "Tagging distribution repo last (master)"
cd "$BASE/yanzi"
branch="$(git rev-parse --abbrev-ref HEAD)"
if [ "$branch" != "master" ]; then
  echo "Switching yanzi to master branch"
  git fetch origin master
  git checkout master
fi

git pull --ff-only origin master
tag_if_missing "yanzi" "$VERSION"

echo "QA release tags pushed."
