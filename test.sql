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
