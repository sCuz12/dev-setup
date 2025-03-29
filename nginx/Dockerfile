# Use an official Nginx base image
FROM nginx:alpine

# Copy your custom nginx.conf to the container
# Overwriting the default config (or placing it in conf.d)
COPY nginx.conf /etc/nginx/nginx.conf

# If you need static files from your codebase, you can copy or mount them at runtime.
# Example:
# COPY /path/to/your/app/public /usr/share/nginx/html
