create table if not exists training_goals (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null unique,
    goal_type text not null
        check (goal_type = 'distance_event'),
    sport text not null
        check (sport in ('run', 'ride')),
    name text not null
        check (length(trim(name)) > 0),
    target_distance_meters double precision not null
        check (target_distance_meters > 0),
    target_date date not null,
    target_duration_seconds integer
        check (
            target_duration_seconds is null
            or target_duration_seconds > 0
        ),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
