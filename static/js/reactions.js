// Generic reaction handling logic for posts and comments
'use strict';

(function() {
    // Helper function to show inline error messages
    function showPageError(message) {
        const pageErrors = document.getElementById('page-errors');
        if (pageErrors) {
            // SECURITY: Escape message to prevent XSS from reflected error content
            pageErrors.innerHTML = `<p class="error">${window.escapeHtml(message)}</p>`;
            pageErrors.scrollIntoView({ behavior: 'smooth', block: 'center' });
        } else {
            // Fallback for pages without page-errors div
            alert(message);
        }
    }

    function clearPageError() {
        const pageErrors = document.getElementById('page-errors');
        if (pageErrors) pageErrors.innerHTML = '';
    }

    // Function to handle post reactions
    async function handlePostReaction(postId, reactionType) {
        clearPageError();
        try {
            const response = await fetch('/api/reactions', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    target_type: 'post',
                    target_id: postId,
                    type: reactionType
                }),
                credentials: 'include'
            });

            if (response.ok) {
                location.reload(); // Reload to get updated counts
            } else {
                if (response.status === 401) {
                    showPageError('Please login to react to posts');
                } else {
                    const error = await response.json();
                    showPageError(error.error || `Failed to ${reactionType} post`);
                }
            }
        } catch (error) {
            console.error(`Reaction error (${reactionType}):`, error);
            showPageError(`An error occurred while ${reactionType}ing the post`);
        }
    }

    // Function to handle comment reactions
    async function handleCommentReaction(commentId, reactionType) {
        clearPageError();
        try {
            const response = await fetch('/api/reactions', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    target_type: 'comment',
                    target_id: commentId,
                    type: reactionType
                }),
                credentials: 'include'
            });

            if (response.ok) {
                location.reload(); // Reload to get updated counts
            } else {
                if (response.status === 401) {
                    showPageError('Please login to react to comments');
                } else {
                    const error = await response.json();
                    showPageError(error.error || `Failed to ${reactionType} comment`);
                }
            }
        } catch (error) {
            console.error(`Comment reaction error (${reactionType}):`, error);
            showPageError(`An error occurred while ${reactionType}ing the comment`);
        }
    }

    // Event delegation for reaction buttons
    document.addEventListener('DOMContentLoaded', function() {
        document.body.addEventListener('click', async function(e) {
            // Find closest button with reaction class in case click was on emoji/text
            const btn = e.target.closest('button');
            if (!btn) return;

            // Handle post likes
            if (btn.classList.contains('btn-like')) {
                e.preventDefault();
                e.stopPropagation(); // Prevent card click navigation
                const postId = btn.getAttribute('data-post-id');
                if (postId) await handlePostReaction(postId, 'like');
            }

            // Handle post dislikes
            if (btn.classList.contains('btn-dislike')) {
                e.preventDefault();
                e.stopPropagation(); // Prevent card click navigation
                const postId = btn.getAttribute('data-post-id');
                if (postId) await handlePostReaction(postId, 'dislike');
            }

            // Handle comment likes
            if (btn.classList.contains('btn-like-comment')) {
                e.preventDefault();
                const commentId = btn.getAttribute('data-comment-id');
                if (commentId) await handleCommentReaction(commentId, 'like');
            }

            // Handle comment dislikes
            if (btn.classList.contains('btn-dislike-comment')) {
                e.preventDefault();
                const commentId = btn.getAttribute('data-comment-id');
                if (commentId) await handleCommentReaction(commentId, 'dislike');
            }
        });
    });

    // Also expose it to window for cases where it's needed globally (though event delegation is preferred)
    window.handlePostReaction = handlePostReaction;
    window.handleCommentReaction = handleCommentReaction;
})();
