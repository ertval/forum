// Generic reaction handling logic for posts and comments
'use strict';

(function() {
    // Update a button's displayed count by replacing the "(N)" pattern in its text.
    function updateButtonCount(btn, newCount) {
        if (!btn) return;
        btn.textContent = btn.textContent.replace(/\(\d+\)/, `(${newCount})`);
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

    // Function to handle post reactions
    async function handlePostReaction(postId, reactionType, likeBtn, dislikeBtn) {
        window.clearError('page-errors');
        try {
            await window.api.request('/api/reactions', {
                method: 'POST',
                body: JSON.stringify({
                    target_type: 'post',
                    target_id: postId,
                    type: reactionType
                })
            });
            await refreshCounts('post', postId, likeBtn, dislikeBtn);
        } catch (error) {
            if (error && error.status === 401) {
                window.showError('Please login to react to posts', 'page-errors');
                return;
            }
            console.error(`Reaction error (${reactionType}):`, error);
            window.showError(error.message || `An error occurred while ${reactionType}ing the post`, 'page-errors');
        }
    }

    // Function to handle comment reactions
    async function handleCommentReaction(commentId, reactionType, likeBtn, dislikeBtn) {
        window.clearError('page-errors');
        try {
            await window.api.request('/api/reactions', {
                method: 'POST',
                body: JSON.stringify({
                    target_type: 'comment',
                    target_id: commentId,
                    type: reactionType
                })
            });
            await refreshCounts('comment', commentId, likeBtn, dislikeBtn);
        } catch (error) {
            if (error && error.status === 401) {
                window.showError('Please login to react to comments', 'page-errors');
                return;
            }
            console.error(`Comment reaction error (${reactionType}):`, error);
            window.showError(error.message || `An error occurred while ${reactionType}ing the comment`, 'page-errors');
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
                const commentId = btn.getAttribute('data-comment-id');
                if (commentId) {
                    const dislikeBtn = document.querySelector(`.btn-dislike-comment[data-comment-id="${commentId}"]`);
                    await handleCommentReaction(commentId, 'like', btn, dislikeBtn);
                }
            }

            // Handle comment dislikes
            if (btn.classList.contains('btn-dislike-comment')) {
                e.preventDefault();
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
