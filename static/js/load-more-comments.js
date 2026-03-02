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

            const contextDiv = document.createElement('div');
            contextDiv.className = 'comment-context';
            const contextP = document.createElement('p');
            contextP.append('On post: ');
            const postLink = document.createElement('a');
            postLink.className = 'comment-post-link';
            postLink.href = `/posts/${safePostId}`;
            postLink.textContent = safePostTitle;
            contextP.appendChild(postLink);
            contextP.append(' by ');
            const postAuthor = document.createElement('span');
            postAuthor.className = 'comment-post-author';
            postAuthor.textContent = safePostAuthor;
            contextP.appendChild(postAuthor);
            contextDiv.appendChild(contextP);

            const headerDiv = document.createElement('div');
            headerDiv.className = 'comment-header';
            const authorSpan = document.createElement('span');
            authorSpan.className = 'comment-author';
            authorSpan.textContent = safeAuthor;
            const dateSpan = document.createElement('span');
            dateSpan.className = 'comment-date';
            dateSpan.textContent = formattedDate;
            headerDiv.appendChild(authorSpan);
            headerDiv.appendChild(dateSpan);

            const contentDiv = document.createElement('div');
            contentDiv.className = 'comment-content';
            contentDiv.textContent = safeContent;

            const actionsDiv = document.createElement('div');
            actionsDiv.className = 'comment-actions';
            const reactionsDiv = document.createElement('div');
            reactionsDiv.className = 'comment-reactions';

            const likeBtn = document.createElement('button');
            likeBtn.className = 'btn-like-comment';
            likeBtn.setAttribute('data-comment-id', safeCommentId);
            likeBtn.textContent = `👍 (${parseInt(comment.Likes, 10) || 0})`;

            const dislikeBtn = document.createElement('button');
            dislikeBtn.className = 'btn-dislike-comment';
            dislikeBtn.setAttribute('data-comment-id', safeCommentId);
            dislikeBtn.textContent = `👎 (${parseInt(comment.Dislikes, 10) || 0})`;

            reactionsDiv.appendChild(likeBtn);
            reactionsDiv.appendChild(dislikeBtn);

            const ownerActionsDiv = document.createElement('div');
            ownerActionsDiv.className = 'comment-owner-actions';

            const editBtn = document.createElement('button');
            editBtn.className = 'btn btn-secondary btn-edit-comment';
            editBtn.setAttribute('data-comment-id', safeCommentId);
            editBtn.textContent = 'Edit';

            const deleteBtn = document.createElement('button');
            deleteBtn.className = 'btn btn-danger btn-delete-comment';
            deleteBtn.setAttribute('data-comment-id', safeCommentId);
            deleteBtn.textContent = 'Delete';

            ownerActionsDiv.appendChild(editBtn);
            ownerActionsDiv.appendChild(deleteBtn);

            actionsDiv.appendChild(reactionsDiv);
            actionsDiv.appendChild(ownerActionsDiv);

            article.appendChild(contextDiv);
            article.appendChild(headerDiv);
            article.appendChild(contentDiv);
            article.appendChild(actionsDiv);
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

                const comments = await window.api.request(`/api/comments/load-more?${params}`);
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
