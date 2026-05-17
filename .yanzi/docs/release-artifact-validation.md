# Release Artifact Validation

## Candidate

- Tag: `v2.9.1-rc1`
- Commit SHA: `bceb106b0fa97d6574fa1aa5d419f489f3e935c4`
- Release URL: `https://github.com/chuxorg/yanzi/releases/tag/v2.9.1-rc1`

## Artifact Inventory

- `yanzi-darwin-amd64`
  - URL: `https://github.com/chuxorg/yanzi/releases/download/v2.9.1-rc1/yanzi-darwin-amd64`
  - SHA256: `eba6e6d3ec975b2b436632d7a7810f763f338bb387a65cfaf0d3e711162fa4c4`
- `yanzi-darwin-arm64`
  - URL: `https://github.com/chuxorg/yanzi/releases/download/v2.9.1-rc1/yanzi-darwin-arm64`
  - SHA256: `a1cc830efcc0d15b86191f53bbba85455c5993ef5910c6774de28eba98924b55`
- `yanzi-linux-amd64`
  - URL: `https://github.com/chuxorg/yanzi/releases/download/v2.9.1-rc1/yanzi-linux-amd64`
  - SHA256: `b31f64940ef28efd13d68bc3ee8d644db12d2fc1968d73c7d635a10673302104`
- `yanzi-windows-amd64.zip`
  - URL: `https://github.com/chuxorg/yanzi/releases/download/v2.9.1-rc1/yanzi-windows-amd64.zip`
  - SHA256: `c4a8f84ad5cc5fb9f94699d6c1e7ab13ee5e55b2253c34af93c2d689ca53b793`
- `sha256sums.txt`
  - URL: `https://github.com/chuxorg/yanzi/releases/download/v2.9.1-rc1/sha256sums.txt`

## Lineage Observations

- Release tag resolves to the certified candidate SHA.
- Asset URLs are stable and installer-accessible under the tagged release path.
- Artifact checksums are explicitly published for operational verification.
- Release object availability removes mutable-latest ambiguity for pinned installer installs.
