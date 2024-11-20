import { useState, useEffect } from 'react';

export default function Marketplace() {
  const [items, setItems] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [category, setCategory] = useState('');
  const [size, setSize] = useState('');
  const [sortBy, setSortBy] = useState(''); // new sort state
  const [priceRange, setPriceRange] = useState({ min: '', max: '' }); // new price filter

  const categories = ['tops', 'bottoms', 'outerwear', 'footwear', 'accessories'];
  const sizes = ['XS', 'S', 'M', 'L', 'XL'];

  useEffect(() => {
    fetchItems();
  }, [searchQuery, category, size, sortBy, priceRange]);

  const fetchItems = async () => {
    const params = new URLSearchParams({
      q: searchQuery,
      category,
      size,
      min_price: priceRange.min,
      max_price: priceRange.max
    });
    try {
      let response = await fetch(`http://localhost:8080/items/search?${params}`);
      let data = await response.json();

      // Check if data is null or undefined
      if (!data) data = [];

      // Client-side sorting
      if (sortBy && data.length > 0) {
        data = sortItems(data, sortBy);
      }

      setItems(data);
    } catch (error) {
      console.error('Error fetching items:', error);
      setItems([]); // Set empty array on error
    }
  };

  const sortItems = (items, sortType) => {
    return [...items].sort((a, b) => {
      switch (sortType) {
        case 'price-asc':
          return a.price - b.price;
        case 'price-desc':
          return b.price - a.price;
        case 'newest':
          return new Date(b.created_at) - new Date(a.created_at);
        default:
          return 0;
      }
    });
  };

  return (
    <div className="container mx-auto p-4">
      <div className="mb-8 bg-white shadow-lg rounded-lg p-6">
        <h1 className="text-3xl font-bold mb-6">Unisex Clothes Marketplace</h1>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
          <input
            type="text"
            placeholder="Search items..."
            className="border p-2 rounded"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />

          <select
            className="border p-2 rounded"
            value={category}
            onChange={(e) => setCategory(e.target.value)}
          >
            <option value="">All Categories</option>
            {categories.map(cat => (
              <option key={cat} value={cat}>{cat}</option>
            ))}
          </select>

          <select
            className="border p-2 rounded"
            value={size}
            onChange={(e) => setSize(e.target.value)}
          >
            <option value="">All Sizes</option>
            {sizes.map(s => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>

          <select
            className="border p-2 rounded"
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value)}
          >
            <option value="">Sort By</option>
            <option value="price-asc">Price: Low to High</option>
            <option value="price-desc">Price: High to Low</option>
            <option value="newest">Newest First</option>
          </select>
        </div>

        <div className="flex gap-4 items-center">
          <input
            type="number"
            placeholder="Min Price"
            className="border p-2 rounded w-32"
            value={priceRange.min}
            onChange={(e) => setPriceRange(prev => ({ ...prev, min: e.target.value }))}
          />
          <span>to</span>
          <input
            type="number"
            placeholder="Max Price"
            className="border p-2 rounded w-32"
            value={priceRange.max}
            onChange={(e) => setPriceRange(prev => ({ ...prev, max: e.target.value }))}
          />
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {items.map(item => (
          <div key={item.id} className="bg-white rounded-lg shadow-lg overflow-hidden">
            {item.images && item.images.length > 0 && (
              <img
                src={`http://localhost:8080/images?path=${item.images[0]}`}
                alt={item.title}
                className="w-full h-64 object-cover"
              />
            )}
            <div className="p-4">
              <h2 className="text-xl font-bold mb-2">{item.title}</h2>
              <p className="text-gray-600 mb-2">{item.description}</p>
              <div className="flex justify-between items-center">
                <span className="text-lg font-bold">${item.price}</span>
                <div className="space-x-2">
                  <span className="bg-gray-200 px-2 py-1 rounded">{item.size}</span>
                  <span className="bg-gray-200 px-2 py-1 rounded">{item.category}</span>
                </div>
              </div>
              <button
                onClick={() => addToCart(item.id)}
                className="mt-4 w-full bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
              >
                Add to Cart
              </button>
            </div>
          </div>
        ))}
      </div>

      {items.length === 0 && (
        <div className="text-center text-gray-500 mt-8">
          No items found. Try adjusting your search criteria.
        </div>
      )}
    </div>
  );
}
