-- name: AddProduct :one
WITH current_reception AS (
    SELECT id FROM receptions
    WHERE pvz_id = $1 AND status = 'in_progress'
    LIMIT 1
    FOR SHARE
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

-- name: DeleteLastProduct :one
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
