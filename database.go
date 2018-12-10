package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

func executeDatabaseQuery(f func(*sql.DB, error) error) error {
	db, err := sql.Open("postgres", config.Database.ConnectionString())
	if nil != err {
		return err
	}
	defer db.Close()
	return f(db, err)
}

func databaseSetup() error {
	logger.Debug("Setup database...")
	var result string
	err := executeDatabaseQuery(func(db *sql.DB, err error) error {
		if nil != err {
			return err
		}
		query := `

CREATE TABLE IF NOT EXISTS layers (
    layer_name VARCHAR NOT NULL UNIQUE,
    layer_id VARCHAR NOT NULL UNIQUE DEFAULT md5(random()::text || now()::text)::uuid,
    srid INTEGER NOT NULL DEFAULT 4269,
    description VARCHAR,
    created_at TIMESTAMP DEFAULT current_timestamp,
    updated_at TIMESTAMP DEFAULT current_timestamp,
    is_deleted BOOLEAN DEFAULT false,
    PRIMARY KEY(layer_id)
);

-- update triggers
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS layers_update ON layers;
CREATE TRIGGER layers_update BEFORE UPDATE ON layers FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
-- .end


-- https://raw.githubusercontent.com/jawg/blog-resources/master/how-to-make-mvt-with-postgis/bbox.sql
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

		if DEBUG {
			logger.Debug(query)
		}

		row := db.QueryRow(query)
		return row.Scan(&result)
	})

	return err
}

func fetchLayersFromDatabase() (string, error) {
	var result string
	err := executeDatabaseQuery(func(db *sql.DB, err error) error {
		if nil != err {
			return err
		}
		query := `
			SELECT json_agg(c)
		        FROM (
		            SELECT
		                *
		            FROM layers
		            WHERE
		                is_deleted = false
		        ) c;
		`

		if DEBUG {
			logger.Debug(query)
		}

		row := db.QueryRow(query)
		return row.Scan(&result)
	})

	if nil != err {
		logger.Error(err)
	}

	return result, err
}

func deleteLayerFromDatabase(layer_name string) error {
	// normalize
	layer_name = strings.ToLower(layer_name)

	err := executeDatabaseQuery(func(db *sql.DB, err error) error {
		if nil != err {
			return err
		}
		query := fmt.Sprintf(`
			UPDATE layers
				SET is_deleted='t'
			 	WHERE layer_name = '%v';
		`, layer_name)

		if DEBUG {
			logger.Debug(query)
		}

		_, err = db.Exec(query)
		return err
	})

	if nil != err {
		logger.Error(err)
	}

	return err
}

func fetchLayerFromDatabase(layer_name string) (string, error) {
	var result string

	// normalize
	layer_name = strings.ToLower(layer_name)

	err := executeDatabaseQuery(func(db *sql.DB, err error) error {
		if nil != err {
			return err
		}

		query := fmt.Sprintf(`
			SELECT
				row_to_json(c)::jsonb || row_to_json(lyrs.*)::jsonb
			FROM (
				SELECT
			        ST_AsGeoJSON(ST_Extent( ST_Transform( ST_SetSRID(lyr.geom, lyrs.srid), 4269) ))::json AS extent,
					-- ST_AsGeoJSON(ST_Extent(geom))::json AS extent,
					count(*) AS features,
					array_to_json(ARRAY((SELECT column_name::text FROM information_schema.columns WHERE table_name ='%v'))) as properties
				FROM "%v" AS lyr
				INNER JOIN
					layers AS lyrs
						ON layer_name = '%v'
		 	) c
			JOIN
				layers AS lyrs
					ON layer_name = '%v';
		`, layer_name, layer_name, layer_name, layer_name)

		if DEBUG {
			logger.Debug(query)
		}

		row := db.QueryRow(query)
		return row.Scan(&result)
	})

	if nil != err {
		logger.Error(err)
	}

	return result, err
}

func fetchTileFromDatabase(layer_name string, x, y, z uint32, filter string) ([]uint8, error) {
	var tile []uint8

	layer, err := LAYERS.GetLayer(layer_name)
	if nil != err {
		return tile, errors.New("Layer not found")
	}
	srid := layer.SRID

	layer_name = strings.ToLower(layer_name)

	err = executeDatabaseQuery(func(db *sql.DB, err error) error {
		// https://blog.jawg.io/how-to-make-mvt-with-postgis/
		bbox := fmt.Sprintf("BBox(%v, %v, %v)", x, y, z)

		query := ""

		if "" != filter {
			filter = strings.Replace(filter, "WHERE ", "", -1)
			filter = strings.Replace(filter, "where ", "", -1)
			filter = fmt.Sprintf("AND %v", filter)
		}

		// if srid is not 3857 feature geom must be converted
		if 3857 != srid {

			query = fmt.Sprintf(`
			SET work_mem = '2GB';

			WITH features AS (
				SELECT
					row_to_json(lyr)::jsonb - 'geom' AS properties,
					-- ST_Transform( ST_SetSRID(lyr.geom, 4269), 3857) AS geom
					ST_Transform( ST_SetSRID(lyr.geom, %v), 3857) AS geom
				FROM
					"%v" AS lyr
				-- client side filter... stylesheet?
				-- TODO filter with bbox

				WHERE
						fts.geom && %v
					AND
						ST_Intersects(
							fts.geom,
							%v
						)

				%v
			)

			SELECT
				ST_AsMVT(q, 'layer', 4096, 'geom')
			FROM (
				SELECT
					fts.properties,
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
		`, srid, layer_name, bbox, bbox, filter, bbox, bbox, bbox)

		}

		//
		if 3857 == srid {

			// if "" != filter {
			// 	filter = strings.Replace(filter, "WHERE ", "", -1)
			// 	filter = strings.Replace(filter, "where ", "", -1)
			// 	filter = fmt.Sprintf("AND %v", filter)
			// }

			query = fmt.Sprintf(`
			SET work_mem = '2GB';

			SELECT
				ST_AsMVT(q, 'layer', 4096, 'geom')
			FROM (
				SELECT
					row_to_json(fts)::jsonb - 'geom' AS properties,
					ST_AsMvtGeom(
						fts.geom,
						%v,
						4096,
						256,
						true
					) AS geom
				FROM
					"%v" AS fts
				WHERE
						fts.geom && %v
					AND
						ST_Intersects(
							fts.geom,
							%v
						)
					%v
			) AS q;
		`, bbox, layer_name, bbox, bbox, filter)

		}

		if DEBUG {
			logger.Debug(query)
		}

		row := db.QueryRow(query)
		return row.Scan(&tile)
	})

	if nil != err {
		logger.Error(err)
	}

	return tile, err
}
