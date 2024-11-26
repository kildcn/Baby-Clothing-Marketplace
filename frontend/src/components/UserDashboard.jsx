import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';

export default function UserDashboard() {
  const [userItems, setUserItems] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const token = localStorage.getItem('token');
  const navigate = useNavigate();

  useEffect(() => {
    if (!token) {
      navigate('/');
      return;
    }
    fetchUserItems();
  }, [token]);

  const fetchUserItems = async () => {
    try {
      const response = await fetch('http://localhost:8080/user/items', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to fetch items');
      }

      const data = await response.json();
      setUserItems(data || []);
    } catch (error) {
      console.error('Error in fetchUserItems:', error);
      setError(error.message);
      setUserItems([]);
    } finally {
      setIsLoading(false);
    }
  };

  const deleteItem = async (itemId) => {
    if (!confirm('Are you sure you want to delete this item?')) return;

    try {
      const response = await fetch(`http://localhost:8080/items/delete?id=${itemId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Failed to delete item');
      }

      fetchUserItems();
    } catch (error) {
      console.error('Error deleting item:', error);
      alert(error.message);
    }
  };

  if (!token) {
    return (
      <div className="text-center mt-8">
        Please log in to view your dashboard
        <div className="mt-4">
          <Link to="/" className="text-blue-500 hover:text-blue-600">
            Return to Marketplace
          </Link>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return <div className="text-center mt-8">Loading...</div>;
  }

  if (error) {
    return <div className="text-center mt-8 text-red-500">Error: {error}</div>;
  }

  return (
    <div className="container mx-auto p-4">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">My Dashboard</h1>
        <div className="flex gap-2">
          <Link
            to="/"
            className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"
          >
            Post New Item
          </Link>
          <Link
            to="/"
            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          >
            Back to Marketplace
          </Link>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4">
          My Listed Items ({userItems.length})
        </h2>
        {userItems.length === 0 ? (
          <div className="text-center py-8">
            <p className="text-gray-500 mb-4">You haven't listed any items yet.</p>
            <Link
              to="/"
              className="bg-blue-500 text-white px-6 py-2 rounded hover:bg-blue-600"
            >
              Create Your First Listing
            </Link>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {userItems.map(item => (
              <div key={item.id} className={`border rounded-lg overflow-hidden ${
                item.quantity <= 0 ? 'opacity-75' : ''
              }`}>
                {item.images && item.images.length > 0 && (
                  <img
                    src={`http://localhost:8080/images?path=${item.images[0]}`}
                    alt={item.title}
                    className="w-full h-48 object-cover"
                  />
                )}
                <div className="p-4">
                  <h3 className="font-bold">{item.title}</h3>
                  <p className="text-gray-600">${item.price.toFixed(2)}</p>
                  <div className="mt-2 flex items-center justify-between">
                    <span className={`px-2 py-1 rounded text-sm ${
                      item.quantity > 0 ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                    }`}>
                      {item.quantity > 0 ? `${item.quantity} in stock` : 'Out of stock'}
                    </span>
                    <button
                      onClick={() => deleteItem(item.id)}
                      className="text-red-600 hover:text-red-800"
                    >
                      Delete
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
