# flatcar-merit-update-engine

Substitute for the [Flatcar update_engine](https://github.com/kinvolk/update_engine) that enables automated reboots on machines booted
from PXE.

## How it works

The update engine periodically checks the release version reported by
`https://stable.release.flatcar-linux.net/amd64-usr/current/version.txt` (or another, configurable URL) and changes the update status to
`UPDATE_STATUS_UPDATED_NEED_REBOOT` when it differs from the current OS release version. 

It implements the same DBus interface as the original `update_engine`, so it works seamlessly with the `update_engine_client` and reboot
orchestrators like [locksmithd](https://github.com/kinvolk/locksmithd) and
[flatcar-linux-update-operator]((https://github.com/kinvolk/flatcar-linux-update-operator).
