#!/bin/sh

HOST_CMD="nsenter --mount=/proc/1/ns/mnt"
BINARY_NAME="open-object"
CONFIG_DIR=/host/etc/$BINARY_NAME

if [ ! `$HOST_CMD which s3fs` ]; then
    echo "s3fs not found, plz install it first..."
    exit 1
fi

rm -f $CONFIG_DIR/connector.pid
rm -f $CONFIG_DIR/connector.sock

cp -f /fuse-connector.service /host/usr/lib/systemd/system/fuse-connector.service
chmod 755 /host/usr/lib/systemd/system/fuse-connector.service

cp -f /fuse-connector.conf $CONFIG_DIR/fuse-connector.conf
chmod 755 $CONFIG_DIR/fuse-connector.conf

cp -f /bin/$BINARY_NAME $CONFIG_DIR/$BINARY_NAME

$HOST_CMD systemctl daemon-reload
$HOST_CMD systemctl enable fuse-connector.service
$HOST_CMD systemctl restart fuse-connector.service