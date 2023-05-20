#!/bin/bash
cd /gobgp || exit 1
# sleep so that gobgp isn't ready yet
sleep 10
exec ./gobgpd -f gobgp.yaml
