// Generic reaction handling logic for posts and comments
'use strict';

(function() {
    // Update a button's displayed count by replacing the "(N)" pattern in its text.
    function updateButtonCount(btn, newCount) {
        if (!btn) return;

        const currentText = btn.textContent;
        if (/\(\d+\)/.test(currentText)) {
            btn.textContent = currentText.replace(/\(\d+\)/, `(${newCount})`);
        } else if (/\d+/.test(currentText)) {
            btn.textContent = currentText.replace(/\d+(?!.*\d)/, String(newCount));
        } else {
            btn.textContent = `${currentText} (${newCount})`;
        }

        const ariaLabel = btn.getAttribute('aria-label');
        if (ariaLabel) {
            if (/current count:\s*\d+/i.test(ariaLabel)) {
                btn.setAttribute('aria-label', ariaLabel.replace(/current count:\s*\d+/i, `current count: ${newCount}`));
            } else {
                btn.setAttribute('aria-label', `${ariaLabel}, current count: ${newCount}`);
            }
        }
    }

    // Fetch updated counts from the server and apply them to the like/dislike buttons.
    async function refreshCounts(targetType, targetId, likeBtn, dislikeBtn) {
        try {
            const data = await window.api.request(`/api/reactions/${targetType}/${targetId}/count`);
            if (data && typeof data.likes === 'number') {
                updateButtonCount(likeBtn, data.likes);
            }
            if (data && typeof data.dislikes === 'number') {
                updateButtonCount(dislikeBtn, data.dislikes);
            }
        } catch (err) {
            console.error('Failed to refresh reaction counts:', err);
        }
    }

    function applyCountsFromResponse(response, likeBtn, dislikeBtn) {
        if (!response || typeof response !== 'object') {
            return false;
        }

        const hasLikes = typeof response.likes === 'number';
        const hasDislikes = typeof response.dislikes === 'number';
        if (!hasLikes && !hasDislikes) {
            return false;
        }

        if (hasLikes) {
            updateButtonCount(likeBtn, response.likes);
        }
        if (hasDislikes) {
            updateButtonCount(dislikeBtn, response.dislikes);
        }
        return true;
    }

    // Unified reaction handler for both posts and comments
    async function handleReaction(targetType, targetId, reactionType, likeBtn, dislikeBtn) {
        window.clearError('page-errors');
        try {
            const response = await window.api.request('/api/reactions', {
                method: 'POST',
                body: JSON.stringify({
                    target_type: targetType,
                    target_id: targetId,
                    type: reactionType
                })
            });

            const applied = applyCountsFromResponse(response, likeBtn, dislikeBtn);
            if (!applied) {
                await refreshCounts(targetType, targetId, likeBtn, dislikeBtn);
            }
        } catch (error) {
            if (error && error.status === 401) {
                window.showError(`Please login to react to ${targetType}s`, 'page-errors');
                return;
            }
            console.error(`Reaction error (${targetType}/${reactionType}):`, error);
            window.showError(error.message || `An error occurred while ${reactionType}ing the ${targetType}`, 'page-errors');
        }
    }

    // Convenience wrappers for backward compatibility
    function handlePostReaction(postId, reactionType, likeBtn, dislikeBtn) {
        return handleReaction('post', postId, reactionType, likeBtn, dislikeBtn);
    }

    function handleCommentReaction(commentId, reactionType, likeBtn, dislikeBtn) {
        return handleReaction('comment', commentId, reactionType, likeBtn, dislikeBtn);
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
                if (postId) {
                    const dislikeBtn = document.querySelector(`.btn-dislike[data-post-id="${postId}"]`);
                    await handlePostReaction(postId, 'like', btn, dislikeBtn);
                }
            }

            // Handle post dislikes
            if (btn.classList.contains('btn-dislike')) {
                e.preventDefault();
                e.stopPropagation(); // Prevent card click navigation
                const postId = btn.getAttribute('data-post-id');
                if (postId) {
                    const likeBtn = document.querySelector(`.btn-like[data-post-id="${postId}"]`);
                    await handlePostReaction(postId, 'dislike', likeBtn, btn);
                }
            }

            // Handle comment likes
            if (btn.classList.contains('btn-like-comment')) {
                e.preventDefault();
                e.stopPropagation();
                const commentId = btn.getAttribute('data-comment-id');
                if (commentId) {
                    const dislikeBtn = document.querySelector(`.btn-dislike-comment[data-comment-id="${commentId}"]`);
                    await handleCommentReaction(commentId, 'like', btn, dislikeBtn);
                }
            }

            // Handle comment dislikes
            if (btn.classList.contains('btn-dislike-comment')) {
                e.preventDefault();
                e.stopPropagation();
                const commentId = btn.getAttribute('data-comment-id');
                if (commentId) {
                    const likeBtn = document.querySelector(`.btn-like-comment[data-comment-id="${commentId}"]`);
                    await handleCommentReaction(commentId, 'dislike', likeBtn, btn);
                }
            }
        });
    });

    // Also expose it to window for cases where it's needed globally (though event delegation is preferred)
    window.handlePostReaction = handlePostReaction;
    window.handleCommentReaction = handleCommentReaction;
})();
