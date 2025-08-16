"use client";

import { usePopup } from '../../context/PopupContext';
import Popup from './Popup';
import Navbar from './Navbar';

const ClientLayout = ({ children }: { children: React.ReactNode }) => {
  const { popup, hidePopup } = usePopup();

  return (
    <>
      <Navbar />
      <main className="main-content">{children}</main>
      {popup?.visible && (
        <Popup
          message={popup.message}
          type={popup.type}
          onClose={hidePopup}
        />
      )}
    </>
  );
};

export default ClientLayout;
