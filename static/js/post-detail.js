// JavaScript functions for post detail page

// Helper function to show inline error messages
function showPageError(message) {
    const pageErrors = document.getElementById('page-errors');
    if (pageErrors) {
        // SECURITY: Escape message to prevent XSS from reflected error content
        pageErrors.innerHTML = `<p class="error">${window.escapeHtml(message)}</p>`;
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
    // Use guard pattern to prevent redefinition if another script already defined it
    if (!window.deletePost) {
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
    }

    // Function to handle post reactions
    async function handlePostReaction(postId, reactionType) {
        clearPageError();
        try {
            // Simple approach: always send a POST request with the reaction
            // The backend service will handle toggle logic (add, update, or remove based on existing reactions)
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
                credentials: 'include'  // Include cookies with the request for authentication
            });

            if (response.ok) {
                location.reload(); // Reload to get updated counts
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
            // Simple approach: always send a POST request with the reaction
            // The backend service will handle toggle logic (add, update, or remove based on existing reactions)
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
                credentials: 'include'  // Include cookies with the request for authentication
            });

            if (response.ok) {
                location.reload(); // Reload to get updated counts
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
        const actionsDiv = commentElement.querySelector('.comment-actions');
        
        // Defensive check: ensure required elements exist
        if (!contentElement || !actionsDiv) {
            console.error('Comment element missing required children (.comment-content or .comment-actions)');
            return;
        }
        
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
        const originalActions = actionsDiv.innerHTML;
        // Store original actions in a data attribute for later restoration
        actionsDiv.setAttribute('data-original-actions', originalActions);

        // Use consistent button styling like in post edit form
        // SECURITY: Escape commentId to prevent XSS
        const safeCommentId = window.escapeHtml(commentId);
        actionsDiv.innerHTML = `
            <button class="btn btn-primary btn-save-comment" data-comment-id="${safeCommentId}">Save</button>
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