create table if not exists schema_migrations (
    version text primary key,
    applied_at timestamptz not null default now()
);

create extension if not exists pgcrypto;

create table if not exists provider_tokens (
    user_id uuid not null,
    provider text not null,
    provider_user_id text not null,
    access_token text not null,
    refresh_token text not null,
    expires_at timestamptz not null,
    scopes text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (user_id, provider)
);

create table if not exists raw_objects (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null,
    provider text not null,
    provider_object_type text not null,
    provider_object_id text not null,
    object_key text not null unique,
    sha256 text not null,
    content_type text not null,
    size_bytes bigint not null,
    fetched_at timestamptz not null default now(),
    created_at timestamptz not null default now(),
    unique (provider, provider_object_type, provider_object_id)
);

create table if not exists activities (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null,
    provider text not null,
    provider_id text not null,
    activity_type text not null,
    name text not null,
    started_at timestamptz not null,
    duration_seconds integer not null,
    moving_seconds integer not null,
    distance_meters double precision not null,
    elevation_gain_meters double precision,
    average_heartrate_bpm double precision,
    max_heartrate_bpm double precision,
    calories_kcal double precision,
    raw_object_key text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (provider, provider_id)
);
