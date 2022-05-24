variable "release_version" {
  default = "v1.0.0"
}

variable "version_url" {
  default = "https://stable.release.flatcar-linux.net/amd64-usr/current/version.txt"
}

locals {
  vless_release_version = trimprefix(var.release_version, "v")
}
