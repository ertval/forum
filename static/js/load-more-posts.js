// Load More Posts functionality for the forum homepage

document.addEventListener('DOMContentLoaded', function() {
    // Check if Load More button exists
    const loadMoreBtn = document.getElementById('load-more-btn');
    const loadAllBtn = document.getElementById('load-all-btn');

    if (!loadMoreBtn && !loadAllBtn) return;

    // Get initial data from data attributes on the buttons
    let offset = loadMoreBtn ? parseInt(loadMoreBtn.getAttribute('data-offset')) || 0 : 0; // Start after the already displayed posts
    const category = (loadMoreBtn || loadAllBtn).getAttribute('data-category') || '';
    const myPosts = (loadMoreBtn || loadAllBtn).getAttribute('data-my-posts') === 'true';
    const likedPosts = (loadMoreBtn || loadAllBtn).getAttribute('data-liked-posts') === 'true';

    // If we're on a filtered page and no posts were initially loaded,
    // we still want to start at offset 0
    if (offset === 0 && (category || myPosts || likedPosts)) {
        // For filtered views with no initial posts, offset should remain 0
    }

    // Function to load more posts
    async function loadMorePosts() {
        // Show loading state
        loadMoreBtn.textContent = 'Loading...';
        loadMoreBtn.disabled = true;

        try {
            // Build query string with current filters
            const params = new URLSearchParams();
            params.append('offset', offset);

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

            // If no more posts, disable the button
            if (posts.length === 0) {
                loadMoreBtn.textContent = 'No more posts';
                loadMoreBtn.disabled = true;
                return;
            }

            // Get the posts container and append new posts
            const postsContainer = document.querySelector('.posts');

            posts.forEach(post => {
                const postElement = document.createElement('article');
                postElement.className = 'post-card';

                // Format date
                const postDate = new Date(post.CreatedAt);
                const formattedDate = postDate.toLocaleDateString('en-US', {
                    year: 'numeric',
                    month: 'short',
                    day: 'numeric'
                });

                postElement.innerHTML = `
                    <div class="post-header">
                        <h3><a href="/posts/${post.ID}">${post.Title}</a></h3>
                        <div class="post-meta">
                            <span class="author">by ${post.AuthorUsername}</span>
                            <span class="date" style="margin-left: auto;">${formattedDate}</span>
                        </div>
                    </div>

                    ${post.ImageURL ? `
                    <div class="post-image">
                        <img src="${post.ImageURL}" alt="${post.Title}">
                    </div>
                    ` : ''}

                    <div class="post-content">
                        <p>${post.Content}</p>
                    </div>

                    <div class="post-footer">
                        <div class="categories">
                            ${post.Categories.map(cat => `
                                <a class="category-tag" href="/?category=${encodeURIComponent(cat)}">${cat}</a>
                            `).join('')}
                        </div>

                        <div class="post-actions">
                            <span class="likes">👍 ${post.LikeCount}</span>
                            <span class="dislikes">👎 ${post.DislikeCount}</span>
                            <span class="comments">💬 ${post.CommentCount}</span>
                        </div>
                    </div>
                `;

                // Append the new post to the posts container after the existing posts
                postsContainer.insertBefore(postElement, document.querySelector('.load-more-container'));
            });

            // Update offset for next request
            offset += posts.length;

            // Reset button text
            loadMoreBtn.textContent = 'Load More';
            loadMoreBtn.disabled = false;

        } catch (error) {
            console.error('Error loading more posts:', error);
            loadMoreBtn.textContent = 'Load More';
            loadMoreBtn.disabled = false;
            alert('Failed to load more posts. Please try again.');
        }
    }

    // Function to load all posts by repeatedly calling the load-more endpoint
    async function loadAllPosts() {
        // Show loading state
        loadAllBtn.textContent = 'Loading All...';
        loadAllBtn.disabled = true;
        if (loadMoreBtn) {
            loadMoreBtn.disabled = true;
        }

        // Start with the current offset (how many posts are already displayed)
        let currentOffset = parseInt(loadMoreBtn.getAttribute('data-offset')) || document.querySelectorAll('.posts .post-card').length;

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

                // Get the posts container and append new posts
                const postsContainer = document.querySelector('.posts');

                posts.forEach(post => {
                    const postElement = document.createElement('article');
                    postElement.className = 'post-card';

                    // Format date
                    const postDate = new Date(post.CreatedAt);
                    const formattedDate = postDate.toLocaleDateString('en-US', {
                        year: 'numeric',
                        month: 'short',
                        day: 'numeric'
                    });

                    postElement.innerHTML = `
                        <div class="post-header">
                            <h3><a href="/posts/${post.ID}">${post.Title}</a></h3>
                            <div class="post-meta">
                                <span class="author">by ${post.AuthorUsername}</span>
                                <span class="date" style="margin-left: auto;">${formattedDate}</span>
                            </div>
                        </div>

                        ${post.ImageURL ? `
                        <div class="post-image">
                            <img src="${post.ImageURL}" alt="${post.Title}">
                        </div>
                        ` : ''}

                        <div class="post-content">
                            <p>${post.Content}</p>
                        </div>

                        <div class="post-footer">
                            <div class="categories">
                                ${post.Categories.map(cat => `
                                    <a class="category-tag" href="/?category=${encodeURIComponent(cat)}">${cat}</a>
                                `).join('')}
                            </div>

                            <div class="post-actions">
                                <span class="likes">👍 ${post.LikeCount}</span>
                                <span class="dislikes">👎 ${post.DislikeCount}</span>
                                <span class="comments">💬 ${post.CommentCount}</span>
                            </div>
                        </div>
                    `;

                    // Append the new post to the posts container after the existing posts
                    postsContainer.insertBefore(postElement, document.querySelector('.load-more-container'));
                });

                // Update offset for next request
                currentOffset += posts.length;

                // Brief pause to be respectful to the server
                await new Promise(resolve => setTimeout(resolve, 100));
            }

            // Optionally hide the load more button after loading all
            if (loadMoreBtn) {
                loadMoreBtn.style.display = 'none';
            }
            loadAllBtn.textContent = 'All Loaded';
            loadAllBtn.disabled = true;

        } catch (error) {
            console.error('Error loading all posts:', error);
            loadAllBtn.textContent = 'Load All';
            loadAllBtn.disabled = false;
            if (loadMoreBtn) {
                loadMoreBtn.disabled = false;
            }
            alert('Failed to load all posts. Please try again.');
        }
    }

            // Optionally hide the load more button after loading all
            if (loadMoreBtn) {
                loadMoreBtn.style.display = 'none';
            }
            loadAllBtn.textContent = 'All Loaded';
            loadAllBtn.disabled = true;

        } catch (error) {
            console.error('Error loading all posts:', error);
            loadAllBtn.textContent = 'Load All';
            loadAllBtn.disabled = false;
            if (loadMoreBtn) {
                loadMoreBtn.disabled = false;
            }
            alert('Failed to load all posts. Please try again.');
        }
    }

    // Add click event listener to Load More button
    if (loadMoreBtn) {
        loadMoreBtn.addEventListener('click', loadMorePosts);
    }

    // Add click event listener to Load All button
    if (loadAllBtn) {
        loadAllBtn.addEventListener('click', loadAllPosts);
    }
});