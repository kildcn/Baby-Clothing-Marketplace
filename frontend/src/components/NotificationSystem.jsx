import React, { useState, useEffect } from 'react';
import { Bell } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

const NotificationSystem = () => {
  const [notifications, setNotifications] = useState([]);
  const [showNotifications, setShowNotifications] = useState(false);
  const [currentUser, setCurrentUser] = useState(null);
  const token = localStorage.getItem('token');
  const navigate = useNavigate();

  useEffect(() => {
    const fetchCurrentUser = async () => {
      if (!token) return;
      try {
        const response = await fetch('http://localhost:8080/user/current', {
          headers: { 'Authorization': `Bearer ${token}` }
        });
        const userData = await response.json();
        setCurrentUser(userData);
        localStorage.setItem('user_id', userData.id);
      } catch (error) {
        console.error('Error fetching user data:', error);
      }
    };
    fetchCurrentUser();
  }, [token]);

  const getSeenMessagesKey = () => `seenMessages_${currentUser?.id}`;

  const markMessageAsSeen = async (orderId) => {
    const key = getSeenMessagesKey();
    const seenMessages = JSON.parse(localStorage.getItem(key) || '[]');

    try {
      const response = await fetch(`http://localhost:8080/messages/seen`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ order_id: orderId })
      });

      if (!response.ok) throw new Error('Failed to mark messages as seen');

      const orderNotification = notifications.find(n => n.orderId === orderId);
      if (!orderNotification) return;

      const updatedSeenMessages = [...seenMessages, orderNotification.id];
      localStorage.setItem(key, JSON.stringify(updatedSeenMessages));
    } catch (error) {
      console.error('Error marking messages as seen:', error);
    }
  };

  const checkForNewNotifications = async () => {
    if (!token || !currentUser) return;

    try {
      // Check for unread messages
      const messagesResponse = await fetch('http://localhost:8080/messages/unread', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const unreadMessages = await messagesResponse.json();

      // Check for order notifications
      const notificationsResponse = await fetch('http://localhost:8080/notifications/unread', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const orderNotifications = await notificationsResponse.json();

      const messageNotifications = unreadMessages?.map(msg => ({
        id: msg.id,
        orderId: msg.order_id,
        type: 'message',
        title: `New message${msg.count > 1 ? 's' : ''} in Order #${msg.order_id}`,
        message: msg.latest_message.slice(0, 50) + (msg.latest_message.length > 50 ? '...' : ''),
        timestamp: msg.latest_timestamp,
        messageCount: msg.count
      })) || [];

      const systemNotifications = orderNotifications?.map(note => ({
        id: note.id,
        type: note.type,
        orderId: note.reference_id,
        title: note.type === 'order_status' ? 'Order Status Update' : 'New Order',
        message: note.message,
        timestamp: note.created_at
      })) || [];

      setNotifications([...messageNotifications, ...systemNotifications]);
    } catch (error) {
      console.error('Error checking notifications:', error);
    }
  };

  const handleNotificationClick = async (notification) => {
    if (notification.type === 'message') {
      await markMessageAsSeen(notification.orderId);
    } else {
      // Mark order notification as read
      try {
        await fetch(`http://localhost:8080/notifications/seen/${notification.id}`, {
          method: 'POST',
          headers: { 'Authorization': `Bearer ${token}` }
        });
      } catch (error) {
        console.error('Error marking notification as seen:', error);
      }
    }

    setNotifications(prev => prev.filter(n => n.id !== notification.id));
    navigate('/dashboard', { state: { openOrder: notification.orderId, scrollToMessages: true }});
    setShowNotifications(false);
  };

  const clearAllNotifications = async () => {
    // Clear message notifications
    for (const notification of notifications.filter(n => n.type === 'message')) {
      await markMessageAsSeen(notification.orderId);
    }

    // Clear order notifications
    try {
      await fetch('http://localhost:8080/notifications/clear', {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` }
      });
    } catch (error) {
      console.error('Error clearing notifications:', error);
    }

    setNotifications([]);
    setShowNotifications(false);
  };

  useEffect(() => {
    if (currentUser) {
      checkForNewNotifications();
      const interval = setInterval(checkForNewNotifications, 30000);
      return () => clearInterval(interval);
    }
  }, [currentUser]);

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
        <div className="absolute right-0 mt-2 w-96 bg-white rounded-lg shadow-lg z-50">
          <div className="p-4 border-b flex justify-between items-center">
            <h3 className="font-semibold">Notifications</h3>
            {notifications.length > 0 && (
              <button
                onClick={clearAllNotifications}
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
                  {notification.type === 'message' && notification.messageCount > 1 && (
                    <p className="text-xs text-blue-600 mt-1">
                      +{notification.messageCount - 1} more messages
                    </p>
                  )}
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
