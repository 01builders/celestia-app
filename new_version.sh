#!/bin/bash

# new multiplexer binary
make install
cp $(which celestia-appd) ~/downloads/
mv ~/downloads/celestia-appd ~/downloads/celestia-app-v4_Linux_arm64
tar -czvf ~/downloads/celestia-app-v4_Linux_arm64.tar.gz ~/downloads/celestia-app-v4_Linux_arm64
mv ~/downloads/celestia-app-v4_Linux_arm64.tar.gz ~/projects/01builders/celestia-app/internal/embedding/
make install-multiplexer
MULTIPLEXER=true ./local_devnet/scripts/init.sh
celestia-appd start --force-no-bbr

# wait --v2-upgrade-height value
# celestia-appd tx signal signal 3 --from alice --fees 400utia --yes
# celestia-appd tx signal try-upgrade --from alice --fees 400utia --yes
# wait 3 blocks
# celestia-appd tx signal signal 4 --from alice --fees 400utia --yes
# celestia-appd tx signal try-upgrade --from alice --fees 400utia --yes