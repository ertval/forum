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
INSERT OR IGNORE INTO users (public_id, email, username, password_hash, role, oauth_provider, oauth_provider_id, post_count, comment_count, created_at, updated_at, is_active) VALUES
-- Primary test user (use for most tests)
('test-user-0001-0001-000000000001', 'testuser@example.com', 'testuser', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 3, 5, datetime('now', '-7 days'), datetime('now', '-7 days'), 1),
-- Secondary test user (for permission/ownership tests)
('test-user-0002-0002-000000000002', 'testuser2@example.com', 'testuser2', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 2, 4, datetime('now', '-5 days'), datetime('now', '-5 days'), 1),
-- Regular users
('550e8400-e29b-41d4-a716-446655440001', 'alice@example.com', 'alice', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 3, 5, datetime('now', '-7 days'), datetime('now', '-7 days'), 1),
('550e8400-e29b-41d4-a716-446655440002', 'bob@example.com', 'bob', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 2, 4, datetime('now', '-5 days'), datetime('now', '-5 days'), 1),
('550e8400-e29b-41d4-a716-446655440003', 'charlie@example.com', 'charlie', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 2, 3, datetime('now', '-3 days'), datetime('now', '-3 days'), 1),
('550e8400-e29b-41d4-a716-446655440004', 'diana@example.com', 'diana', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 1, 2, datetime('now', '-10 days'), datetime('now', '-10 days'), 1),
-- Moderator account
('550e8400-e29b-41d4-a716-446655440005', 'eve@example.com', 'eve', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'moderator', NULL, NULL, 1, 1, datetime('now', '-8 days'), datetime('now', '-8 days'), 1),
('550e8400-e29b-41d4-a716-446655440006', 'frank@example.com', 'frank', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 1, 1, datetime('now', '-6 days'), datetime('now', '-6 days'), 1),
('550e8400-e29b-41d4-a716-446655440007', 'grace@example.com', 'grace', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'user', NULL, NULL, 1, 1, datetime('now', '-4 days'), datetime('now', '-4 days'), 1),
-- Administrator account
('550e8400-e29b-41d4-a716-446655440008', 'henry@example.com', 'henry', '$2a$10$S/1aDjSzfV5mhi3ViNY0/.BMKLo2CROoQqyrQLn46Ugq1V1JwVu1e', 'administrator', NULL, NULL, 1, 1, datetime('now', '-2 days'), datetime('now', '-2 days'), 1);

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
('post-0001-0001-0001-000000000001', 'Test User Post 1', 'This is the first post by testuser for testing.', (SELECT id FROM users WHERE username = 'testuser'), NULL, datetime('now', '-7 days'), datetime('now', '-7 days')),
('post-0002-0002-0002-000000000002', 'Test User Post 2', 'This is the second post by testuser for testing.', (SELECT id FROM users WHERE username = 'testuser'), NULL, datetime('now', '-6 days'), datetime('now', '-6 days')),
-- Posts by testuser2 (for cross-user tests)
('post-0003-0003-0003-000000000003', 'Test User 2 Post', 'This post belongs to testuser2.', (SELECT id FROM users WHERE username = 'testuser2'), NULL, datetime('now', '-5 days'), datetime('now', '-5 days')),
-- Posts by other users
('750e8400-e29b-41d4-a716-446655440001', 'Welcome to the Forum!', 'This is the first post on our new forum. Feel free to create your own posts and join discussions!', (SELECT id FROM users WHERE username = 'alice'), NULL, datetime('now', '-7 days'), datetime('now', '-7 days')),
('750e8400-e29b-41d4-a716-446655440002', 'Best Programming Languages in 2025', 'What do you think are the best programming languages to learn in 2025? I am currently learning Go and loving it!', (SELECT id FROM users WHERE username = 'bob'), NULL, datetime('now', '-5 days'), datetime('now', '-5 days')),
('750e8400-e29b-41d4-a716-446655440003', 'Favorite Video Games', 'What are your favorite video games of all time? Mine has to be The Legend of Zelda: Breath of the Wild.', (SELECT id FROM users WHERE username = 'charlie'), NULL, datetime('now', '-3 days'), datetime('now', '-3 days')),
('750e8400-e29b-41d4-a716-446655440004', 'AI and Machine Learning Trends', 'The field of AI is evolving rapidly. What trends are you most excited about?', (SELECT id FROM users WHERE username = 'alice'), NULL, datetime('now', '-2 days'), datetime('now', '-2 days')),
('750e8400-e29b-41d4-a716-446655440005', 'Latest Movie Recommendations', 'Just watched an amazing sci-fi movie. What have you been watching lately?', (SELECT id FROM users WHERE username = 'bob'), NULL, datetime('now', '-1 day'), datetime('now', '-1 day')),
('750e8400-e29b-41d4-a716-446655440006', 'Healthy Eating Tips', 'Share your favorite healthy recipes and eating habits!', (SELECT id FROM users WHERE username = 'diana'), NULL, datetime('now', '-6 days'), datetime('now', '-6 days')),
('750e8400-e29b-41d4-a716-446655440007', 'Online Learning Platforms', 'Which online learning platforms do you recommend for skill development?', (SELECT id FROM users WHERE username = 'eve'), NULL, datetime('now', '-4 days'), datetime('now', '-4 days'));

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
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440005'), (SELECT id FROM categories WHERE name = 'Entertainment')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440006'), (SELECT id FROM categories WHERE name = 'Health')),
((SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440007'), (SELECT id FROM categories WHERE name = 'Education'));

-- =============================================================================
-- COMMENTS - For comment display tests and reaction tests
-- =============================================================================
INSERT OR IGNORE INTO comments (public_id, post_id, author_id, content, created_at, updated_at) VALUES
('comment-0001-0001-0001-000000000001', (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), (SELECT id FROM users WHERE username = 'alice'), 'Great post! This is really helpful.', datetime('now', '-6 days'), datetime('now', '-6 days')),
('comment-0002-0002-0002-000000000002', (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), (SELECT id FROM users WHERE username = 'bob'), 'I agree with the above comment.', datetime('now', '-5 days'), datetime('now', '-5 days')),
('comment-0003-0003-0003-000000000003', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), (SELECT id FROM users WHERE username = 'alice'), 'Great question — I think Go and Rust are top choices for systems work.', datetime('now', '-4 days'), datetime('now', '-4 days')),
('comment-0004-0004-0004-000000000004', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), (SELECT id FROM users WHERE username = 'bob'), 'I loved Breath of the Wild too! The world design is amazing.', datetime('now', '-2 days'), datetime('now', '-2 days')),
('comment-0005-0005-0005-000000000005', (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440003'), (SELECT id FROM users WHERE username = 'diana'), 'You should try Hades if you like roguelikes.', datetime('now', '-1 day'), datetime('now', '-1 day'));

-- =============================================================================
-- REACTIONS - For like/dislike count display tests
-- =============================================================================
INSERT OR IGNORE INTO reactions (public_id, user_id, target_id, target_type, type, created_at) VALUES
-- Reactions on testuser's posts (for "see liked posts" test)
('reaction-0001-0001-0001-000000000001', (SELECT id FROM users WHERE username = 'testuser'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440001'), 'post', 'like', datetime('now', '-6 days')),
('reaction-0002-0002-0002-000000000002', (SELECT id FROM users WHERE username = 'testuser'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'post', 'like', datetime('now', '-4 days')),
-- Reactions from other users
('reaction-0003-0003-0003-000000000003', (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), 'post', 'like', datetime('now', '-5 days')),
('reaction-0004-0004-0004-000000000004', (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), 'post', 'like', datetime('now', '-4 days')),
('reaction-0005-0005-0005-000000000005', (SELECT id FROM users WHERE username = 'charlie'), (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), 'post', 'dislike', datetime('now', '-3 days')),
-- Comment reactions for comment like/dislike visibility tests
('reaction-0006-0006-0006-000000000006', (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM comments WHERE public_id = 'comment-0003-0003-0003-000000000003'), 'comment', 'like', datetime('now', '-3 days')),
('reaction-0007-0007-0007-000000000007', (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM comments WHERE public_id = 'comment-0003-0003-0003-000000000003'), 'comment', 'dislike', datetime('now', '-2 days'));

-- =============================================================================
-- NOTIFICATIONS - For notification tests
-- =============================================================================
INSERT OR IGNORE INTO notifications (public_id, user_id, actor_id, target_id, type, message, read, created_at) VALUES
('notif-0001-0001-0001-000000000001', (SELECT id FROM users WHERE username = 'testuser'), (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), 'comment', 'alice commented on your post', 0, datetime('now', '-6 days')),
('notif-0002-0002-0002-000000000002', (SELECT id FROM users WHERE username = 'testuser'), (SELECT id FROM users WHERE username = 'bob'), (SELECT id FROM posts WHERE public_id = 'post-0001-0001-0001-000000000001'), 'reaction', 'bob liked your post', 1, datetime('now', '-4 days'));

-- =============================================================================
-- REPORTS - For moderation tests (optional feature)
-- =============================================================================
INSERT OR IGNORE INTO reports (public_id, reporter_id, moderator_id, target_id, target_type, reason, status, response, created_at, reviewed_at) VALUES
('report-0001-0001-0001-000000000001', (SELECT id FROM users WHERE username = 'alice'), (SELECT id FROM users WHERE username = 'eve'), (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440001'), 'post', 'Test report', 'resolved', 'Report resolved', datetime('now', '-5 days'), datetime('now', '-4 days')),
('report-0002-0002-0002-000000000002', (SELECT id FROM users WHERE username = 'bob'), NULL, (SELECT id FROM posts WHERE public_id = '750e8400-e29b-41d4-a716-446655440002'), 'post', 'Another test report', 'pending', NULL, datetime('now', '-3 days'), NULL);

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
