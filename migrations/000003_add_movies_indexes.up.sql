CREATE INDEX IF NOT EXISTS manga_title_idx ON manga USING GIN (to_tsvector('simple',title));
CREATE INDEX IF NOT EXISTS manga_genres_idx ON manga USING GIN(genres);