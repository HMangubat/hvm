async function loadGallery() {
  try {
    const res = await fetch('/api/photos');
    const photos = await res.json();

    const gallery = document.getElementById('photoGallery');
    gallery.innerHTML = '';
    
    console.log(photos);

    photos.forEach(({ filename, description }) => {
      const container = document.createElement('div');

      const img = document.createElement('img');
      img.src = `/uploads/${filename}`;
      img.style.maxWidth = '150px';
      img.style.display = 'block';
      img.alt = description || 'Photo';

      const caption = document.createElement('p');
      caption.textContent = description || '(No description)';

      container.appendChild(img);
      container.appendChild(caption);

      gallery.appendChild(container);
    });
  } catch (err) {
    console.error('Failed to load gallery:', err);
  }
}

window.addEventListener('DOMContentLoaded', loadGallery);
