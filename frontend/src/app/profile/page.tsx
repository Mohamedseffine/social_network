"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "../../context/AuthContext";
import { API_BASE_URL, getImageUrl } from "../../utils/api";

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

        if (postsRes.ok) setPosts(await postsRes.json());
        else setError(prev => `${prev}Failed to fetch posts. `);

        if (followersRes.ok) setFollowers(await followersRes.json());
        else setError(prev => `${prev}Failed to fetch followers. `);

        if (followingRes.ok) setFollowing(await followingRes.json());
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
      <h1>{user.nickname || `${user.first_name} ${user.last_name}`}</h1>
      <div className="profile-header">
        <img src={getImageUrl(user.avatar)} alt="Avatar" className="profile-avatar" />
        <div className="profile-info">
          <p><strong>Email:</strong> {user.email}</p>
          <p><strong>Date of Birth:</strong> {user.date_of_birth}</p>
          {user.about_me && <p><strong>About Me:</strong> {user.about_me}</p>}
          <div className="privacy-setting">
            <p><strong>Profile Privacy:</strong> {user.profile_is_public ? "Public" : "Private"}</p>
            <button onClick={handleTogglePrivacy}>Make {user.profile_is_public ? "Private" : "Public"}</button>
          </div>
        </div>
      </div>

      {isPageLoading ? (
        <div>Loading profile content...</div>
      ) : error ? (
        <div className="error">{error}</div>
      ) : (
        <div className="profile-content">
          <h2>Posts</h2>
          {posts && posts.length > 0 ? (
            posts.map((post) => (
              <div key={post.id} className="post">
                <p>{post.content}</p>
                {post.image && <img src={getImageUrl(post.image)} alt="Post image" className="post-image" />}
                <small>{new Date(post.created_at).toLocaleString()}</small>
              </div>
            ))
          ) : (
            <p>No posts yet.</p>
          )}
          <div className="follow-lists">
            <div className="followers-list">
              <h2>Followers</h2>
              {followers && followers.length > 0 ? (
                followers.map((follower) => (
                  <div key={follower.id} className="user-item-small">
                    <Link href={`/users/${follower.id}`}>{follower.first_name} {follower.last_name}</Link>
                  </div>
                ))
              ) : (
                <p>No followers yet.</p>
              )}
            </div>
            <div className="following-list">
              <h2>Following</h2>
              {following && following.length > 0 ? (
                following.map((followed) => (
                  <div key={followed.id} className="user-item-small">
                    <Link href={`/users/${followed.id}`}>{followed.first_name} {followed.last_name}</Link>
                  </div>
                ))
              ) : (
                <p>Not following anyone yet.</p>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ProfilePage;
