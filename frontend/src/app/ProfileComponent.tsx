'use client';

import { useEffect, useState } from 'react';
import styles from './css/ProfilePage.module.css';
import { useAuth } from '@/components/AuthContext';
import FollowButton from '@/components/FollowButton';

type UserProfileDTO = {
  id: number;
  username: string;
  firstName: string;
  lastName: string;
  avatarUrl?: string | null;
  email: string;
  gender?: string;
  aboutMe?: string | null;
  privacy?: string;
  joined?: string;
};

export default function ProfilePage({ params }: { params: { username: string } }) {
  const { user: authUser } = useAuth();
  const [user, setUser] = useState<UserProfileDTO | null>(null);
  const [isOwnProfile, setIsOwnProfile] = useState(false);

  const fetchProfile = async () => {
    try {
      const res = await fetch(`/api/profile/another?username=${params.username}`);
      const data = await res.json();
      setUser(data);
    } catch (err) {
      console.error('Failed to load profile', err);
    }
  };

  useEffect(() => {
    fetchProfile();
  }, [params.username]);

  useEffect(() => {
    if (authUser && user) {
      const isMe = authUser.username?.toLowerCase() === user.username?.toLowerCase();
      setIsOwnProfile(isMe);
      console.log("authUser.username:", authUser.username);
      console.log("user.username:", user.username);
      console.log("isOwnProfile:", isMe);
    }
  }, [authUser, user]);

  const fetchFollowers = () => {
    alert("Show Followers clicked");
  };

  const fetchFollowing = () => {
    alert("Show Following clicked");
  };

  if (!user) return <div>Loading...</div>;

  return (
    <div className={styles.profilePage}>
      <div className={styles.profileHeader}>
        <div className={styles.avatar}>
          <span className={styles.initials}>
            {user.firstName?.[0] ?? ''}{user.lastName?.[0] ?? ''}
          </span>
        </div>
        <div>
          <h2>{user.firstName} {user.lastName}</h2>
          <p>@{user.username}</p>
          <p>{user.aboutMe ?? "No bio available."}</p>
        </div>
      </div>

      <div className={styles.profileDetails}>
        <p><strong>Email:</strong> {user.email}</p>
        <p><strong>Gender:</strong> {user.gender ?? "N/A"}</p>
        <p><strong>Privacy:</strong> {user.privacy ?? "public"}</p>
        <p><strong>Joined:</strong> {user.joined ?? "N/A"}</p>
      </div>

      <div className={styles["follow-controls"]}>
        {isOwnProfile ? (
          <>
            <button onClick={fetchFollowers}>Show Followers</button>
            <button onClick={fetchFollowing}>Show Following</button>
          </>
        ) : (
          <FollowButton username={user.username} />
        )}
      </div>
    </div>
  );
}
