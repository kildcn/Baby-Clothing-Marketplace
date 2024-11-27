import React, { useState, useEffect } from 'react';
import { Bell } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

const NotificationSystem = () => {
  const [notifications, setNotifications] = useState([]);
  const [showNotifications, setShowNotifications] = useState(false);
  const token = localStorage.getItem('token');
  const userId = localStorage.getItem('user_id');
  const navigate = useNavigate();

  // Track seen messages in localStorage
  const markMessageAsSeen = (messageId) => {
    const seenMessages = JSON.parse(localStorage.getItem('seenMessages') || '[]');
    if (!seenMessages.includes(messageId)) {
      localStorage.setItem('seenMessages', JSON.stringify([...seenMessages, messageId]));
    }
  };

  const isMessageSeen = (messageId) => {
    const seenMessages = JSON.parse(localStorage.getItem('seenMessages') || '[]');
    return seenMessages.includes(messageId);
  };

  const checkForNewNotifications = async () => {
    if (!token || !userId) return;

    try {
      const ordersResponse = await fetch('http://localhost:8080/user/orders', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const orders = await ordersResponse.json();

      let newNotifications = [];

      // Check messages for each order
      for (const order of orders) {
        const messagesResponse = await fetch(`http://localhost:8080/orders/${order.id}/messages`, {
          headers: { 'Authorization': `Bearer ${token}` }
        });
        const messages = await messagesResponse.json();

        // Filter for unseen messages from others
        const unseenMessages = messages.filter(msg =>
          msg.sender_id !== userId && !isMessageSeen(msg.id)
        );

        // Add notifications for new messages
        newNotifications.push(...unseenMessages.map(msg => ({
          id: msg.id,
          title: 'New message',
          message: msg.message.slice(0, 50) + (msg.message.length > 50 ? '...' : ''),
          type: 'message',
          orderId: order.id,
          timestamp: new Date(msg.created_at).toISOString()
        })));
      }

      // Update notifications state
      if (newNotifications.length > 0) {
        setNotifications(prev => {
          const combined = [...newNotifications, ...prev];
          // Remove duplicates and keep only latest 10
          const unique = Array.from(new Map(combined.map(item => [item.id, item])).values())
            .slice(0, 10);
          return unique;
        });
      }

    } catch (error) {
      console.error('Error checking notifications:', error);
    }
  };

  const clearNotifications = () => {
    // Mark all current message notifications as seen
    notifications.forEach(notification => {
      if (notification.type === 'message') {
        markMessageAsSeen(notification.id);
      }
    });

    setNotifications([]);
    setShowNotifications(false);
  };

  const handleNotificationClick = (notification) => {
    // Mark message as seen if it's a message notification
    if (notification.type === 'message') {
      markMessageAsSeen(notification.id);
    }

    // Remove this notification
    setNotifications(prev => prev.filter(n => n.id !== notification.id));

    // Navigate to the order if there's an order ID
    if (notification.orderId) {
      navigate('/dashboard', {
        state: {
          openOrder: notification.orderId,
          scrollToMessages: true
        }
      });
    }

    setShowNotifications(false);
  };

  useEffect(() => {
    checkForNewNotifications();
    const interval = setInterval(checkForNewNotifications, 10000);
    return () => clearInterval(interval);
  }, [token, userId]);

  return (
    <div className="relative">
      <button
        onClick={() => setShowNotifications(!showNotifications)}
        className="relative p-2 hover:bg-gray-100 rounded-full"
      >
        <Bell size={24} />
        {notifications.length > 0 && (
          <span className="absolute top-0 right-0 bg-red-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center">
            {notifications.length}
          </span>
        )}
      </button>

      {showNotifications && (
        <div className="absolute right-0 mt-2 w-80 bg-white rounded-lg shadow-lg z-50">
          <div className="p-4 border-b flex justify-between items-center">
            <h3 className="font-semibold">Notifications</h3>
            {notifications.length > 0 && (
              <button
                onClick={clearNotifications}
                className="text-sm text-gray-500 hover:text-gray-700"
              >
                Clear all
              </button>
            )}
          </div>
          <div className="max-h-96 overflow-y-auto">
            {notifications.length === 0 ? (
              <p className="p-4 text-gray-500 text-center">No new notifications</p>
            ) : (
              notifications.map(notification => (
                <div
                  key={notification.id}
                  onClick={() => handleNotificationClick(notification)}
                  className="p-4 border-b hover:bg-gray-50 cursor-pointer"
                >
                  <div className="font-semibold text-sm">{notification.title}</div>
                  <p className="text-sm text-gray-600">{notification.message}</p>
                  <small className="text-gray-400">
                    {new Date(notification.timestamp).toLocaleString()}
                  </small>
                </div>
              ))
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default NotificationSystem;
