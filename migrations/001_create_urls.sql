CREATE TABLE IF NOT EXISTS urls (
  id           SERIAL PRIMARY KEY,
  short_code   VARCHAR(10) UNIQUE NOT NULL,
  original_url TEXT NOT NULL,
  click_count  BIGINT DEFAULT 0,
  created_at   TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
