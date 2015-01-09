#!/bin/bash
set -e

redis-cli -p 7000 SET key1 value1
redis-cli -p 7000 SET key2 value2
echo squirrels:
time redis-cli -p 7000 -r 100000 GET key2 > /dev/null
echo nutcracker:
time redis-cli -p 7001 -r 100000 GET key2 > /dev/null
