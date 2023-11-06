#!/bin/bash

# assumes run through make
test -d ./tests/logs && echo "logs directory exists" || exit 1
test -d ./tests/upload && echo "upload directory exists" || exit 2

# for each file time, create 1 hours worth of files (5 minute intervals)
# YYYY-MM-DD.HH.MM.SS
# 2023-10-01.13.00.00
for log_type in DEBUG INFO WARN ERROR FATAL; do
  let minutes=0
  for i in {1..12}; do
    minutes=$(((5 * i) - 5))
    if [ ${minutes} -lt 10 ]; then
      minutes="0${minutes}"
    fi
    log_file_name="./tests/logs/${log_type}.${i}.2023-10-01.13.${minutes}.00.log"
    echo "[2023100113${minutes}.00] ${log_type} ${i} Test 1" >> ${log_file_name}
    echo "[2023100113${minutes}.00] ${log_type} ${i} Test 2" >> ${log_file_name}
    touch -t "23100113${minutes}" ${log_file_name}
  done
done