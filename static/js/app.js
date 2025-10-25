// app.js - Client-side JavaScript for the forum application

// Handle reaction (like/dislike) button clicks
document.addEventListener('DOMContentLoaded', function() {
    // Add event listeners to reaction buttons
    const reactionButtons = document.querySelectorAll('.btn-reaction');
    
    reactionButtons.forEach(button => {
        button.addEventListener('click', handleReaction);
    });
});

// Send reaction to server via AJAX
function handleReaction(event) {
    const button = event.currentTarget;
    const targetType = button.dataset.type;
    const targetId = button.dataset.id;
    const reactionType = button.dataset.reaction;
    
    // Send POST request to server
    fetch('/reaction', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            target_type: targetType,
            target_id: parseInt(targetId),
            reaction_type: parseInt(reactionType)
        })
    })
    .then(response => response.json())
    .then(data => {
        // Update button text with new counts
        if (data.success) {
            updateReactionCounts(targetType, targetId, data.likes, data.dislikes);
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
}

// Update reaction counts in the UI
function updateReactionCounts(targetType, targetId, likes, dislikes) {
    // Find and update the reaction buttons for this target
    const buttons = document.querySelectorAll(`[data-type="${targetType}"][data-id="${targetId}"]`);
    buttons.forEach(button => {
        const reaction = button.dataset.reaction;
        if (reaction === '1') {
            button.textContent = `👍 Like (${likes})`;
        } else if (reaction === '-1') {
            button.textContent = `👎 Dislike (${dislikes})`;
        }
    });
}
