"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "../../../context/AuthContext";

const UserProfilePage = ({ params }: { params: { id: string } }) => {
  const { user: currentUser, isLoading: isAuthLoading } = useAuth();
  const [profileUser, setProfileUser] = useState<any>(null);
  const [posts, setPosts] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const router = useRouter();
  const { id } = params;

  useEffect(() => {
    if (isAuthLoading) {
      return;
    }
    if (currentUser && currentUser.id.toString() === id) {
      router.push("/profile");
      return;
    }

    const fetchUserData = async () => {
      try {
        const userRes = await fetch(`http://localhost:8080/api/users/${id}`, {
          credentials: "include",
        });
        if (userRes.ok) {
          const userData = await userRes.json();
          setProfileUser(userData);
        } else {
          setError("Failed to fetch user data.");
        }

        const postsRes = await fetch(
          `http://localhost:8080/api/users/${id}/posts`,
          {
            credentials: "include",
          }
        );
        if (postsRes.ok) {
          const postsData = await postsRes.json();
          setPosts(postsData);
        } else {
          setError((prev) => `${prev}\nFailed to fetch posts.`);
        }
      } catch (err) {
        setError("An error occurred.");
      } finally {
        setIsLoading(false);
      }
    };

    if (id) {
      fetchUserData();
    }
  }, [id, currentUser, isAuthLoading, router]);

  const handleFollow = async () => {
    try {
      const res = await fetch(`http://localhost:8080/api/users/${id}/follow`, {
        method: "POST",
        credentials: "include",
      });
      if (!res.ok) {
        const errorText = await res.text();
        setError(`Follow request failed: ${errorText}`);
      } else {
        // Optionally, update UI to show "Pending" or "Following"
        alert("Follow request sent!");
      }
    } catch (err: any) {
      setError(`An error occurred: ${err.message}`);
    }
  };

  if (isLoading || isAuthLoading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div className="error">{error}</div>;
  }

  if (!profileUser) {
    return <div>User not found.</div>;
  }

  return (
    <div className="profile-container">
      <h1>{profileUser.nickname || `${profileUser.first_name} ${profileUser.last_name}`}</h1>
      <button onClick={handleFollow}>Follow</button>
      <div className="profile-header">
        <img
          src={profileUser.avatar ? `http://localhost:8080${profileUser.avatar}` : "/default-avatar.png"}
          alt="Avatar"
          className="profile-avatar"
        />
        <div className="profile-info">
          <p>
            <strong>Email:</strong> {profileUser.email}
          </p>
          <p>
            <strong>Date of Birth:</strong> {profileUser.date_of_birth}
          </p>
          {profileUser.about_me && (
            <p>
              <strong>About Me:</strong> {profileUser.about_me}
            </p>
          )}
        </div>
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
          <p>This user has no public posts.</p>
        )}
      </div>
    </div>
  );
};

export default UserProfilePage;
