import {
  id = "projects/${var.project_id}/global/firewalls/allow-app-traffic"
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