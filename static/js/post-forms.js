// JavaScript functions for post creation and editing

document.addEventListener('DOMContentLoaded', function() {
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
                if (file.size > 20 * 1024 * 1024) {  // 20MB limit
                    const formErrors = document.getElementById('form-errors');
                    if (formErrors) formErrors.innerHTML = '<p class="error">Image must be less than 20MB</p>';
                    e.target.value = '';
                    if (preview) preview.innerHTML = '';
                    if (fileNameDisplay) fileNameDisplay.textContent = 'No file chosen';
                    return;
                }

                const reader = new FileReader();
                reader.onload = function(e) {
                    if (preview) {
                        preview.innerHTML = `<img src="${e.target.result}" alt="Preview">`;
                    }
                };
                reader.readAsDataURL(file);
            } else {
                if (preview) preview.innerHTML = '';
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

            const removeImage = document.getElementById('remove-image');
            if (removeImage && removeImage.checked) {
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
        if (!confirm('Are you sure you want to delete this post?')) {
            return;
        }
        
        try {
            const response = await fetch(`/api/posts/${postId}`, {
                method: 'DELETE'
            });
            
            if (response.ok) {
                window.location.href = '/';
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