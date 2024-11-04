# yaml-versioner

This is a small tool for bumping [SemVer](https://semver.org/) versions in YAML files.

**NOTE**: This is a very early version - it works, but it's not particularly
hardened.

Usage:

```shell
$ yaml-versioner increment --filename pkg/versioner/testdata/Chart.yaml --key version --patch=true
```

## TODO

[ ] Metadata in the CLI
[ ] Replace rather than increment (set a specific version)
