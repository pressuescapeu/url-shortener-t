CREATE TABLE IF NOT EXISTS public.url(
    id SERIAL PRIMARY KEY,
    alias TEXT NOT NULL UNIQUE,
    url   TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_alias ON public.url(alias);