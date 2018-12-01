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
        tl_2017_us_puma AS lyr
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
            BBox(14, 24, 6),
            4096,
            256,
            true
        ) AS geom
    FROM
        features AS fts
    WHERE
            fts.geom && BBox(14, 24, 6)
        AND
            ST_Intersects(
                fts.geom,
                BBox(14, 24, 6)
            )
) AS q;











-- https://dba.stackexchange.com/questions/1957/sql-select-all-columns-except-some
SELECT 'SELECT ' || array_to_string(ARRAY(SELECT 'o' || '.' || c.column_name
        FROM information_schema.columns As c
            WHERE table_name = 'tl_2017_us_county'
            AND  c.column_name NOT IN('geom')
    ), ',') || ' FROM tl_2017_us_county As o' As sqlstmt






WITH features AS (
            SELECT
                lyr.*,
                ST_Transform( ST_SetSRID(lyr.geom, 4269), 3857) AS geom
            FROM
                tl_2017_us_county AS lyr
        )

        -- SELECT
            -- ST_AsMVT(q, 'layer', 4096, 'geom')
        -- FROM (
            SELECT
                fts.*,
                ST_AsMvtGeom(
                    fts.geom,
                    BBox(13, 24, 6),
                    4096,
                    256,
                    true
                ) AS geom
            FROM
                features AS fts
            WHERE
                    fts.geom && BBox(13, 24, 6)
                AND
                    ST_Intersects(
                        fts.geom,
                        BBox(13, 24, 6)
                    )
        -- ) AS q;








CREATE TABLE usa_county AS (
    SELECT
        *
    FROM tl_2017_us_county
);


ST_Transform( ST_SetSRID(lyr.geom, 4269), 3857) AS geom
















SELECT
    array_to_string(ARRAY(
        SELECT
            columns.column_name::TEXT
        FROM
            information_schema.columns AS columns
        WHERE
                table_name = 'usa_county'
            AND
                columns.column_name NOT IN('geom')
    ), ',')
FROM usa_county AS uc;







SELECT
    array_to_string(ARRAY(
        SELECT
            uc||'.'||columns.column_name::TEXT
        FROM
            information_schema.columns AS columns
        WHERE
                table_name = 'usa_county'
            AND
                columns.column_name NOT IN('geom')
    ), ',')
FROM usa_county AS uc;










WITH columns AS (
    SELECT
        columns.column_name::TEXT
    FROM
        information_schema.columns AS columns
    WHERE
            table_name = 'usa_county'
        AND
            columns.column_name NOT IN('geom')
)

SELECT
    ARRAY(SELECT uc||'.'||cols.column_name FROM columns AS cols)
FROM usa_county AS uc;



















-- https://dba.stackexchange.com/questions/1957/sql-select-all-columns-except-some
SELECT 'SELECT ' || array_to_string(ARRAY(SELECT 'o' || '.' || c.column_name
        FROM information_schema.columns As c
            WHERE table_name = 'tl_2017_us_county'
            AND  c.column_name NOT IN('geom')
    ), ',') || ' FROM tl_2017_us_county As o' As sqlstmt





    WITH sqlstmt AS (
        SELECT 'SELECT ' || array_to_string(ARRAY(SELECT 'o' || '.' || c.column_name
            FROM information_schema.columns As c
                WHERE table_name = 'tl_2017_us_county'
                AND  c.column_name NOT IN('geom')
        ), ',') || ' FROM tl_2017_us_county As o'
    )

    EXECUTE sqlstmt;
