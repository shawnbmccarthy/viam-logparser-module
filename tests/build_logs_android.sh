#
# meant to run from ./storage
# /storage/logs directory should be created 
# TODO: validate structure
# 

# for each file time, create 1 hours worth of files (5 minute intervals)
# YYYY-MM-DD.HH.MM.SS
# 2024-02-01.13.00.00
for log_type in DEBUG INFO WARN ERROR FATAL; do
  let minutes=0
  for i in 1 2 3 4 5 6 7 8 9 10 11 12; do
    minutes=$(((5 * i) - 5))
    if [ ${minutes} -lt 10 ]; then
      minutes="0${minutes}"
    fi
    log_file_name="./logs/${log_type}.${i}.2024-02-01.13.${minutes}.00.log"
    echo "[2024020113${minutes}.00] ${log_type} ${i} Test 1" >> ${log_file_name}
    echo "[2024020113${minutes}.00] ${log_type} ${i} Test 2" >> ${log_file_name}
    touch -t "24020113${minutes}" ${log_file_name}
  done
done
