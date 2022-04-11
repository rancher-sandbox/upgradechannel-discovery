# upgradechannel-discovery

`ManagedOSVersion` Discovery plugin for `rancheros-operator`.

This plugin allows to discover new `ManagedOSVersion` associated to a release channel and use it in `ManagedOSVersionChannel`.

Currently supports discovering releases only from github releases, but it is flexible enough to allow other syncronization mechanisms.

## Usage

The discovery plugin is meant to be used within a `ManagedOSVersionChannel` spec. 

```yaml
apiVersion: rancheros.cattle.io/v1
kind: ManagedOSVersionChannel
metadata:
  name: testchannel
  namespace: fleet-default
spec:
  options:
    envs:
    - name: "REPOSITORY"
      value: "rancher-sandbox/os2"
    - name: "IMAGE_PREFIX"
      value: "quay.io/costoolkit/os2"
    args:
    - github
    command:
    - /usr/bin/upgradechannel-discovery
    image: ...
  type: custom
```