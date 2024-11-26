import { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import CreateItemForm from './components/CreateItemForm';
import Marketplace from './components/Marketplace';
import UserDashboard from './components/UserDashboard';

export default function App() {
  const [showForm, setShowForm] = useState(false);

  return (
    <Router>
      <div className="min-h-screen bg-gray-100">
        <Routes>
          <Route path="/dashboard" element={<UserDashboard />} />
          <Route
            path="/"
            element={
              showForm ? (
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
              )
            }
          />
        </Routes>
      </div>
    </Router>
  );
}
