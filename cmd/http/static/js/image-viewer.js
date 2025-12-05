// image-viewer.js - Fixed version
document.addEventListener('DOMContentLoaded', function() {
    initializeImageViewer();
});

function initializeImageViewer() {
    // Add click handlers to all images with data-image-view attribute
    document.addEventListener('click', function(e) {
        // Handle image preview clicks
        if (e.target.matches('[data-image-view]') || e.target.closest('[data-image-view]')) {
            const imageElement = e.target.hasAttribute('data-image-view') ? e.target : e.target.closest('[data-image-view]');
            const imageSrc = imageElement.getAttribute('data-image-src');
            const imageName = imageElement.getAttribute('data-image-name') || 'Image';
            openImageModal(imageSrc, imageName);
            e.preventDefault();
        }
        
        // Handle set profile picture button
        if (e.target.matches('[data-set-profile-picture]')) {
            openSetProfilePictureModal();
            e.preventDefault();
        }
        
// In the profile picture choice handler, remove stopPropagation
if (e.target.matches('[data-profile-picture-choice]') || e.target.closest('[data-profile-picture-choice]')) {
    const element = e.target.hasAttribute('data-profile-picture-choice') ? e.target : e.target.closest('[data-profile-picture-choice]');
    const entityId = element.getAttribute('data-entity-id');
    const entityType = element.getAttribute('data-entity-type');
    const fileId = element.getAttribute('data-file-id');
    
    console.log('Setting profile picture:', { entityId, entityType, fileId });
    setProfilePicture(entityId, entityType, fileId);
    e.preventDefault();
    // REMOVED: e.stopPropagation();
}
    });
}

function openImageModal(imageSrc, imageName) {
    // Create or get modal elements
    let modal = document.getElementById('imageModal');
    let modalImage = document.getElementById('modalImage');
    let downloadLink = document.getElementById('downloadLink');
    let modalLabel = document.getElementById('imageModalLabel');
    
    // Create modal if it doesn't exist
    if (!modal) {
        modal = document.createElement('div');
        modal.className = 'modal fade';
        modal.id = 'imageModal';
        modal.tabIndex = -1;
        modal.innerHTML = `
            <div class="modal-dialog modal-xl modal-dialog-centered">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title" id="imageModalLabel">Просмотр изображения</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
                    </div>
                    <div class="modal-body text-center">
                        <img id="modalImage" src="" class="img-fluid" alt="" style="max-height: 80vh;">
                    </div>
                    <div class="modal-footer">
                        <a id="downloadLink" href="#" class="btn btn-primary" download>
                            <i class="fas fa-download"></i> Скачать
                        </a>
                        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Закрыть</button>
                    </div>
                </div>
            </div>
        `;
        document.body.appendChild(modal);
        
        modalImage = document.getElementById('modalImage');
        downloadLink = document.getElementById('downloadLink');
        modalLabel = document.getElementById('imageModalLabel');
    }
    
    // Set modal content
    modalImage.src = imageSrc;
    modalImage.alt = imageName;
    modalLabel.textContent = imageName;
    downloadLink.href = imageSrc.replace('/preview/', '/');
    downloadLink.download = imageName;
    
    // Show modal
    const bootstrapModal = new bootstrap.Modal(modal);
    bootstrapModal.show();
}

function openSetProfilePictureModal() {
    const modal = document.getElementById('setProfilePictureModal');
    if (modal) {
        const bootstrapModal = new bootstrap.Modal(modal);
        bootstrapModal.show();
    }
}

function setProfilePicture(entityId, entityType, fileId) {
    console.log('Setting profile picture:', entityId, entityType, fileId);
    
    // Close the modal first
    const modal = bootstrap.Modal.getInstance(document.getElementById('setProfilePictureModal'));
    if (modal) {
        modal.hide();
    }
    
    // Use HTMX directly instead of manual fetch
    htmx.ajax('POST', `/${entityType}/${entityId}/set-profile-picture/${fileId}`, {
        target: '#content',
        swap: 'innerHTML'
    });
}

// Make functions available globally
window.openImageModal = openImageModal;
window.openSetProfilePictureModal = openSetProfilePictureModal;
window.setProfilePicture = setProfilePicture;