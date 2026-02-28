// Load More Comments functionality for the My Comments page
'use strict';

(function() {
    function init() {
        const loadMoreBtn = document.getElementById('load-more-comments-btn');
        if (!loadMoreBtn) return;

        let offset = parseInt(loadMoreBtn.getAttribute('data-offset')) || 0;
        const category = loadMoreBtn.getAttribute('data-category') || '';
        const dateFilter = loadMoreBtn.getAttribute('data-date-filter') || '';
        const BATCH_SIZE = 20;

        function createCommentElement(comment) {
            // Use HTML <template> element for cloning instead of innerHTML
            const template = document.getElementById('comment-list-template');
            if (!template) {
                return createCommentElementFallback(comment);
            }

            const clone = template.content.cloneNode(true);
            const article = clone.querySelector('article');

            const commentDate = new Date(comment.CreatedAt);
            const formattedDate = commentDate.toLocaleDateString('en-US', {
                year: 'numeric',
                month: 'short',
                day: '2-digit',
                hour: 'numeric',
                minute: '2-digit',
                hour12: true
            });

            // SECURITY: Escape all user-generated content
            const safePostTitle = window.escapeHtml(comment.PostTitle);
            const safePostAuthor = window.escapeHtml(comment.PostAuthorUsername);
            const safeAuthor = window.escapeHtml(comment.AuthorUsername);
            const safeContent = window.escapeHtml(comment.Content);
            const safePostId = window.escapeHtml(comment.PostPublicID);
            const safeCommentId = window.escapeHtml(comment.PublicID);

            article.id = `comment-${safeCommentId}`;

            // Fill template fields
            const postLink = clone.querySelector('[data-field="post-link"]');
            postLink.href = `/posts/${safePostId}`;
            postLink.textContent = safePostTitle;

            clone.querySelector('[data-field="post-author"]').textContent = safePostAuthor;
            clone.querySelector('[data-field="author"]').textContent = safeAuthor;
            clone.querySelector('[data-field="date"]').textContent = formattedDate;
            clone.querySelector('[data-field="content"]').textContent = safeContent;

            const likeBtn = clone.querySelector('[data-field="like-btn"]');
            likeBtn.setAttribute('data-comment-id', safeCommentId);
            likeBtn.textContent = `👍 (${parseInt(comment.Likes, 10) || 0})`;

            const dislikeBtn = clone.querySelector('[data-field="dislike-btn"]');
            dislikeBtn.setAttribute('data-comment-id', safeCommentId);
            dislikeBtn.textContent = `👎 (${parseInt(comment.Dislikes, 10) || 0})`;

            const editBtn = clone.querySelector('[data-field="edit-btn"]');
            editBtn.setAttribute('data-comment-id', safeCommentId);

            const deleteBtn = clone.querySelector('[data-field="delete-btn"]');
            deleteBtn.setAttribute('data-comment-id', safeCommentId);

            return article;
        }

        // Fallback when template element is not available
        function createCommentElementFallback(comment) {
            const article = document.createElement('article');
            article.className = 'comment';
            article.id = `comment-${comment.PublicID}`;

            const commentDate = new Date(comment.CreatedAt);
            const formattedDate = commentDate.toLocaleDateString('en-US', {
                year: 'numeric', month: 'short', day: '2-digit',
                hour: 'numeric', minute: '2-digit', hour12: true
            });

            const safePostTitle = window.escapeHtml(comment.PostTitle);
            const safePostAuthor = window.escapeHtml(comment.PostAuthorUsername);
            const safeAuthor = window.escapeHtml(comment.AuthorUsername);
            const safeContent = window.escapeHtml(comment.Content);
            const safePostId = window.escapeHtml(comment.PostPublicID);
            const safeCommentId = window.escapeHtml(comment.PublicID);

            article.innerHTML = `
                <div class="comment-context">
                    <p>On post: <a class="comment-post-link" href="/posts/${safePostId}">${safePostTitle}</a> by <span class="comment-post-author">${safePostAuthor}</span></p>
                </div>
                <div class="comment-header">
                    <span class="comment-author">${safeAuthor}</span>
                    <span class="comment-date">${formattedDate}</span>
                </div>
                <div class="comment-content">${safeContent}</div>
                <div class="comment-actions">
                    <div class="comment-reactions">
                        <button class="btn-like-comment" data-comment-id="${safeCommentId}">👍 (${parseInt(comment.Likes, 10) || 0})</button>
                        <button class="btn-dislike-comment" data-comment-id="${safeCommentId}">👎 (${parseInt(comment.Dislikes, 10) || 0})</button>
                    </div>
                    <div class="comment-owner-actions">
                        <button class="btn btn-secondary btn-edit-comment" data-comment-id="${safeCommentId}">Edit</button>
                        <button class="btn btn-danger btn-delete-comment" data-comment-id="${safeCommentId}">Delete</button>
                    </div>
                </div>
            `;
            return article;
        }

        async function loadMoreComments() {
            loadMoreBtn.textContent = 'Loading...';
            loadMoreBtn.disabled = true;

            try {
                const params = new URLSearchParams();
                params.append('offset', offset);
                params.append('limit', BATCH_SIZE);
                if (category) params.append('category', category);
                if (dateFilter) params.append('date_filter', dateFilter);

                const response = await fetch(`/api/comments/load-more?${params}`);
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }

                const comments = await response.json();
                if (!comments || comments.length === 0) {
                    loadMoreBtn.textContent = 'No more comments';
                    loadMoreBtn.disabled = true;
                    return;
                }

                const commentsList = document.querySelector('.comments-list');
                const loadMoreContainer = document.querySelector('.load-more-container');

                comments.forEach(comment => {
                    const el = createCommentElement(comment);
                    if (loadMoreContainer) {
                        commentsList.insertBefore(el, loadMoreContainer);
                    } else {
                        commentsList.appendChild(el);
                    }
                });

                offset += comments.length;
                loadMoreBtn.textContent = 'Show More';
                loadMoreBtn.disabled = false;

                // If we received less than BATCH_SIZE, there are no more comments
                if (comments.length < BATCH_SIZE) {
                    loadMoreBtn.textContent = 'No more comments';
                    loadMoreBtn.disabled = true;
                }
            } catch (err) {
                console.error('Error loading more comments:', err);
                loadMoreBtn.textContent = 'Show More';
                loadMoreBtn.disabled = false;

                const pageErrors = document.getElementById('page-errors');
                if (pageErrors) {
                    pageErrors.innerHTML = '<p class="error">Failed to load more comments. Please try again.</p>';
                    pageErrors.scrollIntoView({ behavior: 'smooth', block: 'center' });
                }
            }
        }

        loadMoreBtn.addEventListener('click', loadMoreComments);
    }

    // Run init when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
