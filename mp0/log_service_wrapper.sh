#!/bin/bash

# log=home/jalen/go/src/ece428/logging_service.log
log=logging_service.log
# export PATH=/usr/local/bin:/usr/local/sbin:/bin:/sbin:/usr/bin:/usr/sbin
cd /go/src/ece428
exec &>$log
echo $(date +"%D %T")" Starting..."
( exec logly --server --config=config.json &>>$log ) &
exit 0