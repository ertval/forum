-- =============================================================================
-- Comprehensive Seed Data for Forum Testing
-- This provides all data needed to test functionality per audit.md requirements
-- =============================================================================

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

-- =============================================================================
-- USERS - Various roles and test accounts
-- Password for all test users is "password123" (bcrypt hash below)
-- =============================================================================
INSERT OR IGNORE INTO users (public_id, email, username, password_hash, role, oauth_provider, oauth_provider_id, post_count, comment_count, reaction_count, created_at, updated_at, is_active) VALUES
-- Primary test user (use for most tests)
('test-user-0001-0001-000000000001', 'testuser@example.com', 'Test User', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 3, 5, 0, datetime('now', '-7 days'), datetime('now', '-7 days'), 1),
-- Secondary test user (for permission/ownership tests)
('test-user-0002-0002-000000000002', 'testuser2@example.com', 'Second User', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 2, 4, 0, datetime('now', '-5 days'), datetime('now', '-5 days'), 1),
-- Regular users
('550e8400-e29b-41d4-a716-446655440001', 'alice@example.com', 'Alice Smith', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 3, 5, 0, datetime('now', '-7 days'), datetime('now', '-7 days'), 1),
('550e8400-e29b-41d4-a716-446655440002', 'bob@example.com', 'Bob Johnson', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 2, 4, 0, datetime('now', '-5 days'), datetime('now', '-5 days'), 1),
('550e8400-e29b-41d4-a716-446655440003', 'charlie@example.com', 'Charlie Brown', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 2, 3, 0, datetime('now', '-3 days'), datetime('now', '-3 days'), 1),
('550e8400-e29b-41d4-a716-446655440004', 'diana@example.com', 'Diana Ross', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 1, 2, 0, datetime('now', '-10 days'), datetime('now', '-10 days'), 1),
-- Moderator account
('550e8400-e29b-41d4-a716-446655440005', 'eve@example.com', 'Eve Williams', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'moderator', NULL, NULL, 1, 1, 0, datetime('now', '-8 days'), datetime('now', '-8 days'), 1),
('550e8400-e29b-41d4-a716-446655440006', 'frank@example.com', 'Frank Miller', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 1, 1, 0, datetime('now', '-6 days'), datetime('now', '-6 days'), 1),
('550e8400-e29b-41d4-a716-446655440007', 'grace@example.com', 'Grace Taylor', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 1, 1, 0, datetime('now', '-4 days'), datetime('now', '-4 days'), 1),
-- Administrator account
('550e8400-e29b-41d4-a716-446655440008', 'henry@example.com', 'Henry Admin', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'admin', NULL, NULL, 1, 1, 0, datetime('now', '-2 days'), datetime('now', '-2 days'), 1);

-- =============================================================================
-- CATEGORIES - Required for post creation tests
-- =============================================================================
INSERT OR IGNORE INTO categories (public_id, name, description, created_at) VALUES
('cat-0001-0001-0001-000000000001', 'General', 'General discussion topics', datetime('now', '-8 days')),
('cat-0002-0002-0002-000000000002', 'Technology', 'Tech-related posts and discussions', datetime('now', '-6 days')),
('cat-0003-0003-0003-000000000003', 'Gaming', 'Video games and gaming culture', datetime('now', '-4 days')),
('cat-0004-0004-0004-000000000004', 'Science', 'Science and research discussions', datetime('now', '-2 days')),
('cat-0005-0005-0005-000000000005', 'Entertainment', 'Movies, TV shows, and entertainment', datetime('now', '-1 day')),
('cat-0006-0006-0006-000000000006', 'Sports', 'Sports news and discussions', datetime('now', '-9 days')),
('cat-0007-0007-0007-000000000007', 'Health', 'Health and wellness topics', datetime('now', '-7 days')),
('cat-0008-0008-0008-000000000008', 'Education', 'Learning and educational content', datetime('now', '-5 days')),
('cat-0009-0009-0009-000000000009', 'Tests', 'Category for automated testing', datetime('now'));

