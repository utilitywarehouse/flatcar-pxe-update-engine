variable "interval_initial" {
  default = "7m"
}

variable "interval_periodic" {
  default = "45m"
}

variable "interval_fuzz" {
  default = "20m"
}

variable "release_version" {
  default = "0.1.0"
}

variable "version_url" {
  default = "https://stable.release.flatcar-linux.net/amd64-usr/current/version.txt"
}
