'use client';

import { useEffect, useState } from 'react';
import styles from './css/ProfilePage.module.css';

type UserProfileDTO = {
  id: number;
  username: string;
  firstName: string;
  lastName: string;
  avatarUrl?: string | null;
  email: string;
  aboutMe?: string | null;
  privacyStatus: string;
  gender: string;
  createdAt: string;
};

interface Props {
  username?: string;
}

export default function ProfilePage({ username }: Props) {
  const [profile, setProfile] = useState<UserProfileDTO | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchProfile() {
      try {
        const endpoint = username
          ? `http://localhost:8080/api/users/profile/${username}`
          : `http://localhost:8080/api/profile`;

        const res = await fetch(endpoint, { credentials: 'include' });

        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const data: UserProfileDTO = await res.json();

        setProfile(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load profile');
      } finally {
        setLoading(false);
      }
    }

    fetchProfile();
  }, [username]);

  if (loading) return <p>Loading profile...</p>;
  if (error) return <p style={{ color: 'red' }}>{error}</p>;
  if (!profile) return <p>Profile not found.</p>;

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <div className={styles.avatarBox}>
          {profile.avatarUrl ? (
            <img src={`http://localhost:8080/${profile.avatarUrl}`} alt="avatar" className={styles.avatar} />
          ) : (
            <div className={styles.avatarPlaceholder}>
              {profile.firstName[0]}
              {profile.lastName[0]}
            </div>
          )}
        </div>
        <div className={styles.info}>
          <h1>{profile.firstName} {profile.lastName}</h1>
          <h2>@{profile.username}</h2>
          <p>{profile.aboutMe || 'No bio available.'}</p>
        </div>
      </div>

      <div className={styles.details}>
        <div><strong>Email:</strong> {profile.email}</div>
        <div><strong>Gender:</strong> {profile.gender}</div>
        <div><strong>Privacy:</strong> {profile.privacyStatus}</div>
        <div><strong>Joined:</strong> {profile.createdAt ? new Date(profile.createdAt).toLocaleDateString() : 'N/A'}</div>
      </div>
    </div>
  );
}
