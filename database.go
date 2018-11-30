package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// https://c.tile.openstreetmap.org/8/75/95.png
// https://raw.githubusercontent.com/jawg/blog-resources/master/how-to-make-mvt-with-postgis/bbox.sql
var bbox_psql_function = `
    CREATE OR REPLACE FUNCTION BBox(x integer, y integer, zoom integer)
        RETURNS geometry AS
    $BODY$
    DECLARE
        max numeric := 6378137 * pi();
        res numeric := max * 2 / 2^zoom;
        bbox geometry;
    BEGIN
        return ST_MakeEnvelope(
            -max + (x * res),
            max - (y * res),
            -max + (x * res) + res,
            max - (y * res) - res,
            3857);
    END;
    $BODY$
      LANGUAGE plpgsql IMMUTABLE;
`

func fetchTile(layer_name string, x, y, z int) ([]uint8, error) {

	var tile []uint8

	db, err := sql.Open("postgres", "postgres://stefan:geolRocks@localhost/stefan")
	if nil != err {
		return tile, err
	}
	defer db.Close()

	bbox := fmt.Sprintf("BBox(%v, %v, %v)", x, y, z)

	// https://blog.jawg.io/how-to-make-mvt-with-postgis/
	row := db.QueryRow(fmt.Sprintf(`
        WITH features AS (
            SELECT
                lyr.gid,
                lyr.statefp10,
                lyr.pumace10,
                lyr.geoid10,
                lyr.namelsad10,
                lyr.mtfcc10,
                ST_Transform( ST_SetSRID(lyr.geom, 4269), 3857) AS geom
            FROM
                %v AS lyr
        )

        SELECT
            ST_AsMVT(q, 'layer', 4096, 'geom')
        FROM (
            SELECT
                fts.gid,
                fts.statefp10,
                fts.pumace10,
                fts.geoid10,
                fts.namelsad10,
                fts.mtfcc10,
                ST_AsMvtGeom(
                    fts.geom,
                    %v,
                    4096,
                    256,
                    true
                ) AS geom
            FROM
                features AS fts
            WHERE
                    fts.geom && %v
                AND
                    ST_Intersects(
                        fts.geom,
                        %v
                    )
        ) AS q;
    `, layer_name, bbox, bbox, bbox))

	err = row.Scan(&tile)
	if nil != err {
		return tile, err
	}

	return tile, nil
}
