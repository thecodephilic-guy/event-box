CREATE EXTENSION IF NOT EXISTS citext;

-- 1. USERS
CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    role text NOT NULL CHECK (role IN ('organizer', 'customer')),
    version integer NOT NULL DEFAULT 1
);

-- 2. TOKENS (Opaque Stateful Tokens)
CREATE TABLE IF NOT EXISTS tokens (
    hash bytea PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    expiry timestamp(0) with time zone NOT NULL,
    scope text NOT NULL
);

-- 3. EVENTS
CREATE TABLE IF NOT EXISTS events (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    organizer_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    title text NOT NULL,
    description text NOT NULL,
    date timestamp(0) with time zone NOT NULL,
    total_tickets integer NOT NULL,
    available_tickets integer NOT NULL CHECK (available_tickets >= 0),
    version integer NOT NULL DEFAULT 1
);

-- 4. BOOKINGS
CREATE TABLE IF NOT EXISTS bookings (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    event_id bigint NOT NULL REFERENCES events ON DELETE CASCADE,
    customer_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    UNIQUE(event_id, customer_id) -- Prevents a customer from booking the exact same event twice accidentally
);