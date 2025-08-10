"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "../../context/AuthContext";
import { API_BASE_URL, getImageUrl } from "../../utils/api";

const ProfilePage = () => {
  const { user } = useAuth(); // Simplified: only need user
  const [posts, setPosts] = useState<any[]>([]);
  const [followers, setFollowers] = useState<any[]>([]);
  const [following, setFollowing] = useState<any[]>([]);
  const [error, setError] = useState("");
  const [isPageLoading, setIsPageLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const fetchProfileData = async () => {
      if (!user) return; // Should not happen if logic below is correct, but good guard
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

    if (!user) {
      // Auth context is still loading or user is not logged in.
      // The context itself will handle redirects if necessary after its own loading.
      return;
    }
    fetchProfileData();
  }, [user, router]); // Dependency on user is correct.

  const handleCreatePost = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const content = formData.get("content");
    const privacy = formData.get("privacy");
    const imageFile = formData.get("image") as File;

    let imagePath = "";
    if (imageFile && imageFile.size > 0) {
      try {
        const uploadFormData = new FormData();
        uploadFormData.append("image", imageFile);
        const uploadRes = await fetch(`${API_BASE_URL}/upload`, {
          method: "POST", body: uploadFormData, credentials: "include",
        });
        if (uploadRes.ok) {
          imagePath = (await uploadRes.json()).path;
        } else {
          setError(`Image upload failed: ${await uploadRes.text()}`);
          return;
        }
      } catch (err: any) {
        setError(`Image upload failed: ${err.message}`);
        return;
      }
    }

    try {
      const res = await fetch(`${API_BASE_URL}/posts`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content, privacy, image: imagePath }),
        credentials: "include",
      });
      if (res.ok) {
        form.reset();
        // fetchProfileData(); // Refetch all profile data
      } else {
        setError(`Failed to create post: ${await res.text()}`);
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    }
  };

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

  if (!user) {
    // This will show while the AuthContext is performing its initial session check.
    return <div>Loading...</div>;
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

      <div className="create-post">
        <h2>Create a Post</h2>
        <form onSubmit={handleCreatePost}>
          <textarea name="content" placeholder="What's on your mind?" required></textarea>
          <input type="file" name="image" accept="image/*" />
          <select name="privacy" defaultValue="public">
            <option value="public">Public</option>
            <option value="private">Followers Only</option>
          </select>
          <button type="submit">Post</button>
        </form>
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
