#!/bin/bash
set -xe

sql=$(shp2pgsql -I "~/Repos/mvt-server/data/inrix_shapefiles/copenhagen/Oresundsanalys_draft_180807.shp" "test")

# if [ -z "$sql" ] then
#     echo "failed to import file"
#     exit 1
# fi;

echo $sql | PGPASSWORD=dev psql -d geodev -U geodev -h localhost -p 5432

PGPASSWORD=dev psql -d geodev -U geodev -h localhost -p 5432 -c "
        INSERT INTO layers (layer_name, description, srid) VALUES ('test', 'roads of denmark inrix', 4326)
    "



# {"method":"upload","file_path": "/home/stefan/Repos/mvt-server/data/inrix_shapefiles/copenhagen/Oresundsanalys_draft_180807.shp", "table_name":"test", "description":"roads of denmark inrix", "srid": 4326}
