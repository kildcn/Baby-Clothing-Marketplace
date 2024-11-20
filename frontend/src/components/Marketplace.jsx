import { useState, useEffect } from 'react';

export default function Marketplace() {
  const [items, setItems] = useState([]);
  const [cartItems, setCartItems] = useState([]); // Initialize as empty array
  const [searchQuery, setSearchQuery] = useState('');
  const [category, setCategory] = useState('');
  const [size, setSize] = useState('');
  const [sortBy, setSortBy] = useState('');
  const [priceRange, setPriceRange] = useState({ min: '', max: '' });
  const [showCart, setShowCart] = useState(false);

  const categories = ['tops', 'bottoms', 'outerwear', 'footwear', 'accessories'];
  const sizes = ['XS', 'S', 'M', 'L', 'XL'];

  const fetchCart = async () => {
    try {
      const response = await fetch('http://localhost:8080/cart?user_id=123');
      const data = await response.json();
      setCartItems(data || []); // Ensure we always set an array
    } catch (error) {
      console.error('Error fetching cart:', error);
      setCartItems([]); // Set empty array on error
    }
  };

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

      if (!data) data = [];

      if (sortBy && data.length > 0) {
        data = sortItems(data, sortBy);
      }

      setItems(data);
    } catch (error) {
      console.error('Error fetching items:', error);
      setItems([]);
    }
  };

  const addToCart = async (itemId) => {
    try {
      const response = await fetch('http://localhost:8080/cart/add', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ user_id: '123', item_id: itemId })
      });

      if (!response.ok) {
        const error = await response.text();
        alert(error);
        return;
      }

      fetchCart();
    } catch (error) {
      console.error('Error adding to cart:', error);
    }
  };

  const removeFromCart = async (itemId) => {
    try {
      const response = await fetch('http://localhost:8080/cart/remove', {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: '123',
          item_id: itemId
        })
      });

      if (!response.ok) {
        throw new Error('Failed to remove item from cart');
      }

      fetchCart();
    } catch (error) {
      console.error('Error removing from cart:', error);
    }
  };

  const handleCheckout = async () => {
    alert('Thank you for your purchase!');
    setCartItems([]);
    setShowCart(false);
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

  useEffect(() => {
    fetchItems();
  }, [searchQuery, category, size, sortBy, priceRange]);

  useEffect(() => {
    fetchCart();
  }, []);

  const CartView = () => (
    <div className="fixed right-0 top-0 h-full w-80 bg-white shadow-lg p-4 overflow-y-auto z-50">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-bold">Shopping Cart</h2>
        <button
          onClick={() => setShowCart(false)}
          className="text-gray-500 hover:text-gray-700"
        >
          Close
        </button>
      </div>

      {cartItems.length === 0 ? (
        <p>Your cart is empty</p>
      ) : (
        <>
          {cartItems.map((item, index) => (
            <div key={`${item.id}-${index}`} className="border-b py-2">
              <div className="flex justify-between items-center">
                <div>
                  <h3 className="font-bold">{item.title}</h3>
                  <p className="text-gray-600">${item.price}</p>
                </div>
                <button
                  onClick={() => removeFromCart(item.id)}
                  className="text-red-500 hover:text-red-700 px-2"
                >
                  ×
                </button>
              </div>
            </div>
          ))}
          <div className="mt-4">
            <p className="font-bold">
              Total: ${cartItems.reduce((sum, item) => sum + item.price, 0).toFixed(2)}
            </p>
            <button
              onClick={handleCheckout}
              className="w-full bg-green-500 text-white py-2 rounded mt-2 hover:bg-green-600"
            >
              Checkout
            </button>
          </div>
        </>
      )}
    </div>
  );

  return (
    <div className="container mx-auto p-4">
      <button
        onClick={() => setShowCart(true)}
        className="fixed top-4 right-4 bg-blue-500 text-white px-4 py-2 rounded z-40"
      >
        Cart ({cartItems.length})
      </button>

      {showCart && <CartView />}

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
