import React, { useState, useEffect } from 'react';
import { Bell } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

const NotificationSystem = () => {
  const [notifications, setNotifications] = useState([]);
  const [showNotifications, setShowNotifications] = useState(false);
  const token = localStorage.getItem('token');
  const userId = localStorage.getItem('user_id');
  const navigate = useNavigate();

  const getSeenMessagesKey = () => `seenMessages_${userId}`;

  const markMessageAsSeen = async (orderId) => {
    const key = getSeenMessagesKey();
    const seenMessages = JSON.parse(localStorage.getItem(key) || '[]');

    try {
      const messagesResponse = await fetch(`http://localhost:8080/orders/${orderId}/messages`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const messages = await messagesResponse.json();

      // Find the latest message in current notifications for this order
      const orderNotification = notifications.find(n => n.orderId === orderId);
      if (!orderNotification) return;

      // Mark all messages up to this timestamp as seen
      const latestTimestamp = new Date(orderNotification.timestamp);
      const messageIdsToMark = messages
        .filter(msg => new Date(msg.created_at) <= latestTimestamp)
        .map(msg => msg.id);

      const updatedSeenMessages = [...new Set([...seenMessages, ...messageIdsToMark])];
      localStorage.setItem(key, JSON.stringify(updatedSeenMessages));
    } catch (error) {
      console.error('Error marking messages as seen:', error);
    }
  };

  const isMessageSeen = (messageId) => {
    const key = getSeenMessagesKey();
    const seenMessages = JSON.parse(localStorage.getItem(key) || '[]');
    return seenMessages.includes(messageId);
  };

  const checkForNewNotifications = async () => {
    if (!token || !userId) return;

    try {
      const ordersResponse = await fetch('http://localhost:8080/user/orders', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      const orders = await ordersResponse.json();

      const allNotifications = [];

      for (const order of orders) {
        // Check if user is involved in this order
        const isBuyer = order.user_id === userId;
        const isSeller = order.items.some(item => item.seller_id === userId);

        if (!isBuyer && !isSeller) continue;

        const messagesResponse = await fetch(`http://localhost:8080/orders/${order.id}/messages`, {
          headers: { 'Authorization': `Bearer ${token}` }
        });
        const messages = await messagesResponse.json();

        console.log('Processing order:', {
          orderId: order.id,
          isBuyer,
          isSeller,
          userId,
          orderUserId: order.user_id,
          sellerIds: order.items.map(item => item.seller_id)
        });

        // Filter for unread messages where user is recipient
        const unseenMessages = messages.filter(msg => {
          // Skip seen messages first
          if (isMessageSeen(msg.id)) {
            console.log('Skipping seen message:', msg.id);
            return false;
          }

          // If user is buyer, only show seller messages
          if (isBuyer) {
            const isFromSeller = order.items.some(item => item.seller_id === msg.sender_id);
            console.log('Buyer message check:', {
              messageId: msg.id,
              senderId: msg.sender_id,
              isFromSeller,
              isSentByCurrentUser: msg.sender_id === userId
            });
            return isFromSeller && msg.sender_id !== userId;
          }

          // If user is seller, only show buyer messages
          if (isSeller) {
            const isFromBuyer = msg.sender_id === order.user_id;
            console.log('Seller message check:', {
              messageId: msg.id,
              senderId: msg.sender_id,
              buyerId: order.user_id,
              isFromBuyer,
              isSentByCurrentUser: msg.sender_id === userId
            });
            return isFromBuyer && msg.sender_id !== userId;
          }

          return false;
        });

        if (unseenMessages.length > 0) {
          // Sort messages by date (newest first)
          const sortedMessages = unseenMessages.sort((a, b) =>
            new Date(b.created_at) - new Date(a.created_at)
          );

          const latestMessage = sortedMessages[0];
          console.log('Adding notification:', {
            messageId: latestMessage.id,
            orderId: order.id,
            messageCount: sortedMessages.length
          });

          allNotifications.push({
            id: latestMessage.id,
            title: `New message${sortedMessages.length > 1 ? 's' : ''} in Order #${order.id}`,
            message: latestMessage.message.slice(0, 50) + (latestMessage.message.length > 50 ? '...' : ''),
            type: 'message',
            orderId: order.id,
            timestamp: new Date(latestMessage.created_at).toISOString(),
            messageCount: sortedMessages.length
          });
        }
      }

      if (allNotifications.length > 0) {
        console.log('Setting notifications:', allNotifications);
        setNotifications(
          allNotifications
            .sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp))
            .slice(0, 10)
        );
      }
    } catch (error) {
      console.error('Error checking notifications:', error);
    }
  };

  const clearNotifications = async () => {
    for (const notification of notifications) {
      await markMessageAsSeen(notification.orderId);
    }
    setNotifications([]);
    setShowNotifications(false);
  };

  const handleNotificationClick = async (notification) => {
    await markMessageAsSeen(notification.orderId);
    setNotifications(prev => prev.filter(n => n.orderId !== notification.orderId));

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
    console.log('Auth State:', {
      token: localStorage.getItem('token'),
      userId: localStorage.getItem('user_id'),
      name: localStorage.getItem('name'),
      email: localStorage.getItem('email')
    });
  }, []);

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
                  {notification.messageCount > 1 && (
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
