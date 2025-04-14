-- name: CreateOrGetReception :one
WITH existing_reception AS (
    SELECT id FROM receptions
    WHERE receptions.pvz_id = $1 AND receptions.status = 'in_progress'
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

-- name: CloseReception :one
WITH reception_to_close AS (
    SELECT id, date_time, pvz_id, status 
    FROM receptions
    WHERE receptions.pvz_id = $1 AND receptions.status = 'in_progress'
    ORDER BY date_time DESC
    LIMIT 1
    FOR UPDATE NOWAIT
),
updated_reception AS (
    UPDATE receptions r
    SET status = 'close', updated_at = NOW()
    FROM reception_to_close rtc
    WHERE r.id = rtc.id
    RETURNING r.id, r.date_time, r.pvz_id, r.status
)
SELECT * FROM updated_reception
UNION ALL
SELECT id, date_time, pvz_id, status
FROM reception_to_close
WHERE NOT EXISTS (SELECT 1 FROM updated_reception)
LIMIT 1;
