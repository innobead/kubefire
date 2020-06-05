#!/usr/bin/env sh

# Force-remove all running VMs
ignite rm -f $(ignite ps -aq)
# Remove the data directory
rm -r /var/lib/firecracker
# Remove the ignite and ignited binaries
rm /usr/local/bin/ignite{,d}