# Chunk

## Motivation

- 冗長性、シーケンシャルのアクセス性能を提供する
- 既存の論文における...
  - [Google FileSystem](http://research.google.com/archive/gfs.html) の Chunk
  - [Windows Azure Storage](https://azure.microsoft.com/ja-jp/blog/sosp-paper-windows-azure-storage-a-highly-available-cloud-storage-service-with-strong-consistency/) の Block / Extent
  - [Ceph](https://docs.ceph.com/docs/mimic/architecture/#data-striping) の stripe unit
