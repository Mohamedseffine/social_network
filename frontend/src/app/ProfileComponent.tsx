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

type FollowerUser = {
  id: number;
  username: string;
  firstName: string;
  lastName: string;
};

export default function ProfilePage({ username }: { username: string | undefined }) {
  const { user: authUser } = useAuth();
  const [user, setUser] = useState<UserProfileDTO | null>(null);
  const [isOwnProfile, setIsOwnProfile] = useState(false);
  const [followers, setFollowers] = useState<FollowerUser[]>([]);
  const [following, setFollowing] = useState<FollowerUser[]>([]);

  const fetchProfile = async () => {
    try {
      const res = await fetch(`/api/profile/another?username=${username}`);
      const data = await res.json();
      setUser(data);
    } catch (err) {
      console.error('Failed to load profile', err);
    }
  };

  useEffect(() => {
    fetchProfile();
  }, [username]);

  useEffect(() => {
    if (authUser && user) {
      const isMe = authUser.username?.toLowerCase() === user.username?.toLowerCase();
      setIsOwnProfile(isMe);
    }
  }, [authUser, user]);

  const fetchFollowers = async () => {
    try {
      const res = await fetch(`/api/follow/followers?username=${user?.username}`);
      const data = await res.json();
      if (Array.isArray(data)) {
        setFollowers(data);
      }
    } catch (err) {
      console.error('Error fetching followers', err);
    }
  };

  const fetchFollowing = async () => {
    try {
      const res = await fetch(`/api/follow/following?username=${user?.username}`);
      const data = await res.json();
      if (Array.isArray(data)) {
        setFollowing(data);
      }
    } catch (err) {
      console.error('Error fetching following', err);
    }
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
            <button onClick={fetchFollowers}>Show Followers ({followers.length})</button>
            <button onClick={fetchFollowing}>Show Following ({following.length})</button>
          </>
        ) : (
            <>
            <button onClick={fetchFollowers}>Show Followers ({followers.length})</button>
            <button onClick={fetchFollowing}>Show Following ({following.length})</button>
            <FollowButton id={user.id} />
          </>
        )}
      </div>

      {followers.length > 0 && (
        <div className={styles.followList}>
          <h3>Followers</h3>
          <ul>
            {followers.map((f) => (
              <li key={f.id}>{f.firstName} {f.lastName} (@{f.username})</li>
            ))}
          </ul>
        </div>
      )}

      {following.length > 0 && (
        <div className={styles.followList}>
          <h3>Following</h3>
          <ul>
            {following.map((f) => (
              <li key={f.id}>{f.firstName} {f.lastName} (@{f.username})</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
