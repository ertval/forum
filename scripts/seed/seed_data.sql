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

-- Insert test users (ID auto-increments, public_id is UUID)
INSERT OR IGNORE INTO users (public_id, email, username, password_hash, role, created_at, updated_at, is_active) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'alice@example.com', 'alice', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-7 days'), datetime('now', '-7 days'), 1),
('550e8400-e29b-41d4-a716-446655440002', 'bob@example.com', 'bob', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-5 days'), datetime('now', '-5 days'), 1),
('550e8400-e29b-41d4-a716-446655440003', 'charlie@example.com', 'charlie', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-3 days'), datetime('now', '-3 days'), 1),
('550e8400-e29b-41d4-a716-446655440004', 'diana@example.com', 'diana', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-10 days'), datetime('now', '-10 days'), 1),
('550e8400-e29b-41d4-a716-446655440005', 'eve@example.com', 'eve', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'moderator', datetime('now', '-8 days'), datetime('now', '-8 days'), 1),
('550e8400-e29b-41d4-a716-446655440006', 'frank@example.com', 'frank', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-6 days'), datetime('now', '-6 days'), 1),
('550e8400-e29b-41d4-a716-446655440007', 'grace@example.com', 'grace', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-4 days'), datetime('now', '-4 days'), 1),
('550e8400-e29b-41d4-a716-446655440008', 'henry@example.com', 'henry', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'administrator', datetime('now', '-2 days'), datetime('now', '-2 days'), 1);

-- Insert test categories (ID auto-increments, public_id is UUID)
INSERT OR IGNORE INTO categories (public_id, name, description, created_at) VALUES
('650e8400-e29b-41d4-a716-446655440001', 'General', 'General discussion topics', datetime('now', '-8 days')),
('650e8400-e29b-41d4-a716-446655440002', 'Technology', 'Tech-related posts and discussions', datetime('now', '-6 days')),
('650e8400-e29b-41d4-a716-446655440003', 'Gaming', 'Video games and gaming culture', datetime('now', '-4 days')),
('650e8400-e29b-41d4-a716-446655440004', 'Science', 'Science and research discussions', datetime('now', '-2 days')),
('650e8400-e29b-41d4-a716-446655440005', 'Entertainment', 'Movies, TV shows, and entertainment', datetime('now', '-1 day')),
('650e8400-e29b-41d4-a716-446655440006', 'Sports', 'Sports news and discussions', datetime('now', '-9 days')),
('650e8400-e29b-41d4-a716-446655440007', 'Health', 'Health and wellness topics', datetime('now', '-7 days')),
('650e8400-e29b-41d4-a716-446655440008', 'Education', 'Learning and educational content', datetime('now', '-5 days')),
('650e8400-e29b-41d4-a716-446655440009', 'Tests', 'Automated test posts category', datetime('now'));

-- Insert test posts (ID auto-increments, public_id is UUID, author_id is INT referencing users.id)
INSERT OR IGNORE INTO posts (public_id, title, content, author_id, created_at, updated_at) VALUES
('750e8400-e29b-41d4-a716-446655440001', 'Welcome to the Forum!', 'This is the first post on our new forum. Feel free to create your own posts and join discussions!', 1, datetime('now', '-7 days'), datetime('now', '-7 days')),
('750e8400-e29b-41d4-a716-446655440002', 'Best Programming Languages in 2025', 'What do you think are the best programming languages to learn in 2025? I am currently learning Go and loving it!', 2, datetime('now', '-5 days'), datetime('now', '-5 days')),
('750e8400-e29b-41d4-a716-446655440003', 'Favorite Video Games', 'What are your favorite video games of all time? Mine has to be The Legend of Zelda: Breath of the Wild.', 3, datetime('now', '-3 days'), datetime('now', '-3 days')),
('750e8400-e29b-41d4-a716-446655440004', 'AI and Machine Learning Trends', 'The field of AI is evolving rapidly. What trends are you most excited about?', 1, datetime('now', '-2 days'), datetime('now', '-2 days')),
('750e8400-e29b-41d4-a716-446655440005', 'Latest Movie Recommendations', 'Just watched an amazing sci-fi movie. What have you been watching lately?', 2, datetime('now', '-1 day'), datetime('now', '-1 day')),
('750e8400-e29b-41d4-a716-446655440006', 'Climate Change Research', 'Recent studies show significant progress in renewable energy. Lets discuss!', 3, datetime('now', '-12 hours'), datetime('now', '-12 hours')),
('750e8400-e29b-41d4-a716-446655440007', 'Healthy Eating Tips', 'Share your favorite healthy recipes and eating habits!', 4, datetime('now', '-6 days'), datetime('now', '-6 days')),
('750e8400-e29b-41d4-a716-446655440008', 'Online Learning Platforms', 'Which online learning platforms do you recommend for skill development?', 5, datetime('now', '-4 days'), datetime('now', '-4 days')),
('750e8400-e29b-41d4-a716-446655440009', 'Favorite Sports Teams', 'Which sports teams do you support and why?', 6, datetime('now', '-2 days'), datetime('now', '-2 days')),
('750e8400-e29b-41d4-a716-446655440010', 'Tech Gadgets of 2025', 'What new tech gadgets are you excited about this year?', 7, datetime('now', '-1 day'), datetime('now', '-1 day'));

-- Associate posts with categories (using internal INT IDs)
INSERT OR IGNORE INTO post_categories (post_id, category_id) VALUES
(1, 1),   -- Welcome post in General
(2, 2),   -- Programming languages in Technology  
(3, 3),   -- Video games in Gaming
(4, 2),   -- AI in Technology
(5, 5),   -- Movies in Entertainment
(6, 4),   -- Climate in Science
(7, 7),   -- Healthy eating in Health
(8, 8),   -- Online learning in Education
(9, 6),   -- Sports teams in Sports
(10, 2);  -- Tech gadgets in Technology

-- Insert some test reactions (ID auto-increments, public_id is UUID, user_id and target_id are INT)
INSERT OR IGNORE INTO reactions (public_id, user_id, target_id, target_type, type, created_at) VALUES
('850e8400-e29b-41d4-a716-446655440001', 2, 1, 'post', 'like', datetime('now', '-6 days')),    -- bob likes post 1
('850e8400-e29b-41d4-a716-446655440002', 3, 1, 'post', 'like', datetime('now', '-6 days')),    -- charlie likes post 1
('850e8400-e29b-41d4-a716-446655440003', 1, 2, 'post', 'like', datetime('now', '-4 days')),    -- alice likes post 2
('850e8400-e29b-41d4-a716-446655440004', 3, 2, 'post', 'like', datetime('now', '-4 days')),    -- charlie likes post 2
('850e8400-e29b-41d4-a716-446655440005', 1, 3, 'post', 'like', datetime('now', '-2 days')),    -- alice likes post 3
('850e8400-e29b-41d4-a716-446655440006', 2, 3, 'post', 'dislike', datetime('now', '-2 days')), -- bob dislikes post 3
('850e8400-e29b-41d4-a716-446655440007', 2, 4, 'post', 'like', datetime('now', '-1 day')),     -- bob likes post 4
('850e8400-e29b-41d4-a716-446655440008', 3, 5, 'post', 'like', datetime('now', '-12 hours')),  -- charlie likes post 5
('850e8400-e29b-41d4-a716-446655440009', 4, 6, 'post', 'like', datetime('now', '-10 hours')),  -- diana likes post 6
('850e8400-e29b-41d4-a716-446655440010', 5, 7, 'post', 'like', datetime('now', '-5 days')),    -- eve likes post 7
('850e8400-e29b-41d4-a716-446655440011', 6, 8, 'post', 'like', datetime('now', '-3 days')),    -- frank likes post 8
('850e8400-e29b-41d4-a716-446655440012', 7, 9, 'post', 'dislike', datetime('now', '-1 day')),  -- grace dislikes post 9
('850e8400-e29b-41d4-a716-446655440013', 8, 10, 'post', 'like', datetime('now', '-12 hours')); -- henry likes post 10

-- Insert some test notifications (ID auto-increments, public_id is UUID, all IDs are INT)
INSERT OR IGNORE INTO notifications (public_id, user_id, actor_id, target_id, type, message, read, created_at) VALUES
('950e8400-e29b-41d4-a716-446655440001', 1, 2, 1, 'comment', 'Bob commented on your post "Welcome to the Forum!"', 0, datetime('now', '-6 days')),
('950e8400-e29b-41d4-a716-446655440002', 2, 1, 2, 'comment', 'Alice commented on your post "Best Programming Languages in 2025"', 1, datetime('now', '-4 days')),
('950e8400-e29b-41d4-a716-446655440003', 3, 1, 3, 'comment', 'Alice commented on your post "Favorite Video Games"', 0, datetime('now', '-2 days'));

-- Verify data insertion
SELECT 'Users inserted: ' || COUNT(*) FROM users;
SELECT 'Categories inserted: ' || COUNT(*) FROM categories;
SELECT 'Posts inserted: ' || COUNT(*) FROM posts;
SELECT 'Comments inserted: ' || COUNT(*) FROM comments;
SELECT 'Reactions inserted: ' || COUNT(*) FROM reactions;
SELECT 'Sessions inserted: ' || COUNT(*) FROM sessions;
SELECT 'Reports inserted: ' || COUNT(*) FROM reports;
SELECT 'Notifications inserted: ' || COUNT(*) FROM notifications;
