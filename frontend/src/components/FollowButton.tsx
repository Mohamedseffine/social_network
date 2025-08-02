import React, { useState } from 'react';
import styles from './css/FollowButton.module.css';

interface FollowButtonProps {
  id: number;
}

const FollowButton: React.FC<FollowButtonProps> = ({ id }) => {
  const [isFollowing, setIsFollowing] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleFollow = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch('http://localhost:8080/api/follow', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ following_id: id }),
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to follow user');
      }

      setIsFollowing(true);
    } catch (error: any) {
      setError(error.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <button
      className={`${styles.followButton} ${isFollowing ? styles.following : ''}`}
      onClick={handleFollow}
      disabled={isLoading || isFollowing}
    >
      {isLoading ? 'Following...' : isFollowing ? 'Following' : 'Follow'}
      {error && <p className={styles.error}>{error}</p>}
    </button>
  );
};

export default FollowButton;