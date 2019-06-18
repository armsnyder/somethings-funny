#!/usr/bin/env bash
set -e
[[ -e .env ]] && source .env
cd terraform
terraform apply
