"use client";

import React, { createContext, useContext, useState, ReactNode } from 'react';

type PopupType = 'success' | 'error';

interface PopupContextType {
  showPopup: (message: string, type: PopupType) => void;
  hidePopup: () => void;
  popup: {
    message: string;
    type: PopupType;
    visible: boolean;
  } | null;
}

const PopupContext = createContext<PopupContextType | undefined>(undefined);

export const PopupProvider = ({ children }: { children: ReactNode }) => {
  const [popup, setPopup] = useState<{ message: string; type: PopupType; visible: boolean } | null>(null);

  const showPopup = (message: string, type: PopupType) => {
    setPopup({ message, type, visible: true });
  };

  const hidePopup = () => {
    setPopup(null);
  };

  return (
    <PopupContext.Provider value={{ showPopup, hidePopup, popup }}>
      {children}
    </PopupContext.Provider>
  );
};

export const usePopup = () => {
  const context = useContext(PopupContext);
  if (context === undefined) {
    throw new Error('usePopup must be used within a PopupProvider');
  }
  return context;
};
