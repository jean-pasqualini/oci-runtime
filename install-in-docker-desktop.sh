#!/bin/bash

make build
mv /host/bin/runc /host/bin/runc.old
cp /tmp/oci-runtime /host/bin/runc
ls -al /tmp/oci-runtime

read -r -p "Did it work? (yes/no): " answer
if [[ "$answer" == "yes" || "$answer" == "y" ]]; then
    echo "Nice, it worked!"
    exit 0
else
    echo "Alright, handling failure case..."
    cp /host/bin/runc.old /host/bin/runc
    exit 0
fi