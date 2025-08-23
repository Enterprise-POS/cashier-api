#!/bin/bash

# setup.bashrc only containing local environment variable
# example when creating setup.bashrc, 
# refer to .env.example to see what env should be prepare before run application
# 
# #!/bin/bash
# export MODE=dev
if [ -f "setup.bashrc" ]; then
  source ./setup.bashrc
else
  echo "[ERROR] setup.bashrc file does not exist. Please create setup.bashrc"
  exit
fi


if [ -f "main.go" ]; then
  go run main.go
else
  echo "[ERROR] main.go file does not exist."
fi
