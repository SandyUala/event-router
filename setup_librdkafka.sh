#!/bin/bash

echo "Hello!  This script will setup librdkafka so you can actually build this program!"
echo "If you don't actually have g++ or make installed, this script will break!"
echo "It's up to you to setup g++/make/python/whatever to build this!"
echo "Also it requires sudo!"
echo ""

read -n1 -r -p "Press any key to continue..." key

if [[ ! -d librdkafka-0.11.0 ]]; then
	if [[ ! -f v0.11.0.tar.gz ]]; then
		wget https://github.com/edenhill/librdkafka/archive/v0.11.0.tar.gz
	fi
	tar -xf v0.11.0.tar.gz
fi
cd librdkafka-0.11.0
./configure
make && sudo make install
cd ..
rm -rf librdkafka-0.11.0
rm v0.11.0.tar.gz