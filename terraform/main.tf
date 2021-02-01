data "ignition_file" "merit_update_engine" {
  mode       = 493
  filesystem = "root"
  path       = "/opt/bin/merit-update-engine"

  source {
    source = "https://github.com/utilitywarehouse/flatcar-merit-update-engine/releases/download/${var.release_version}/flatcar-merit-update-engine_${var.release_version}_linux_amd64"
  }
}

data "ignition_systemd_unit" "merit_update_engine" {
  name = "update-engine.service"
  content = templatefile("${path.module}/resources/update-engine.service",
    {
      version_url = var.version_url
    }
  )
}


data "ignition_systemd_unit" "locksmithd" {
  name    = "locksmithd.service"
  content = file("${path.module}/resources/locksmithd.service")
}
