"use client";

import { useAuth } from "@/context/AuthContext";
import Notifications from "../components/Notifications";

const NotificationsPage = () => {
  const {user} = useAuth()
  return (
  user != null &&(  <div className="container">
      <Notifications />
    </div>)
  );
};

export default NotificationsPage;
