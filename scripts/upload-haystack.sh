#!/bin/bash

base_dir=$(cd "$(dirname "$0")" && pwd)
dist_dir="$base_dir/../src/dist"
remote_host="haystack"
remote_dir="/opt/dockers/nginx/webroot/haystack/"

if [ ! -d "$dist_dir" ]; then
    echo "Error: dist directory not found"
    exit 1
fi

scp $dist_dir/haystack-linux-* "$remote_host:$remote_dir/linux"
scp $dist_dir/haystack-darwin-* "$remote_host:$remote_dir/darwin"
scp $dist_dir/haystack-windows-* "$remote_host:$remote_dir/windows"
scp $base_dir/src/VERSION "$remote_host:$remote_dir/VERSION"

echo "Upload completed"
