#!/bin/bash

if [ "$CLUSTER_TYPE" == "openshift" ]; then
  echo "openshift cluster"
  exit
fi

chroot_path="/host"

function clean_services() {
  # Remove switchdev service files
  rm -f $chroot_path/etc/udev/switchdev-vf-link-name.sh

  # clean ovs-vswitchd services
  ovs_service=$chroot_path/usr/lib/systemd/system/ovs-vswitchd.service

  if [ -f $ovs_service ]; then
    sed -i.bak '/hw-offload/d' $ovs_service
  fi
}

clean_services
