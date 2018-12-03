package main

import (
	"database/sql"
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
		row := db.QueryRow(query)
		return row.Scan(&result)
	})

	if nil != err {
		logger.Error(err)
	}

	return result, err
}

func fetchLayerFromDatabase(layer_name string) (string, error) {
	var result string

	layer_name = strings.ToLower(layer_name)

	err := executeDatabaseQuery(func(db *sql.DB, err error) error {
		if nil != err {
			return err
		}
		query := fmt.Sprintf(`
			SELECT
				row_to_json(c)
			FROM (
				SELECT
					ST_AsGeoJSON(ST_Extent(geom))::json AS extent,
					count(*) AS features
				FROM "%v"
		 	) c;
		`, layer_name)
		row := db.QueryRow(query)
		return row.Scan(&result)
	})

	if nil != err {
		logger.Error(err)
	}

	return result, err
}

func getFrcStylesheetHACK(z uint32) string {
	switch {
	case z < 7:
		return `
			WHERE
				frc IN ('0')
		`
	case z < 8:
		return `
			WHERE
				frc IN ('0', '1')
		`
	case z < 10:
		return `
			WHERE
				frc IN ('0', '1', '2')
		`
	case z < 12:
		return `
			WHERE
				frc IN ('0', '1', '2', '3')
		`
	default:
		return ""
	}
}

func fetchTileFromDatabase(layer_name string, x, y, z uint32) ([]uint8, error) {
	var tile []uint8

	layer_name = strings.ToLower(layer_name)

	err := executeDatabaseQuery(func(db *sql.DB, err error) error {
		// https://blog.jawg.io/how-to-make-mvt-with-postgis/
		bbox := fmt.Sprintf("BBox(%v, %v, %v)", x, y, z)

		query := fmt.Sprintf(`
			WITH features AS (
				SELECT
					row_to_json(lyr)::jsonb - 'geom' AS properties,
					ST_Transform( ST_SetSRID(lyr.geom, 4269), 3857) AS geom
				FROM
					"%v" AS lyr

				-- TODO: stylesheet?
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
		`, layer_name, getFrcStylesheetHACK(z), bbox, bbox, bbox)

		row := db.QueryRow(query)
		return row.Scan(&tile)
	})

	if nil != err {
		logger.Error(err)
	}

	return tile, nil
}
