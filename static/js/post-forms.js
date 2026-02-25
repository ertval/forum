// JavaScript functions for post creation and editing

document.addEventListener('DOMContentLoaded', function() {
    // Track if image should be removed on form submission
    let shouldRemoveImage = false;

    // Handle content preview functionality for post creation
    const contentTextarea = document.getElementById('content');
    const contentPreview = document.getElementById('content-preview');
    
    if (contentTextarea && contentPreview) {
        // Initialize preview with existing content (for edit form)
        if (contentTextarea.value) {
            contentPreview.textContent = contentTextarea.value;
        }

        contentTextarea.addEventListener('input', function() {
            contentPreview.textContent = this.value;
        });
    }

    // Handle image preview functionality
    const imageInput = document.getElementById('image');
    if (imageInput) {
        imageInput.addEventListener('change', function(e) {
            const file = e.target.files[0];
            const preview = document.getElementById('image-preview');
            const fileNameDisplay = document.getElementById('file-name-display');

            // Update file name display
            if (fileNameDisplay) {
                fileNameDisplay.textContent = file ? file.name : 'No file chosen';
            }

            if (file) {
                // Read maxImageSize from form data attribute (bytes)
                const form = document.getElementById('post-create-form') || document.getElementById('post-edit-form');
                const maxImageSize = form ? parseInt(form.getAttribute('data-max-image-size'), 10) : 20971520;
                const maxImageSizeMB = maxImageSize / (1024 * 1024);
                
                if (file.size > maxImageSize) {
                    const formErrors = document.getElementById('form-errors');
                    if (formErrors) formErrors.innerHTML = `<p class="error">Image must be less than ${maxImageSizeMB}MB</p>`;
                    e.target.value = '';
                    if (preview) preview.innerHTML = '';
                    if (fileNameDisplay) fileNameDisplay.textContent = 'No file chosen';
                    return;
                }

                const reader = new FileReader();
                reader.onload = function(e) {
                    if (preview) {
                        preview.innerHTML = `
                            <img src="${e.target.result}" alt="Preview">
                            <button type="button" class="btn-remove-image" id="remove-preview-image" title="Remove image">
                                <span class="remove-icon">×</span> Remove Image
                            </button>
                        `;
                        // Attach remove handler to the new button
                        attachPreviewRemoveHandler();
                    }
                };
                reader.readAsDataURL(file);
            } else {
                if (preview) preview.innerHTML = '';
            }
        });
    }

    // Handle removing preview image (newly selected file)
    function attachPreviewRemoveHandler() {
        const removePreviewBtn = document.getElementById('remove-preview-image');
        if (removePreviewBtn) {
            removePreviewBtn.addEventListener('click', function() {
                const preview = document.getElementById('image-preview');
                const fileNameDisplay = document.getElementById('file-name-display');
                const imageInput = document.getElementById('image');

                // Clear the file input
                if (imageInput) imageInput.value = '';
                // Clear the preview
                if (preview) preview.innerHTML = '';
                // Reset file name display
                if (fileNameDisplay) fileNameDisplay.textContent = 'No file chosen';
            });
        }
    }

    // Handle removing current/existing image (in edit form)
    const removeCurrentBtn = document.getElementById('remove-current-image');
    if (removeCurrentBtn) {
        removeCurrentBtn.addEventListener('click', function() {
            const currentImageContainer = document.getElementById('current-image-container');
            if (currentImageContainer) {
                // Mark for removal
                shouldRemoveImage = true;
                // Hide the container
                currentImageContainer.style.display = 'none';
                // Update the label
                const imageLabel = document.querySelector('label[for="image"]');
                if (imageLabel) {
                    imageLabel.textContent = 'Add Image (Optional)';
                }
            }
        });
    }
    
    // Handle post creation form submission
    const postCreateForm = document.getElementById('post-create-form');
    if (postCreateForm) {
        postCreateForm.addEventListener('submit', async function(e) {
            e.preventDefault();

            const formErrors = document.getElementById('form-errors');
            if (formErrors) formErrors.innerHTML = '';

            const title = document.getElementById('title').value.trim();
            const content = document.getElementById('content').value.trim();
            const categoryCheckboxes = document.querySelectorAll('input[name="categories"]:checked');
            const categories = Array.from(categoryCheckboxes).map(cb => cb.value);

            // Validation
            if (!title) {
                if (formErrors) formErrors.innerHTML = '<p class="error">Title is required</p>';
                return;
            }

            if (!content) {
                if (formErrors) formErrors.innerHTML = '<p class="error">Content is required</p>';
                return;
            }

            if (categories.length === 0) {
                if (formErrors) formErrors.innerHTML = '<p class="error">Please select at least one category</p>';
                return;
            }

            const formData = new FormData();
            formData.append('title', title);
            formData.append('content', content);
            categories.forEach(cat => formData.append('categories[]', cat));

            const imageFile = document.getElementById('image').files[0];
            if (imageFile) {
                formData.append('image', imageFile);
            }

            try {
                const response = await fetch('/api/posts', {
                    method: 'POST',
                    body: formData
                });

                if (response.ok) {
                    const result = await response.json();
                    window.location.href = `/posts/${result.id}`;
                } else {
                    const error = await response.json();
                    if (formErrors) formErrors.innerHTML = `<p class="error">${error.error || 'Failed to create post'}</p>`;
                }
            } catch (error) {
                if (formErrors) formErrors.innerHTML = '<p class="error">Network error. Please try again.</p>';
            }
        });
    }
    
    // Handle post editing form submission
    const postEditForm = document.getElementById('post-edit-form');
    if (postEditForm) {
        // Get the post ID from the data attribute set in the template
        const postId = postEditForm.getAttribute('data-post-id');

        postEditForm.addEventListener('submit', async function(e) {
            e.preventDefault();

            const formErrors = document.getElementById('form-errors');
            if (formErrors) formErrors.innerHTML = '';

            const title = document.getElementById('title').value.trim();
            const content = document.getElementById('content').value.trim();
            const categoryCheckboxes = document.querySelectorAll('input[name="categories"]:checked');
            const categories = Array.from(categoryCheckboxes).map(cb => cb.value);

            // Validation
            if (!title) {
                if (formErrors) formErrors.innerHTML = '<p class="error">Title is required</p>';
                return;
            }

            if (!content) {
                if (formErrors) formErrors.innerHTML = '<p class="error">Content is required</p>';
                return;
            }

            if (categories.length === 0) {
                if (formErrors) formErrors.innerHTML = '<p class="error">Please select at least one category</p>';
                return;
            }

            const formData = new FormData();
            formData.append('title', title);
            formData.append('content', content);
            categories.forEach(cat => formData.append('categories[]', cat));

            // Add remove_image flag if user clicked remove button
            if (shouldRemoveImage) {
                formData.append('remove_image', 'true');
            }

            const imageFile = document.getElementById('image').files[0];
            if (imageFile) {
                formData.append('image', imageFile);
            }

            try {
                const response = await fetch(`/api/posts/${postId}`, {
                    method: 'PUT',
                    body: formData
                });

                if (response.ok) {
                    window.location.href = `/posts/${postId}`;
                } else {
                    const error = await response.json();
                    if (formErrors) formErrors.innerHTML = `<p class="error">${error.error || 'Failed to update post'}</p>`;
                }
            } catch (error) {
                if (formErrors) formErrors.innerHTML = '<p class="error">Network error. Please try again.</p>';
            }
        });
    }
    
    // Handle post deletion (for edit form)
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
                const formErrors = document.getElementById('form-errors');
                if (formErrors) formErrors.innerHTML = `<p class="error">${error.error || 'Failed to delete post'}</p>`;
            }
        } catch (error) {
            console.error('Delete error:', error);
            const formErrors = document.getElementById('form-errors');
            if (formErrors) formErrors.innerHTML = '<p class="error">An error occurred while deleting the post</p>';
        }
    };
});