import {
  id = "projects/${var.project_id}/global/firewalls/allow-http-ssh-kafka"
  to = google_compute_firewall.default
}

import {
  id = "projects/${var.project_id}/zones/europe-west4-a/instances/votegolang-vm"
  to = google_compute_instance.app_server
}

import {
  id = "projects/${var.project_id}/regions/europe-west4/addresses/votegolang-static-ip"
  to = google_compute_address.static_ip
}

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

variable "project_id" {
  description = "The GCP Project ID"
}

# 1. Use a Data Source for the Network
# This ensures we don't try to "re-create" the default network if it exists.
data "google_compute_network" "default" {
  name = "default"
}

# 2. Static IP with a lifecycle protection
resource "google_compute_address" "static_ip" {
  name   = "votegolang-static-ip"
  region = "europe-west4"

  # If you manually delete the VM, this prevents the IP from being nuked
  lifecycle {
    prevent_destroy = false
  }
}

# 3. Firewall - Allow required ports
resource "google_compute_firewall" "default" {
  name    = "allow-http-ssh-kafka"
  network = data.google_compute_network.default.name

  allow {
    protocol = "tcp"
    ports    = ["22", "80", "8080", "5601", "8081", "9092", "9200"]
  }

  source_ranges = ["0.0.0.0/0"]
}

# 4. The VM Instance
resource "google_compute_instance" "app_server" {
  name         = "votegolang-vm"
  machine_type = "e2-standard-4"
  zone         = "europe-west4-a"

  # This allows updating the VM (like changing machine type) without deleting it
  allow_stopping_for_update = true

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
      size  = 50
    }
  }

  network_interface {
    network = data.google_compute_network.default.name
    access_config {
      nat_ip = google_compute_address.static_ip.address
    }
  }

  metadata_startup_script = <<-EOT
    #!/bin/bash
    if ! command -v docker &> /dev/null; then
      sudo apt-get update
      sudo apt-get install -y ca-certificates curl gnupg
      sudo install -m 0755 -d /etc/apt/keyrings
      curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
      sudo chmod a+r /etc/apt/keyrings/docker.gpg
      echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
      sudo apt-get update
      sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
      sudo systemctl enable docker
      sudo systemctl start docker
    fi
  EOT
}

output "instance_ip" {
  value = google_compute_address.static_ip.address
}
