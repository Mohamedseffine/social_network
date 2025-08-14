"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "../../../context/AuthContext";
import { API_BASE_URL, getImageUrl } from "../../../utils/api";

const UserProfilePage = ({ params }: { params: { id: string } }) => {
  const { user: currentUser } = useAuth();
  const [profileUser, setProfileUser] = useState<any>(null);
  const [posts, setPosts] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [followStatus, setFollowStatus] = useState("");
  const router = useRouter();
  const { id } = params;

  const fetchUserData = useCallback(async () => {
    setIsLoading(true);
    setError("");
    try {
      const userRes = await fetch(`${API_BASE_URL}/users/${id}`, {
        credentials: "include",
      });
      if (!userRes.ok) throw new Error("Failed to fetch user data.");
      const userData = await userRes.json();
      setProfileUser(userData);
      setFollowStatus(userData.follow_status || "not_following");

      if (userData.email) {
        const postsRes = await fetch(`${API_BASE_URL}/users/${id}/posts`, {
          credentials: "include",
        });
        if (postsRes.ok) {
            const postData = await postsRes.json();
            setPosts(postData || []);
        }
      }
    } catch (err: any) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  }, [id]);

  useEffect(() => {
    if (currentUser && currentUser.id.toString() === id) {
      router.push("/profile");
      return;
    }
    if (id) {
      fetchUserData();
    }
  }, [id, currentUser, router, fetchUserData]);

  const handleFollowAction = async () => {
    const action = (followStatus === 'accepted' || followStatus === 'pending') ? 'unfollow' : 'follow';
    try {
      const res = await fetch(`${API_BASE_URL}/users/${id}/${action}`, {
        method: "POST",
        credentials: "include",
      });
      if (!res.ok) {
        alert(`Request failed: ${await res.text()}`);
      } else {
        // Refetch user data to get the latest follow status
        fetchUserData();
      }
    } catch (err: any) {
      alert(`An error occurred: ${err.message}`);
    }
  };

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  if (!profileUser) {
    return <div>User not found.</div>;
  }

  const getButtonText = () => {
    switch (followStatus) {
      case "accepted":
        return "Unfollow";
      case "pending":
        return "Pending";
      default:
        return "Follow";
    }
  };

  return (
    <div className="profile-container">
      <h1>{profileUser.nickname || `${profileUser.first_name} ${profileUser.last_name}`}</h1>
      {followStatus !== "is_self" && (
        <button onClick={handleFollowAction} disabled={isLoading}>
          {getButtonText()}
        </button>
      )}
      <div className="profile-header">
        <img src={getImageUrl(profileUser.avatar)} alt="Avatar" className="profile-avatar" />
        <div className="profile-info">
          {profileUser.email ? (
            <>
              <p><strong>Email:</strong> {profileUser.email}</p>
              <p><strong>Date of Birth:</strong> {profileUser.date_of_birth}</p>
              {profileUser.about_me && <p><strong>About Me:</strong> {profileUser.about_me}</p>}
            </>
          ) : (
            <p>This profile is private.</p>
          )}
        </div>
      </div>

      <div className="profile-content">
        <h2>Posts</h2>
        {profileUser.email ? (
            posts.length > 0 ? (
            posts.map((post) => (
                <div key={post.id} className="post">
                <p>{post.content}</p>
                {post.image && <img src={getImageUrl(post.image)} alt="Post image" className="post-image" />}
                <small>{new Date(post.created_at).toLocaleString()}</small>
                </div>
            ))
            ) : (
            <p>This user has no posts.</p>
            )
        ) : (
            <p>This profile is private. Follow them to see their posts.</p>
        )}
      </div>
    </div>
  );
};

export default UserProfilePage;
