import { useState } from 'react';
import CreateItemForm from './components/CreateItemForm';
import Marketplace from './components/Marketplace';

export default function App() {
  const [showForm, setShowForm] = useState(false);

  return (
    <div className="min-h-screen bg-gray-100">
      {showForm ? (
        <div className="p-4">
          <button
            onClick={() => setShowForm(false)}
            className="mb-4 bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600"
          >
            Back to Marketplace
          </button>
          <CreateItemForm />
        </div>
      ) : (
        <div className="p-4">
          <button
            onClick={() => setShowForm(true)}
            className="mb-4 bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          >
            Post New Item
          </button>
          <Marketplace />
        </div>
      )}
    </div>
  );
}
