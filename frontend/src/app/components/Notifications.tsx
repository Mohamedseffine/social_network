"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useAuth } from "../../context/AuthContext";
import { usePopup } from "../../context/PopupContext";
import { API_BASE_URL } from "../../utils/api";
import { useRouter } from "next/navigation";
const Notifications = () => {
  const router = useRouter()
  const { showPopup } = usePopup();
  const [notifications, setNotifications] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const { fetchNotifications: fetchUnreadCount } = useAuth();

  const fetchNotificationsData = async () => {
    setLoading(true);
    try {
      // First, mark all notifications as read on the backend.
      await fetch(`${API_BASE_URL}/notifications/read-all`, {
        method: "POST",
        credentials: "include",
      });

      // After marking as read, fetch the fresh unread count (should be 0)
      fetchUnreadCount();

      // Then, fetch the notifications themselves to display.
      const res = await fetch(`${API_BASE_URL}/notifications`, {
        credentials: "include",
      });

      if (res.ok) {
        const data = await res.json();
        setNotifications(data);
      } else {
        const errorText = await res.text();
        if (res.status == 401){
          router.push("/")
        }
        setError(`Failed to fetch notifications: ${errorText}`);
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchNotificationsData();
  }, []); // Dependency array is empty, so it runs once on mount.

  const handleAction = async (url: string) => {
    try {
      const res = await fetch(url, {
        method: "POST",
        credentials: "include",
      });

      if (res.ok) {
        fetchNotificationsData(); // Refresh notifications and the count
      } else {
        const errorText = await res.text();
        if (res.status==401){
          router.push("/")
        }
        showPopup(`Action failed: ${errorText}`, 'error');
      }
    } catch (err: any) {
      showPopup(`An error occurred: ${err.message}`, 'error');
    }
  };

  const handleDelete = async (notificationId: number) => {
    try {
      const res = await fetch(`${API_BASE_URL}/notifications/${notificationId}`, {
        method: "DELETE",
        credentials: "include",
      });

      if (res.ok) {
        setNotifications(notifications.filter(n => n.id !== notificationId));
        fetchUnreadCount(); // The number of unread might have changed
      } else {
        const errorText = await res.text();
        showPopup(`Failed to delete notification: ${errorText}`, 'error');
      }
    } catch (err: any) {
      showPopup(`An error occurred: ${err.message}`, 'error');
    }
  };

  const renderActions = (notification: any) => {
    let acceptUrl = "";
    let declineUrl = "";

    switch (notification.type) {
      case "follow_request":
        acceptUrl = `${API_BASE_URL}/requests/${notification.related_id}/accept`;
        declineUrl = `${API_BASE_URL}/requests/${notification.related_id}/decline`;
        break;
      case "group_invite":
        acceptUrl = `${API_BASE_URL}/groups/invites/${notification.related_id}/accept`;
        declineUrl = `${API_BASE_URL}/groups/invites/${notification.related_id}/decline`;
        break;
      case "group_join_request":
        acceptUrl = `${API_BASE_URL}/groups/requests/${notification.related_id}/accept`;
        declineUrl = `${API_BASE_URL}/groups/requests/${notification.related_id}/decline`;
        break;
      case "group_request_accepted":
        return (
          <div className="notification-actions">
            <Link href={`/groups/${notification.related_id}`} className="btn-accept">
              Go to Group
            </Link>
          </div>
        );
      default:
        return null;
    }

    return (
      <div className="notification-actions">
        <button onClick={() => handleAction(acceptUrl)} className="btn-accept">Accept</button>
        <button onClick={() => handleAction(declineUrl)} className="btn-decline">Decline</button>
      </div>
    );
  };


  if (loading) {
    return <div>Loading notifications...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  return (
    <div className="notifications-container">
      <h2>Notifications</h2>
      {notifications && notifications.length > 0 ? (
        notifications.map((notification) => (
          <div
            key={notification.id}
            className={`notification-item card ${
              notification.is_read ? "read" : "unread"
            }`}
          >
            <button className="delete-notification" onClick={() => handleDelete(notification.id)}>X</button>
            <p>{notification.message}</p>
            {renderActions(notification)}
          </div>
        ))
      ) : (
        <p>No notifications.</p>
      )}
    </div>
  );
};

export default Notifications;
