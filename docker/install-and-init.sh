#!/bin/bash

set -Eeuo pipefail

python3 -m venv /exabgp/venv
source /exabgp/venv/bin/activate

if [[ "${EXABGP_VERSION}" == "main" ]]; then
    echo "Installing exabgp from git"
    env EXABGP_VERSION="$(date +%Y.%m.%d)" pip3 install git+https://github.com/Exa-Networks/exabgp
else
    pip3 install "exabgp==${EXABGP_VERSION}"
fi

exec /init
