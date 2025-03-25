#!/bin/bash

# wait --v2-upgrade-height value
celestia-appd tx signal signal 3 --from alice --fees 400utia --yes
sleep 6
celestia-appd tx signal try-upgrade --from alice --fees 400utia --yes
sleep 20
# wait 3 blocks
celestia-appd tx signal signal 4 --from alice --fees 400utia --yes
sleep 6
celestia-appd tx signal try-upgrade --from alice --fees 400utia --yes