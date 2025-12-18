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
  region  = "europe-west4" # Netherlands
  zone    = "europe-west4-a"
}

variable "project_id" {
  description = "The GCP Project ID"
}

# Reserve a static external IP address
resource "google_compute_address" "static_ip" {
  name   = "votegolang-static-ip"
  region = "europe-west4"
}

# Firewall rule to allow HTTP, Nginx, and SSH
resource "google_compute_firewall" "default" {
  name    = "allow-http-ssh-kafka"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["22", "80", "8080", "5601", "8081"]
  }

  source_ranges = ["0.0.0.0/0"]
}

# The VM Instance
resource "google_compute_instance" "app_server" {
  name         = "votegolang-vm"
  machine_type = "e2-standard-4" # ELK & Kafka require significant RAM (min 8GB-16GB recommended)
  zone         = "europe-west4-a"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-12"
      size  = 50 # GB
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.static_ip.address
    }
  }

  metadata_startup_script = <<-EOT
    #!/bin/bash
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
  EOT
}

output "instance_ip" {
  value = google_compute_address.static_ip.address
}