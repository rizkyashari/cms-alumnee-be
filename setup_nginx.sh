#!/bin/bash

# Update package lists
sudo apt update

# Install nginx
sudo apt install -y nginx

# Allow necessary ports in the firewall (if applicable)
sudo ufw allow 80
sudo ufw allow 81
sudo ufw allow 83

# Allow SSH traffic (port 22) in the firewall
sudo ufw allow ssh

# Create a copy nginx.conf file
sudo cp ./nginx.conf /etc/nginx/nginx.conf

# Reload Nginx to apply changes
sudo systemctl reload nginx

# Display status
echo "Nginx has been set up and is now listening on ports 8080, 8083, and 1337."
echo "SSH traffic on port 22 is allowed in the firewall."
