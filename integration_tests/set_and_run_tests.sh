#!/bin/bash

SETUP_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd "$SETUP_DIR/set_up"
chmod +x ./start_node.sh
echo "Starting the node..."
source ./start_node.sh

cd "$SETUP_DIR/set_up"
chmod +x ./start_bdjuno.sh
echo "Starting BDJuno..."
source ./start_bdjuno.sh 
