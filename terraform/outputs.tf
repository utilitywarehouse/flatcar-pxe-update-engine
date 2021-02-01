output "file" {
  value = data.ignition_file.merit_update_engine.rendered
}

output "unit" {
  value = data.ignition_systemd_unit.merit_update_engine.rendered
}

output "locksmithd_unit" {
  value = data.ignition_systemd_unit.locksmithd.rendered
}
