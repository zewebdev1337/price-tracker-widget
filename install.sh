#!/bin/bash

# This script is used to compile the widget and move the resulting binary 
# to /usr/local/bin/, install and start a systemd service for it.

go build

# Check if the binary was successfully built
if [ -f ./price-tracker-widget ]; then
  echo "Build complete."

  mv price-tracker-widget /usr/local/bin/

  if [ ! -d ~/.config/systemd/user ]; then
    mkdir -p ~/.config/systemd/user
  fi

  # This makes the systemd service available for the user
  cp prices.service ~/.config/systemd/user/
  
  # Enable and start the service for the user
  # The --now flag immediately starts the service after enabling it
  systemctl enable --now prices.service --user
else
  # If the 'price-tracker-widget' binary file was not successfully built, output an error message
  echo "Build process failed!"
fi
