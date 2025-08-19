import React, { useEffect } from 'react';
import './Popup.css';

interface PopupProps {
  message: string;
  type: 'success' | 'error';
  onClose: () => void;
}

const Popup: React.FC<PopupProps> = ({ message, type, onClose }) => {
  useEffect(() => {
    const timer = setTimeout(() => {
      onClose();
    }, 3000); // Auto-close after 5 seconds

    return () => {
      clearTimeout(timer);
    };
  }, [onClose]);

  if (!message) {
    return null;
  }

  return (
    <div className="popup-overlay">
      <div className={`popup-content ${type}`}>
        <button onClick={onClose} className="popup-close-button">
          &times;
        </button>
        <p>{message}</p>
      </div>
    </div>
  );
};

export default Popup;
