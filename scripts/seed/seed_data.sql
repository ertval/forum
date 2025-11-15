-- Seed data for testing the forum homepage
-- This script assumes a clean database with proper schema

-- Disable foreign key constraints temporarily
PRAGMA foreign_keys = OFF;

-- Clear existing data in the right order (respecting foreign key constraints)
DELETE FROM notifications;
DELETE FROM reports;
DELETE FROM reactions;
DELETE FROM comments;
DELETE FROM post_categories;
DELETE FROM posts;
DELETE FROM categories;
DELETE FROM sessions;
DELETE FROM users;

-- Enable foreign key constraints back
PRAGMA foreign_keys = ON;

-- Insert test users (INTEGER PRIMARY KEY - auto-increment)
INSERT OR IGNORE INTO users (email, username, password_hash, role, created_at, updated_at, is_active) VALUES
('alice@example.com', 'alice', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-7 days'), datetime('now', '-7 days'), 1),
('bob@example.com', 'bob', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-5 days'), datetime('now', '-5 days'), 1),
('charlie@example.com', 'charlie', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-3 days'), datetime('now', '-3 days'), 1),
('diana@example.com', 'diana', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-10 days'), datetime('now', '-10 days'), 1),
('eve@example.com', 'eve', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'moderator', datetime('now', '-8 days'), datetime('now', '-8 days'), 1),
('frank@example.com', 'frank', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-6 days'), datetime('now', '-6 days'), 1),
('grace@example.com', 'grace', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-4 days'), datetime('now', '-4 days'), 1),
('henry@example.com', 'henry', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'administrator', datetime('now', '-2 days'), datetime('now', '-2 days'), 1);

-- Insert test categories (TEXT PRIMARY KEY)
INSERT OR IGNORE INTO categories (id, name, description, created_at) VALUES
('general', 'General', 'General discussion topics', datetime('now', '-8 days')),
('technology', 'Technology', 'Tech-related posts and discussions', datetime('now', '-6 days')),
('gaming', 'Gaming', 'Video games and gaming culture', datetime('now', '-4 days')),
('science', 'Science', 'Science and research discussions', datetime('now', '-2 days')),
('entertainment', 'Entertainment', 'Movies, TV shows, and entertainment', datetime('now', '-1 day')),
('sports', 'Sports', 'Sports news and discussions', datetime('now', '-9 days')),
('health', 'Health', 'Health and wellness topics', datetime('now', '-7 days')),
('education', 'Education', 'Learning and educational content', datetime('now', '-5 days'));
-- Add Tests category for automated test posts
INSERT OR IGNORE INTO categories (id, name, description, created_at) VALUES
('tests', 'Tests', 'Automated test posts category', datetime('now'));

-- Insert test posts (TEXT PRIMARY KEY, INTEGER author_id)
INSERT OR IGNORE INTO posts (id, title, content, author_id, created_at, updated_at) VALUES
('post-1', 'Welcome to the Forum!', 'This is the first post on our new forum. Feel free to create your own posts and join discussions!', 1, datetime('now', '-7 days'), datetime('now', '-7 days')),
('post-2', 'Best Programming Languages in 2025', 'What do you think are the best programming languages to learn in 2025? I am currently learning Go and loving it!', 2, datetime('now', '-5 days'), datetime('now', '-5 days')),
('post-3', 'Favorite Video Games', 'What are your favorite video games of all time? Mine has to be The Legend of Zelda: Breath of the Wild.', 3, datetime('now', '-3 days'), datetime('now', '-3 days')),
('post-4', 'AI and Machine Learning Trends', 'The field of AI is evolving rapidly. What trends are you most excited about?', 1, datetime('now', '-2 days'), datetime('now', '-2 days')),
('post-5', 'Latest Movie Recommendations', 'Just watched an amazing sci-fi movie. What have you been watching lately?', 2, datetime('now', '-1 day'), datetime('now', '-1 day')),
('post-6', 'Climate Change Research', 'Recent studies show significant progress in renewable energy. Lets discuss!', 3, datetime('now', '-12 hours'), datetime('now', '-12 hours')),
('post-7', 'Healthy Eating Tips', 'Share your favorite healthy recipes and eating habits!', 4, datetime('now', '-6 days'), datetime('now', '-6 days')),
('post-8', 'Online Learning Platforms', 'Which online learning platforms do you recommend for skill development?', 5, datetime('now', '-4 days'), datetime('now', '-4 days')),
('post-9', 'Favorite Sports Teams', 'Which sports teams do you support and why?', 6, datetime('now', '-2 days'), datetime('now', '-2 days')),
('post-10', 'Tech Gadgets of 2025', 'What new tech gadgets are you excited about this year?', 7, datetime('now', '-1 day'), datetime('now', '-1 day'));

-- Associate posts with categories
INSERT OR IGNORE INTO post_categories (post_id, category_id) VALUES
('post-1', 'general'),   -- Welcome post in General
('post-2', 'technology'), -- Programming languages in Technology  
('post-3', 'gaming'),     -- Video games in Gaming
('post-4', 'technology'), -- AI in Technology
('post-5', 'entertainment'), -- Movies in Entertainment
('post-6', 'science'),    -- Climate in Science
('post-7', 'health'),     -- Healthy eating in Health
('post-8', 'education'),  -- Online learning in Education
('post-9', 'sports'),     -- Sports teams in Sports
('post-10', 'technology'); -- Tech gadgets in Technology

-- Insert some test reactions (TEXT target_id referencing posts.id)
INSERT OR IGNORE INTO reactions (user_id, target_id, target_type, type, created_at) VALUES
(2, 'post-1', 'post', 'like', datetime('now', '-6 days')),    -- bob likes post 1
(3, 'post-1', 'post', 'like', datetime('now', '-6 days')),    -- charlie likes post 1
(1, 'post-2', 'post', 'like', datetime('now', '-4 days')),    -- alice likes post 2
(3, 'post-2', 'post', 'like', datetime('now', '-4 days')),    -- charlie likes post 2
(1, 'post-3', 'post', 'like', datetime('now', '-2 days')),    -- alice likes post 3
(2, 'post-3', 'post', 'dislike', datetime('now', '-2 days')), -- bob dislikes post 3
(2, 'post-4', 'post', 'like', datetime('now', '-1 day')),     -- bob likes post 4
(3, 'post-5', 'post', 'like', datetime('now', '-12 hours')),  -- charlie likes post 5
(4, 'post-6', 'post', 'like', datetime('now', '-10 hours')),  -- diana likes post 6
(5, 'post-7', 'post', 'like', datetime('now', '-5 days')),    -- eve likes post 7
(6, 'post-8', 'post', 'like', datetime('now', '-3 days')),    -- frank likes post 8
(7, 'post-9', 'post', 'dislike', datetime('now', '-1 day')),  -- grace dislikes post 9
(8, 'post-10', 'post', 'like', datetime('now', '-12 hours')), -- henry likes post 10
(1, 'comment-1', 'comment', 'like', datetime('now', '-5 days')), -- alice likes comment 1
(2, 'comment-2', 'comment', 'dislike', datetime('now', '-5 days')), -- bob dislikes comment 2
(3, 'comment-3', 'comment', 'like', datetime('now', '-3 days')); -- charlie likes comment 3

-- Insert some test notifications
INSERT OR IGNORE INTO notifications (id, user_id, actor_id, target_id, type, message, read, created_at) VALUES
('notif-1', '1', '2', 'post-1', 'comment', 'Bob commented on your post "Welcome to the Forum!"', 0, datetime('now', '-6 days')),
('notif-2', '2', '1', 'post-2', 'comment', 'Alice commented on your post "Best Programming Languages in 2025"', 1, datetime('now', '-4 days')),
('notif-3', '3', '1', 'post-3', 'comment', 'Alice commented on your post "Favorite Video Games"', 0, datetime('now', '-2 days'));

-- Verify data insertion
SELECT 'Users inserted: ' || COUNT(*) FROM users;
SELECT 'Categories inserted: ' || COUNT(*) FROM categories;
SELECT 'Posts inserted: ' || COUNT(*) FROM posts;
SELECT 'Comments inserted: ' || COUNT(*) FROM comments;
SELECT 'Reactions inserted: ' || COUNT(*) FROM reactions;
SELECT 'Sessions inserted: ' || COUNT(*) FROM sessions;
SELECT 'Reports inserted: ' || COUNT(*) FROM reports;
SELECT 'Notifications inserted: ' || COUNT(*) FROM notifications;
