import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useState, useEffect } from 'react';
import NotificationSystem from '../components/NotificationSystem';
import { MessageCircle, Package, ShoppingBag, Tag } from 'lucide-react';

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
  const [userNames, setUserNames] = useState({});
  const [trackingNumber, setTrackingNumber] = useState('');
  const [userId, setUserId] = useState(null);
  const [userName, setUserName] = useState('');
  const [trackingNumbers, setTrackingNumbers] = useState({});

  const token = localStorage.getItem('token');
  const navigate = useNavigate();
  const location = useLocation();

  // Handle opening orders from notifications
  useEffect(() => {
    if (location.state?.openOrder) {
      const order = orders.find(o => o.id === location.state.openOrder);
      if (order) {
        setSelectedOrder(order);
        fetchMessages(order.id);
      }
    }
  }, [orders, location.state]);

  // Initialize dashboard data
  useEffect(() => {
    if (!token) {
      navigate('/');
      return;
    }
    const id = getUserIdFromToken(token);
    setUserId(id);

    // Fetch current user's name
    const fetchCurrentUserName = async () => {
      try {
        const response = await fetch(`http://localhost:8080/users/${id}`, {
          headers: { 'Authorization': `Bearer ${token}` }
        });
        const data = await response.json();
        setUserName(data.name);
      } catch (error) {
        console.error('Error fetching user name:', error);
      }
    };

    fetchCurrentUserName();
    fetchUserItems();
    fetchOrders();
  }, [token]);

  const fetchUserName = async (userId) => {
    if (userNames[userId]) return;
    try {
      const response = await fetch(`http://localhost:8080/users/${userId}`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const data = await response.json();
      setUserNames(prev => ({
        ...prev,
        [userId]: data.name
      }));
    } catch (error) {
      console.error('Error fetching user name:', error);
    }
  };

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

      // Fetch names for all users in messages
      const uniqueUserIds = [...new Set(data.map(msg => msg.sender_id))];
      uniqueUserIds.forEach(fetchUserName);
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

  const confirmOrderDelivery = async (orderId) => {
    try {
      await updateOrderStatus(orderId, 'delivered');
      fetchOrders();
    } catch (error) {
      console.error('Error confirming delivery:', error);
    }
  };

  const cancelOrder = async (orderId, reason) => {
    try {
      await fetch(`http://localhost:8080/orders/update?order_id=${orderId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          status: 'cancelled',
          message: `Order cancelled. Reason: ${reason}`
        })
      });
      fetchOrders();
    } catch (error) {
      console.error('Error cancelling order:', error);
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

      if (trackingNumber) {
        setTrackingNumbers(prev => ({
          ...prev,
          [orderId]: trackingNumber
        }));
        setTrackingNumber('');
      }
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

  const filterPurchasedOrders = (order) => String(order.user_id) === String(userId);
  const filterSoldOrders = (order) => {
    console.log('Filtering order:', order);
    const isSold = order.items.some(item => String(item.seller_id) === String(userId));
    console.log('Is sold:', isSold);
    return isSold;
  };

  return (
    <div className="container mx-auto p-4">
      <div className="flex justify-between items-center mb-6">
        <div className="flex items-center gap-4">
          <h1 className="text-2xl font-bold">Dashboard</h1>
          {userName && <p className="text-gray-600">Welcome back, {userName}!</p>}
        </div>
        <div className="flex items-center gap-4">
          <NotificationSystem />
          <Link to="/" className="bg-blue-500 text-white px-4 py-2 rounded">
            Back to Marketplace
          </Link>
        </div>
      </div>

      <div className="flex space-x-2 mb-6">
        <button
          onClick={() => setActiveTab('purchased')}
          className={`px-4 py-2 rounded flex items-center gap-2 ${
            activeTab === 'purchased' ? 'bg-blue-500 text-white' : 'bg-gray-200'
          }`}
        >
          <ShoppingBag size={20} />
          Purchases
        </button>
        <button
          onClick={() => setActiveTab('sold')}
          className={`px-4 py-2 rounded flex items-center gap-2 ${
            activeTab === 'sold' ? 'bg-blue-500 text-white' : 'bg-gray-200'
          }`}
        >
          <Package size={20} />
          Sales
        </button>
        <button
          onClick={() => setActiveTab('listings')}
          className={`px-4 py-2 rounded flex items-center gap-2 ${
            activeTab === 'listings' ? 'bg-blue-500 text-white' : 'bg-gray-200'
          }`}
        >
          <Tag size={20} />
          My Listings
        </button>
      </div>

      {/* Purchased Orders Tab */}
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
              order.status === 'cancelled' ? 'bg-red-100 text-red-800' :
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
          <div className="mt-4 text-sm text-gray-600">
            <p className="font-semibold">Shipping Address:</p>
            <p>{order.address.firstName} {order.address.lastName}</p>
            <p>{order.address.street}</p>
            <p>{order.address.city}, {order.address.state} {order.address.zipCode}</p>
            <p>{order.address.country}</p>
          </div>
          {order.status === 'shipped' && (
            <div className="mt-4 space-y-2">
              {order.trackingNumber && (
                <div className="text-sm text-gray-600">
                  <p className="font-semibold">Tracking Number:</p>
                  <p>{order.trackingNumber}</p>
                </div>
              )}
              <button
                onClick={() => confirmOrderDelivery(order.id)}
                className="w-full bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"
              >
                Confirm Order Received
              </button>
            </div>
          )}
          {order.status === 'cancelled' && order.cancelReason && (
            <div className="mt-4 text-sm text-red-600">
              <p className="font-semibold">Cancellation Reason:</p>
              <p>{order.cancelReason}</p>
            </div>
          )}
          <button
            onClick={() => {
              setSelectedOrder(order);
              fetchMessages(order.id);
            }}
            className="mt-4 text-blue-500 flex items-center gap-2 hover:text-blue-600"
          >
            <MessageCircle size={16} />
            View Details & Messages
          </button>
        </div>
      ))}
      {orders.filter(filterPurchasedOrders).length === 0 && (
        <div className="text-center text-gray-500 py-8">
          You haven't made any purchases yet.
        </div>
      )}
    </div>
  </div>
)}

      {/* Sales Tab */}
      {activeTab === 'sold' && (
  <div className="bg-white rounded-lg shadow p-6">
    <h2 className="text-xl font-semibold mb-4">My Sales</h2>
    <div className="space-y-4">
      {orders.filter(filterSoldOrders).length === 0 ? (
        <p className="text-gray-500 text-center py-4">No sales found</p>
      ) : (
        orders.filter(filterSoldOrders).map(order => (
          <div key={order.id} className="border rounded p-4">
            <div className="flex justify-between mb-2">
              <span className="font-semibold">Order #{order.id}</span>
              <span className={`px-2 py-1 rounded text-sm ${
                order.status === 'delivered' ? 'bg-green-100 text-green-800' :
                order.status === 'shipped' ? 'bg-blue-100 text-blue-800' :
                order.status === 'cancelled' ? 'bg-red-100 text-red-800' :
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
            <div className="mt-4 text-sm text-gray-600">
              <p className="font-semibold">Shipping Address:</p>
              <p>{order.address.firstName} {order.address.lastName}</p>
              <p>{order.address.street}</p>
              <p>{order.address.city}, {order.address.state} {order.address.zipCode}</p>
              <p>{order.address.country}</p>
            </div>
            {order.status === 'pending' && (
              <div className="mt-4 space-y-3">
                <div className="flex gap-2">
                  <input
                    type="text"
                    placeholder="Enter tracking number"
                    className="flex-1 border rounded p-2"
                    value={trackingNumber}
                    onChange={(e) => setTrackingNumber(e.target.value)}
                  />
                  <button
                    onClick={() => updateOrderStatus(order.id, 'shipped')}
                    className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
                    disabled={!trackingNumber.trim()}
                  >
                    Mark as Shipped
                  </button>
                </div>
                <button
                  onClick={() => {
                    const reason = prompt('Please enter cancellation reason:');
                    if (reason) {
                      cancelOrder(order.id, reason);
                    }
                  }}
                  className="w-full text-red-500 hover:text-red-700 text-sm font-medium"
                >
                  Cancel Order
                </button>
              </div>
            )}
            {order.status === 'shipped' && trackingNumbers[order.id] && (
              <div className="mt-2 text-sm text-gray-600">
                <p className="font-semibold">Tracking Number:</p>
                <p>{trackingNumbers[order.id]}</p>
              </div>
            )}
            <button
              onClick={() => {
                setSelectedOrder(order);
                fetchMessages(order.id);
              }}
              className="mt-4 text-blue-500 flex items-center gap-2 hover:text-blue-600"
            >
              <MessageCircle size={16} />
              View Details & Messages
            </button>
          </div>
        ))
      )}
    </div>
  </div>
)}

      {/* Listings Tab */}
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

      {/* Order Details Modal */}
      {selectedOrder && (
  <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div className="bg-white rounded-lg p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-semibold">Order Details & Messages</h2>
        <button onClick={() => setSelectedOrder(null)} className="text-gray-500 hover:text-gray-700">
          ×
        </button>
      </div>
      <div className="mb-6 space-y-4">
        <div>
          <h3 className="font-semibold">Order Status: {selectedOrder.status}</h3>
          <p className="text-sm text-gray-500">Order #{selectedOrder.id}</p>
        </div>

        <div className="bg-gray-50 p-4 rounded-lg">
          <h4 className="font-semibold mb-2">Shipping Details</h4>
          <div className="space-y-1">
            <p className="font-medium">{selectedOrder.address.first_name} {selectedOrder.address.last_name}</p>
            <p>{selectedOrder.address.street}</p>
            <p>{selectedOrder.address.city}, {selectedOrder.address.state} {selectedOrder.address.zip_code}</p>
            <p>{selectedOrder.address.country}</p>
          </div>
        </div>

        <div className="bg-gray-50 p-4 rounded-lg">
          <h4 className="font-semibold mb-2">Order Items</h4>
          <div className="space-y-2">
            {selectedOrder.items.map(item => (
              <div key={item.id} className="flex justify-between">
                <span>{item.title}</span>
                <span className="font-medium">${item.price}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="border-t pt-4">
        <h3 className="font-semibold mb-2">Messages</h3>
        <div className="space-y-2 mb-4 max-h-96 overflow-y-auto">
          {messages.map(msg => (
            <div
              key={msg.id}
              className={`p-2 rounded ${
                msg.sender_id === userId
                  ? 'bg-blue-100 ml-8'
                  : 'bg-gray-100 mr-8'
              }`}
            >
              <div className="flex justify-between items-center mb-1">
                <span className="font-semibold text-sm">
                  {msg.sender_id === userId ? 'You' : userNames[msg.sender_id] || 'User'}
                </span>
                <small className="text-gray-500">
                  {new Date(msg.created_at).toLocaleString()}
                </small>
              </div>
              <p>{msg.message}</p>
            </div>
          ))}
        </div>
        <div className="flex gap-2">
          <input
            type="text"
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            onKeyPress={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                sendMessage(selectedOrder.id);
              }
            }}
            placeholder="Type a message..."
            className="flex-1 border rounded p-2"
          />
          <button
            onClick={() => sendMessage(selectedOrder.id)}
            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
            disabled={!message.trim()}
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
