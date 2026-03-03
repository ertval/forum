-- Migration: Create posts and categories tables
-- Module: post
-- Version: 003

-- +migrate Up
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    public_id TEXT UNIQUE NOT NULL,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    public_id TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    author_id INTEGER NOT NULL,
    image_path TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS post_categories (
    post_id INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    PRIMARY KEY (post_id, category_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

CREATE INDEX idx_categories_public_id ON categories(public_id);
CREATE INDEX idx_posts_public_id ON posts(public_id);
CREATE INDEX idx_posts_author ON posts(author_id);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX idx_post_categories_category ON post_categories(category_id);
-- Categories: name is queried case-insensitively (WHERE LOWER(name) = LOWER(?))
-- The existing UNIQUE constraint index doesn't optimize case-insensitive lookups.
CREATE INDEX IF NOT EXISTS idx_categories_name_nocase ON categories(name COLLATE NOCASE);

-- +migrate Down
DROP INDEX IF EXISTS idx_post_categories_category;
DROP INDEX IF EXISTS idx_posts_created_at;
DROP INDEX IF EXISTS idx_posts_author;
DROP INDEX IF EXISTS idx_posts_public_id;
DROP INDEX IF EXISTS idx_categories_public_id;
DROP TABLE IF EXISTS post_categories;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS categories;
DROP INDEX IF EXISTS idx_categories_name_nocase;
