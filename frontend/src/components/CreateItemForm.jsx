import { useState } from 'react';

export default function CreateItemForm() {
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    price: '',
    size: 'M',
    category: 'tops'
  });
  const [images, setImages] = useState([]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    const data = new FormData();
    const itemData = {
        ...formData,
        price: parseFloat(formData.price)
    };
    console.log('Form data being sent:', itemData);
    data.append('item', JSON.stringify(itemData));

    for (let image of images) {
        console.log('Adding image:', image.name);
        data.append('images', image);
    }

    try {
        const response = await fetch('http://localhost:8080/items/create', {
            method: 'POST',
            body: data
        });
        const result = await response.text();
        console.log('Server response:', result);

        if (response.ok) {
            alert('Item created successfully!');
            setFormData({
                title: '',
                description: '',
                price: '',
                size: 'M',
                category: 'tops'
            });
            setImages([]);
        }
    } catch (error) {
        console.error('Error:', error);
    }
};

  return (
    <div className="max-w-md mx-auto bg-white p-6 rounded-lg shadow-lg">
      <h2 className="text-2xl font-bold mb-4">Create New Listing</h2>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block mb-1">Title</label>
          <input
            type="text"
            value={formData.title}
            onChange={e => setFormData({...formData, title: e.target.value})}
            className="w-full border rounded p-2"
            required
          />
        </div>

        <div>
          <label className="block mb-1">Description</label>
          <textarea
            value={formData.description}
            onChange={e => setFormData({...formData, description: e.target.value})}
            className="w-full border rounded p-2"
            required
          />
        </div>

        <div>
          <label className="block mb-1">Price</label>
          <input
            type="number"
            value={formData.price}
            onChange={e => setFormData({...formData, price: e.target.value})}
            className="w-full border rounded p-2"
            required
          />
        </div>

        <div>
          <label className="block mb-1">Size</label>
          <select
            value={formData.size}
            onChange={e => setFormData({...formData, size: e.target.value})}
            className="w-full border rounded p-2"
          >
            {['XS', 'S', 'M', 'L', 'XL'].map(size => (
              <option key={size} value={size}>{size}</option>
            ))}
          </select>
        </div>

        <div>
          <label className="block mb-1">Category</label>
          <select
            value={formData.category}
            onChange={e => setFormData({...formData, category: e.target.value})}
            className="w-full border rounded p-2"
          >
            {['tops', 'bottoms', 'outerwear', 'footwear', 'accessories'].map(cat => (
              <option key={cat} value={cat}>{cat}</option>
            ))}
          </select>
        </div>

        <div>
          <label className="block mb-1">Images (max 3)</label>
          <input
            type="file"
            onChange={e => setImages(Array.from(e.target.files))}
            multiple
            accept="image/*"
            className="w-full border rounded p-2"
            required
          />
        </div>

        <button type="submit" className="w-full bg-blue-500 text-white py-2 rounded hover:bg-blue-600">
          Create Listing
        </button>
      </form>
    </div>
  );
}
