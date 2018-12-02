#!/bin/bash

set -x

SHAPE_FILE="$1"
DB_TABLE="$2"

# TODO: DB_TABLE from SHAPE_FILE

shp2pgsql -I "$SHAPE_FILE" "$DB_TABLE" | PGPASSWORD=dev psql -d geodev -U geodev
