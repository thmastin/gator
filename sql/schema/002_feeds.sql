-- +goose up
    CREATE TABLE feeds (
        id uuid PRIMARY KEY,
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP NOT NULL,
        name TEXT NOT NULL,
        url TEXT UNIQUE NOT NULL,
        user_id TEXT NOT NULL references users(id) ON DELETE CASCADE
    );

-- +goose down
    DROP TABLE feeds;