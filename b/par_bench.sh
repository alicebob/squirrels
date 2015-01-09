#!/bin/bash
set -e

redis-cli -p 7000 SET key1 value1
redis-cli -p 7000 SET key2 value2

echo squirrel start: $(date)
redis-cli -p 7000 -r 100000 GET key1 > /dev/null &
redis-cli -p 7000 -r 100000 GET key2 > /dev/null &
redis-cli -p 7000 -r 100000 GET key2 > /dev/null &
redis-cli -p 7000 -r 100000 GET key2 > /dev/null &
redis-cli -p 7000 -r 100000 GET key3 > /dev/null &
wait
echo squirrel done: $(date)

echo nutcracker start: $(date)
redis-cli -p 7001 -r 100000 GET key1 > /dev/null &
redis-cli -p 7001 -r 100000 GET key2 > /dev/null &
redis-cli -p 7001 -r 100000 GET key2 > /dev/null &
redis-cli -p 7001 -r 100000 GET key2 > /dev/null &
redis-cli -p 7001 -r 100000 GET key3 > /dev/null &
wait
echo nutcracker done: $(date)
