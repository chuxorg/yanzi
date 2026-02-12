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
  echo "Tagging $repo with $VERSION"
  cd "$BASE/$repo"
  tag_if_missing "$repo" "$VERSION"
done

echo "Tagging distribution repo last"
cd "$BASE/yanzi"
tag_if_missing "yanzi" "$VERSION"

echo "Release triggered."
