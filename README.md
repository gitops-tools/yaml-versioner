# yaml-versioner

This is a small tool for bumping [SemVer](https://semver.org/) versions in YAML files.

**NOTE**: This is a very early version - it works, but it's not particularly
hardened.

## Usage

```shell
$ yaml-versioner increment --filename pkg/versioner/testdata/Chart.yaml --key version --patch=true
```

This would update the file `pkg/versioner/testdata/Chart.yaml` as a YAML file, incrementing the [SemVer](https://semver.org/) in the top-level `.version` key, but only incrementing the _patch_ level.

If the file looked like this:

```yaml
apiVersion: v2
name: test-controller
description: A Helm chart for Kubernetes
type: application
version: 1.0.1
appVersion: "1.0.0"
```

Then it would get changed to look like this:

```yaml
apiVersion: v2
name: test-controller
description: A Helm chart for Kubernetes
type: application
version: 1.0.2
appVersion: "1.0.0"
```
With a diff of

```diff
type: application
- version: 1.0.1
+ version: 1.0.2
appVersion: "1.0.0"
```

## Incrementing the minor version

Incrementing the _minor_ level will result in the patch level being reset to zero.

$ yaml-versioner increment --filename pkg/versioner/testdata/Chart.yaml --key version --minor=true

```diff
type: application
- version: 1.0.1
+ version: 1.1.0
appVersion: "1.0.0"
```

## TODO

* [ ] [Pre-release data](https://semver.org/#spec-item-9) in the CLI
* [ ] [Build metadata](https://semver.org/#spec-item-10) in the CLI
* [ ] Replace rather than increment (set a specific version) for example, this would allow bumping the `.appVersion` in the YAML above
