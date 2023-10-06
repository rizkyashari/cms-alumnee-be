#!/bin/bash

# Define the path to the webpack.config.js file
config_path="node_modules/@strapi/admin/webpack.config.js"

# Define the string to search for and the replacement string
search_string='minimize: optimize,'
replace_string='minimize: false,'

# Backup the original file (optional but recommended)
cp $config_path $config_path.backup

# Use sed to perform the replacement
sed -i "s/$search_string/$replace_string/g" $config_path

echo "Config file updated successfully!"
