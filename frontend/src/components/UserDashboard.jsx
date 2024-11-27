import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';

const getUserIdFromToken = (token) => {
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const payload = JSON.parse(window.atob(base64));
    return payload.user_id;
  } catch (error) {
    console.error('Error parsing token:', error);
    return null;
  }
};

export default function UserDashboard() {
  const [activeTab, setActiveTab] = useState('purchased');
  const [userItems, setUserItems] = useState([]);
  const [orders, setOrders] = useState([]);
  const [selectedOrder, setSelectedOrder] = useState(null);
  const [message, setMessage] = useState('');
  const [messages, setMessages] = useState([]);
  const [trackingNumber, setTrackingNumber] = useState('');
  const [userId, setUserId] = useState(null);
  const token = localStorage.getItem('token');
  const navigate = useNavigate();

  useEffect(() => {
    if (!token) {
      navigate('/');
      return;
    }
    const id = getUserIdFromToken(token);
    setUserId(id);
    fetchUserItems();
    fetchOrders();
  }, [token]);

  const fetchUserItems = async () => {
    try {
      const response = await fetch('http://localhost:8080/user/items', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const data = await response.json();
      setUserItems(data || []);
    } catch (error) {
      console.error('Error fetching items:', error);
    }
  };

  const fetchOrders = async () => {
    try {
      const response = await fetch('http://localhost:8080/user/orders', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const data = await response.json();
      console.log('Orders fetched:', data);
      console.log('Current userId:', userId);
      setOrders(data || []);
    } catch (error) {
      console.error('Error fetching orders:', error);
    }
  };

  const fetchMessages = async (orderId) => {
    try {
      const response = await fetch(`http://localhost:8080/orders/${orderId}/messages`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const data = await response.json();
      setMessages(data || []);
    } catch (error) {
      console.error('Error fetching messages:', error);
    }
  };

  const sendMessage = async (orderId) => {
    if (!message.trim()) return;

    try {
      await fetch(`http://localhost:8080/orders/${orderId}/messages`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ message })
      });
      setMessage('');
      fetchMessages(orderId);
    } catch (error) {
      console.error('Error sending message:', error);
    }
  };

  const updateOrderStatus = async (orderId, status) => {
    try {
      await fetch(`http://localhost:8080/orders/update?order_id=${orderId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          status,
          message: trackingNumber ? `Tracking number: ${trackingNumber}` : undefined
        })
      });
      setTrackingNumber('');
      fetchOrders();
    } catch (error) {
      console.error('Error updating order:', error);
    }
  };

  const deleteItem = async (itemId) => {
    if (!confirm('Delete this item?')) return;
    try {
      const response = await fetch(`http://localhost:8080/items/delete?id=${itemId}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` }
      });
      if (response.ok) {
        fetchUserItems();
      } else {
        throw new Error('Failed to delete item');
      }
    } catch (error) {
      console.error('Error deleting item:', error);
      alert('Failed to delete item');
    }
  };

  const filterPurchasedOrders = (order) => {
    console.log("Checking purchase order:", order, "userId:", userId);
    return String(order.user_id) === String(userId);
  };

  const filterSoldOrders = (order) => {
    console.log("Checking sold order:", order, "userId:", userId);
    return order.items.some(item => String(item.seller_id) === String(userId));
  };

  return (
    <div className="container mx-auto p-4">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <Link to="/" className="bg-blue-500 text-white px-4 py-2 rounded">
          Back to Marketplace
        </Link>
      </div>

      <div className="flex space-x-2 mb-6">
        <button
          onClick={() => setActiveTab('purchased')}
          className={`px-4 py-2 rounded ${
            activeTab === 'purchased' ? 'bg-blue-500 text-white' : 'bg-gray-200'
          }`}
        >
          Purchases
        </button>
        <button
          onClick={() => setActiveTab('sold')}
          className={`px-4 py-2 rounded ${
            activeTab === 'sold' ? 'bg-blue-500 text-white' : 'bg-gray-200'
          }`}
        >
          Sales
        </button>
        <button
          onClick={() => setActiveTab('listings')}
          className={`px-4 py-2 rounded ${
            activeTab === 'listings' ? 'bg-blue-500 text-white' : 'bg-gray-200'
          }`}
        >
          My Listings
        </button>
      </div>

      {activeTab === 'purchased' && (
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">My Purchases</h2>
          <div className="space-y-4">
            {orders.filter(filterPurchasedOrders).map(order => (
              <div key={order.id} className="border rounded p-4">
                <div className="flex justify-between mb-2">
                  <span className="font-semibold">Order #{order.id}</span>
                  <span className={`px-2 py-1 rounded text-sm ${
                    order.status === 'delivered' ? 'bg-green-100 text-green-800' :
                    order.status === 'shipped' ? 'bg-blue-100 text-blue-800' :
                    'bg-yellow-100 text-yellow-800'
                  }`}>
                    {order.status}
                  </span>
                </div>
                <div className="space-y-2">
                  {order.items.map(item => (
                    <div key={item.id} className="flex justify-between">
                      <span>{item.title}</span>
                      <span>${item.price}</span>
                    </div>
                  ))}
                </div>
                <button
                  onClick={() => {
                    setSelectedOrder(order);
                    fetchMessages(order.id);
                  }}
                  className="mt-2 text-blue-500"
                >
                  View Details & Messages
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {activeTab === 'sold' && (
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">My Sales</h2>
          <div className="space-y-4">
            {orders.filter(filterSoldOrders).map(order => (
              <div key={order.id} className="border rounded p-4">
                <div className="flex justify-between mb-2">
                  <span className="font-semibold">Order #{order.id}</span>
                  <span className={`px-2 py-1 rounded text-sm ${
                    order.status === 'delivered' ? 'bg-green-100 text-green-800' :
                    order.status === 'shipped' ? 'bg-blue-100 text-blue-800' :
                    'bg-yellow-100 text-yellow-800'
                  }`}>
                    {order.status}
                  </span>
                </div>
                <div className="space-y-2">
                  {order.items.filter(item => item.seller_id === userId).map(item => (
                    <div key={item.id} className="flex justify-between">
                      <span>{item.title}</span>
                      <span>${item.price}</span>
                    </div>
                  ))}
                </div>
                {order.status === 'pending' && (
                  <div className="mt-2">
                    <input
                      type="text"
                      placeholder="Enter tracking number"
                      className="border rounded p-2 mr-2"
                      value={trackingNumber}
                      onChange={(e) => setTrackingNumber(e.target.value)}
                    />
                    <button
                      onClick={() => updateOrderStatus(order.id, 'shipped')}
                      className="bg-blue-500 text-white px-4 py-2 rounded"
                    >
                      Mark as Shipped
                    </button>
                  </div>
                )}
                <button
                  onClick={() => {
                    setSelectedOrder(order);
                    fetchMessages(order.id);
                  }}
                  className="mt-2 text-blue-500"
                >
                  View Details & Messages
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {activeTab === 'listings' && (
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">My Listed Items ({userItems.length})</h2>
            <Link to="/" className="bg-green-500 text-white px-4 py-2 rounded">
              Post New Item
            </Link>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {userItems.map(item => (
              <div key={item.id} className={`border rounded-lg overflow-hidden ${
                item.quantity <= 0 ? 'opacity-75' : ''
              }`}>
                {item.images?.[0] && (
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
        </div>
      )}

      {selectedOrder && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
          <div className="bg-white rounded-lg p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-semibold">Order Details & Messages</h2>
              <button onClick={() => setSelectedOrder(null)} className="text-gray-500">
                ×
              </button>
            </div>
            <div className="mb-4">
              <h3 className="font-semibold">Order Status: {selectedOrder.status}</h3>
              <p>Shipping Address:</p>
              <p>{selectedOrder.address.street}</p>
              <p>{selectedOrder.address.city}, {selectedOrder.address.state} {selectedOrder.address.zip_code}</p>
              <p>{selectedOrder.address.country}</p>
            </div>
            <div className="border-t pt-4">
              <h3 className="font-semibold mb-2">Messages</h3>
              <div className="space-y-2 mb-4">
                {messages.map(msg => (
                  <div key={msg.id} className={`p-2 rounded ${
                    msg.sender_id === userId
                      ? 'bg-blue-100 ml-8'
                      : 'bg-gray-100 mr-8'
                  }`}>
                    <p>{msg.message}</p>
                    <small className="text-gray-500">
                      {new Date(msg.created_at).toLocaleString()}
                    </small>
                  </div>
                ))}
              </div>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  placeholder="Type a message..."
                  className="flex-1 border rounded p-2"
                />
                <button
                  onClick={() => sendMessage(selectedOrder.id)}
                  className="bg-blue-500 text-white px-4 py-2 rounded"
                >
                  Send
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
