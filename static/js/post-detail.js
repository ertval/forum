// JavaScript functions for post detail page
'use strict';

// Helper function to show inline error messages (delegates to shared utility)
function showPageError(message) {
    window.showError(message, 'page-errors');
}

function clearPageError() {
    window.clearError('page-errors');
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
    // Event delegation for comment and post actions
    document.body.addEventListener('click', async function(e) {
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
                const result = await window.api.request(`/api/comments/posts/${postId}`, {
                    method: 'POST',
                    body: JSON.stringify({ content: content })
                });

                commentForm.reset();

                // Inject new comment into DOM instead of reloading
                const commentsList = document.querySelector('.comments-list');
                if (commentsList && result) {
                    const commentEl = createNewCommentElement(result);
                    // Insert at the end of the comments list
                    commentsList.appendChild(commentEl);

                    // Update comments count in section header
                    const commentsHeader = document.querySelector('.comments-section h2');
                    if (commentsHeader) {
                        const match = commentsHeader.textContent.match(/Comments \((\d+)\)/);
                        if (match) {
                            const newCount = parseInt(match[1]) + 1;
                            commentsHeader.textContent = `Comments (${newCount})`;
                        }
                    }

                    // Update user stats
                    updateUserCommentCount(1);

                    // Scroll to new comment
                    commentEl.scrollIntoView({ behavior: 'smooth', block: 'center' });
                }
            } catch (error) {
                showPageError(error.message || 'An error occurred while posting the comment');
            }
        });
    }

    // Create a new comment element from API response data
    function createNewCommentElement(data) {
        // Try to use HTML template element
        const template = document.getElementById('comment-template');
        const commentDate = new Date(data.created_at);
        const formattedDate = commentDate.toLocaleDateString('en-US', {
            year: 'numeric', month: 'short', day: '2-digit',
            hour: 'numeric', minute: '2-digit', hour12: true
        });
        const safeCommentId = window.escapeHtml(data.id);
        const safeContent = window.escapeHtml(data.content);

        // Get current user's username from the page
        const currentUsername = document.querySelector('.user-menu-dropdown .username')?.textContent?.trim()
            || document.querySelector('.welcome-text')?.textContent?.replace('Welcome, ', '')?.trim()
            || 'You';
        const safeAuthor = window.escapeHtml(currentUsername);

        if (template) {
            const clone = template.content.cloneNode(true);
            const article = clone.querySelector('article');
            article.id = `comment-${safeCommentId}`;

            clone.querySelector('[data-field="author"]').textContent = safeAuthor;
            clone.querySelector('[data-field="date"]').textContent = formattedDate;
            clone.querySelector('[data-field="content"]').textContent = safeContent;

            const likeBtn = clone.querySelector('[data-field="like-btn"]');
            likeBtn.setAttribute('data-comment-id', safeCommentId);

            const dislikeBtn = clone.querySelector('[data-field="dislike-btn"]');
            dislikeBtn.setAttribute('data-comment-id', safeCommentId);

            const editBtn = clone.querySelector('[data-field="edit-btn"]');
            editBtn.setAttribute('data-comment-id', safeCommentId);

            const deleteBtn = clone.querySelector('[data-field="delete-btn"]');
            deleteBtn.setAttribute('data-comment-id', safeCommentId);

            return article;
        }

        // Fallback: build element manually
        const article = document.createElement('article');
        article.className = 'comment';
        article.id = `comment-${safeCommentId}`;

        const header = document.createElement('div');
        header.className = 'comment-header';
        const author = document.createElement('span');
        author.className = 'comment-author';
        author.textContent = safeAuthor;
        const date = document.createElement('span');
        date.className = 'comment-date';
        date.textContent = formattedDate;
        header.appendChild(author);
        header.appendChild(date);

        const content = document.createElement('div');
        content.className = 'comment-content';
        content.textContent = safeContent;

        const actions = document.createElement('div');
        actions.className = 'comment-actions';

        const reactions = document.createElement('div');
        reactions.className = 'comment-reactions';
        const likeBtn = document.createElement('button');
        likeBtn.className = 'btn-like-comment';
        likeBtn.setAttribute('data-comment-id', safeCommentId);
        likeBtn.textContent = '👍 (0)';
        const dislikeBtn = document.createElement('button');
        dislikeBtn.className = 'btn-dislike-comment';
        dislikeBtn.setAttribute('data-comment-id', safeCommentId);
        dislikeBtn.textContent = '👎 (0)';
        reactions.appendChild(likeBtn);
        reactions.appendChild(dislikeBtn);

        const ownerActions = document.createElement('div');
        ownerActions.className = 'comment-owner-actions';
        const editBtn = document.createElement('button');
        editBtn.className = 'btn btn-secondary btn-edit-comment';
        editBtn.setAttribute('data-comment-id', safeCommentId);
        editBtn.textContent = 'Edit';
        const deleteBtn = document.createElement('button');
        deleteBtn.className = 'btn btn-danger btn-delete-comment';
        deleteBtn.setAttribute('data-comment-id', safeCommentId);
        deleteBtn.textContent = 'Delete';
        ownerActions.appendChild(editBtn);
        ownerActions.appendChild(deleteBtn);

        actions.appendChild(reactions);
        actions.appendChild(ownerActions);

        article.appendChild(header);
        article.appendChild(content);
        article.appendChild(actions);
        return article;
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
                await window.api.request(`/api/posts/${postId}`, {
                    method: 'DELETE'
                });
                window.location.href = '/board?my_posts=true';
            } catch (error) {
                showPageError(error.message || 'Failed to delete post');
            }
        };
    }


    // Function to delete a comment
    async function deleteComment(commentId) {
        const confirmed = await confirmDelete('Comment');
        if (!confirmed) {
            return;
        }

        clearPageError();
        try {
            await window.api.request(`/api/comments/${commentId}`, {
                method: 'DELETE'
            });

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
        } catch (error) {
            showPageError(error.message || 'An error occurred while deleting the comment');
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
            const result = await window.api.request(`/api/comments/${commentId}`, {
                method: 'PUT',
                body: JSON.stringify({ content: newContent })
            });

            // Update comment content in place instead of reloading
            const commentElement = document.getElementById(`comment-${commentId}`);
            if (commentElement) {
                const contentElement = commentElement.querySelector('.comment-content');
                const actionsDiv = commentElement.querySelector('.comment-actions');

                // Update content with escaped text
                if (contentElement) {
                    contentElement.textContent = result?.content || newContent;
                    contentElement.removeAttribute('data-original-content');
                }

                // Restore original actions
                if (actionsDiv) {
                    const originalActions = actionsDiv.getAttribute('data-original-actions');
                    if (originalActions) {
                        actionsDiv.innerHTML = originalActions;
                        actionsDiv.removeAttribute('data-original-actions');
                    }
                }
            }
        } catch (error) {
            showPageError(error.message || 'An error occurred while updating the comment');
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