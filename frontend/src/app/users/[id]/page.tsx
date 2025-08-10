"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "../../../context/AuthContext";
import { API_BASE_URL, getImageUrl } from "../../../utils/api";

const UserProfilePage = ({ params }: { params: { id: string } }) => {
  const { user: currentUser } = useAuth(); // Simplified: only need currentUser
  const [profileUser, setProfileUser] = useState<any>(null);
  const [posts, setPosts] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const router = useRouter();
  const { id } = params;

  useEffect(() => {
    // Redirect to own profile page if that's what the user is trying to view
    if (currentUser && currentUser.id.toString() === id) {
      router.push("/profile");
      return;
    }

    const fetchUserData = async () => {
      setIsLoading(true);
      setError("");
      try {
        const userRes = await fetch(`${API_BASE_URL}/users/${id}`, {
          credentials: "include",
        });
        if (!userRes.ok) throw new Error("Failed to fetch user data.");
        const userData = await userRes.json();
        setProfileUser(userData);

        // Only fetch posts if the profile is public or if we are following them
        // The GetUserHandler already returns a limited profile, so we check for a field that only full profiles have, like 'email'
        if (userData.email) {
            const postsRes = await fetch(`${API_BASE_URL}/users/${id}/posts`, {
                credentials: "include",
            });
            if (postsRes.ok) setPosts(await postsRes.json());
        }

      } catch (err: any) {
        setError(err.message);
      } finally {
        setIsLoading(false);
      }
    };

    if (id) {
      fetchUserData();
    }
  }, [id, currentUser, router]);

  const handleFollow = async () => {
    try {
      const res = await fetch(`${API_BASE_URL}/users/${id}/follow`, {
        method: "POST",
        credentials: "include",
      });
      if (!res.ok) {
        alert(`Follow request failed: ${await res.text()}`);
      } else {
        alert("Follow request sent!");
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

  return (
    <div className="profile-container">
      <h1>{profileUser.nickname || `${profileUser.first_name} ${profileUser.last_name}`}</h1>
      <button onClick={handleFollow}>Follow</button>
      <div className="profile-header">
        <img src={getImageUrl(profileUser.avatar)} alt="Avatar" className="profile-avatar" />
        <div className="profile-info">
          {/* Only show details if we have them (i.e., profile is public or we are a follower) */}
          {profileUser.email && (
            <>
              <p><strong>Email:</strong> {profileUser.email}</p>
              <p><strong>Date of Birth:</strong> {profileUser.date_of_birth}</p>
              {profileUser.about_me && <p><strong>About Me:</strong> {profileUser.about_me}</p>}
            </>
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
