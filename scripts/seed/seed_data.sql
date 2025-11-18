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
('550e8400-e29b-41d4-a716-446655440008', 'henry@example.com', 'henry', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'administrator', datetime('now', '-2 days'), datetime('now', '-2 days'), 1),
('550e8400-e29b-41d4-a716-446655440009', 'user9@example.com', 'user9', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-1 day'), datetime('now', '-1 day'), 1),
('550e8400-e29b-41d4-a716-446655440010', 'user10@example.com', 'user10', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-12 hours'), datetime('now', '-12 hours'), 1),
('550e8400-e29b-41d4-a716-446655440011', 'user11@example.com', 'user11', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-6 hours'), datetime('now', '-6 hours'), 1),
('550e8400-e29b-41d4-a716-446655440012', 'user12@example.com', 'user12', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-3 hours'), datetime('now', '-3 hours'), 1),
('550e8400-e29b-41d4-a716-446655440013', 'user13@example.com', 'user13', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-2 hours'), datetime('now', '-2 hours'), 1),
('550e8400-e29b-41d4-a716-446655440014', 'user14@example.com', 'user14', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-1 hour'), datetime('now', '-1 hour'), 1),
('550e8400-e29b-41d4-a716-446655440015', 'user15@example.com', 'user15', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-30 minutes'), datetime('now', '-30 minutes'), 1),
('550e8400-e29b-41d4-a716-446655440016', 'user16@example.com', 'user16', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-15 minutes'), datetime('now', '-15 minutes'), 1),
('550e8400-e29b-41d4-a716-446655440017', 'user17@example.com', 'user17', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-10 minutes'), datetime('now', '-10 minutes'), 1),
('550e8400-e29b-41d4-a716-446655440018', 'user18@example.com', 'user18', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-5 minutes'), datetime('now', '-5 minutes'), 1),
('550e8400-e29b-41d4-a716-446655440019', 'user19@example.com', 'user19', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-2 minutes'), datetime('now', '-2 minutes'), 1),
('550e8400-e29b-41d4-a716-446655440020', 'user20@example.com', 'user20', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now', '-1 minute'), datetime('now', '-1 minute'), 1);

-- Insert test sessions (use lookups for user IDs)
INSERT OR IGNORE INTO sessions (public_id, user_id, token, expires_at, created_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', (SELECT id FROM users WHERE username = 'alice'), 'session_token_alice_001', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440002', (SELECT id FROM users WHERE username = 'bob'), 'session_token_bob_002', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440003', (SELECT id FROM users WHERE username = 'charlie'), 'session_token_charlie_003', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440004', (SELECT id FROM users WHERE username = 'diana'), 'session_token_diana_004', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440005', (SELECT id FROM users WHERE username = 'eve'), 'session_token_eve_005', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440006', (SELECT id FROM users WHERE username = 'frank'), 'session_token_frank_006', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440007', (SELECT id FROM users WHERE username = 'grace'), 'session_token_grace_007', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440008', (SELECT id FROM users WHERE username = 'henry'), 'session_token_henry_008', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440009', (SELECT id FROM users WHERE username = 'user9'), 'session_token_user9_009', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440010', (SELECT id FROM users WHERE username = 'user10'), 'session_token_user10_010', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440011', (SELECT id FROM users WHERE username = 'user11'), 'session_token_user11_011', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440012', (SELECT id FROM users WHERE username = 'user12'), 'session_token_user12_012', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440013', (SELECT id FROM users WHERE username = 'user13'), 'session_token_user13_013', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440014', (SELECT id FROM users WHERE username = 'user14'), 'session_token_user14_014', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440015', (SELECT id FROM users WHERE username = 'user15'), 'session_token_user15_015', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440016', (SELECT id FROM users WHERE username = 'user16'), 'session_token_user16_016', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440017', (SELECT id FROM users WHERE username = 'user17'), 'session_token_user17_017', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440018', (SELECT id FROM users WHERE username = 'user18'), 'session_token_user18_018', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440019', (SELECT id FROM users WHERE username = 'user19'), 'session_token_user19_019', datetime('now', '+24 hours'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440020', (SELECT id FROM users WHERE username = 'user20'), 'session_token_user20_020', datetime('now', '+24 hours'), datetime('now'));

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
('650e8400-e29b-41d4-a716-446655440009', 'Tests', 'Automated test posts category', datetime('now')),
('650e8400-e29b-41d4-a716-446655440010', 'Travel', 'Travel experiences, tips and trip reports', datetime('now', '-3 days')),
('650e8400-e29b-41d4-a716-446655440011', 'Photography', 'Techniques, gear and photo sharing', datetime('now', '-2 days')),
('650e8400-e29b-41d4-a716-446655440012', 'Books', 'Book recommendations, reviews and reading lists', datetime('now', '-1 day')),
('650e8400-e29b-41d4-a716-446655440013', 'Finance', 'Personal finance, investing and frugal living', datetime('now', '-12 hours')),
('650e8400-e29b-41d4-a716-446655440014', 'Career', 'Career advice, resumes and interviews', datetime('now', '-6 hours')),
('650e8400-e29b-41d4-a716-446655440015', 'DIY', 'Do-it-yourself projects and maker resources', datetime('now', '-3 hours')),
('650e8400-e29b-41d4-a716-446655440016', 'Music', 'Music, instruments and practice tips', datetime('now', '-2 hours')),
('650e8400-e29b-41d4-a716-446655440017', 'Art', 'Drawing, painting and visual arts', datetime('now', '-1 hour')),
('650e8400-e29b-41d4-a716-446655440018', 'Food', 'Recipes, restaurants and food culture', datetime('now', '-30 minutes')),
('650e8400-e29b-41d4-a716-446655440019', 'Parenting', 'Parenting experiences and resources', datetime('now', '-15 minutes')),
('650e8400-e29b-41d4-a716-446655440020', 'Environment', 'Conservation, sustainability and climate action', datetime('now', '-5 minutes'));

-- Insert test posts (ID auto-increments, public_id is UUID, author_id is INT referencing users.id)
-- Insert posts using subqueries to look up current internal IDs (robust against existing sequences)
INSERT OR IGNORE INTO posts (public_id, title, content, author_id, created_at, updated_at) VALUES
('750e8400-e29b-41d4-a716-446655440001', 'Welcome to the Forum!', 'This is the first post on our new forum. Feel free to create your own posts and join discussions!', (SELECT id FROM users WHERE username = 'alice'), datetime('now', '-7 days'), datetime('now', '-7 days')),
('750e8400-e29b-41d4-a716-446655440002', 'Best Programming Languages in 2025', 'What do you think are the best programming languages to learn in 2025? I am currently learning Go and loving it!', (SELECT id FROM users WHERE username = 'bob'), datetime('now', '-5 days'), datetime('now', '-5 days')),
('750e8400-e29b-41d4-a716-446655440003', 'Favorite Video Games', 'What are your favorite video games of all time? Mine has to be The Legend of Zelda: Breath of the Wild.', (SELECT id FROM users WHERE username = 'charlie'), datetime('now', '-3 days'), datetime('now', '-3 days')),
('750e8400-e29b-41d4-a716-446655440004', 'AI and Machine Learning Trends', 'The field of AI is evolving rapidly. What trends are you most excited about?', (SELECT id FROM users WHERE username = 'alice'), datetime('now', '-2 days'), datetime('now', '-2 days')),
('750e8400-e29b-41d4-a716-446655440005', 'Latest Movie Recommendations', 'Just watched an amazing sci-fi movie. What have you been watching lately?', (SELECT id FROM users WHERE username = 'bob'), datetime('now', '-1 day'), datetime('now', '-1 day')),
('750e8400-e29b-41d4-a716-446655440006', 'Climate Change Research', 'Recent studies show significant progress in renewable energy. Lets discuss!', (SELECT id FROM users WHERE username = 'charlie'), datetime('now', '-12 hours'), datetime('now', '-12 hours')),
('750e8400-e29b-41d4-a716-446655440007', 'Healthy Eating Tips', 'Share your favorite healthy recipes and eating habits!', (SELECT id FROM users WHERE username = 'diana'), datetime('now', '-6 days'), datetime('now', '-6 days')),
('750e8400-e29b-41d4-a716-446655440008', 'Online Learning Platforms', 'Which online learning platforms do you recommend for skill development?', (SELECT id FROM users WHERE username = 'eve'), datetime('now', '-4 days'), datetime('now', '-4 days')),
('750e8400-e29b-41d4-a716-446655440009', 'Favorite Sports Teams', 'Which sports teams do you support and why?', (SELECT id FROM users WHERE username = 'frank'), datetime('now', '-2 days'), datetime('now', '-2 days')),
('750e8400-e29b-41d4-a716-446655440010', 'Tech Gadgets of 2025', 'What new tech gadgets are you excited about this year?', (SELECT id FROM users WHERE username = 'grace'), datetime('now', '-1 day'), datetime('now', '-1 day'));

-- Associate posts with categories (use lookups so we don't depend on internal numeric IDs)
INSERT OR IGNORE INTO post_categories (post_id, category_id) VALUES
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440001'), (SELECT id FROM categories WHERE name = 'General')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), (SELECT id FROM categories WHERE name = 'Technology')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), (SELECT id FROM categories WHERE name = 'Gaming')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440004'), (SELECT id FROM categories WHERE name = 'Technology')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440005'), (SELECT id FROM categories WHERE name = 'Entertainment')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440006'), (SELECT id FROM categories WHERE name = 'Science')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440007'), (SELECT id FROM categories WHERE name = 'Health')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440008'), (SELECT id FROM categories WHERE name = 'Education')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440009'), (SELECT id FROM categories WHERE name = 'Sports')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440010'), (SELECT id FROM categories WHERE name = 'Technology'));

-- Additional posts to create more cases and ensure some users have multiple posts (use lookups for author)
INSERT OR IGNORE INTO posts (public_id, title, content, author_id, created_at, updated_at) VALUES
('750e8400-e29b-41d4-a716-446655440011', 'Deep Dive into Go Concurrency', 'Goroutines and channels are great—let''s discuss patterns and anti-patterns for concurrent programs.', (SELECT id FROM users WHERE username = 'alice'), datetime('now', '-3 days'), datetime('now', '-3 days')),
('750e8400-e29b-41d4-a716-446655440012', 'Indie Game Dev Tips', 'Starting an indie game studio? Here are lessons learned and resources.', (SELECT id FROM users WHERE username = 'charlie'), datetime('now', '-2 days'), datetime('now', '-2 days')),
('750e8400-e29b-41d4-a716-446655440013', 'Exercise Routines', 'Quick at-home exercises that boost energy and concentration.', (SELECT id FROM users WHERE username = 'diana'), datetime('now', '-1 day'), datetime('now', '-1 day')),
('750e8400-e29b-41d4-a716-446655440014', 'Backpacking Through Patagonia', 'I just returned from a three-week backpacking trip through Patagonia — here are my route notes, highlights, and tips for anyone planning a similar adventure.', (SELECT id FROM users WHERE username = 'user9'), datetime('now', '-12 hours'), datetime('now', '-12 hours')),
('750e8400-e29b-41d4-a716-446655440015', 'Mastering Low-Light Photography', 'Low-light photography can be challenging. I share my workflow for shooting handheld at night and how I process raw images for clarity and color.', (SELECT id FROM users WHERE username = 'user10'), datetime('now', '-6 hours'), datetime('now', '-6 hours')),
('750e8400-e29b-41d4-a716-446655440016', 'Creating a Budget That Works', 'A practical guide to building a monthly budget, tracking expenses, and small changes that compound into big savings over a year.', (SELECT id FROM users WHERE username = 'user11'), datetime('now', '-3 hours'), datetime('now', '-3 hours')),
('750e8400-e29b-41d4-a716-446655440017', 'Balancing Work and Career Growth', 'How I negotiated a promotion while maintaining work-life balance — tips on goal-setting, skill development, and conversations with managers.', (SELECT id FROM users WHERE username = 'user12'), datetime('now', '-2 hours'), datetime('now', '-2 hours')),
('750e8400-e29b-41d4-a716-446655440018', 'Beginner Guitar Practice Routine', 'A 20-minute daily practice routine for beginners that focuses on chord changes, rhythm, and one short melody exercise.', (SELECT id FROM users WHERE username = 'user13'), datetime('now', '-1 hour'), datetime('now', '-1 hour')),
('750e8400-e29b-41d4-a716-446655440019', 'Urban Sketching: Materials & Tips', 'Sharing my favorite sketching kit for city sketchwalks and a few techniques for quick perspective and shading on the go.', (SELECT id FROM users WHERE username = 'user14'), datetime('now', '-30 minutes'), datetime('now', '-30 minutes')),
('750e8400-e29b-41d4-a716-446655440020', 'Zero-Waste Kitchen Basics', 'Small changes to reduce kitchen waste: composting tips, storage swaps, and recipes that use up leftovers.', (SELECT id FROM users WHERE username = 'user15'), datetime('now', '-15 minutes'), datetime('now', '-15 minutes'));

-- Categories for the newly added posts (use lookups)
INSERT OR IGNORE INTO post_categories (post_id, category_id) VALUES
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440011'), (SELECT id FROM categories WHERE name = 'Technology')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440012'), (SELECT id FROM categories WHERE name = 'Gaming')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440013'), (SELECT id FROM categories WHERE name = 'Health')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440014'), (SELECT id FROM categories WHERE name = 'Travel')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440015'), (SELECT id FROM categories WHERE name = 'Photography')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440016'), (SELECT id FROM categories WHERE name = 'Books')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440017'), (SELECT id FROM categories WHERE name = 'Finance')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440018'), (SELECT id FROM categories WHERE name = 'Career')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440019'), (SELECT id FROM categories WHERE name = 'Music')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440020'), (SELECT id FROM categories WHERE name = 'Art'));


-- Insert test reactions (use lookups for user/post IDs)
INSERT OR IGNORE INTO reactions (public_id, user_id, target_id, target_type, type, created_at) VALUES
('850e8400-e29b-41d4-a716-446655440001', (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440001'), 'post', 'like', datetime('now', '-6 days')),
('850e8400-e29b-41d4-a716-446655440002', (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'post', 'like', datetime('now', '-4 days')),
('850e8400-e29b-41d4-a716-446655440003', (SELECT id FROM users WHERE username = 'charlie'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), 'post', 'like', datetime('now', '-2 days')),
('850e8400-e29b-41d4-a716-446655440004', (SELECT id FROM users WHERE username = 'diana'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440004'), 'post', 'like', datetime('now', '-5 days')),
('850e8400-e29b-41d4-a716-446655440005', (SELECT id FROM users WHERE username = 'eve'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440005'), 'post', 'like', datetime('now', '-12 hours')),
('850e8400-e29b-41d4-a716-446655440006', (SELECT id FROM users WHERE username = 'frank'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440006'), 'post', 'like', datetime('now', '-1 hour')),
('850e8400-e29b-41d4-a716-446655440007', (SELECT id FROM users WHERE username = 'grace'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440007'), 'post', 'like', datetime('now', '-5 days')),
('850e8400-e29b-41d4-a716-446655440008', (SELECT id FROM users WHERE username = 'henry'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440008'), 'post', 'like', datetime('now', '-3 days')),
('850e8400-e29b-41d4-a716-446655440009', (SELECT id FROM users WHERE username = 'user9'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440009'), 'post', 'dislike', datetime('now', '-1 day')),
('850e8400-e29b-41d4-a716-446655440010', (SELECT id FROM users WHERE username = 'user10'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440010'), 'post', 'like', datetime('now', '-12 hours')),
('850e8400-e29b-41d4-a716-446655440011', (SELECT id FROM users WHERE username = 'user11'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440011'), 'post', 'like', datetime('now', '-2 days')),
('850e8400-e29b-41d4-a716-446655440012', (SELECT id FROM users WHERE username = 'user12'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440012'), 'post', 'like', datetime('now', '-1 day')),
('850e8400-e29b-41d4-a716-446655440013', (SELECT id FROM users WHERE username = 'user13'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440013'), 'post', 'like', datetime('now', '-2 seconds')),
('850e8400-e29b-41d4-a716-446655440014', (SELECT id FROM users WHERE username = 'user14'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440014'), 'post', 'like', datetime('now', '-12 hours')),
('850e8400-e29b-41d4-a716-446655440015', (SELECT id FROM users WHERE username = 'user15'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440015'), 'post', 'like', datetime('now', '-6 hours')),
('850e8400-e29b-41d4-a716-446655440016', (SELECT id FROM users WHERE username = 'user16'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440016'), 'post', 'like', datetime('now', '-3 hours')),
('850e8400-e29b-41d4-a716-446655440017', (SELECT id FROM users WHERE username = 'user17'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440017'), 'post', 'like', datetime('now', '-1 hour')),
('850e8400-e29b-41d4-a716-446655440018', (SELECT id FROM users WHERE username = 'user18'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440018'), 'post', 'like', datetime('now', '-30 minutes')),
('850e8400-e29b-41d4-a716-446655440019', (SELECT id FROM users WHERE username = 'user19'), (SELECT id FROM comments WHERE public_id = 'a50e8400-e29b-41d4-a716-446655440001'), 'comment', 'like', datetime('now', '-2 days')),
('850e8400-e29b-41d4-a716-446655440020', (SELECT id FROM users WHERE username = 'user20'), (SELECT id FROM comments WHERE public_id = 'a50e8400-e29b-41d4-a716-446655440002'), 'comment', 'dislike', datetime('now', '-3 days'));

-- Insert test comments (use lookups for post and author ids)
INSERT OR IGNORE INTO comments (public_id, post_id, author_id, content, created_at, updated_at) VALUES
('a50e8400-e29b-41d4-a716-446655440001', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), (SELECT id FROM users WHERE username = 'alice'), 'Great question — I think Go and Rust are top choices for systems and concurrency work.', datetime('now', '-4 days'), datetime('now', '-4 days')),
('a50e8400-e29b-41d4-a716-446655440002', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), (SELECT id FROM users WHERE username = 'bob'), 'I loved Breath of the Wild too! The world design is amazing.', datetime('now', '-2 days'), datetime('now', '-2 days')),
('a50e8400-e29b-41d4-a716-446655440003', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), (SELECT id FROM users WHERE username = 'alice'), 'You should try Hades if you like roguelikes — amazing combat loop.', datetime('now', '-1 day'), datetime('now', '-1 day')),
('a50e8400-e29b-41d4-a716-446655440004', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440001'), (SELECT id FROM users WHERE username = 'bob'), 'Congrats on the launch! Looking forward to more content.', datetime('now', '-6 days'), datetime('now', '-6 days')),
('a50e8400-e29b-41d4-a716-446655440005', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), (SELECT id FROM users WHERE username = 'charlie'), 'I prefer Python for quick scripts, but Go has great performance.', datetime('now', '-3 days'), datetime('now', '-3 days')),
('a50e8400-e29b-41d4-a716-446655440006', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440011'), (SELECT id FROM users WHERE username = 'bob'), 'Nice deep dive — concurrency is tricky, thanks for the examples.', datetime('now', '-2 days'), datetime('now', '-2 days')),
('a50e8400-e29b-41d4-a716-446655440007', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440011'), (SELECT id FROM users WHERE username = 'alice'), 'Thanks for sharing, I have a question about channel buffering.', datetime('now', '-1 day'), datetime('now', '-1 day')),
('a50e8400-e29b-41d4-a716-446655440008', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440012'), (SELECT id FROM users WHERE username = 'charlie'), 'Good tips for indie devs! Would love to hear about monetization strategies.', datetime('now', '-1 day'), datetime('now', '-1 day')),
('a50e8400-e29b-41d4-a716-446655440009', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440005'), (SELECT id FROM users WHERE username = 'alice'), 'I saw that movie too — here are a few recommendations based on that tone.', datetime('now', '-12 hours'), datetime('now', '-12 hours')),
('a50e8400-e29b-41d4-a716-446655440010', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440014'), (SELECT id FROM users WHERE username = 'user16'), 'Amazing photos — which trail did you find most challenging?', datetime('now', '-10 minutes'), datetime('now', '-10 minutes')),
('a50e8400-e29b-41d4-a716-446655440011', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440015'), (SELECT id FROM users WHERE username = 'user17'), 'Great tips — any recommendations for lenses for low-light handheld shots?', datetime('now', '-5 minutes'), datetime('now', '-5 minutes')),
('a50e8400-e29b-41d4-a716-446655440012', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440016'), (SELECT id FROM users WHERE username = 'user18'), 'Nice framework — do you use an app to track expenses or a spreadsheet?', datetime('now', '-2 minutes'), datetime('now', '-2 minutes')),
('a50e8400-e29b-41d4-a716-446655440013', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440017'), (SELECT id FROM users WHERE username = 'user19'), 'Congrats on the promotion — could you share how you structured the negotiation conversation?', datetime('now', '-1 minute'), datetime('now', '-1 minute')),
('a50e8400-e29b-41d4-a716-446655440014', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440018'), (SELECT id FROM users WHERE username = 'user20'), 'Thanks — any simple songs you recommend for absolute beginners?', datetime('now', '-30 seconds'), datetime('now', '-30 seconds')),
('a50e8400-e29b-41d4-a716-446655440015', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440019'), (SELECT id FROM users WHERE username = 'alice'), 'Love the kit — which pencils do you use for quick shading?', datetime('now', '-15 seconds'), datetime('now', '-15 seconds')),
('a50e8400-e29b-41d4-a716-446655440016', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440020'), (SELECT id FROM users WHERE username = 'bob'), 'Great tips — what compost method do you recommend for apartment kitchens?', datetime('now', '-10 seconds'), datetime('now', '-10 seconds')),
('a50e8400-e29b-41d4-a716-446655440017', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440001'), (SELECT id FROM users WHERE username = 'charlie'), 'Excited to be here — looking forward to thoughtful discussions and helpful threads.', datetime('now', '-5 seconds'), datetime('now', '-5 seconds')),
('a50e8400-e29b-41d4-a716-446655440018', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), (SELECT id FROM users WHERE username = 'diana'), 'I''ve found Go''s tooling and simplicity hard to beat for web services — great choice!', datetime('now', '-2 seconds'), datetime('now', '-2 seconds')),
('a50e8400-e29b-41d4-a716-446655440019', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), (SELECT id FROM users WHERE username = 'eve'), 'If you like open-world exploration, Horizon Zero Dawn has stunning environments and a great story.', datetime('now', '-1 second'), datetime('now', '-1 second')),
('a50e8400-e29b-41d4-a716-446655440020', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440004'), (SELECT id FROM users WHERE username = 'frank'), 'Edge ML and on-device privacy are exciting — seeing more real-world use cases lately.', datetime('now'), datetime('now'));

-- Additional notifications for new comments (use lookups)
INSERT OR IGNORE INTO notifications (public_id, user_id, actor_id, target_id, type, message, read, created_at) VALUES
('950e8400-e29b-41d4-a716-446655440004', (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'comment', 'Alice commented on your post "Best Programming Languages in 2025"', 0, datetime('now', '-4 days')),
('950e8400-e29b-41d4-a716-446655440005', (SELECT id FROM users WHERE username = 'charlie'), (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), 'comment', 'Bob commented on your post "Favorite Video Games"', 0, datetime('now', '-2 days')),
('950e8400-e29b-41d4-a716-446655440006', (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440011'), 'comment', 'Bob commented on your post "Deep Dive into Go Concurrency"', 0, datetime('now', '-2 days')),
('950e8400-e29b-41d4-a716-446655440007', (SELECT id FROM users WHERE username = 'diana'), (SELECT id FROM users WHERE username = 'eve'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440007'), 'reaction', 'Eve liked your post "Healthy Eating Tips"', 0, datetime('now', '-5 days')),
('950e8400-e29b-41d4-a716-446655440008', (SELECT id FROM users WHERE username = 'eve'), (SELECT id FROM users WHERE username = 'frank'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440008'), 'reaction', 'Frank liked your post "Online Learning Platforms"', 1, datetime('now', '-3 days')),
('950e8400-e29b-41d4-a716-446655440009', (SELECT id FROM users WHERE username = 'frank'), (SELECT id FROM users WHERE username = 'grace'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440009'), 'reaction', 'Grace disliked your post "Favorite Sports Teams"', 0, datetime('now', '-1 day')),
('950e8400-e29b-41d4-a716-446655440010', (SELECT id FROM users WHERE username = 'grace'), (SELECT id FROM users WHERE username = 'henry'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440010'), 'reaction', 'Henry liked your post "Tech Gadgets of 2025"', 0, datetime('now', '-12 hours')),
('950e8400-e29b-41d4-a716-446655440011', (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM users WHERE username = 'user9'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440001'), 'reaction', 'user9 liked your post "Welcome to the Forum!"', 0, datetime('now', '-6 days')),
('950e8400-e29b-41d4-a716-446655440012', (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM users WHERE username = 'user10'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'reaction', 'user10 liked your post "Best Programming Languages in 2025"', 1, datetime('now', '-4 days')),
('950e8400-e29b-41d4-a716-446655440013', (SELECT id FROM users WHERE username = 'charlie'), (SELECT id FROM users WHERE username = 'user11'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), 'reaction', 'user11 liked your post "Favorite Video Games"', 0, datetime('now', '-2 days')),
('950e8400-e29b-41d4-a716-446655440014', (SELECT id FROM users WHERE username = 'diana'), (SELECT id FROM users WHERE username = 'user12'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440007'), 'reaction', 'user12 liked your post "Healthy Eating Tips"', 0, datetime('now', '-6 days')),
('950e8400-e29b-41d4-a716-446655440015', (SELECT id FROM users WHERE username = 'eve'), (SELECT id FROM users WHERE username = 'user13'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440008'), 'reaction', 'user13 liked your post "Online Learning Platforms"', 1, datetime('now', '-4 days')),
('950e8400-e29b-41d4-a716-446655440016', (SELECT id FROM users WHERE username = 'frank'), (SELECT id FROM users WHERE username = 'user14'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440009'), 'reaction', 'user14 disliked your post "Favorite Sports Teams"', 0, datetime('now', '-2 days')),
('950e8400-e29b-41d4-a716-446655440017', (SELECT id FROM users WHERE username = 'grace'), (SELECT id FROM users WHERE username = 'user15'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440010'), 'reaction', 'user15 liked your post "Tech Gadgets of 2025"', 0, datetime('now', '-1 day')),
('950e8400-e29b-41d4-a716-446655440018', (SELECT id FROM users WHERE username = 'henry'), (SELECT id FROM users WHERE username = 'user16'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440001'), 'reaction', 'user16 liked your post "Welcome to the Forum!"', 0, datetime('now', '-7 days')),
('950e8400-e29b-41d4-a716-446655440019', (SELECT id FROM users WHERE username = 'user9'), (SELECT id FROM users WHERE username = 'user17'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440014'), 'reaction', 'user17 liked your post "Backpacking Through Patagonia"', 0, datetime('now', '-12 hours')),
('950e8400-e29b-41d4-a716-446655440020', (SELECT id FROM users WHERE username = 'user10'), (SELECT id FROM users WHERE username = 'user18'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440015'), 'reaction', 'user18 liked your post "Mastering Low-Light Photography"', 1, datetime('now', '-6 hours')),
('950e8400-e29b-41d4-a716-446655440021', (SELECT id FROM users WHERE username = 'user11'), (SELECT id FROM users WHERE username = 'user19'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440016'), 'reaction', 'user19 liked your post "Creating a Budget That Works"', 0, datetime('now', '-3 hours')),
('950e8400-e29b-41d4-a716-446655440022', (SELECT id FROM users WHERE username = 'user12'), (SELECT id FROM users WHERE username = 'user20'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440017'), 'reaction', 'user20 liked your post "Balancing Work and Career Growth"', 0, datetime('now', '-1 hour')),
('950e8400-e29b-41d4-a716-446655440023', (SELECT id FROM users WHERE username = 'user13'), (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440018'), 'reaction', 'alice liked your post "Beginner Guitar Practice Routine"', 1, datetime('now', '-30 minutes'));

-- Insert test reports (use lookups for user/post IDs)
INSERT OR IGNORE INTO reports (public_id, reporter_id, moderator_id, target_id, target_type, reason, status, response, created_at, reviewed_at) VALUES
('c50e8400-e29b-41d4-a716-446655440001', (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM users WHERE username = 'eve'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440001'), 'post', 'Inappropriate content', 'resolved', 'Content reviewed and approved', datetime('now', '-5 days'), datetime('now', '-4 days')),
('c50e8400-e29b-41d4-a716-446655440002', (SELECT id FROM users WHERE username = 'bob'), NULL, (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'post', 'Spam', 'pending', NULL, datetime('now', '-3 days'), NULL),
('c50e8400-e29b-41d4-a716-446655440003', (SELECT id FROM users WHERE username = 'charlie'), (SELECT id FROM users WHERE username = 'eve'), (SELECT id FROM comments WHERE public_id = 'a50e8400-e29b-41d4-a716-446655440001'), 'comment', 'Harassment', 'resolved', 'Comment removed', datetime('now', '-2 days'), datetime('now', '-1 day')),
('c50e8400-e29b-41d4-a716-446655440004', (SELECT id FROM users WHERE username = 'diana'), NULL, (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), 'post', 'Off-topic', 'pending', NULL, datetime('now', '-1 day'), NULL),
('c50e8400-e29b-41d4-a716-446655440005', (SELECT id FROM users WHERE username = 'eve'), (SELECT id FROM users WHERE username = 'henry'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440004'), 'post', 'Copyright violation', 'resolved', 'Content edited', datetime('now', '-6 hours'), datetime('now', '-3 hours')),
('c50e8400-e29b-41d4-a716-446655440006', (SELECT id FROM users WHERE username = 'frank'), NULL, (SELECT id FROM comments WHERE public_id = 'a50e8400-e29b-41d4-a716-446655440002'), 'comment', 'Spam', 'pending', NULL, datetime('now', '-4 hours'), NULL),
('c50e8400-e29b-41d4-a716-446655440007', (SELECT id FROM users WHERE username = 'grace'), (SELECT id FROM users WHERE username = 'eve'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440005'), 'post', 'Inappropriate language', 'resolved', 'Warning issued', datetime('now', '-2 hours'), datetime('now', '-1 hour')),
('c50e8400-e29b-41d4-a716-446655440008', (SELECT id FROM users WHERE username = 'henry'), NULL, (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440006'), 'post', 'Misinformation', 'pending', NULL, datetime('now', '-1 hour'), NULL),
('c50e8400-e29b-41d4-a716-446655440009', (SELECT id FROM users WHERE username = 'user9'), (SELECT id FROM users WHERE username = 'eve'), (SELECT id FROM comments WHERE public_id = 'a50e8400-e29b-41d4-a716-446655440003'), 'comment', 'Trolling', 'resolved', 'Comment hidden', datetime('now', '-30 minutes'), datetime('now', '-15 minutes')),
('c50e8400-e29b-41d4-a716-446655440010', (SELECT id FROM users WHERE username = 'user10'), NULL, (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440007'), 'post', 'Duplicate content', 'pending', NULL, datetime('now', '-20 minutes'), NULL),
('c50e8400-e29b-41d4-a716-446655440011', (SELECT id FROM users WHERE username = 'user11'), (SELECT id FROM users WHERE username = 'henry'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440008'), 'post', 'Advertising', 'resolved', 'Post removed', datetime('now', '-10 minutes'), datetime('now', '-5 minutes')),
('c50e8400-e29b-41d4-a716-446655440012', (SELECT id FROM users WHERE username = 'user12'), NULL, (SELECT id FROM comments WHERE public_id = 'a50e8400-e29b-41d4-a716-446655440004'), 'comment', 'Inappropriate content', 'pending', NULL, datetime('now', '-5 minutes'), NULL),
('c50e8400-e29b-41d4-a716-446655440013', (SELECT id FROM users WHERE username = 'user13'), (SELECT id FROM users WHERE username = 'eve'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440009'), 'post', 'Hate speech', 'resolved', 'User banned', datetime('now', '-2 minutes'), datetime('now', '-1 minute')),
('c50e8400-e29b-41d4-a716-446655440014', (SELECT id FROM users WHERE username = 'user14'), NULL, (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440010'), 'post', 'Spam', 'pending', NULL, datetime('now', '-1 minute'), NULL),
('c50e8400-e29b-41d4-a716-446655440015', (SELECT id FROM users WHERE username = 'user15'), (SELECT id FROM users WHERE username = 'henry'), (SELECT id FROM comments WHERE public_id = 'a50e8400-e29b-41d4-a716-446655440005'), 'comment', 'Harassment', 'resolved', 'Comment deleted', datetime('now', '-30 seconds'), datetime('now', '-15 seconds')),
('c50e8400-e29b-41d4-a716-446655440016', (SELECT id FROM users WHERE username = 'user16'), NULL, (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440011'), 'post', 'Off-topic', 'pending', NULL, datetime('now', '-20 seconds'), NULL),
('c50e8400-e29b-41d4-a716-446655440017', (SELECT id FROM users WHERE username = 'user17'), (SELECT id FROM users WHERE username = 'eve'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440012'), 'post', 'Inappropriate content', 'resolved', 'Content moderated', datetime('now', '-10 seconds'), datetime('now', '-5 seconds')),
('c50e8400-e29b-41d4-a716-446655440018', (SELECT id FROM users WHERE username = 'user18'), NULL, (SELECT id FROM comments WHERE public_id = 'a50e8400-e29b-41d4-a716-446655440006'), 'comment', 'Spam', 'pending', NULL, datetime('now', '-5 seconds'), NULL),
('c50e8400-e29b-41d4-a716-446655440019', (SELECT id FROM users WHERE username = 'user19'), (SELECT id FROM users WHERE username = 'henry'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440013'), 'post', 'Misinformation', 'resolved', 'Fact check added', datetime('now', '-2 seconds'), datetime('now', '-1 second')),
('c50e8400-e29b-41d4-a716-446655440020', (SELECT id FROM users WHERE username = 'user20'), NULL, (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440014'), 'post', 'Copyright violation', 'pending', NULL, datetime('now'), NULL);

-- Verify data insertion
SELECT 'Users inserted: ' || COUNT(*) FROM users;
SELECT 'Categories inserted: ' || COUNT(*) FROM categories;
SELECT 'Posts inserted: ' || COUNT(*) FROM posts;
SELECT 'Comments inserted: ' || COUNT(*) FROM comments;
SELECT 'Reactions inserted: ' || COUNT(*) FROM reactions;
SELECT 'Sessions inserted: ' || COUNT(*) FROM sessions;
SELECT 'Reports inserted: ' || COUNT(*) FROM reports;
SELECT 'Notifications inserted: ' || COUNT(*) FROM notifications;