-- =============================================================================
-- POSTS - For testing post viewing, filtering, reactions
-- =============================================================================
INSERT OR IGNORE INTO posts (public_id, title, content, author_id, image_path, created_at, updated_at) VALUES
-- Posts by testuser (for ownership tests)
('post-0001-0001-0001-000000000001', 'Test User Post 1', 'This is the first post by testuser for testing.', (SELECT id FROM users WHERE username = 'Test User'), NULL, datetime('now', '-7 days'), datetime('now', '-7 days')),
('post-0002-0002-0002-000000000002', 'Test User Post 2', 'This is the second post by testuser for testing.', (SELECT id FROM users WHERE username = 'Test User'), NULL, datetime('now', '-6 days'), datetime('now', '-6 days')),
-- Posts by testuser2 (for cross-user tests)
('post-0003-0003-0003-000000000003', 'Test User 2 Post', 'This post belongs to testuser2.', (SELECT id FROM users WHERE username = 'Second User'), NULL, datetime('now', '-5 days'), datetime('now', '-5 days')),
-- Posts by other users
('750e8400-e29b-41d4-a716-446655440001', 'Welcome to the Forum!', 'This is the first post on our new forum. Feel free to create your own posts and join discussions!', (SELECT id FROM users WHERE username = 'Alice Smith'), NULL, datetime('now', '-7 days'), datetime('now', '-7 days')),
('750e8400-e29b-41d4-a716-446655440002', 'Best Programming Languages in 2025', 'What do you think are the best programming languages to learn in 2025? I am currently learning Go and loving it!', (SELECT id FROM users WHERE username = 'Bob Johnson'), NULL, datetime('now', '-5 days'), datetime('now', '-5 days')),
('750e8400-e29b-41d4-a716-446655440003', 'Favorite Video Games', 'What are your favorite video games of all time? Mine has to be The Legend of Zelda: Breath of the Wild.', (SELECT id FROM users WHERE username = 'Charlie Brown'), '003ee9e6-2bc4-476e-8f81-2345eb090599.jpg', datetime('now', '-3 days'), datetime('now', '-3 days')),
('750e8400-e29b-41d4-a716-446655440004', 'AI and Machine Learning Trends', 'The field of AI is evolving rapidly. What trends are you most excited about?', (SELECT id FROM users WHERE username = 'Alice Smith'), '00e71cfe-cfff-4107-8f9c-cd199468d6a0.png', datetime('now', '-2 days'), datetime('now', '-2 days')),
('750e8400-e29b-41d4-a716-446655440005', 'Latest Movie Recommendations', 'Just watched an amazing sci-fi movie. What have you been watching lately?', (SELECT id FROM users WHERE username = 'Bob Johnson'), NULL, datetime('now', '-1 day'), datetime('now', '-1 day')),
('750e8400-e29b-41d4-a716-446655440006', 'Healthy Eating Tips', 'Share your favorite healthy recipes and eating habits!', (SELECT id FROM users WHERE username = 'Diana Ross'), '0f2e4b54-66f8-414b-8615-2c80b2489f46.png', datetime('now', '-6 days'), datetime('now', '-6 days')),
('750e8400-e29b-41d4-a716-446655440007', 'Online Learning Platforms', 'Which online learning platforms do you recommend for skill development?', (SELECT id FROM users WHERE username = 'Eve Williams'), NULL, datetime('now', '-4 days'), datetime('now', '-4 days')),
-- Additional posts with images for image testing
('750e8400-e29b-41d4-a716-446655440008', 'Amazing Nature Photography', 'Check out this beautiful landscape I captured during my hiking trip!', (SELECT id FROM users WHERE username = 'Frank Miller'), '10462764-3698-4dba-aa8f-a7e51c48b7e5.png', datetime('now', '-3 days'), datetime('now', '-3 days')),
('750e8400-e29b-41d4-a716-446655440009', 'My Coding Setup 2025', 'Finally got my dream development workspace together. What do you think?', (SELECT id FROM users WHERE username = 'Grace Taylor'), '13d40610-9ea0-4711-8378-40e6e479870c.jpg', datetime('now', '-2 days'), datetime('now', '-2 days')),
('750e8400-e29b-41d4-a716-446655440010', 'Funny Meme of the Day', 'Found this hilarious meme, had to share!', (SELECT id FROM users WHERE username = 'Test User'), '00cc6d7c-5a29-4cfd-8651-79ed4c8da555.gif', datetime('now', '-1 day'), datetime('now', '-1 day')),
('750e8400-e29b-41d4-a716-446655440011', 'Sports Highlights', 'What an incredible game last night! Here are some highlights.', (SELECT id FROM users WHERE username = 'Henry Admin'), NULL, datetime('now'), datetime('now')),
-- Additional older posts for comprehensive testing
('750e8400-e29b-41d4-a716-446655440012', 'Classic Gaming Discussion', 'Remember the golden age of gaming? Lets discuss our favorite retro games and memories.', (SELECT id FROM users WHERE username = 'Alice Smith'), NULL, datetime('now', '-15 days'), datetime('now', '-15 days')),
('750e8400-e29b-41d4-a716-446655440013', 'Web Development Best Practices 2024', 'A comprehensive guide to modern web development patterns and tools that every developer should know.', (SELECT id FROM users WHERE username = 'Bob Johnson'), NULL, datetime('now', '-20 days'), datetime('now', '-20 days')),
('750e8400-e29b-41d4-a716-446655440014', 'Fitness Journey Update', 'Its been 3 months since I started my fitness journey. Here are my progress photos and what Ive learned.', (SELECT id FROM users WHERE username = 'Diana Ross'), NULL, datetime('now', '-25 days'), datetime('now', '-25 days')),
('750e8400-e29b-41d4-a716-446655440015', 'Climate Change and Technology', 'How technology can help combat climate change. Discussing renewable energy, carbon capture, and sustainable practices.', (SELECT id FROM users WHERE username = 'Charlie Brown'), NULL, datetime('now', '-30 days'), datetime('now', '-30 days')),
('750e8400-e29b-41d4-a716-446655440016', 'Book Recommendations for Winter', 'Cozy up with these amazing books! My top 10 recommendations for the winter season.', (SELECT id FROM users WHERE username = 'Eve Williams'), NULL, datetime('now', '-45 days'), datetime('now', '-45 days')),
('750e8400-e29b-41d4-a716-446655440017', 'Cryptocurrency Market Analysis', 'An in-depth look at the current state of cryptocurrency markets and future predictions.', (SELECT id FROM users WHERE username = 'Frank Miller'), NULL, datetime('now', '-60 days'), datetime('now', '-60 days')),
('750e8400-e29b-41d4-a716-446655440018', 'Travel Destinations 2025', 'Planning your next adventure? Here are the must-visit destinations for 2025.', (SELECT id FROM users WHERE username = 'Grace Taylor'), NULL, datetime('now', '-90 days'), datetime('now', '-90 days'));

