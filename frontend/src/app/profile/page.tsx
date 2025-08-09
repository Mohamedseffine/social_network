"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "../../context/AuthContext";

const ProfilePage = () => {
  const { user, isLoading } = useAuth();
  const [posts, setPosts] = useState<any[]>([]);
  const [followers, setFollowers] = useState<any[]>([]);
  const [following, setFollowing] = useState<any[]>([]);
  const [error, setError] = useState("");
  const router = useRouter();

  useEffect(() => {
    if (isLoading) {
      return;
    }
    if (!user) {
      router.push("/login");
      return;
    }

    const fetchData = async () => {
      try {
        // Fetch posts
        const postsRes = await fetch(
          `http://localhost:8080/api/users/${user.id}/posts`,
          { credentials: "include" }
        );
        if (postsRes.ok) setPosts(await postsRes.json());
        else setError("Failed to fetch posts");

        // Fetch followers
        const followersRes = await fetch(
          `http://localhost:8080/api/users/${user.id}/followers`,
          { credentials: "include" }
        );
        if (followersRes.ok) setFollowers(await followersRes.json());
        else setError("Failed to fetch followers");

        // Fetch following
        const followingRes = await fetch(
          `http://localhost:8080/api/users/${user.id}/following`,
          { credentials: "include" }
        );
        if (followingRes.ok) setFollowing(await followingRes.json());
        else setError("Failed to fetch following");

      } catch (err: any) {
        setError(`An error occurred: ${err.message}`);
      }
    };

    fetchData();
  }, [user, isLoading, router]);

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
        const uploadRes = await fetch(`http://localhost:8080/api/upload`, {
          method: "POST",
          body: uploadFormData,
          credentials: "include",
        });

        if (uploadRes.ok) {
          const uploadData = await uploadRes.json();
          imagePath = uploadData.path;
        } else {
          const error = await uploadRes.text();
          setError(`Image upload failed: ${error}`);
          return;
        }
      } catch (err: any) {
        setError(`Image upload failed: ${err.message}`);
        return;
      }
    }

    const postData = { content, privacy, image: imagePath };

    try {
      const res = await fetch(`http://localhost:8080/api/posts`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(postData),
        credentials: "include",
      });

      if (res.ok) {
        form.reset();
        // Re-fetch posts
        if (user) {
            const postsRes = await fetch(
                `http://localhost:8080/api/users/${user.id}/posts`,
                {
                  credentials: "include",
                }
              );
              if (postsRes.ok) {
                const postsData = await postsRes.json();
                setPosts(postsData);
              }
        }
      } else {
        const errorText = await res.text();
        setError(`Failed to create post: ${errorText}`);
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    }
  };

  const handleTogglePrivacy = async () => {
    if (!user) return;
    try {
      const res = await fetch(`http://localhost:8080/api/profile/privacy`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ is_public: !user.profile_is_public }),
        credentials: "include",
      });

      if (res.ok) {
        // The context doesn't automatically update, so for now we can just reload the page
        // A better solution would be to update the user in the context
        window.location.reload();
      } else {
        const errorText = await res.text();
        setError(`Failed to update privacy: ${errorText}`);
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    }
  };

  if (isLoading || !user) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  return (
    <div className="profile-container">
      <h1>{user.nickname || `${user.first_name} ${user.last_name}`}</h1>
      <div className="profile-header">
        <img
          src={user.avatar ? `http://localhost:8080${user.avatar}` : "/default-avatar.png"}
          alt="Avatar"
          className="profile-avatar"
        />
        <div className="profile-info">
          <p>
            <strong>Email:</strong> {user.email}
          </p>
          <p>
            <strong>Date of Birth:</strong> {user.date_of_birth}
          </p>
          {user.about_me && (
            <p>
              <strong>About Me:</strong> {user.about_me}
            </p>
          )}
          <div className="privacy-setting">
            <p>
              <strong>Profile Privacy:</strong>{" "}
              {user.profile_is_public ? "Public" : "Private"}
            </p>
            <button onClick={handleTogglePrivacy}>
              Make {user.profile_is_public ? "Private" : "Public"}
            </button>
          </div>
        </div>
      </div>

      <div className="create-post">
        <h2>Create a Post</h2>
        <form onSubmit={handleCreatePost}>
          <textarea
            name="content"
            placeholder="What's on your mind?"
            required
          ></textarea>
          <input type="file" name="image" accept="image/*" />
          <select name="privacy" defaultValue="public">
            <option value="public">Public</option>
            <option value="private">Private</option>
          </select>
          <button type="submit">Post</button>
        </form>
      </div>

      <div className="profile-content">
        <h2>Posts</h2>
        {posts && posts.length > 0 ? (
          posts.map((post) => (
            <div key={post.id} className="post">
              <p>{post.content}</p>
              {post.image && (
                <img
                  src={`http://localhost:8080${post.image}`}
                  alt="Post image"
                  className="post-image"
                />
              )}
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
    </div>
  );
};

export default ProfilePage;
