#!/bin/bash
# this is an `standalone` mode exporter that should be able to successfully call exabgpcli listening on a different port
exec /exabgp/exabgp_exporter --web.listen-address=":9570" --log.format="json" standalone --exabgp.cli.command="/exabgp/venv/bin/exabgpcli" --exabgp.root="/exabgp/"
