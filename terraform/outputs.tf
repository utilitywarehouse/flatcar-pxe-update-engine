output "file" {
  value = data.ignition_file.pxe_update_engine.rendered
}

output "unit" {
  value = data.ignition_systemd_unit.pxe_update_engine.rendered
}

output "locksmithd_unit" {
  value = data.ignition_systemd_unit.locksmithd.rendered
}
