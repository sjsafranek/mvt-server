package geodatabase

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"

	"github.com/sjsafranek/ligneous"
)

var (
	logger = ligneous.NewLogger()
)

func NewGeoDatabase(DataSourceName string) (*GeoDatabase, error) {
	db, err := sql.Open("postgres", DataSourceName)
	if nil != err {
		return &GeoDatabase{}, err
	}

	geodb := GeoDatabase{DataSourceName: DataSourceName, db: db}
	geodb.Init()

	return &geodb, nil
}

type GeoDatabase struct {
	DataSourceName string
	db             *sql.DB
	Debug          bool
}

func (self *GeoDatabase) executeQuery(f func(*sql.DB, error) error) error {
	db, err := sql.Open("postgres", self.DataSourceName)
	if nil != err {
		return err
	}

	defer db.Close()
	return f(db, err)
}

func (self *GeoDatabase) executeQueryWithContext(ctx context.Context, f func(*sql.Conn, error) error) error {
	conn, err := self.db.Conn(ctx)
	if nil != err {
		return err
	}

	defer conn.Close()
	return f(conn, err)
}

func (self *GeoDatabase) QueryRow(query string, results ...interface{}) error {
	err := self.executeQuery(func(db *sql.DB, err error) error {
		if nil != err {
			return err
		}

		if self.Debug {
			logger.Debug(query)
		}

		row := db.QueryRow(query)
		return row.Scan(results...)
	})

	if nil != err {
		logger.Error(err)
		logger.Debug(query)
	}

	return err
}

func (self *GeoDatabase) QueryRowWithContext(ctx context.Context, query string, results ...interface{}) error {
	err := self.executeQueryWithContext(ctx, func(db *sql.Conn, err error) error {
		if nil != err {
			return err
		}

		if self.Debug {
			logger.Debug(query)
		}

		row := db.QueryRowContext(ctx, query)
		return row.Scan(results...)
	})

	if nil != err {
		logger.Error(err)
		logger.Debug(query)
	}

	return err
}

func (self *GeoDatabase) QueryRowJSON(query string) (string, error) {
	var result string
	err := self.executeQuery(func(db *sql.DB, err error) error {
		if nil != err {
			return err
		}

		if self.Debug {
			logger.Debug(query)
		}

		row := db.QueryRow(query)
		return row.Scan(&result)
	})

	if nil != err {
		logger.Error(err)
		logger.Debug(query)
	}

	return result, err
}

func (self *GeoDatabase) Init() error {
	logger.Debug("Setup database...")
	var result string
	err := self.executeQuery(func(db *sql.DB, err error) error {
		if nil != err {
			return err
		}
		query := `

CREATE TABLE IF NOT EXISTS layers (
    layer_name VARCHAR NOT NULL UNIQUE,
    layer_id VARCHAR NOT NULL UNIQUE DEFAULT md5(random()::text || now()::text)::uuid,
    srid INTEGER NOT NULL DEFAULT 4269,
    description VARCHAR,
	attribution VARCHAR,
	is_updatable BOOLEAN DEFAULT false,
	-- extent JSONB,
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

		if self.Debug {
			logger.Debug(query)
		}

		row := db.QueryRow(query)
		return row.Scan(&result)
	})

	return err
}

func (self *GeoDatabase) FetchLayers() (string, error) {
	// var result string
	var result sql.NullString

	err := self.QueryRow(`
	SELECT json_agg(c)
		FROM (
			SELECT
				*
			FROM layers
			WHERE
				is_deleted = false
		) c;
	`, &result)

	// return result, err
	if result.Valid {
		return result.String, err
	}
	return "", err
}

func (self *GeoDatabase) DeleteLayer(layer_name string) error {
	// normalize
	layer_name = strings.ToLower(layer_name)

	err := self.executeQuery(func(db *sql.DB, err error) error {
		if nil != err {
			return err
		}
		query := fmt.Sprintf(`
	UPDATE layers
		SET is_deleted='t'
	 	WHERE layer_name = '%v';
	`, layer_name)

		if self.Debug {
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

func (self *GeoDatabase) FetchLayer(layer_name string) (string, error) {
	var result string

	// normalize
	layer_name = strings.ToLower(layer_name)

	// -- array_to_json(ARRAY((SELECT column_name::text FROM information_schema.columns WHERE table_name ='%v'))) AS properties,

	err := self.QueryRow(fmt.Sprintf(`
	SELECT
		row_to_json(c)::jsonb || row_to_json(lyrs.*)::jsonb
	FROM (
		SELECT
			ST_AsGeoJSON(ST_Extent( ST_Transform( ST_SetSRID(lyr.geom, lyrs.srid), 4269) ))::json AS extent,
			count(*) AS features,
			array_to_json(ARRAY((
				SELECT ('{"column_id":"'||column_name::text||'","type":"'||udt_name::text||'"}')::jsonb FROM information_schema.columns WHERE table_name ='%v'
			))) AS columns,
			array_to_json(ARRAY((SELECT DISTINCT GeometryType(geom) FROM "%v"))) AS geometry_types
		FROM "%v" AS lyr
		INNER JOIN
			layers AS lyrs
				ON layer_name = '%v'
	) c
	JOIN
		layers AS lyrs
			ON layer_name = '%v';
	`, layer_name, layer_name, layer_name, layer_name, layer_name), &result)

	return result, err
}

func (self *GeoDatabase) getTileSQL(layer_name string, x, y, z uint32, srid int64, filter string) string {
	layer_name = strings.ToLower(layer_name)

	// https://blog.jawg.io/how-to-make-mvt-with-postgis/
	bbox := fmt.Sprintf("BBox(%v, %v, %v)", x, y, z)

	if "" != filter {
		filter = strings.Replace(filter, "WHERE ", "", -1)
		filter = strings.Replace(filter, "where ", "", -1)
		filter = fmt.Sprintf("AND %v", filter)
	}

	// if srid is not 3857 feature geom must be converted
	if 3857 != srid {

		return fmt.Sprintf(`
	SET work_mem = '2GB';

	WITH features AS (
		SELECT
			row_to_json(lyr)::jsonb - 'geom' AS properties,
			ST_Transform( ST_SetSRID(lyr.geom, %v), 3857) AS geom
		FROM
			"%v" AS lyr

		WHERE
				ST_Transform( ST_SetSRID(lyr.geom, %v), 3857) && %v
			AND
				ST_Intersects(
					ST_Transform( ST_SetSRID(lyr.geom, %v), 3857),
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
	) AS q;
	`, srid, layer_name, srid, bbox, srid, bbox, filter, bbox)

	}

	// no need to convert srid
	return fmt.Sprintf(`
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

func (self *GeoDatabase) FetchTile(layer_name string, x, y, z uint32, srid int64, filter string) ([]uint8, error) {
	var tile []uint8
	query := self.getTileSQL(layer_name, x, y, z, srid, filter)
	err := self.QueryRow(query, &tile)
	return tile, err
}

func (self *GeoDatabase) FetchTileWithContext(ctx context.Context, layer_name string, x, y, z uint32, srid int64, filter string) ([]uint8, error) {
	var tile []uint8
	query := self.getTileSQL(layer_name, x, y, z, srid, filter)
	err := self.QueryRowWithContext(ctx, query, &tile)
	return tile, err
}
