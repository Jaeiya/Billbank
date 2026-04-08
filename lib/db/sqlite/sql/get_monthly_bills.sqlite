SELECT
    bm.id AS id,
    bm.is_active,
    b.amount,
    b.due_date,
    b.period,
    bt.name AS bill_type
FROM bills_monthly AS bm
JOIN bills AS b
    ON bm.bill_id = b.id
JOIN bill_types AS bt
    ON b.type_id = bt.id
