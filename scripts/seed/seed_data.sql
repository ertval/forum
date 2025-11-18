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
('750e8400-e29b-41d4-a716-446655440013', 'Exercise Routines', 'Quick at-home exercises that boost energy and concentration.', (SELECT id FROM users WHERE username = 'diana'), datetime('now', '-1 day'), datetime('now', '-1 day'));

-- Categories for the newly added posts (use lookups)
INSERT OR IGNORE INTO post_categories (post_id, category_id) VALUES
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440011'), (SELECT id FROM categories WHERE name = 'Technology')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440012'), (SELECT id FROM categories WHERE name = 'Gaming')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440013'), (SELECT id FROM categories WHERE name = 'Health'));


-- Insert some test reactions (use lookups for user/post IDs)
INSERT OR IGNORE INTO reactions (public_id, user_id, target_id, target_type, type, created_at) VALUES
('850e8400-e29b-41d4-a716-446655440001', (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440001'), 'post', 'like', datetime('now', '-6 days')),
('850e8400-e29b-41d4-a716-446655440002', (SELECT id FROM users WHERE username = 'charlie'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440001'), 'post', 'like', datetime('now', '-6 days')),
('850e8400-e29b-41d4-a716-446655440003', (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'post', 'like', datetime('now', '-4 days')),
('850e8400-e29b-41d4-a716-446655440004', (SELECT id FROM users WHERE username = 'charlie'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'post', 'like', datetime('now', '-4 days')),
('850e8400-e29b-41d4-a716-446655440005', (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), 'post', 'like', datetime('now', '-2 days')),
('850e8400-e29b-41d4-a716-446655440006', (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), 'post', 'dislike', datetime('now', '-2 days')),
('850e8400-e29b-41d4-a716-446655440007', (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440004'), 'post', 'like', datetime('now', '-1 day')),
('850e8400-e29b-41d4-a716-446655440008', (SELECT id FROM users WHERE username = 'charlie'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440005'), 'post', 'like', datetime('now', '-12 hours')),
('850e8400-e29b-41d4-a716-446655440009', (SELECT id FROM users WHERE username = 'diana'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440006'), 'post', 'like', datetime('now', '-10 hours')),
('850e8400-e29b-41d4-a716-446655440010', (SELECT id FROM users WHERE username = 'eve'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440007'), 'post', 'like', datetime('now', '-5 days')),
('850e8400-e29b-41d4-a716-446655440011', (SELECT id FROM users WHERE username = 'frank'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440008'), 'post', 'like', datetime('now', '-3 days')),
('850e8400-e29b-41d4-a716-446655440012', (SELECT id FROM users WHERE username = 'grace'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440009'), 'post', 'dislike', datetime('now', '-1 day')),
('850e8400-e29b-41d4-a716-446655440013', (SELECT id FROM users WHERE username = 'henry'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440010'), 'post', 'like', datetime('now', '-12 hours'));

-- Insert some test notifications (use lookups for user/post IDs)
INSERT OR IGNORE INTO notifications (public_id, user_id, actor_id, target_id, type, message, read, created_at) VALUES
('950e8400-e29b-41d4-a716-446655440001', (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440001'), 'comment', 'Bob commented on your post "Welcome to the Forum!"', 0, datetime('now', '-6 days')),
('950e8400-e29b-41d4-a716-446655440002', (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'comment', 'Alice commented on your post "Best Programming Languages in 2025"', 1, datetime('now', '-4 days')),
('950e8400-e29b-41d4-a716-446655440003', (SELECT id FROM users WHERE username = 'charlie'), (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), 'comment', 'Alice commented on your post "Favorite Video Games"', 0, datetime('now', '-2 days'));

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
('a50e8400-e29b-41d4-a716-446655440009', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440005'), (SELECT id FROM users WHERE username = 'alice'), 'I saw that movie too — here are a few recommendations based on that tone.', datetime('now', '-12 hours'), datetime('now', '-12 hours'));

-- Additional notifications for new comments (use lookups)
INSERT OR IGNORE INTO notifications (public_id, user_id, actor_id, target_id, type, message, read, created_at) VALUES
('950e8400-e29b-41d4-a716-446655440004', (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'comment', 'Alice commented on your post "Best Programming Languages in 2025"', 0, datetime('now', '-4 days')),
('950e8400-e29b-41d4-a716-446655440005', (SELECT id FROM users WHERE username = 'charlie'), (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), 'comment', 'Bob commented on your post "Favorite Video Games"', 0, datetime('now', '-2 days')),
('950e8400-e29b-41d4-a716-446655440006', (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440011'), 'comment', 'Bob commented on your post "Deep Dive into Go Concurrency"', 0, datetime('now', '-2 days'));

-- Verify data insertion
SELECT 'Users inserted: ' || COUNT(*) FROM users;
SELECT 'Categories inserted: ' || COUNT(*) FROM categories;
SELECT 'Posts inserted: ' || COUNT(*) FROM posts;
SELECT 'Comments inserted: ' || COUNT(*) FROM comments;
SELECT 'Reactions inserted: ' || COUNT(*) FROM reactions;
SELECT 'Sessions inserted: ' || COUNT(*) FROM sessions;
SELECT 'Reports inserted: ' || COUNT(*) FROM reports;
SELECT 'Notifications inserted: ' || COUNT(*) FROM notifications;
