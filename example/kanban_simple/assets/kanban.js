// Kanban Board Helper Functions
// This file contains utility functions for the kanban board

// Global upload function for file attachments
window.uploadFiles = function(files, boardID, cardID) {
    if (!cardID || cardID === '') {
        alert('Please save the card first before adding attachments');
        return;
    }
    
    var formData = new FormData();
    var validFiles = 0;
    
    for (var i = 0; i < files.length; i++) {
        if (files[i].size > 5 * 1024 * 1024) {
            alert('File ' + files[i].name + ' is too large (max 5MB)');
            continue;
        }
        formData.append('files', files[i]);
        validFiles++;
    }
    
    if (validFiles === 0) {
        return;
    }
    
    // Show progress if element exists
    var progressEl = document.getElementById('upload-progress');
    var statusEl = document.getElementById('upload-status');
    if (progressEl) {
        progressEl.style.display = 'block';
        if (statusEl) {
            statusEl.textContent = 'Uploading ' + validFiles + ' file(s)...';
        }
    }
    
    // Upload via AJAX
    fetch('/api/upload/' + boardID + '/' + cardID, {
        method: 'POST',
        body: formData
    })
    .then(function(response) { 
        return response.json(); 
    })
    .then(function(data) {
        if (progressEl) {
            progressEl.style.display = 'none';
        }
        if (data.success) {
            // Notify via WebSocket to refresh attachments
            if (typeof send_event === 'function') {
                send_event('kanban_board', 'RefreshAttachments', JSON.stringify({
                    cardID: cardID,
                    files: data.files
                }));
            }
            alert(data.message);
        } else {
            alert('Upload failed: ' + (data.error || 'Unknown error'));
        }
    })
    .catch(function(error) {
        if (progressEl) {
            progressEl.style.display = 'none';
        }
        alert('Upload error: ' + error.message);
    });
};

console.log('[Kanban] Upload helper loaded');