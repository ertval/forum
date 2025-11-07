-- Seed data for testing the forum homepage

-- Insert test users
INSERT OR IGNORE INTO users (id, email, username, password, role, created_at, updated_at) VALUES
('user-1', 'alice@example.com', 'alice', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now'), datetime('now')),
('user-2', 'bob@example.com', 'bob', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now'), datetime('now')),
('user-3', 'charlie@example.com', 'charlie', '$2a$10$XqjW4P1YZF5lQ3Y6N0YGYuZdT7ZiYzQW9P7YZF5lQ3Y6N0YGYuZdT', 'user', datetime('now'), datetime('now'));

-- Insert test categories
INSERT OR IGNORE INTO categories (id, name, description, created_at) VALUES
('cat-1', 'General', 'General discussion topics', datetime('now')),
('cat-2', 'Technology', 'Tech-related posts and discussions', datetime('now')),
('cat-3', 'Gaming', 'Video games and gaming culture', datetime('now')),
('cat-4', 'Science', 'Science and research discussions', datetime('now')),
('cat-5', 'Entertainment', 'Movies, TV shows, and entertainment', datetime('now'));

-- Insert test posts
INSERT OR IGNORE INTO posts (id, title, content, author_id, image_path, created_at, updated_at) VALUES
('post-1', 'Welcome to the Forum!', 'This is the first post on our new forum. Feel free to create your own posts and join discussions!', 'user-1', NULL, datetime('now', '-7 days'), datetime('now', '-7 days')),
('post-2', 'Best Programming Languages in 2025', 'What do you think are the best programming languages to learn in 2025? I am currently learning Go and loving it!', 'user-2', NULL, datetime('now', '-5 days'), datetime('now', '-5 days')),
('post-3', 'Favorite Video Games', 'What are your favorite video games of all time? Mine has to be The Legend of Zelda: Breath of the Wild.', 'user-3', NULL, datetime('now', '-3 days'), datetime('now', '-3 days')),
('post-4', 'AI and Machine Learning Trends', 'The field of AI is evolving rapidly. What trends are you most excited about?', 'user-1', NULL, datetime('now', '-2 days'), datetime('now', '-2 days')),
('post-5', 'Latest Movie Recommendations', 'Just watched an amazing sci-fi movie. What have you been watching lately?', 'user-2', NULL, datetime('now', '-1 day'), datetime('now', '-1 day')),
('post-6', 'Climate Change Research', 'Recent studies show significant progress in renewable energy. Lets discuss!', 'user-3', NULL, datetime('now', '-12 hours'), datetime('now', '-12 hours'));

-- Associate posts with categories
INSERT OR IGNORE INTO post_categories (post_id, category_id) VALUES
('post-1', 'cat-1'),
('post-2', 'cat-2'),
('post-3', 'cat-3'),
('post-4', 'cat-2'),
('post-5', 'cat-5'),
('post-6', 'cat-4');

-- Insert some test reactions
INSERT OR IGNORE INTO reactions (id, user_id, target_id, target_type, type, created_at) VALUES
('react-1', 'user-2', 'post-1', 'post', 'like', datetime('now', '-6 days')),
('react-2', 'user-3', 'post-1', 'post', 'like', datetime('now', '-6 days')),
('react-3', 'user-1', 'post-2', 'post', 'like', datetime('now', '-4 days')),
('react-4', 'user-3', 'post-2', 'post', 'like', datetime('now', '-4 days')),
('react-5', 'user-1', 'post-3', 'post', 'like', datetime('now', '-2 days')),
('react-6', 'user-2', 'post-3', 'post', 'dislike', datetime('now', '-2 days')),
('react-7', 'user-2', 'post-4', 'post', 'like', datetime('now', '-1 day')),
('react-8', 'user-3', 'post-5', 'post', 'like', datetime('now', '-12 hours'));

-- Insert some test comments
INSERT OR IGNORE INTO comments (id, post_id, author_id, content, created_at, updated_at) VALUES
('comment-1', 'post-1', 'user-2', 'Great to be here!', datetime('now', '-6 days'), datetime('now', '-6 days')),
('comment-2', 'post-1', 'user-3', 'Looking forward to interesting discussions.', datetime('now', '-6 days'), datetime('now', '-6 days')),
('comment-3', 'post-2', 'user-1', 'I agree, Go is fantastic!', datetime('now', '-4 days'), datetime('now', '-4 days')),
('comment-4', 'post-2', 'user-3', 'Python is still my favorite though.', datetime('now', '-4 days'), datetime('now', '-4 days')),
('comment-5', 'post-3', 'user-1', 'BOTW is amazing! Have you tried Tears of the Kingdom?', datetime('now', '-2 days'), datetime('now', '-2 days'));
