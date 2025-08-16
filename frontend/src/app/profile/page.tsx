"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "../../context/AuthContext";
import { API_BASE_URL, getImageUrl } from "../../utils/api";

const timeAgo = (dateString: string) => {
  const date = new Date(dateString);
  const now = new Date();
  const seconds = Math.round((now.getTime() - date.getTime()) / 1000);
  const minutes = Math.round(seconds / 60);
  const hours = Math.round(minutes / 60);
  const days = Math.round(hours / 24);

  if (seconds < 60) return `${seconds}s ago`;
  if (minutes < 60) return `${minutes}m ago`;
  if (hours < 24) return `${hours}h ago`;
  return `${days}d ago`;
};

const ProfilePage = () => {
  const { user, isLoading } = useAuth();
  const router = useRouter();
  const [posts, setPosts] = useState<any[]>([]);
  const [followers, setFollowers] = useState<any[]>([]);
  const [following, setFollowing] = useState<any[]>([]);
  const [error, setError] = useState("");
  const [isPageLoading, setIsPageLoading] = useState(true);

  useEffect(() => {
    if (isLoading) return; // Wait until session check is complete

    if (!user) {
      router.push('/'); // Redirect to login if not authenticated
      return;
    }

    const fetchProfileData = async () => {
      setIsPageLoading(true);
      setError("");
      try {
        const [postsRes, followersRes, followingRes] = await Promise.all([
          fetch(`${API_BASE_URL}/users/${user.id}/posts`, { credentials: "include" }),
          fetch(`${API_BASE_URL}/users/${user.id}/followers`, { credentials: "include" }),
          fetch(`${API_BASE_URL}/users/${user.id}/following`, { credentials: "include" }),
        ]);

        if (postsRes.ok) setPosts(await postsRes.json() || []);
        else setError(prev => `${prev}Failed to fetch posts. `);

        if (followersRes.ok) setFollowers(await followersRes.json() || []);
        else setError(prev => `${prev}Failed to fetch followers. `);

        if (followingRes.ok) setFollowing(await followingRes.json() || []);
        else setError(prev => `${prev}Failed to fetch following. `);

      } catch (err: any) {
        setError(`An error occurred: ${err.message}`);
      } finally {
        setIsPageLoading(false);
      }
    };
    fetchProfileData();
  }, [user, isLoading, router]);

  const handleTogglePrivacy = async () => {
    if (!user) return;
    try {
      const res = await fetch(`${API_BASE_URL}/profile/privacy`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ is_public: !user.profile_is_public }),
        credentials: "include",
      });
      if (res.ok) {
        window.location.reload();
      } else {
        setError(`Failed to update privacy: ${await res.text()}`);
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    }
  };

  if (isLoading || !user) {
    // Show a loading indicator while the session is being checked, or if redirecting.
    return <div className="loading-container">Loading...</div>;
  }

  return (
    <div className="profile-container">
      <div className="profile-header">
        <img src={getImageUrl(user.avatar)} alt="Avatar" className="profile-avatar" />
        <div className="profile-info">
          <div className="profile-title">
            <h1>{user.nickname || `${user.first_name} ${user.last_name}`}</h1>
            <button onClick={handleTogglePrivacy} className="privacy-toggle-btn">
              Make {user.profile_is_public ? "Private" : "Public"}
            </button>
          </div>
          <div className="profile-stats">
            <div className="stat-item"><strong>{posts.length}</strong> posts</div>
            <div className="stat-item"><strong>{followers.length}</strong> followers</div>
            <div className="stat-item"><strong>{following.length}</strong> following</div>
          </div>
          <div className="profile-bio">
            {user.about_me && <p>{user.about_me}</p>}
            <p><strong>Email:</strong> {user.email}</p>
            <p><strong>Profile is:</strong> {user.profile_is_public ? "Public" : "Private"}</p>
          </div>
        </div>
      </div>

      <div className="profile-content">
        <hr className="divider" />
        {isPageLoading ? (
          <div>Loading posts...</div>
        ) : error ? (
          <div className="error">{error}</div>
        ) : (
          <div className="post-grid">
            {posts && posts.length > 0 ? (
              posts.map((post) => (
                <div key={post.id} className="post-grid-item">
                  {post.image ? (
                    <img src={getImageUrl(post.image)} alt="Post" />
                  ) : (
                    <div className="post-text-preview styled">{post.content}</div>
                  )}
                  <div className="post-grid-overlay">
                    <div className="overlay-info">
                      <span>{post.privacy}</span>
                      <span>{new Date(post.created_at).toLocaleString()}</span>

                    </div>
                  </div>
                </div>
              ))
            ) : (
              <p>No posts to display.</p>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default ProfilePage;
