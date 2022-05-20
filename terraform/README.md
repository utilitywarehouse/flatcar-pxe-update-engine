# flatcar-pxe-update-engine/terraform

## Input Variables

Refer to [variables.tf](variables.tf).

## Outputs

At a minimum you should include these outputs in your ignition config:

- `unit` - ignition systemd unit
- `file` - pxe-update-agent binary ignition file

If you intend to use `locksmithd` to orchestrate updates, then you should use the
following:

- `locksmithd_unit` - ignition systemd unit for `locksmithd.service` that
  supports PXE booted hosts

## Usage

```hcl
module "pxe_update_engine" {
  source = "github.com/utilitywarehouse/flatcar-pxe-update-engine//terraform?ref=master"

  version_url = "http://my-flatcar-mirror.example.com/assets/flatcar/stable/version.txt"
}

data "ignition_config" "node" {
  files = [
    module.pxe_update_engine.file,
  ]
  systemd = [
    module.pxe_update_engine.unit,
    module.pxe_update_engine.locksmithd_unit,
  ]
}
```
