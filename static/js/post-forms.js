// JavaScript functions for post creation and editing
'use strict';

document.addEventListener('DOMContentLoaded', function() {
    // Track if image should be removed on form submission
    let shouldRemoveImage = false;

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

                // SEC-2 fix: Validate file type to prevent upload of non-image files
                const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
                if (!allowedTypes.includes(file.type)) {
                    const formErrors = document.getElementById('form-errors');
                    if (formErrors) formErrors.innerHTML = '<p class="error">Only JPEG, PNG, GIF, and WebP images are allowed</p>';
                    e.target.value = '';
                    if (preview) preview.innerHTML = '';
                    if (fileNameDisplay) fileNameDisplay.textContent = 'No file chosen';
                    return;
                }

                const reader = new FileReader();
                reader.onload = function(e) {
                    if (preview) {
                        // Build preview with DOM methods (avoids innerHTML with data URL)
                        preview.innerHTML = '';
                        const previewImg = document.createElement('img');
                        previewImg.src = e.target.result;
                        previewImg.alt = 'Preview';
                        const removeBtn = document.createElement('button');
                        removeBtn.type = 'button';
                        removeBtn.className = 'btn-remove-image';
                        removeBtn.id = 'remove-preview-image';
                        removeBtn.title = 'Remove image';
                        const removeIcon = document.createElement('span');
                        removeIcon.className = 'remove-icon';
                        removeIcon.textContent = '\u00d7';
                        removeBtn.appendChild(removeIcon);
                        removeBtn.appendChild(document.createTextNode(' Remove Image'));
                        preview.appendChild(previewImg);
                        preview.appendChild(removeBtn);
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
                const result = await window.api.request('/api/posts', {
                    method: 'POST',
                    body: formData
                });
                window.location.href = `/posts/${result.id}`;
            } catch (error) {
                if (formErrors) formErrors.innerHTML = `<p class="error">${window.escapeHtml(error.message || 'Failed to create post')}</p>`;
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
                await window.api.request(`/api/posts/${postId}`, {
                    method: 'PUT',
                    body: formData
                });
                window.location.href = `/posts/${postId}`;
            } catch (error) {
                if (formErrors) formErrors.innerHTML = `<p class="error">${window.escapeHtml(error.message || 'Failed to update post')}</p>`;
            }
        });
    }
    
    // Handle post deletion (for edit form)
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
                const formErrors = document.getElementById('form-errors');
                if (formErrors) formErrors.innerHTML = `<p class="error">${window.escapeHtml(error.message || 'Failed to delete post')}</p>`;
            }
        };
    }
});