-- Normalize any legacy image paths from previous seed/data formats
UPDATE posts SET image_path = REPLACE(image_path, 'uploads/', '') WHERE image_path LIKE 'uploads/%';
UPDATE posts SET image_path = NULL WHERE image_path LIKE '%seed-placeholder%';

-- =============================================================================
-- POST CATEGORIES - Link posts to categories (required for category filter tests)
-- =============================================================================
INSERT OR IGNORE INTO post_categories (post_id, category_id) VALUES
((SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), (SELECT id FROM categories WHERE name = 'General')),
((SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), (SELECT id FROM categories WHERE name = 'Tests')),
((SELECT id FROM posts WHERE public_id = 'post-0002-0002-0002-000000000002'), (SELECT id FROM categories WHERE name = 'Technology')),
((SELECT id FROM posts WHERE public_id = 'post-0003-0003-0003-000000000003'), (SELECT id FROM categories WHERE name = 'General')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440001'), (SELECT id FROM categories WHERE name = 'General')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), (SELECT id FROM categories WHERE name = 'Technology')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), (SELECT id FROM categories WHERE name = 'Gaming')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440004'), (SELECT id FROM categories WHERE name = 'Technology')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440004'), (SELECT id FROM categories WHERE name = 'Science')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440005'), (SELECT id FROM categories WHERE name = 'Entertainment')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440006'), (SELECT id FROM categories WHERE name = 'Health')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440007'), (SELECT id FROM categories WHERE name = 'Education')),
-- Categories for new posts with images
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440008'), (SELECT id FROM categories WHERE name = 'General')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440009'), (SELECT id FROM categories WHERE name = 'Technology')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440010'), (SELECT id FROM categories WHERE name = 'Entertainment')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440010'), (SELECT id FROM categories WHERE name = 'General')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440011'), (SELECT id FROM categories WHERE name = 'Sports'));

