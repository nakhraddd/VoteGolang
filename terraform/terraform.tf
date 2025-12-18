terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = "europe-west4"
  zone    = "europe-west4-a"
}

# Variables
variable "project_id" {}
variable "ssh_public_key" {}
variable "instance_name" { default = "votegolang-vm" }

# Connect to existing IP if it exists
resource "google_compute_address" "static_ip" {
  name   = "votegolang-static-ip"
  region = "europe-west4"
}

# Firewall - Merged SSH and App ports
resource "google_compute_firewall" "default" {
  name    = "allow-app-traffic"
  network = "default"
  allow {
    protocol = "tcp"
    ports    = ["22", "80", "8080", "5601", "8081"]
  }
  source_ranges = ["0.0.0.0/0"]
}

resource "google_compute_instance" "app_server" {
  name         = var.instance_name
  machine_type = "e2-standard-4" # Required for ELK/Kafka
  zone         = "europe-west4-a"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
      size  = 50
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.static_ip.address
    }
  }

  metadata = {
    "ssh-keys" = "gcp-user:${var.ssh_public_key}"
  }

  metadata_startup_script = <<-EOF
    #!/bin/bash
    set -e
    # Wait for apt locks
    while fuser /var/lib/dpkg/lock >/dev/null 2>&1 ; do sleep 1; done
    
    # Install Docker
    if ! command -v docker &> /dev/null; then
      apt-get update
      apt-get install -y ca-certificates curl gnupg
      install -m 0755 -d /etc/apt/keyrings
      curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
      chmod a+r /etc/apt/keyrings/docker.gpg
      echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
      apt-get update
      apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
      systemctl enable docker
      systemctl start docker
    fi
    # Ensure gcp-user can use docker
    usermod -aG docker gcp-user
  EOF
}

output "instance_ip" {
  value = google_compute_address.static_ip.address
}