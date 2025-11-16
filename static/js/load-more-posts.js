// Load More Posts functionality for the forum homepage

document.addEventListener('DOMContentLoaded', function() {
    // Support both board and home buttons
    const loadMoreBtn = document.getElementById('load-more-btn');
    // board uses id 'load-all-btn', home uses 'load-all-posts-btn'
    const loadAllBtn = document.getElementById('load-all-btn') || document.getElementById('load-all-posts-btn');

    if (!loadMoreBtn && !loadAllBtn) return;

    // Helper to read attributes from whichever button exists
    const sourceBtn = loadMoreBtn || loadAllBtn;
    let offset = sourceBtn ? parseInt(sourceBtn.getAttribute('data-offset')) || 0 : 0;
    const category = sourceBtn ? sourceBtn.getAttribute('data-category') || '' : '';
    const myPosts = sourceBtn ? sourceBtn.getAttribute('data-my-posts') === 'true' : false;
    const likedPosts = sourceBtn ? sourceBtn.getAttribute('data-liked-posts') === 'true' : false;

    // Detect whether we're on the home page (compact layout)
    const additionalPostsContainer = document.getElementById('additional-posts-container');
    const isHomeCompact = !!additionalPostsContainer;

    function createPostElement(post, compact) {
        const el = document.createElement('article');
        if (compact) {
            el.className = 'post-card-compact';
            const postDate = new Date(post.CreatedAt);
            const formattedDate = postDate.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
            el.innerHTML = `
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
                        ${post.Categories.map(cat => `<a class="category-tag-compact" href="?category=${encodeURIComponent(cat)}">${cat}</a>`).join('')}
                    </div>

                    <div class="post-actions-compact">
                        <span class="likes-compact">👍 ${post.LikeCount}</span>
                        <span class="dislikes-compact">👎 ${post.DislikeCount}</span>
                        <span class="comments-compact">💬 ${post.CommentCount}</span>
                    </div>
                </div>
            `;
        } else {
            el.className = 'post-card';
            const postDate = new Date(post.CreatedAt);
            const formattedDate = postDate.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
            el.innerHTML = `
                <div class="post-header">
                    <h3><a href="/posts/${post.ID}">${post.Title}</a></h3>
                    <div class="post-meta">
                        <span class="author">by ${post.AuthorUsername}</span>
                        <span class="date">${formattedDate}</span>
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
                        ${post.Categories.map(cat => `<a class="category-tag" href="/board?category=${encodeURIComponent(cat)}">${cat}</a>`).join('')}
                    </div>

                    <div class="post-actions">
                        <span class="likes">👍 ${post.LikeCount}</span>
                        <span class="dislikes">👎 ${post.DislikeCount}</span>
                        <span class="comments">💬 ${post.CommentCount}</span>
                    </div>
                </div>
            `;
        }
        return el;
    }

    async function fetchPosts(params) {
        const response = await fetch(`/api/posts/load-more?${params}`);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        return await response.json();
    }

    const BATCH_SIZE = 20; // load 20 posts per click

    // Load more (single page load)
    async function loadMorePosts() {
        if (!loadMoreBtn) return;
        loadMoreBtn.textContent = 'Loading...';
        loadMoreBtn.disabled = true;

        try {
            const params = new URLSearchParams();
            params.append('offset', offset);
            params.append('limit', BATCH_SIZE);
            if (category) params.append('category', category);
            if (myPosts) params.append('my_posts', 'true');
            if (likedPosts) params.append('liked_posts', 'true');

            const posts = await fetchPosts(params);
            if (!posts || posts.length === 0) {
                // No more posts: keep button visible at end but disable it and update text
                loadMoreBtn.textContent = 'No more posts';
                loadMoreBtn.disabled = true;
                return;
            }

            const postsContainer = document.querySelector('.posts');
            posts.forEach(post => {
                const el = createPostElement(post, isHomeCompact);
                if (isHomeCompact && additionalPostsContainer) {
                    additionalPostsContainer.appendChild(el);
                } else if (postsContainer) {
                    const loadMoreContainer = document.querySelector('.load-more-container');
                    postsContainer.insertBefore(el, loadMoreContainer);
                }
            });

            offset += posts.length;
            loadMoreBtn.textContent = 'Show More';
            loadMoreBtn.disabled = false;
        } catch (err) {
            console.error('Error loading more posts:', err);
            loadMoreBtn.textContent = 'Show More';
            loadMoreBtn.disabled = false;
            alert('Failed to load more posts. Please try again.');
        }
    }

    // Map any load-all button to load the next batch (do not auto-load all)
    if (loadMoreBtn) loadMoreBtn.addEventListener('click', loadMorePosts);
    if (loadAllBtn) loadAllBtn.addEventListener('click', loadMorePosts);
});