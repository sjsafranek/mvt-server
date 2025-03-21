
# http://overpass-api.de/api/interpreter?data=[out:json];way(id:1000,100000418,100000419,100001669);out tags;


COPY(
    SELECT
        DISTINCT
            sseg_id AS osm_way_id
        FROM "gbr_osm_segments_02-31-2019"
        GROUP BY sseg_id
) TO '/tmp/gbr_osm_way_ids.csv'
WITH CSV HEADER NULL AS '';



cp /tmp/gbr_osm_way_ids.csv .


rsync -avh --progress --partial stefan@tileserver.db4iot.com:gbr_osm_way_ids.csv .
