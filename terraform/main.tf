data "ignition_file" "pxe_update_engine" {
  mode       = 493
  filesystem = "root"
  path       = "/opt/bin/pxe-update-engine"

  source {
    source = "https://github.com/utilitywarehouse/flatcar-pxe-update-engine/releases/download/${var.release_version}/flatcar-pxe-update-engine_${local.vless_release_version}_linux_amd64"
  }
}

data "ignition_systemd_unit" "pxe_update_engine" {
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
