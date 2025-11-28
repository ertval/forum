// JavaScript functions for post detail page

// Helper function to show inline error messages
function showPageError(message) {
    const pageErrors = document.getElementById('page-errors');
    if (pageErrors) {
        pageErrors.innerHTML = `<p class="error">${message}</p>`;
        pageErrors.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
}

function clearPageError() {
    const pageErrors = document.getElementById('page-errors');
    if (pageErrors) pageErrors.innerHTML = '';
}

document.addEventListener('DOMContentLoaded', function() {
    // Handle post reactions
    document.body.addEventListener('click', async function(e) {
        // Handle post likes
        if (e.target.classList.contains('btn-like')) {
            e.preventDefault();
            const postId = e.target.getAttribute('data-post-id');
            await handlePostReaction(postId, 'like');
        }

        // Handle post dislikes
        if (e.target.classList.contains('btn-dislike')) {
            e.preventDefault();
            const postId = e.target.getAttribute('data-post-id');
            await handlePostReaction(postId, 'dislike');
        }

        // Handle comment likes
        if (e.target.classList.contains('btn-like-comment')) {
            e.preventDefault();
            const commentId = e.target.getAttribute('data-comment-id');
            await handleCommentReaction(commentId, 'like');
        }

        // Handle comment dislikes
        if (e.target.classList.contains('btn-dislike-comment')) {
            e.preventDefault();
            const commentId = e.target.getAttribute('data-comment-id');
            await handleCommentReaction(commentId, 'dislike');
        }

        // Handle comment deletion
        if (e.target.classList.contains('btn-delete-comment')) {
            e.preventDefault();
            const commentId = e.target.getAttribute('data-comment-id');
            await deleteComment(commentId);
        }
    });

    // Handle comment form submission
    const commentForm = document.getElementById('comment-form');
    if (commentForm) {
        commentForm.addEventListener('submit', async function(e) {
            e.preventDefault();
            clearPageError();

            const postId = commentForm.getAttribute('data-post-id');
            const content = commentForm.querySelector('textarea[name="content"]').value.trim();

            if (!content) {
                showPageError('Comment content is required');
                return;
            }

            try {
                const formData = new FormData();
                formData.append('content', content);

                const response = await fetch(`/posts/${postId}/comments`, {
                    method: 'POST',
                    body: formData
                });

                if (response.ok) {
                    commentForm.reset();
                    // In a real implementation, we would update the UI to show the new comment
                    // without page reload
                    location.reload(); // Simple approach to refresh comments
                } else {
                    const error = await response.json();
                    showPageError(error.error || 'Failed to post comment');
                }
            } catch (error) {
                console.error('Comment error:', error);
                showPageError('An error occurred while posting the comment');
            }
        });
    }

    // Handle the global deletePost function that is called from inline onclick
    window.deletePost = async function(postId) {
        if (!confirm('Are you sure you want to delete this post?')) {
            return;
        }

        try {
            const response = await fetch(`/posts/${postId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                window.location.href = '/';
            } else {
                const error = await response.json();
                showPageError(error.error || 'Failed to delete post');
            }
        } catch (error) {
            console.error('Delete error:', error);
            showPageError('An error occurred while deleting the post');
        }
    };

    // Function to handle post reactions
    async function handlePostReaction(postId, reactionType) {
        clearPageError();
        try {
            const response = await fetch(`/posts/${postId}/reactions`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ type: reactionType })
            });

            if (response.ok) {
                // Update the UI to reflect the new reaction count
                location.reload(); // Simple approach; in a real app we'd update the UI dynamically
            } else {
                const error = await response.json();
                showPageError(error.error || `Failed to ${reactionType} post`);
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
            const response = await fetch(`/comments/${commentId}/reactions`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ type: reactionType })
            });

            if (response.ok) {
                // Update the UI to reflect the new reaction count
                location.reload(); // Simple approach; in a real app we'd update the UI dynamically
            } else {
                const error = await response.json();
                showPageError(error.error || `Failed to ${reactionType} comment`);
            }
        } catch (error) {
            console.error(`Comment reaction error (${reactionType}):`, error);
            showPageError(`An error occurred while ${reactionType}ing the comment`);
        }
    }

    // Function to delete a comment
    async function deleteComment(commentId) {
        if (!confirm('Are you sure you want to delete this comment?')) {
            return;
        }

        clearPageError();
        try {
            const response = await fetch(`/comments/${commentId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                // Remove the comment from the UI
                const commentElement = document.getElementById(`comment-${commentId}`);
                if (commentElement) {
                    commentElement.remove();
                }
            } else {
                const error = await response.json();
                showPageError(error.error || 'Failed to delete comment');
            }
        } catch (error) {
            console.error('Delete comment error:', error);
            showPageError('An error occurred while deleting the comment');
        }
    }
});