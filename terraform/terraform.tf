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


# Resource names must match the 'to' field in the import blocks above
resource "google_compute_address" "static_ip" {
  name   = "votegolang-static-ip"
  region = "europe-west4"
}

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
  machine_type = "e2-standard-4"
  zone         = "europe-west4-a"

  allow_stopping_for_update = true

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

    # 1. More aggressive wait for apt locks
    echo "Waiting for system updates to finish..."
    while fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1 ; do sleep 5; done

    # 2. Install Docker if missing
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

    # 3. CRITICAL: Add user to group AND ensure it takes effect
    usermod -aG docker gcp-user
    # Force the group change for the current session (though SSH usually handles this on next login)
    newgrp docker || true
  EOF
}