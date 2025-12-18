terraform {
  backend "gcs" {
    bucket = "vote_golang_bucket"
    prefix = "terraform/state"
  }
}