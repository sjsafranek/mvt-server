
-- Enable PostGIS (includes raster)
CREATE EXTENSION postgis;
-- Enable Topology
CREATE EXTENSION postgis_topology;
-- Enable PostGIS Advanced 3D
-- and other geoprocessing algorithms
-- sfcgal not available with all distributions
CREATE EXTENSION postgis_sfcgal;
-- fuzzy matching needed for Tiger
CREATE EXTENSION fuzzystrmatch;
-- rule based standardizer
CREATE EXTENSION address_standardizer;
-- example rule data set
CREATE EXTENSION address_standardizer_data_us;
-- Enable US Tiger Geocoder
CREATE EXTENSION postgis_tiger_geocoder;

-- Upgrade PostGIS (includes raster) to latest version
ALTER EXTENSION postgis UPDATE;
ALTER EXTENSION postgis_topology UPDATE;

-- Enable pgcrypto for passwords
CREATE EXTENSION pgcrypto;


CREATE TABLE IF NOT EXISTS layers (
    layer_name VARCHAR NOT NULL UNIQUE,
    layer_id VARCHAR NOT NULL UNIQUE DEFAULT md5(random()::text || now()::text)::uuid,
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
