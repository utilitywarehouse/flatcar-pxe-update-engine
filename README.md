# flatcar-merit-update-engine

Substitute for the [Flatcar
update_engine](https://github.com/kinvolk/update_engine) that enables automated
reboots on machines booted from PXE.

Instead of downloading the new release, it only signals for a reboot. Flatcar's PXE
image does not provide an option to update itself:
https://www.flatcar.org/docs/latest/installing/bare-metal/booting-with-ipxe/#update-process

## How it works

The update engine periodically checks the release version reported by
`https://stable.release.flatcar-linux.net/amd64-usr/current/version.txt` (or
another, configurable URL) and changes the update status to
`UPDATE_STATUS_UPDATED_NEED_REBOOT` when it differs from the current OS release
version.

It implements the same DBus interface as the original `update_engine`, so it
works seamlessly with the `update_engine_client` and reboot orchestrators like
[locksmithd](https://github.com/kinvolk/locksmithd) and
[flatcar-linux-update-operator]((https://github.com/kinvolk/flatcar-linux-update-operator).

## Deploy

See [terraform/](terraform/) for a Terraform module that provides ignition
config for running the merit update engine as a systemd service. Refer to the
[README](terraform/README.md) for instructions.
