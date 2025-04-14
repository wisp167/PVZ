-- name: CreatePVZ :one
INSERT INTO pvz (
    city
) VALUES (
    $1
)
RETURNING id, registration_date, city;


-- name: GetPVZsWithReceptions :many
WITH pvz_paginated AS (
    SELECT id, registration_date, city
    FROM pvz
    ORDER BY registration_date DESC, id
    LIMIT $3 OFFSET (($4 - 1) * $3)
),
reception_data AS (
    SELECT 
        r.pvz_id,
        r.id AS reception_id,
        r.date_time,
        r.status,
        (SELECT json_agg(json_build_object(
            'id', p.id,
            'dateTime', p.date_time,
            'type', p.type,
            'receptionId', p.reception_id
        ) ORDER BY p.sequence DESC)
        FROM products p
        WHERE p.reception_id = r.id
        ) AS products
    FROM receptions r
    WHERE 
        r.date_time BETWEEN COALESCE($1, '-infinity'::timestamp) 
                         AND COALESCE($2, 'infinity'::timestamp)
        AND r.pvz_id IN (SELECT id FROM pvz_paginated)
)
SELECT 
    p.id AS pvz_id,
    p.registration_date,
    p.city,
    COALESCE(
        (SELECT json_agg(json_build_object(
            'reception', json_build_object(
                'id', rd.reception_id,
                'dateTime', rd.date_time,
                'status', rd.status,
                'pvzId', rd.pvz_id
            ),
            'products', rd.products
        )) FROM reception_data rd WHERE rd.pvz_id = p.id),
        '[]'::json
    )::text AS receptions_json
FROM pvz_paginated p;
