-- In
INSERT INTO users (email, password_hash, role) 
VALUES ($1, crypt($2, gen_salt('bf', 8)), $3)
RETURNING id, email, role;

-- Use index-only scan with fast hash comparison
SELECT id, email, role 
FROM users 
WHERE email = $1 AND password_hash = crypt($2, password_hash);

-- Use advisory lock to prevent city-based contention
SELECT pg_advisory_xact_lock(hashtext($1));
INSERT INTO pvz (city) 
VALUES ($1) 
RETURNING id, registration_date, city;


-- Materialized CTE with parallel scan
WITH pvz_paginated AS (
    SELECT id, registration_date, city
    FROM pvz
    WHERE id > $4  -- Cursor pagination for better performance
    ORDER BY id
    LIMIT $3  -- Page size
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
            'type', p.type
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
                'status', rd.status
            ),
            'products', rd.products
        )) FROM reception_data rd WHERE rd.pvz_id = p.id),
        '[]'::json
    ) AS receptions
FROM pvz_paginated p;


-- Use SKIP LOCKED to handle high contention
WITH existing_reception AS (
    SELECT id FROM receptions
    WHERE pvz_id = $1 AND status = 'in_progress'
    LIMIT 1
    FOR UPDATE SKIP LOCKED
),
new_reception AS (
    INSERT INTO receptions (pvz_id, status)
    SELECT $1, 'in_progress'
    WHERE NOT EXISTS (SELECT 1 FROM existing_reception)
    RETURNING id, date_time, pvz_id, status
)
SELECT * FROM new_reception
UNION ALL
SELECT 
    r.id, 
    r.date_time, 
    r.pvz_id, 
    r.status
FROM existing_reception er
JOIN receptions r ON er.id = r.id;

-- Use explicit row lock with NOWAIT
WITH reception_to_close AS (
    SELECT id FROM receptions
    WHERE pvz_id = $1 AND status = 'in_progress'
    ORDER BY date_time DESC
    LIMIT 1
    FOR UPDATE NOWAIT
)
UPDATE receptions r
SET status = 'close', updated_at = NOW()
FROM reception_to_close rtc
WHERE r.id = rtc.id
RETURNING r.id, r.date_time, r.pvz_id, r.status;

-- Use RETURNING with CTE for single roundtrip
WITH current_reception AS (
    SELECT id FROM receptions
    WHERE pvz_id = $1 AND status = 'in_progress'
    LIMIT 1
    FOR SHARE  -- Less restrictive lock
),
product_insert AS (
    INSERT INTO products (type, reception_id)
    SELECT $2, id FROM current_reception
    RETURNING id, date_time, type, reception_id
)
SELECT * FROM product_insert
UNION ALL
SELECT NULL, NULL, NULL, NULL
WHERE NOT EXISTS (SELECT 1 FROM product_insert);

-- Use index-only scan for sequence
WITH product_to_delete AS (
    SELECT p.id
    FROM products p
    JOIN receptions r ON p.reception_id = r.id
    WHERE r.pvz_id = $1 AND r.status = 'in_progress'
    ORDER BY p.sequence DESC
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
DELETE FROM products
WHERE id IN (SELECT id FROM product_to_delete)
RETURNING id;
