// Load All Posts functionality for the home page

document.addEventListener('DOMContentLoaded', function() {
    // Check if Load All Posts button exists
    const loadAllBtn = document.getElementById('load-all-posts-btn');

    if (!loadAllBtn) return;

    // Get initial data from data attributes on the button
    let offset = parseInt(loadAllBtn.getAttribute('data-offset')) || 0; // Start after the already displayed posts
    const category = loadAllBtn.getAttribute('data-category') || '';
    const myPosts = loadAllBtn.getAttribute('data-my-posts') === 'true';
    const likedPosts = loadAllBtn.getAttribute('data-liked-posts') === 'true';

    // Function to load all posts by repeatedly calling the load-more endpoint
    async function loadAllPosts() {
        // Show loading state
        loadAllBtn.textContent = 'Loading All...';
        loadAllBtn.disabled = true;

        // Start with the current offset (how many posts are already displayed)
        let currentOffset = offset;

        try {
            // Keep loading more posts until no more are available
            let hasMorePosts = true;
            while (hasMorePosts) {
                // Build query string with current filters
                const params = new URLSearchParams();
                params.append('offset', currentOffset);
                params.append('limit', 50); // Load in batches of 50

                if (category) {
                    params.append('category', category);
                }

                if (myPosts) {
                    params.append('my_posts', 'true');
                }

                if (likedPosts) {
                    params.append('liked_posts', 'true');
                }

                const response = await fetch(`/api/posts/load-more?${params}`);

                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }

                const posts = await response.json();

                // If no more posts, we're done
                if (posts.length === 0) {
                    hasMorePosts = false;
                    break;
                }

                // Get the posts container and append new posts in compact format
                const postsContainer = document.getElementById('additional-posts-container');

                posts.forEach(post => {
                    const postElement = document.createElement('article');
                    postElement.className = 'post-card-compact';

                    // Format date
                    const postDate = new Date(post.CreatedAt);
                    const formattedDate = postDate.toLocaleDateString('en-US', {
                        year: 'numeric',
                        month: 'short',
                        day: 'numeric'
                    });

                    postElement.innerHTML = `
                        <div class="post-header-compact">
                            <h3><a href="/posts/${post.ID}">${post.Title}</a></h3>
                            <div class="post-meta-compact">
                                <span class="author-compact">by ${post.AuthorUsername}</span>
                                <span class="date-compact">${formattedDate}</span>
                            </div>
                        </div>

                        ${post.ImageURL ? `
                        <div class="post-image-compact">
                            <img src="${post.ImageURL}" alt="${post.Title}">
                        </div>
                        ` : ''}

                        <div class="post-content-compact">
                            <p>${post.Content}</p>
                        </div>

                        <div class="post-footer-compact">
                            <div class="categories-compact">
                                ${post.Categories.map(cat => `
                                    <a class="category-tag-compact" href="/board?category=${encodeURIComponent(cat)}">${cat}</a>
                                `).join('')}
                            </div>

                            <div class="post-actions-compact">
                                <span class="likes-compact">👍 ${post.LikeCount}</span>
                                <span class="dislikes-compact">👎 ${post.DislikeCount}</span>
                                <span class="comments-compact">💬 ${post.CommentCount}</span>
                            </div>
                        </div>
                    `;

                    // Append the new post to the additional posts container
                    postsContainer.appendChild(postElement);
                });

                // Update offset for next request
                currentOffset += posts.length;

                // Brief pause to be respectful to the server
                await new Promise(resolve => setTimeout(resolve, 100));
            }

            // Update button text when all posts are loaded
            loadAllBtn.textContent = 'All Posts Loaded';
            loadAllBtn.disabled = true;

        } catch (error) {
            console.error('Error loading all posts:', error);
            loadAllBtn.textContent = 'View All Posts';
            loadAllBtn.disabled = false;
            alert('Failed to load all posts. Please try again.');
        }
    }

    // Add click event listener to Load All button
    loadAllBtn.addEventListener('click', loadAllPosts);
});