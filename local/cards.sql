SELECT
    cards.id,
    cards.link_skill1_id,
    cards.link_skill2_id,
    cards.link_skill3_id,
    cards.link_skill4_id,
    cards.link_skill5_id,
    cards.link_skill6_id,
    cards.link_skill7_id,
    cards.rarity,
    cards.skill_lv_max,
    cards.optimal_awakening_grow_type,
    cards.potential_board_id,
    card_exps.exp_total,
    CASE
        WHEN cards.optimal_awakening_grow_type IS NULL THEN cards.skill_lv_max
        ELSE (
            SELECT
                optimal_awakening_growths.skill_lv_max
            FROM
                optimal_awakening_growths
            WHERE
                optimal_awakening_growths.optimal_awakening_grow_type = cards.optimal_awakening_grow_type
            ORDER BY
                optimal_awakening_growths.step DESC
            LIMIT
                1
        )
    END AS True_skill_lv,
    CASE
        WHEN cards.optimal_awakening_grow_type IS NULL THEN cards.lv_max
        ELSE (
            SELECT
                optimal_awakening_growths.lv_max
            FROM
                optimal_awakening_growths
            WHERE
                optimal_awakening_growths.optimal_awakening_grow_type = cards.optimal_awakening_grow_type
            ORDER BY
                optimal_awakening_growths.step DESC
            LIMIT
                1
        )
    END AS True_lv,
    CASE
        WHEN cards.rarity != 4 THEN NULL
        ELSE (
            SELECT
                card_decorations.id
            FROM
                card_decorations
            WHERE
                card_decorations.card_id == (cards.leader_skill_set_id || '1')
        )
    END as card_deco_id,
    CASE
        WHEN cards.optimal_awakening_grow_type IS NULL THEN null
        ELSE (
            SELECT
                optimal_awakening_growths.step
            FROM
                optimal_awakening_growths
            WHERE
                optimal_awakening_growths.optimal_awakening_grow_type = cards.optimal_awakening_grow_type
            ORDER BY
                optimal_awakening_growths.step DESC
            LIMIT
                1
        )
    END AS optimal_awakening_step,
    strftime('%s', cards.updated_at) as updated_at,
    strftime('%s', cards.created_at) as created_at
FROM
    cards
    JOIN card_exps ON cards.exp_type == card_exps.exp_type
WHERE
    (
        cards.id LIKE '1%'
        OR cards.id LIKE '4%'
    )
    AND (cards.id LIKE '%1')
    AND (card_exps.lv == cards.lv_max)
ORDER BY
    cards.id ASC