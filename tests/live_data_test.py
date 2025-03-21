#!/usr/bin/python3
# -*- coding: utf-8 -*-

import json
import time
import random
import psycopg2
import sys


conn = None

CREATE_TABLE_SQL = """

    CREATE TABLE IF NOT EXISTS test_geojson (
        gid SERIAL,
        username VARCHAR NOT NULL,
        create_at TIMESTAMP DEFAULT NOW(),
        update_at TIMESTAMP DEFAULT NOW(),
        properties JSONB
    );

    SELECT AddGeometryColumn('test_geojson','geom',4326,'POINT',2);

    CREATE INDEX test_geojson__geoidx
        ON test_geojson
        USING GIST (geom);

"""


def makeFeature():
    return {
        "type": "Feature",
        "geometry": {
            "type": "Point",
            "coordinates": [
                random.uniform(-180, 180),
                random.uniform(-90, 90)
            ]
        },
        "properties": {
            "timestamp": time.time()
        }
    }



geojson = json.dumps({
    "type": "FeatureCollection",
    "features": [
        makeFeature() for i in range(100)
    ]
})

INSERT_QUERY = """

    WITH data AS (SELECT '{0}'::json AS fc)

    INSERT INTO test_geojson (username, properties, geom) ((
        SELECT
            'stefan_path' AS username,
            feat->'properties' AS properties,
            ST_SetSRID(ST_GeomFromGeoJSON(feat->>'geometry'), 4326) AS geom
        FROM (
            SELECT json_array_elements(fc->'features') AS feat
            FROM data
        ) AS q
    ));


""".format(geojson)



try:
    conn = psycopg2.connect("host='localhost' dbname='geodev' user='geodev' password='dev'")
    cursor = conn.cursor()
    # cur.execute(CREATE_TABLE_SQL)
    cursor.execute(INSERT_QUERY)
    conn.commit()
except Exception as e:
    if conn:
        conn.rollback()
    print(e)
    sys.exit(1)

finally:
    if conn:
        conn.close()
