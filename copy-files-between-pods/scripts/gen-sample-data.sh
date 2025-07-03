#!/usr/bin/env bash
set -euo pipefail

# Usage: ./generate_sample_data.sh [TARGET_DIR]
# Default target directory: ./src_data

TARGET_DIR="${1:-./src_data}"

# Clean up any existing data
echo "Creating sample source directory at: $TARGET_DIR"
rm -rf "$TARGET_DIR"

# Create directory structure
mkdir -p \
  "$TARGET_DIR"/dir1/sub1 \
  "$TARGET_DIR"/dir1/sub2 \
  "$TARGET_DIR"/dir2/sub1 \
  "$TARGET_DIR"/dir2/sub2 \
  "$TARGET_DIR"/dir3

# Create files in root
cat <<EOF > "$TARGET_DIR"/root_file1.txt
This is root_file1.txt
Sample data line 1
Sample data line 2
EOF

cat <<EOF > "$TARGET_DIR"/root_file2.csv
id,name,quantity
1,Apple,10
2,Banana,20
EOF

# Create files in dir1
cat <<EOF > "$TARGET_DIR"/dir1/file1.log
[INFO] dir1/file1.log generated at $(date -u)
[DEBUG] Sample debug message
EOF

cat <<EOF > "$TARGET_DIR"/dir1/sub1/deep_file1.json
{"id": 100, "status": "ok", "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"}
EOF

# Create files in dir2
echo -e "line A\nline B\nline C" > "$TARGET_DIR"/dir2/sub2/text_data.txt

# Create an empty placeholder file
touch "$TARGET_DIR"/dir3/placeholder.txt

# Print summary
echo "Sample directory structure created:"
find "$TARGET_DIR" -type f | sed 's/^/  - /'
