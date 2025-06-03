
document.getElementById('uploadForm').addEventListener('submit', async (e) => {
  e.preventDefault();

  const formData = new FormData();
  const file = document.getElementById('photo').files[0];
  const userId = document.getElementById('userIdInput').value.trim();
  const description = document.getElementById('description').value.trim();

  if (!userId) {
    alert("Please enter a user ID.");
    return;
  }

  formData.append('photo', file);
  formData.append('user_id', userId);
  formData.append('description', description);

  try {
    const res = await fetch('/api/upload-photo', {
      method: 'POST',
      body: formData,
    });

    const msg = await res.text();
    document.getElementById('uploadStatus').textContent = msg;

    // Show preview
    const preview = document.getElementById('previewImage');
    preview.src = `/uploads/${file.name}`;
    preview.style.display = 'block';

    // Refresh gallery
    loadGallery();
  } catch (err) {
    document.getElementById('uploadStatus').textContent = 'Error: ' + err.message;
  }
});

async function loadGallery() {
  try {
    const res = await fetch('/api/photos');
    const photos = await res.json();

    const gallery = document.getElementById('photoGallery');
    gallery.innerHTML = '';

    photos.forEach(photo => {
      const wrapper = document.createElement('div');
      wrapper.style.display = 'inline-block';
      wrapper.style.margin = '10px';
      wrapper.style.textAlign = 'center';

      const img = document.createElement('img');
      img.src = `/uploads/${photo.filename}`;
      img.style.maxWidth = '150px';
      img.style.display = 'block';
      img.alt = photo.filename;

      const info = document.createElement('p');
      info.textContent = `${photo.description}`;

      // Delete button
      const delBtn = document.createElement('button');
      delBtn.textContent = 'Delete';
      delBtn.onclick = () => {
        if (confirm(`Delete photo "${photo.filename}"?`)) {
          deletePhoto(photo.id);
        }
      };

      // Update button
      const updateBtn = document.createElement('button');
      updateBtn.textContent = 'Update';
      updateBtn.style.marginLeft = '5px';
      updateBtn.onclick = () => {
        const newDescription = prompt("Enter new description:", photo.description);
        if (newDescription === null) return; // Cancelled
        updatePhoto(photo.id, newDescription);
      };


      wrapper.appendChild(img);
      wrapper.appendChild(info);
      wrapper.appendChild(delBtn);
      wrapper.appendChild(updateBtn);

      gallery.appendChild(wrapper);
    });
  } catch (err) {
    console.error('Failed to load gallery:', err);
  }
}

async function deletePhoto(id) {
  try {
    const res = await fetch(`/api/delete?id=${id}`, {
      method: 'DELETE',
    });
    const msg = await res.text();
    alert(msg);
    loadGallery();
  } catch (err) {
    alert('Delete failed: ' + err.message);
  }
}

async function updatePhoto(id, description) {
  try {
    const res = await fetch('/api/photos/update', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        id,
        description: description,
      }),
    });

    const msg = await res.text();
    alert(msg);
    loadGallery();
  } catch (err) {
    alert('Update failed: ' + err.message);
  }
}

window.addEventListener('DOMContentLoaded', loadGallery);
