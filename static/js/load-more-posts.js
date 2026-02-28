// Load More Posts functionality for the forum homepage
'use strict';

// Initialize as soon as possible - handle both cases where DOM is already ready or not
(function() {
    function init() {
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
        const dislikedPosts = sourceBtn ? sourceBtn.getAttribute('data-disliked-posts') === 'true' : false;
        const commentedPosts = sourceBtn ? sourceBtn.getAttribute('data-commented-posts') === 'true' : false;
        const activityType = sourceBtn ? sourceBtn.getAttribute('data-activity-type') || 'all' : 'all';
        const reactionType = sourceBtn ? sourceBtn.getAttribute('data-reaction-type') || 'all' : 'all';
        const commenter = sourceBtn ? sourceBtn.getAttribute('data-commenter') || '' : '';
        const dateFilter = sourceBtn ? sourceBtn.getAttribute('data-date-filter') || '' : '';

        // Detect whether we're on the home page (compact layout)
        const additionalPostsContainer = document.getElementById('additional-posts-container');
        const isHomeCompact = !!additionalPostsContainer;

        function createPostElement(post, compact) {
            const el = document.createElement('article');
            
            // SECURITY: Escape all user-generated content to prevent XSS
            const safePostId = window.escapeHtml(post.PublicID);
            const safeTitle = window.escapeHtml(post.Title);
            const safeAuthor = window.escapeHtml(post.AuthorUsername);
            const safeContent = window.escapeHtml(post.Content);
            // ImageURL is server-generated, but escape for safety
            const safeImageURL = post.ImageURL ? window.escapeHtml(post.ImageURL) : '';
            
            const postDate = new Date(post.CreatedAt);
            const formattedDate = postDate.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
            
            // Safely render categories
            const categoriesHtml = (post.Categories || []).map(cat => {
                const safeCat = window.escapeHtml(cat);
                if (compact) {
                    return `<a class="category-tag-compact" href="?category=${encodeURIComponent(cat)}">${safeCat}</a>`;
                }
                return `<a class="category-tag" href="/board?category=${encodeURIComponent(cat)}">${safeCat}</a>`;
            }).join('');
            
            // Parse numeric values safely
            const likeCount = parseInt(post.LikeCount, 10) || 0;
            const dislikeCount = parseInt(post.DislikeCount, 10) || 0;
            const commentCount = parseInt(post.CommentCount, 10) || 0;
            
            if (compact) {
                el.className = 'post-card-compact clickable-card';
                el.setAttribute('data-href', `/posts/${safePostId}`);
                el.innerHTML = `
                    <div class="post-header-compact">
                        <h3><a href="/posts/${safePostId}">${safeTitle}</a></h3>
                        <div class="post-meta-compact">
                            <span class="author-compact">by ${safeAuthor}</span>
                            <span class="date-compact">${formattedDate}</span>
                        </div>
                    </div>

                    ${safeImageURL ? `
                    <div class="post-image-compact">
                        <img src="${safeImageURL}" alt="${safeTitle}">
                    </div>
                    ` : ''}

                    <div class="post-content-compact">
                        <p>${safeContent}</p>
                    </div>

                    <div class="post-footer-compact">
                        <div class="categories-compact">
                            ${categoriesHtml}
                        </div>

                        <div class="post-actions-compact">
                            <button class="btn-like" data-post-id="${safePostId}" aria-label="Like this post" title="Like">👍 ${likeCount}</button>
                            <button class="btn-dislike" data-post-id="${safePostId}" aria-label="Dislike this post" title="Dislike">👎 ${dislikeCount}</button>
                            <span class="comments-compact">💬 ${commentCount}</span>
                        </div>
                    </div>
                `;
            } else {
                el.className = 'post-card clickable-card';
                el.setAttribute('data-href', `/posts/${safePostId}`);
                el.innerHTML = `
                    <div class="post-header">
                        <h3><a href="/posts/${safePostId}">${safeTitle}</a></h3>
                        <div class="post-meta">
                            <span class="author">by ${safeAuthor}</span>
                            <span class="date">${formattedDate}</span>
                        </div>
                    </div>

                    ${safeImageURL ? `
                    <div class="post-image">
                        <img src="${safeImageURL}" alt="${safeTitle}">
                    </div>
                    ` : ''}

                    <div class="post-content">
                        <p>${safeContent}</p>
                    </div>

                    <div class="post-footer">
                        <div class="categories">
                            ${categoriesHtml}
                        </div>

                        <div class="post-actions">
                            <button class="btn-like" data-post-id="${safePostId}" aria-label="Like this post" title="Like">👍 ${likeCount}</button>
                            <button class="btn-dislike" data-post-id="${safePostId}" aria-label="Dislike this post" title="Dislike">👎 ${dislikeCount}</button>
                            <span class="comments">💬 ${commentCount}</span>
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
                if (activityType && activityType !== 'all') params.append('activity_type', activityType);
                if (reactionType && reactionType !== 'all') params.append('reaction_type', reactionType);
                if (myPosts) params.append('my_posts', 'true');
                if (likedPosts) params.append('liked_posts', 'true');
                if (dislikedPosts) params.append('disliked_posts', 'true');
                if (commentedPosts) params.append('commented_posts', 'true');
                if (commenter) params.append('commenter', commenter);
                if (dateFilter) params.append('date_filter', dateFilter);

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
                const pageErrors = document.getElementById('page-errors');
                if (pageErrors) {
                    pageErrors.innerHTML = '<p class="error">Failed to load more posts. Please try again.</p>';
                    pageErrors.scrollIntoView({ behavior: 'smooth', block: 'center' });
                }
            }
        }

        // Map any load-all button to load the next batch (do not auto-load all)
        if (loadMoreBtn) loadMoreBtn.addEventListener('click', loadMorePosts);
        if (loadAllBtn) loadAllBtn.addEventListener('click', loadMorePosts);
    }

    // Run init when DOM is ready - handle both already-loaded and loading cases
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        // DOM is already ready, run immediately
        init();
    }
})();