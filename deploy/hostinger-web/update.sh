#!/bin/bash
# Update SIPKEU dari GitHub + restart
set -euo pipefail
APP_DIR="${HOME}/sipkeu"
bash "$APP_DIR/deploy/hostinger-web/install.sh"
