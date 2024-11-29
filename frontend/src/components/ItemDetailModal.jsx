import React, { useState } from 'react';
import PropTypes from 'prop-types';

const ItemDetailModal = ({ item, onClose, currentUser, token, setShowLogin }) => {
 const [message, setMessage] = useState('');

 const sendMessage = async () => {
   try {
     const response = await fetch(`http://localhost:8080/items/${item.id}/message`, {
       method: 'POST',
       headers: {
         'Content-Type': 'application/json',
         'Authorization': `Bearer ${token}`
       },
       body: JSON.stringify({ message })
     });

     if (response.ok) {
       setMessage('');
       alert('Message sent to seller');
     }
   } catch (error) {
     console.error('Error sending message:', error);
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

     alert('Item added to cart!');
     onClose();
   } catch (error) {
     console.error('Error adding to cart:', error);
   }
 };

 return (
   <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" onClick={onClose}>
     <div className="bg-white p-6 rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
       <div className="flex justify-between items-start mb-4">
         <h2 className="text-2xl font-bold">{item.title}</h2>
         <button onClick={onClose} className="text-gray-500 hover:text-gray-700">×</button>
       </div>

       <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
         <div>
           {item.images?.[0] && (
             <img
               src={`http://localhost:8080/images?path=${item.images[0]}`}
               alt={item.title}
               className="w-full h-64 object-cover rounded"
             />
           )}
         </div>

         <div>
           <p className="text-xl font-bold mb-2">${item.price.toFixed(2)}</p>
           <p className="mb-2 text-gray-600">Seller: {item.seller_name}</p>
           <p className="text-gray-600 mb-4">{item.description}</p>
           <div className="flex gap-2 mb-4">
             <span className="bg-gray-200 px-2 py-1 rounded">{item.size}</span>
             <span className="bg-gray-200 px-2 py-1 rounded">{item.category}</span>
           </div>

           {currentUser?.id !== item.seller_id && item.quantity > 0 && (
             <button
               onClick={() => addToCart(item.id)}
               className="w-full bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 mb-4"
             >
               Add to Cart
             </button>
           )}

           {currentUser?.id !== item.seller_id && (
             <div className="mt-4">
               <h3 className="font-bold mb-2">Ask Seller a Question</h3>
               <textarea
                 value={message}
                 onChange={(e) => setMessage(e.target.value)}
                 className="w-full border rounded p-2 mb-2"
                 placeholder="Type your question here..."
                 rows={3}
               />
               <button
                 onClick={sendMessage}
                 className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
                 disabled={!message.trim()}
               >
                 Send Message
               </button>
             </div>
           )}
         </div>
       </div>
     </div>
   </div>
 );
};

ItemDetailModal.propTypes = {
 item: PropTypes.shape({
   id: PropTypes.string.isRequired,
   title: PropTypes.string.isRequired,
   description: PropTypes.string,
   price: PropTypes.number.isRequired,
   size: PropTypes.string,
   category: PropTypes.string,
   seller_id: PropTypes.string.isRequired,
   seller_name: PropTypes.string.isRequired,
   quantity: PropTypes.number.isRequired,
   images: PropTypes.arrayOf(PropTypes.string)
 }).isRequired,
 onClose: PropTypes.func.isRequired,
 currentUser: PropTypes.shape({
   id: PropTypes.string
 }),
 token: PropTypes.string.isRequired,
 setShowLogin: PropTypes.func.isRequired
};

export default ItemDetailModal;
