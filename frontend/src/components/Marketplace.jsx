import { useState, useEffect } from 'react';

export default function Marketplace() {
  // Core state
  const [items, setItems] = useState([]);
  const [cartItems, setCartItems] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [category, setCategory] = useState('');
  const [size, setSize] = useState('');
  const [sortBy, setSortBy] = useState('');
  const [priceRange, setPriceRange] = useState({ min: '', max: '' });
  const [showCart, setShowCart] = useState(false);

  // Auth state
  const [token, setToken] = useState(localStorage.getItem('token'));
  const [showLogin, setShowLogin] = useState(false);
  const [showSignup, setShowSignup] = useState(false);
  const [loginData, setLoginData] = useState({ email: '', password: '' });
  const [signupData, setSignupData] = useState({ email: '', password: '', name: '' });
  const [currentUser, setCurrentUser] = useState(null);

  // Constants
  const categories = ['tops', 'bottoms', 'outerwear', 'footwear', 'accessories'];
  const sizes = ['XS', 'S', 'M', 'L', 'XL'];

  // Auth handlers
  const signup = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch('http://localhost:8080/signup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(signupData)
      });

      if (!response.ok) {
        const errorData = await response.text();
        throw new Error(errorData || 'Signup failed');
      }

      const data = await response.json();

      if (!data.token) {
        throw new Error('No token received');
      }

      localStorage.setItem('token', data.token);
      localStorage.setItem('user_id', data.user_id);
      localStorage.setItem('name', data.name);
      localStorage.setItem('email', data.email);
      setToken(data.token);
      setCurrentUser({ id: data.user_id, name: data.name, email: data.email });
      setShowSignup(false);
      alert('Signup successful!');
    } catch (error) {
      console.error('Signup error:', error);
      alert('Signup failed: ' + error.message);
    }
  };

  const login = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch('http://localhost:8080/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(loginData)
      });

      const data = await response.json();
      localStorage.setItem('token', data.token);
      setToken(data.token);
      setShowLogin(false);
      alert('Login successful!');
    } catch (error) {
      alert('Login failed: ' + error.message);
    }
  };

  const logout = () => {
    localStorage.removeItem('token');
    setToken(null);
    setCartItems([]);
  };

  // Cart handlers
  const fetchCart = async () => {
    if (!token) return;
    try {
      const response = await fetch('http://localhost:8080/cart', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const data = await response.json();
      setCartItems(data || []);
    } catch (error) {
      console.error('Error fetching cart:', error);
      setCartItems([]);
    }
  };

  const addToCart = async (itemId) => {
    if (!token) {
      setShowLogin(true);
      return;
    }

    try {
      const response = await fetch('http://localhost:8080/cart/add', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ item_id: itemId })
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
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ item_id: itemId })
      });

      if (!response.ok) throw new Error('Failed to remove item');
      fetchCart();
    } catch (error) {
      console.error('Error removing from cart:', error);
    }
  };

  // Item handlers
  const fetchItems = async () => {
    const params = new URLSearchParams({
      q: searchQuery,
      category,
      size,
      min_price: priceRange.min,
      max_price: priceRange.max
    });

    try {
      const response = await fetch(`http://localhost:8080/items/search?${params}`);
      const data = await response.json();

      if (sortBy && data?.length) {
        setItems(sortItems(data, sortBy));
      } else {
        setItems(data || []);
      }
    } catch (error) {
      console.error('Error fetching items:', error);
      setItems([]);
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

  // Effects
  useEffect(() => {
    fetchItems();
  }, [searchQuery, category, size, sortBy, priceRange]);

  useEffect(() => {
    if (token) {
      fetchCart();
    }
  }, [token]);

  return (
    <div className="container mx-auto p-4">
      {/* Header with auth buttons */}
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Unisex Clothes Marketplace</h1>
        <div className="flex gap-4">
          {token ? (
            <button
              onClick={logout}
              className="bg-red-500 text-white px-4 py-2 rounded"
            >
              Logout
            </button>
          ) : (
            <div className="flex gap-2">
              <button
                onClick={() => setShowLogin(true)}
                className="bg-blue-500 text-white px-4 py-2 rounded"
              >
                Login
              </button>
              <button
                onClick={() => setShowSignup(true)}
                className="bg-green-500 text-white px-4 py-2 rounded"
              >
                Sign Up
              </button>
            </div>
          )}
          <button
            onClick={() => setShowCart(true)}
            className="bg-blue-500 text-white px-4 py-2 rounded"
          >
            Cart ({cartItems.length})
          </button>
        </div>
      </div>

      {/* Search and filters */}
      <div className="mb-8 bg-white shadow-lg rounded-lg p-6">
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

      {/* Items grid */}
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
  {items.map(item => (
    <div
      key={item.id}
      className={`bg-white rounded-lg shadow-lg overflow-hidden ${
        item.quantity <= 0 ? 'opacity-50' : ''
      }`}
    >
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
        {item.quantity <= 0 ? (
          <button
            disabled
            className="mt-4 w-full bg-gray-400 text-white px-4 py-2 rounded cursor-not-allowed"
          >
            Out of Stock
          </button>
        ) : (
          <button
            onClick={() => addToCart(item.id)}
            className="mt-4 w-full bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          >
            Add to Cart
          </button>
        )}
      </div>
    </div>
  ))}
</div>

      {items.length === 0 && (
        <div className="text-center text-gray-500 mt-8">
          No items found. Try adjusting your search criteria.
        </div>
      )}

      {/* Login Modal */}
      {showLogin && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-8 rounded-lg shadow-xl max-w-md w-full">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-2xl font-bold">Login</h2>
              <button onClick={() => setShowLogin(false)} className="text-gray-500">×</button>
            </div>
            <form onSubmit={login} className="space-y-4">
              <div>
                <label className="block mb-1">Email</label>
                <input
                  type="email"
                  className="w-full border rounded p-2"
                  value={loginData.email}
                  onChange={e => setLoginData({...loginData, email: e.target.value})}
                  required
                />
              </div>
              <div>
                <label className="block mb-1">Password</label>
                <input
                  type="password"
                  className="w-full border rounded p-2"
                  value={loginData.password}
                  onChange={e => setLoginData({...loginData, password: e.target.value})}
                  required
                />
              </div>
              <button type="submit" className="w-full bg-blue-500 text-white py-2 rounded">
                Login
              </button>
              <p className="text-center text-sm">
                Don't have an account?{' '}
                <button
                  type="button"
                  onClick={() => {
                    setShowLogin(false);
                    setShowSignup(true);
                  }}
                  className="text-blue-500"
                >
                  Sign up
                </button>
              </p>
            </form>
          </div>
        </div>
      )}

      {/* Signup Modal */}
      {showSignup && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white p-8 rounded-lg shadow-xl max-w-md w-full">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-2xl font-bold">Sign Up</h2>
              <button onClick={() => setShowSignup(false)} className="text-gray-500">×</button>
            </div>
            <form onSubmit={signup} className="space-y-4">
              <div>
                <label className="block mb-1">Name</label>
                <input
                  type="text"
                  className="w-full border rounded p-2"
                  value={signupData.name}
                  onChange={e => setSignupData({...signupData, name: e.target.value})}
                  required
                />
              </div>
              <div>
                <label className="block mb-1">Email</label>
                <input
                  type="email"
                  className="w-full border rounded p-2"
                  value={signupData.email}
                  onChange={e => setSignupData({...signupData, email: e.target.value})}
                  required
                />
              </div>
              <div>
                <label className="block mb-1">Password</label>
                <input
                  type="password"
                  className="w-full border rounded p-2"
                  value={signupData.password}
                  onChange={e => setSignupData({...signupData, password: e.target.value})}
                  required
                />
              </div>
              <button type="submit" className="w-full bg-blue-500 text-white py-2 rounded">
                Sign Up
              </button>
              <p className="text-center text-sm">
                Already have an account?{' '}
                <button
                  type="button"
                  onClick={() => {
                    setShowSignup(false);
                    setShowLogin(true);
                  }}
                  className="text-blue-500"
                >
                  Login
                </button>
              </p>
            </form>
          </div>
        </div>
      )}

      {/* Cart Modal */}
     {showCart && (
       <div className="fixed right-0 top-0 h-full w-80 bg-white shadow-lg p-4 overflow-y-auto z-50">
         <div className="flex justify-between items-center mb-4">
           <h2 className="text-xl font-bold">Shopping Cart</h2>
           <button onClick={() => setShowCart(false)} className="text-gray-500">
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
                     className="text-red-500"
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
                 onClick={() => {
                   if (!token) {
                     setShowLogin(true);
                   } else {
                     alert('Thank you for your purchase!');
                     setCartItems([]);
                     setShowCart(false);
                   }
                 }}
                 className="w-full bg-green-500 text-white py-2 rounded mt-2"
               >
                 Checkout
               </button>
             </div>
           </>
         )}
       </div>
     )}
   </div>
 );
}