-- =============================================================================
-- COMMENTS - For comment display tests and reaction tests
-- =============================================================================
INSERT OR IGNORE INTO comments (public_id, post_id, author_id, content, created_at, updated_at) VALUES
-- Comments on testuser's posts
('comment-0001-0001-0001-000000000001', (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), (SELECT id FROM users WHERE username = 'Alice Smith'), 'Great post! This is really helpful.', datetime('now', '-6 days'), datetime('now', '-6 days')),
('comment-0002-0002-0002-000000000002', (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), (SELECT id FROM users WHERE username = 'Bob Johnson'), 'I agree with the above comment.', datetime('now', '-5 days'), datetime('now', '-5 days')),
('comment-0010-0010-0010-000000000010', (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), (SELECT id FROM users WHERE username = 'Charlie Brown'), 'Thanks for sharing this information!', datetime('now', '-4 days'), datetime('now', '-4 days')),
('comment-0011-0011-0011-000000000011', (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), (SELECT id FROM users WHERE username = 'Diana Ross'), 'Very insightful, learned something new today.', datetime('now', '-3 days'), datetime('now', '-3 days')),
-- Comments on programming languages post
('comment-0003-0003-0003-000000000003', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), (SELECT id FROM users WHERE username = 'Alice Smith'), 'Great question — I think Go and Rust are top choices for systems work.', datetime('now', '-4 days'), datetime('now', '-4 days')),
('comment-0012-0012-0012-000000000012', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), (SELECT id FROM users WHERE username = 'Test User'), 'Python is still king for AI/ML development.', datetime('now', '-3 days'), datetime('now', '-3 days')),
('comment-0013-0013-0013-000000000013', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), (SELECT id FROM users WHERE username = 'Charlie Brown'), 'Don''t sleep on TypeScript - it has come a long way!', datetime('now', '-2 days'), datetime('now', '-2 days')),
('comment-0014-0014-0014-000000000014', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), (SELECT id FROM users WHERE username = 'Eve Williams'), 'Rust is amazing but the learning curve is steep.', datetime('now', '-1 day'), datetime('now', '-1 day')),
-- Comments on gaming post
('comment-0004-0004-0004-000000000004', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), (SELECT id FROM users WHERE username = 'Bob Johnson'), 'I loved Breath of the Wild too! The world design is amazing.', datetime('now', '-2 days'), datetime('now', '-2 days')),
('comment-0005-0005-0005-000000000005', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), (SELECT id FROM users WHERE username = 'Diana Ross'), 'You should try Hades if you like roguelikes.', datetime('now', '-1 day'), datetime('now', '-1 day')),
('comment-0015-0015-0015-000000000015', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), (SELECT id FROM users WHERE username = 'Alice Smith'), 'Elden Ring is my GOTY for sure. Incredible open world!', datetime('now', '-1 day'), datetime('now', '-1 day')),
('comment-0016-0016-0016-000000000016', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), (SELECT id FROM users WHERE username = 'Frank Miller'), 'Classic titles like Mario and Zelda never get old.', datetime('now'), datetime('now')),
-- Comments on AI post
('comment-0017-0017-0017-000000000017', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440004'), (SELECT id FROM users WHERE username = 'Bob Johnson'), 'LLMs are changing everything. Exciting but also concerning.', datetime('now', '-1 day'), datetime('now', '-1 day')),
('comment-0018-0018-0018-000000000018', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440004'), (SELECT id FROM users WHERE username = 'Test User'), 'I''m most excited about AI in healthcare and drug discovery.', datetime('now', '-1 day'), datetime('now', '-1 day')),
('comment-0019-0019-0019-000000000019', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440004'), (SELECT id FROM users WHERE username = 'Eve Williams'), 'We need better AI safety research alongside capability research.', datetime('now'), datetime('now')),
-- Comments on movie post
('comment-0020-0020-0020-000000000020', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440005'), (SELECT id FROM users WHERE username = 'Alice Smith'), 'Just watched Dune Part 2 - absolutely stunning visuals!', datetime('now'), datetime('now')),
('comment-0021-0021-0021-000000000021', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440005'), (SELECT id FROM users WHERE username = 'Charlie Brown'), 'Try "Everything Everywhere All At Once" if you haven''t seen it.', datetime('now'), datetime('now')),
-- Comments on healthy eating post
('comment-0022-0022-0022-000000000022', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440006'), (SELECT id FROM users WHERE username = 'Test User'), 'Meal prepping on Sundays changed my life!', datetime('now', '-5 days'), datetime('now', '-5 days')),
('comment-0023-0023-0023-000000000023', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440006'), (SELECT id FROM users WHERE username = 'Grace Taylor'), 'Mediterranean diet is both delicious and healthy.', datetime('now', '-4 days'), datetime('now', '-4 days')),
-- Comments on online learning post
('comment-0024-0024-0024-000000000024', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440007'), (SELECT id FROM users WHERE username = 'Bob Johnson'), 'Coursera has amazing university courses for free!', datetime('now', '-3 days'), datetime('now', '-3 days')),
('comment-0025-0025-0025-000000000025', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440007'), (SELECT id FROM users WHERE username = 'Second User'), 'freeCodeCamp is great for learning to code.', datetime('now', '-2 days'), datetime('now', '-2 days'));

-- =============================================================================
-- REACTIONS - For like/dislike count display tests
-- =============================================================================
INSERT OR IGNORE INTO reactions (public_id, user_id, target_id, target_type, type, created_at) VALUES
-- Reactions on testuser's posts (for "see liked posts" test)
('reaction-0001-0001-0001-000000000001', (SELECT id FROM users WHERE username = 'Test User'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440001'), 'post', 'like', datetime('now', '-6 days')),
('reaction-0002-0002-0002-000000000002', (SELECT id FROM users WHERE username = 'Test User'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'post', 'like', datetime('now', '-4 days')),
-- Reactions from other users on testuser's posts
('reaction-0003-0003-0003-000000000003', (SELECT id FROM users WHERE username = 'Alice Smith'), (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), 'post', 'like', datetime('now', '-5 days')),
('reaction-0004-0004-0004-000000000004', (SELECT id FROM users WHERE username = 'Bob Johnson'), (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), 'post', 'like', datetime('now', '-4 days')),
('reaction-0005-0005-0005-000000000005', (SELECT id FROM users WHERE username = 'Charlie Brown'), (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), 'post', 'dislike', datetime('now', '-3 days')),
-- Comment reactions for comment like/dislike visibility tests
('reaction-0006-0006-0006-000000000006', (SELECT id FROM users WHERE username = 'Alice Smith'), (SELECT id FROM comments WHERE public_id = 'comment-0003-0003-0003-000000000003'), 'comment', 'like', datetime('now', '-3 days')),
('reaction-0007-0007-0007-000000000007', (SELECT id FROM users WHERE username = 'Bob Johnson'), (SELECT id FROM comments WHERE public_id = 'comment-0003-0003-0003-000000000003'), 'comment', 'dislike', datetime('now', '-2 days')),
-- More reactions on posts for popularity testing
('reaction-0008-0008-0008-000000000008', (SELECT id FROM users WHERE username = 'Diana Ross'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'post', 'like', datetime('now', '-3 days')),
('reaction-0009-0009-0009-000000000009', (SELECT id FROM users WHERE username = 'Eve Williams'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'post', 'like', datetime('now', '-2 days')),
('reaction-0010-0010-0010-000000000010', (SELECT id FROM users WHERE username = 'Frank Miller'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'post', 'like', datetime('now', '-1 day')),
('reaction-0011-0011-0011-000000000011', (SELECT id FROM users WHERE username = 'Grace Taylor'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'post', 'like', datetime('now')),
-- Gaming post gets mixed reactions
('reaction-0012-0012-0012-000000000012', (SELECT id FROM users WHERE username = 'Test User'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), 'post', 'like', datetime('now', '-2 days')),
('reaction-0013-0013-0013-000000000013', (SELECT id FROM users WHERE username = 'Alice Smith'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), 'post', 'like', datetime('now', '-1 day')),
('reaction-0014-0014-0014-000000000014', (SELECT id FROM users WHERE username = 'Bob Johnson'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), 'post', 'dislike', datetime('now')),
-- AI post reactions
('reaction-0015-0015-0015-000000000015', (SELECT id FROM users WHERE username = 'Charlie Brown'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440004'), 'post', 'like', datetime('now', '-1 day')),
('reaction-0016-0016-0016-000000000016', (SELECT id FROM users WHERE username = 'Second User'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440004'), 'post', 'like', datetime('now')),
-- Comment reactions
('reaction-0017-0017-0017-000000000017', (SELECT id FROM users WHERE username = 'Test User'), (SELECT id FROM comments WHERE public_id = 'comment-0001-0001-0001-000000000001'), 'comment', 'like', datetime('now', '-5 days')),
('reaction-0018-0018-0018-000000000018', (SELECT id FROM users WHERE username = 'Charlie Brown'), (SELECT id FROM comments WHERE public_id = 'comment-0001-0001-0001-000000000001'), 'comment', 'like', datetime('now', '-4 days')),
('reaction-0019-0019-0019-000000000019', (SELECT id FROM users WHERE username = 'Diana Ross'), (SELECT id FROM comments WHERE public_id = 'comment-0004-0004-0004-000000000004'), 'comment', 'like', datetime('now', '-1 day')),
('reaction-0020-0020-0020-000000000020', (SELECT id FROM users WHERE username = 'Alice Smith'), (SELECT id FROM comments WHERE public_id = 'comment-0012-0012-0012-000000000012'), 'comment', 'like', datetime('now', '-2 days')),
('reaction-0021-0021-0021-000000000021', (SELECT id FROM users WHERE username = 'Eve Williams'), (SELECT id FROM comments WHERE public_id = 'comment-0017-0017-0017-000000000017'), 'comment', 'like', datetime('now'));

-- =============================================================================
-- NOTIFICATIONS - For notification tests
-- =============================================================================
INSERT OR IGNORE INTO notifications (public_id, user_id, actor_id, target_id, type, message, read, created_at) VALUES
('notif-0001-0001-0001-000000000001', (SELECT id FROM users WHERE username = 'Test User'), (SELECT id FROM users WHERE username = 'Alice Smith'), (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), 'comment', 'alice commented on your post', 0, datetime('now', '-6 days')),
('notif-0002-0002-0002-000000000002', (SELECT id FROM users WHERE username = 'Test User'), (SELECT id FROM users WHERE username = 'Bob Johnson'), (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), 'reaction', 'bob liked your post', 1, datetime('now', '-4 days'));

-- =============================================================================
-- REPORTS - For moderation tests (optional feature)
-- =============================================================================
INSERT OR IGNORE INTO reports (public_id, reporter_id, moderator_id, target_id, target_type, reason, status, response, created_at, reviewed_at) VALUES
('report-0001-0001-0001-000000000001', (SELECT id FROM users WHERE username = 'Alice Smith'), (SELECT id FROM users WHERE username = 'Eve Williams'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440001'), 'post', 'Test report', 'resolved', 'Report resolved', datetime('now', '-5 days'), datetime('now', '-4 days')),
('report-0002-0002-0002-000000000002', (SELECT id FROM users WHERE username = 'Bob Johnson'), NULL, (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'post', 'Another test report', 'pending', NULL, datetime('now', '-3 days'), NULL);

-- =============================================================================
-- Verify data insertion
-- =============================================================================
SELECT '=== Seed Data Summary ===' AS message;
SELECT 'Users: ' || COUNT(*) FROM users;
SELECT 'Categories: ' || COUNT(*) FROM categories;
SELECT 'Posts: ' || COUNT(*) FROM posts;
SELECT 'Post-Categories: ' || COUNT(*) FROM post_categories;
SELECT 'Comments: ' || COUNT(*) FROM comments;
SELECT 'Reactions: ' || COUNT(*) FROM reactions;
SELECT 'Notifications: ' || COUNT(*) FROM notifications;
SELECT 'Reports: ' || COUNT(*) FROM reports;
SELECT '=== Test Credentials ===' AS message;
SELECT 'Email: testuser@example.com | Password: password123' AS primary_user;
SELECT 'Email: testuser2@example.com | Password: password123' AS secondary_user;
