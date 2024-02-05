#!/bin/bash

if [ "$CLUSTER_TYPE" == "openshift" ]; then
  echo "openshift cluster"
  exit
fi

chroot_path="/host"

function clean_services() {
  ovs_service=$chroot_path/usr/lib/systemd/system/ovs-vswitchd.service

  if [ -f $ovs_service ]; then
    sed -i.bak '/hw-offload/d' $ovs_service
  fi
}

clean_services
# Reload host services
chroot $chroot_path /bin/bash -c systemctl daemon-reload >/dev/null 2>&1 || true

# Restart system services
chroot $chroot_path /bin/bash -c systemctl restart ovs-vswitchd.service >/dev/null 2>&1 || true
