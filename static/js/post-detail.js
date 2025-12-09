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

// Helper function to update user comment count in sidebar and dropdown
function updateUserCommentCount(delta) {
    // Update in sidebar user-card
    const sidebarStats = document.querySelectorAll('.sidebar-right .user-card .stat-item');
    sidebarStats.forEach(function(statItem) {
        const label = statItem.querySelector('.stat-label');
        if (label && label.textContent.trim() === 'Comments') {
            const valueEl = statItem.querySelector('.stat-value');
            if (valueEl) {
                const currentValue = parseInt(valueEl.textContent) || 0;
                valueEl.textContent = Math.max(0, currentValue + delta);
            }
        }
    });

    // Update in dropdown menu
    const dropdownStats = document.querySelectorAll('.user-menu-dropdown .stat-item');
    dropdownStats.forEach(function(statItem) {
        const label = statItem.querySelector('.stat-label');
        if (label && label.textContent.trim() === 'Comments') {
            const valueEl = statItem.querySelector('.stat-value');
            if (valueEl) {
                const currentValue = parseInt(valueEl.textContent) || 0;
                valueEl.textContent = Math.max(0, currentValue + delta);
            }
        }
    });
}

document.addEventListener('DOMContentLoaded', function() {
    // Initialize reaction button states based on current user's reactions
    initializeReactionButtons();

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

        // Handle comment editing
        if (e.target.classList.contains('btn-edit-comment')) {
            e.preventDefault();
            const commentId = e.target.getAttribute('data-comment-id');
            const commentElement = document.getElementById(`comment-${commentId}`);
            if (commentElement) {
                startEditingComment(commentElement, commentId);
            }
        }

        // Handle post deletion
        if (e.target.classList.contains('btn-delete-post')) {
            e.preventDefault();
            const postId = e.target.getAttribute('data-post-id');
            await deletePost(postId);
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

                const response = await fetch(`/api/comments/posts/${postId}`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ content: content })
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
        const confirmed = await confirmDelete('Post');
        if (!confirmed) {
            return;
        }

        try {
            const response = await fetch(`/api/posts/${postId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                window.location.href = '/board?my_posts=true';
            } else {
                const error = await response.json();
                showPageError(error.error || 'Failed to delete post');
            }
        } catch (error) {
            console.error('Delete error:', error);
            showPageError('An error occurred while deleting the post');
        }
    };

    // Function to initialize reaction button states based on user's reactions
    async function initializeReactionButtons() {
        // Get all unique post IDs and check reactions for each
        const postIds = new Set();
        document.querySelectorAll('.btn-like[data-post-id], .btn-dislike[data-post-id]').forEach(button => {
            postIds.add(button.getAttribute('data-post-id'));
        });

        // Process each post to determine user's reaction
        for (const postId of postIds) {
            await checkUserReactionToTarget(postId, 'post');
        }

        // Get all unique comment IDs and check reactions for each
        const commentIds = new Set();
        document.querySelectorAll('.btn-like-comment[data-comment-id], .btn-dislike-comment[data-comment-id]').forEach(button => {
            commentIds.add(button.getAttribute('data-comment-id'));
        });

        // Process each comment to determine user's reaction
        for (const commentId of commentIds) {
            await checkUserReactionToTarget(commentId, 'comment');
        }
    }

    // Function to check a user's reaction for a specific target and update button UI
    async function checkUserReactionToTarget(targetId, targetType) {
        try {
            // Use the new API endpoint that returns only the current user's reaction to the target
            const response = await fetch(`/api/reactions/${targetType}/${targetId}/user`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include'  // Include session cookie
            });

            if (response.ok) {
                const userReaction = await response.json();

                // Update our local tracking of the user's reaction
                if (targetType === 'post') {
                    userReactions.posts[targetId] = userReaction.type || null;
                } else if (targetType === 'comment') {
                    userReactions.comments[targetId] = userReaction.type || null;
                }

                // Update the button UI to reflect the user's existing reaction
                if (userReaction.type === 'like') {
                    let likeButton;
                    if (targetType === 'post') {
                        likeButton = document.querySelector(`.btn-like[data-post-id="${targetId}"]`);
                    } else if (targetType === 'comment') {
                        likeButton = document.querySelector(`.btn-like-comment[data-comment-id="${targetId}"]`);
                    }

                    if (likeButton) {
                        likeButton.classList.add('active');
                    }
                } else if (userReaction.type === 'dislike') {
                    let dislikeButton;
                    if (targetType === 'post') {
                        dislikeButton = document.querySelector(`.btn-dislike[data-post-id="${targetId}"]`);
                    } else if (targetType === 'comment') {
                        dislikeButton = document.querySelector(`.btn-dislike-comment[data-comment-id="${targetId}"]`);
                    }

                    if (dislikeButton) {
                        dislikeButton.classList.add('active');
                    }
                }
            }
        } catch (error) {
            console.error(`Error checking user reaction to ${targetType} ${targetId}:`, error);
        }
    }

    // Object to track user's reactions to targets
    const userReactions = {
        posts: {},
        comments: {}
    };

    // Function to initialize reaction button states based on user's reactions
    async function initializeReactionButtons() {
        // This function requires backend changes to provide this data in the HTML
        // For now, we'll rely on the user clicking first to establish the state
        // This initialization would happen when we have an API to get all user reactions
    }

    // Function to get a user's specific reaction to a target
    async function getUserReactionToTarget(targetId, targetType) {
        try {
            // Get all reactions for this target
            const response = await fetch(`/api/reactions/${targetType}/${targetId}`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json'
                },
                credentials: 'include'  // Include session cookie
            });

            if (response.ok) {
                const reactions = await response.json();
                // In the current implementation, the backend doesn't identify which user made each reaction
                // To properly implement this feature, we would need an endpoint that specifically returns
                // the current user's reaction to a target, but since we can't change the backend,
                // we'll implement a solution that tracks the state after the first interaction
                return null;
            } else {
                return null;
            }
        } catch (error) {
            console.error('Error getting user reaction:', error);
            return null;
        }
    }

    // Function to handle post reactions - simplified version since toggle is handled by backend
    async function handlePostReaction(postId, reactionType) {
        clearPageError();

        try {
            // Get the current button elements to identify the current state
            const likeButton = document.querySelector(`.btn-like[data-post-id="${postId}"]`);
            const dislikeButton = document.querySelector(`.btn-dislike[data-post-id="${postId}"]`);

            // Check the current state of the buttons from our tracked state
            const currentReaction = userReactions.posts[postId] || null;
            const isLikeButtonActive = currentReaction === 'like';
            const isDislikeButtonActive = currentReaction === 'dislike';

            // Make a POST request to add or modify reaction (toggle handled by backend)
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
                // Update our tracking of user's reaction based on the action
                // If clicking the same reaction type as currently active, we're removing it
                if ((reactionType === 'like' && isLikeButtonActive) ||
                    (reactionType === 'dislike' && isDislikeButtonActive)) {
                    userReactions.posts[postId] = null;

                    // Visually update the button state for removal
                    if (reactionType === 'like' && likeButton) {
                        likeButton.classList.remove('active');
                        // Update the count (decrement by 1)
                        const currentCount = parseInt(likeButton.textContent.match(/\((\d+)\)/)?.[1] || '0');
                        likeButton.textContent = `👍 Like (${Math.max(0, currentCount - 1)})`;
                    } else if (reactionType === 'dislike' && dislikeButton) {
                        dislikeButton.classList.remove('active');
                        // Update the count (decrement by 1)
                        const currentCount = parseInt(dislikeButton.textContent.match(/\((\d+)\)/)?.[1] || '0');
                        dislikeButton.textContent = `👎 Dislike (${Math.max(0, currentCount - 1)})`;
                    }
                } else {
                    // Add or change reaction
                    userReactions.posts[postId] = reactionType;

                    // Visually update the button state for addition
                    if (reactionType === 'like' && likeButton) {
                        likeButton.classList.add('active');
                        // Update the count (increment by 1 if not previously active, or keep same if switching)
                        const currentCount = parseInt(likeButton.textContent.match(/\((\d+)\)/)?.[1] || '0');
                        likeButton.textContent = `👍 Like (${currentCount + 1})`;

                        // If switching from dislike to like, update the dislike button
                        if (isDislikeButtonActive && dislikeButton) {
                            const currentDislikeCount = parseInt(dislikeButton.textContent.match(/\((\d+)\)/)?.[1] || '0');
                            dislikeButton.textContent = `👎 Dislike (${Math.max(0, currentDislikeCount - 1)})`;
                            dislikeButton.classList.remove('active');
                        }
                    } else if (reactionType === 'dislike' && dislikeButton) {
                        dislikeButton.classList.add('active');
                        // Update the count (increment by 1 if not previously active, or keep same if switching)
                        const currentCount = parseInt(dislikeButton.textContent.match(/\((\d+)\)/)?.[1] || '0');
                        dislikeButton.textContent = `👎 Dislike (${currentCount + 1})`;

                        // If switching from like to dislike, update the like button
                        if (isLikeButtonActive && likeButton) {
                            const currentLikeCount = parseInt(likeButton.textContent.match(/\((\d+)\)/)?.[1] || '0');
                            likeButton.textContent = `👍 Like (${Math.max(0, currentLikeCount - 1)})`;
                            likeButton.classList.remove('active');
                        }
                    }
                }
            } else {
                const error = await response.json();
                showPageError(error.error || 'Failed to process reaction');
                location.reload(); // Reload to reset to server state on error
            }
        } catch (error) {
            console.error(`Reaction error (${reactionType}):`, error);
            showPageError(`An error occurred while processing the reaction`);
            // Reload to reset to server state
            location.reload();
        }
    }

    // Function to handle comment reactions - simplified version since toggle is handled by backend
    async function handleCommentReaction(commentId, reactionType) {
        clearPageError();

        try {
            // Get the current button elements to identify the current state
            const likeButton = document.querySelector(`.btn-like-comment[data-comment-id="${commentId}"]`);
            const dislikeButton = document.querySelector(`.btn-dislike-comment[data-comment-id="${commentId}"]`);

            // Check the current state of the buttons from our tracked state
            const currentReaction = userReactions.comments[commentId] || null;
            const isLikeButtonActive = currentReaction === 'like';
            const isDislikeButtonActive = currentReaction === 'dislike';

            // Make a POST request to add or modify reaction (toggle handled by backend)
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
                // Update our tracking of user's reaction based on the action
                // If clicking the same reaction type as currently active, we're removing it
                if ((reactionType === 'like' && isLikeButtonActive) ||
                    (reactionType === 'dislike' && isDislikeButtonActive)) {
                    userReactions.comments[commentId] = null;

                    // Visually update the button state for removal
                    if (reactionType === 'like' && likeButton) {
                        likeButton.classList.remove('active');
                        // Update the count (decrement by 1)
                        const currentCount = parseInt(likeButton.textContent.match(/\((\d+)\)/)?.[1] || '0');
                        likeButton.textContent = `👍 (${Math.max(0, currentCount - 1)})`;
                    } else if (reactionType === 'dislike' && dislikeButton) {
                        dislikeButton.classList.remove('active');
                        // Update the count (decrement by 1)
                        const currentCount = parseInt(dislikeButton.textContent.match(/\((\d+)\)/)?.[1] || '0');
                        dislikeButton.textContent = `👎 (${Math.max(0, currentCount - 1)})`;
                    }
                } else {
                    // Add or change reaction
                    userReactions.comments[commentId] = reactionType;

                    // Visually update the button state for addition
                    if (reactionType === 'like' && likeButton) {
                        likeButton.classList.add('active');
                        // Update the count (increment by 1 if not previously active, or keep same if switching)
                        const currentCount = parseInt(likeButton.textContent.match(/\((\d+)\)/)?.[1] || '0');
                        likeButton.textContent = `👍 (${currentCount + 1})`;

                        // If switching from dislike to like, update the dislike button
                        if (isDislikeButtonActive && dislikeButton) {
                            const currentDislikeCount = parseInt(dislikeButton.textContent.match(/\((\d+)\)/)?.[1] || '0');
                            dislikeButton.textContent = `👎 (${Math.max(0, currentDislikeCount - 1)})`;
                            dislikeButton.classList.remove('active');
                        }
                    } else if (reactionType === 'dislike' && dislikeButton) {
                        dislikeButton.classList.add('active');
                        // Update the count (increment by 1 if not previously active, or keep same if switching)
                        const currentCount = parseInt(dislikeButton.textContent.match(/\((\d+)\)/)?.[1] || '0');
                        dislikeButton.textContent = `👎 (${currentCount + 1})`;

                        // If switching from like to dislike, update the like button
                        if (isLikeButtonActive && likeButton) {
                            const currentLikeCount = parseInt(likeButton.textContent.match(/\((\d+)\)/)?.[1] || '0');
                            likeButton.textContent = `👍 (${Math.max(0, currentLikeCount - 1)})`;
                            likeButton.classList.remove('active');
                        }
                    }
                }
            } else {
                const error = await response.json();
                showPageError(error.error || 'Failed to process reaction');
                location.reload(); // Reload to reset to server state on error
            }
        } catch (error) {
            console.error(`Comment reaction error (${reactionType}):`, error);
            showPageError(`An error occurred while processing the reaction`);
            // Reload to reset to server state
            location.reload();
        }
    }

    // Function to delete a comment
    async function deleteComment(commentId) {
        const confirmed = await confirmDelete('Comment');
        if (!confirmed) {
            return;
        }

        clearPageError();
        try {
            const response = await fetch(`/api/comments/${commentId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                // Remove the comment from the UI
                const commentElement = document.getElementById(`comment-${commentId}`);
                if (commentElement) {
                    commentElement.remove();
                }

                // Update the comments count in the section header
                const commentsHeader = document.querySelector('.comments-section h2');
                if (commentsHeader) {
                    const match = commentsHeader.textContent.match(/Comments \((\d+)\)/);
                    if (match) {
                        const newCount = Math.max(0, parseInt(match[1]) - 1);
                        commentsHeader.textContent = `Comments (${newCount})`;
                    }
                }

                // Update user stats in sidebar and dropdown (comment count)
                updateUserCommentCount(-1);
            } else {
                const error = await response.json();
                showPageError(error.error || 'Failed to delete comment');
            }
        } catch (error) {
            console.error('Delete comment error:', error);
            showPageError('An error occurred while deleting the comment');
        }
    }

    // Function to start editing a comment
    function startEditingComment(commentElement, commentId) {
        const contentElement = commentElement.querySelector('.comment-content');
        const currentContent = contentElement.textContent.trim();

        // Create form structure similar to post edit form
        const formGroup = document.createElement('div');
        formGroup.className = 'form-group';

        const textarea = document.createElement('textarea');
        textarea.className = 'edit-comment-textarea';
        textarea.value = currentContent;
        textarea.rows = 4;
        textarea.required = true;
        textarea.placeholder = "Edit your comment...";

        formGroup.appendChild(textarea);

        // Save original content for reference
        const originalContent = contentElement.innerHTML;
        contentElement.setAttribute('data-original-content', originalContent);

        // Replace content with form structure
        contentElement.innerHTML = '';
        contentElement.appendChild(formGroup);

        // Create save and cancel buttons with consistent styling
        const actionsDiv = commentElement.querySelector('.comment-actions');
        const originalActions = actionsDiv.innerHTML;
        // Store original actions in a data attribute for later restoration
        actionsDiv.setAttribute('data-original-actions', originalActions);

        // Use consistent button styling like in post edit form
        actionsDiv.innerHTML = `
            <button class="btn btn-primary btn-save-comment" data-comment-id="${commentId}">Save</button>
            <button class="btn btn-secondary btn-cancel-edit">Cancel</button>
        `;

        // Focus on the textarea
        textarea.focus();
    }

    // Function to update a comment
    async function updateComment(commentId, newContent) {
        try {
            const response = await fetch(`/api/comments/${commentId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ content: newContent })
            });

            if (response.ok) {
                // Reload the page to reflect the updated comment
                location.reload();
            } else {
                const error = await response.json();
                showPageError(error.error || 'Failed to update comment');
                return false;
            }
        } catch (error) {
            console.error('Update comment error:', error);
            showPageError('An error occurred while updating the comment');
            return false;
        }
        return true;
    }

    // Add event listener for save and cancel buttons
    document.body.addEventListener('click', async function(e) {
        if (e.target.classList.contains('btn-save-comment')) {
            e.preventDefault();
            const commentId = e.target.getAttribute('data-comment-id');
            const commentElement = document.getElementById(`comment-${commentId}`);
            if (commentElement) {
                const textarea = commentElement.querySelector('textarea');
                if (textarea) {
                    const newContent = textarea.value.trim();
                    if (!newContent) {
                        showPageError('Comment content cannot be empty');
                        return;
                    }
                    await updateComment(commentId, newContent);
                }
            }
        }

        if (e.target.classList.contains('btn-cancel-edit')) {
            e.preventDefault();
            const commentElement = e.target.closest('.comment');
            if (commentElement) {
                const contentElement = commentElement.querySelector('.comment-content');
                const actionsDiv = commentElement.querySelector('.comment-actions');

                // Restore original content
                contentElement.innerHTML = contentElement.getAttribute('data-original-content') || '';

                // Restore original actions
                actionsDiv.innerHTML = actionsDiv.getAttribute('data-original-actions') || '';
            }
        }
    });
});