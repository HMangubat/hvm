document.getElementById('uploadForm').addEventListener('submit', async (e) => {
  e.preventDefault();

  const formData = new FormData();
  const file = document.getElementById('photo').files[0];
  formData.append('photo', file);

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
  } catch (err) {
    document.getElementById('uploadStatus').textContent = 'Error: ' + err.message;
  }
});
