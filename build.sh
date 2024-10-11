#!/bin/bash

while getopts ":u:p:h:" opt; do
  case $opt in
    u) U=$OPTARG ;;
    p) P=$OPTARG ;;
    h) H=$OPTARG ;;
    \?) echo "$(date "+%Y-%m-%d %H:%M:%S") [Error] invalid option"; exit 1;;
    :) echo "$(date "+%Y-%m-%d %H:%M:%S") [Error] option requires an argument."; exit 1;;
  esac
done

echo "$(date "+%Y-%m-%d %H:%M:%S") [Info] Coping to remote host: $H..."

expect -c "
  log_user 0
  set username \"$U\"
  set password \"$P\"
  set hostname \"$H\"

  spawn rsync -rlptvz --exclude-from=.gitignore . \$username@\$hostname:~/build
  expect \"password:\"
  send -- \"\$password\r\"
  expect \"total\"
  expect eof
"

echo "$(date "+%Y-%m-%d %H:%M:%S") [Info] Building on remote host: $H..."

expect -c "
  log_user 0
  set username \"$U\"
  set password \"$P\"
  set hostname \"$H\"

  spawn ssh \$username@\$hostname
  expect \"password:\"
  send -- \"\$password\r\"
  expect \":~\"
  send -- \"cd ./build\r\"
  send -- \"make\r\"
  expect \":~\"
  send -- \"exit\r\"
  expect eof
"
echo "$(date "+%Y-%m-%d %H:%M:%S") [Info] Downloading from remote host: $H..."

expect -c "
  log_user 0
  set username \"$U\"
  set password \"$P\"
  set hostname \"$H\"

  spawn scp -r \$username@\$hostname:~/build/eagleeye ./eagleeye
  expect \"password:\"
  send -- \"\$password\r\"
  expect \"100%\"
  expect eof
"

echo "$(date "+%Y-%m-%d %H:%M:%S") [Info] Done!"